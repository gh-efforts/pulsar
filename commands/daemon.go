package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/bitrainforest/pulsar/internal/service/subscriber"

	metricsprometheus "github.com/ipfs/go-metrics-prometheus"

	"github.com/filecoin-project/go-jsonrpc"
	"github.com/filecoin-project/lotus/api"

	"github.com/bitrainforest/pulsar/modules"

	"github.com/filecoin-project/lotus/node/repo"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/filecoin-project/go-paramfetch"
	lotusbuild "github.com/filecoin-project/lotus/build"
	"github.com/filecoin-project/lotus/chain/stmgr"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/lotus/lib/peermgr"
	"github.com/filecoin-project/lotus/node"
	lotusmodules "github.com/filecoin-project/lotus/node/modules"
	"github.com/filecoin-project/lotus/node/modules/dtypes"
	"github.com/mitchellh/go-homedir"
	"github.com/multiformats/go-multiaddr"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"
)

type Daemon struct {
	cliCtx   *cli.Context
	finishCh <-chan struct{}
	sub      *subscriber.Core
}

func NewDaemon(cliCtx *cli.Context, sub *subscriber.Core) *Daemon {
	return &Daemon{
		cliCtx:   cliCtx,
		finishCh: make(chan struct{}),
		sub:      sub,
	}
}

func (d *Daemon) Start(ctx context.Context) error {
	c := d.cliCtx
	isLite := c.Bool("lite")
	lotuslog.SetupLogLevels()

	var err error
	repoPath, err := homedir.Expand(c.String("repo"))
	if err != nil {
		log.Warnw("could not expand repo location", "error", err)
	} else {
		log.Infof("bony repo: %s", repoPath)
	}

	r, err := repo.NewFS(repoPath)
	if err != nil {
		return xerrors.Errorf("opening fs repo: %w", err) //nolint
	}

	if daemonFlags.config == "" {
		daemonFlags.config = filepath.Join(repoPath, "config.toml")
	} else {
		daemonFlags.config, err = homedir.Expand(daemonFlags.config)
		if err != nil {
			log.Warnw("could not expand repo location", "error", err)
		} else {
			log.Infof("bony config: %s", daemonFlags.config)
		}
	}

	r.SetConfigPath(daemonFlags.config)

	err = r.Init(repo.FullNode)
	if err != nil && err != repo.ErrRepoExists {
		return xerrors.Errorf("repo init error: %w", err) //nolint
	}

	if !isLite {
		if err = paramfetch.GetParams(lcli.ReqContext(c), lotusbuild.ParametersJSON(), lotusbuild.SrsJSON(), 0); err != nil {
			return xerrors.Errorf("fetching proof parameters: %w", err) //nolint
		}
	}

	var genBytes []byte
	if c.String("genesis") != "" {
		genBytes, err = ioutil.ReadFile(daemonFlags.genesis)
		if err != nil {
			return xerrors.Errorf("reading genesis: %w", err) //nolint
		}
	} else {
		genBytes = lotusbuild.MaybeGenesis()
	}

	genesis := node.Options()
	if len(genBytes) > 0 {
		genesis = node.Override(new(lotusmodules.Genesis), lotusmodules.LoadGenesis(genBytes))
	}

	//isBootstrapper := false
	shutdown := make(chan struct{})
	liteModeDeps := node.Options()

	if isLite {
		gapi, closer, err := lcli.GetGatewayAPI(c) //nolint
		if err != nil {
			return err
		}

		defer closer()
		liteModeDeps = node.Override(new(api.Gateway), gapi)
	}

	if err := metricsprometheus.Inject(); err != nil { //nolint
		log.Warnf("unable to inject prometheus ipfs/go-metrics exporter; some metrics will be unavailable; err: %s", err)
	}

	var api api.FullNode

	stop, err := node.New(ctx,
		// Start Sentinel Dep injection
		node.FullAPI(&api, node.Lite(isLite)),

		node.Override(new(dtypes.ShutdownChan), shutdown),
		node.Base(),
		node.Repo(r),
		node.Override(new(*stmgr.StateManager), modules.StateManager),
		// replace with our own exec monitor
		//node.Override(new(stmgr.ExecMonitor), modules.NewBufferedExecMonitor),
		node.Override(new(stmgr.ExecMonitor), d.sub),

		// End custom StateManager injection.
		genesis,
		liteModeDeps,

		node.ApplyIf(func(s *node.Settings) bool { return c.IsSet("api") },
			node.Override(node.SetApiEndpointKey, func(lr repo.LockedRepo) error {
				apima, err := multiaddr.NewMultiaddr(clientAPIFlags.apiAddr) //nolint
				if err != nil {
					return err
				}
				return lr.SetAPIEndpoint(apima)
			})),
		node.ApplyIf(func(s *node.Settings) bool { return !daemonFlags.bootstrap },
			node.Unset(node.RunPeerMgrKey),
			node.Unset(new(*peermgr.PeerMgr)),
		),
	)
	if err != nil {
		return xerrors.Errorf("initializing node: %w", err) //nolint
	}

	if daemonFlags.archive {
		if daemonFlags.archiveModelStorage == "" {
			stop(ctx)                                                  //nolint
			return xerrors.Errorf("archive model storage must be set") //nolint
		}
		if daemonFlags.archiveFileStorage == "" {
			stop(ctx)                                                 //nolint
			return xerrors.Errorf("archive file storage must be set") //nolint
		}
	}

	endpoint, err := r.APIEndpoint()
	if err != nil {
		return xerrors.Errorf("getting api endpoint: %w", err) //nolint
	}

	//
	// Instantiate JSON-RPC endpoint.
	// ----

	// Populate JSON-RPC options.
	serverOptions := make([]jsonrpc.ServerOption, 0)
	if maxRequestSize := c.Int("api-max-req-size"); maxRequestSize != 0 {
		serverOptions = append(serverOptions, jsonrpc.WithMaxRequestSize(int64(maxRequestSize)))
	}

	// Instantiate the full node handler.
	h, err := node.FullNodeHandler(api, true, serverOptions...)
	if err != nil {
		return fmt.Errorf("failed to instantiate rpc handler: %s", err)
	}

	// Serve the RPC.
	rpcStopper, err := node.ServeRPC(h, "lotus-daemon", endpoint)
	if err != nil {
		return fmt.Errorf("failed to start json-rpc endpoint: %s", err)
	}

	// Monitor for shutdown.
	d.finishCh = node.MonitorShutdown(shutdown,
		node.ShutdownHandler{Component: "rpc server", StopFunc: rpcStopper},
		node.ShutdownHandler{Component: "node", StopFunc: stop},
	)
	return nil
}

func (d *Daemon) Stop(ctx context.Context) error {
	<-d.finishCh
	d.sub.Stop()
	return nil
}
