package commands

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

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

var clientAPIFlags struct {
	apiAddr  string
	apiToken string
}

var clientAPIFlag = &cli.StringFlag{
	Name:        "api",
	Usage:       "Address of bony api in multiaddr format.",
	EnvVars:     []string{"BONY_API"},
	Value:       "/ip4/127.0.0.1/tcp/1234",
	Destination: &clientAPIFlags.apiAddr,
}

var clientTokenFlag = &cli.StringFlag{
	Name:        "api-token",
	Usage:       "Authentication token for bony api.",
	EnvVars:     []string{"BONY_API_TOKEN"},
	Value:       "",
	Destination: &clientAPIFlags.apiToken,
}

var repoFlag = &cli.StringFlag{
	Name:    "repo",
	Usage:   "Specify path where bony should store chain state.",
	EnvVars: []string{"BONY_REPO"},
	Value:   "~/.pulsar",
}

// clientAPIFlagSet are used by commands that act as clients of a daemon's API
var clientAPIFlagSet = []cli.Flag{
	clientAPIFlag,
	clientTokenFlag,
	repoFlag,
}

type daemonOpts struct {
	bootstrap           bool // TODO: is this necessary - do we want to run bony in this mode?
	config              string
	genesis             string
	archive             bool
	archiveModelStorage string
	archiveFileStorage  string
}

var daemonFlags daemonOpts

var cacheFlags struct {
	BlockstoreCacheSize uint // number of raw blocks to cache in memory
	StatestoreCacheSize uint // number of decoded actor states to cache in memory
}

var DaemonCmd = &cli.Command{
	Name:        "daemon",
	Usage:       "Start a bony daemon process.",
	Description: `daemon is the main command you use to run a bony node.`,
	Flags: flagSet(
		clientAPIFlagSet,
		[]cli.Flag{
			&cli.BoolFlag{
				Name: "bootstrap",
				// TODO: usage description
				EnvVars:     []string{"BONY_BOOTSTRAP"},
				Value:       true,
				Destination: &daemonFlags.bootstrap,
				Hidden:      true, // hide until we decide if we want to keep this.
			},
			&cli.StringFlag{
				Name:        "config",
				Usage:       "Specify path of config file to use.",
				EnvVars:     []string{"BONY_CONFIG"},
				Destination: &daemonFlags.config,
			},
		}),
	Action: func(c *cli.Context) error {
		isLite := c.Bool("lite")

		lotuslog.SetupLogLevels()

		ctx := context.Background()
		var err error
		repoPath, err := homedir.Expand(c.String("repo"))
		if err != nil {
			log.Warnw("could not expand repo location", "error", err)
		} else {
			log.Infof("bony repo: %s", repoPath)
		}

		r, err := repo.NewFS(repoPath)
		if err != nil {
			return xerrors.Errorf("opening fs repo: %w", err)
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
		//if err := config.EnsureExists(daemonFlags.config); err != nil {
		//	return xerrors.Errorf("ensuring config is present at %q: %w", daemonFlags.config, err)
		//}

		r.SetConfigPath(daemonFlags.config)

		err = r.Init(repo.FullNode)
		if err != nil && err != repo.ErrRepoExists {
			return xerrors.Errorf("repo init error: %w", err)
		}

		if !isLite {
			if err := paramfetch.GetParams(lcli.ReqContext(c), lotusbuild.ParametersJSON(), lotusbuild.SrsJSON(), 0); err != nil {
				return xerrors.Errorf("fetching proof parameters: %w", err)
			}
		}

		var genBytes []byte
		if c.String("genesis") != "" {
			genBytes, err = ioutil.ReadFile(daemonFlags.genesis)
			if err != nil {
				return xerrors.Errorf("reading genesis: %w", err)
			}
		} else {
			genBytes = lotusbuild.MaybeGenesis()
		}

		genesis := node.Options()
		if len(genBytes) > 0 {
			genesis = node.Override(new(lotusmodules.Genesis), lotusmodules.LoadGenesis(genBytes))
		}

		isBootstrapper := false
		shutdown := make(chan struct{})
		liteModeDeps := node.Options()

		if isLite {
			gapi, closer, err := lcli.GetGatewayAPI(c)
			if err != nil {
				return err
			}

			defer closer()
			liteModeDeps = node.Override(new(api.Gateway), gapi)
		}

		// some libraries like ipfs/go-ds-measure and ipfs/go-ipfs-blockstore
		// use ipfs/go-metrics-interface. This injects a Prometheus exporter
		// for those. Metrics are exported to the default registry.
		if err := metricsprometheus.Inject(); err != nil {
			log.Warnf("unable to inject prometheus ipfs/go-metrics exporter; some metrics will be unavailable; err: %s", err)
		}

		var api api.FullNode

		stop, err := node.New(ctx,
			// Start Sentinel Dep injection
			node.FullAPI(&api, node.Lite(isLite)),

			node.Override(new(dtypes.Bootstrapper), isBootstrapper),
			node.Override(new(dtypes.ShutdownChan), shutdown),
			node.Base(),
			node.Repo(r),
			node.Override(new(*stmgr.StateManager), modules.StateManager),
			node.Override(new(stmgr.ExecMonitor), modules.NewBufferedExecMonitor),
			// End custom StateManager injection.
			genesis,
			liteModeDeps,

			node.ApplyIf(func(s *node.Settings) bool { return c.IsSet("api") },
				node.Override(node.SetApiEndpointKey, func(lr repo.LockedRepo) error {
					apima, err := multiaddr.NewMultiaddr(clientAPIFlags.apiAddr)
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
			return xerrors.Errorf("initializing node: %w", err)
		}

		if daemonFlags.archive {
			if daemonFlags.archiveModelStorage == "" {
				stop(ctx)
				return xerrors.Errorf("archive model storage must be set")
			}
			if daemonFlags.archiveFileStorage == "" {
				stop(ctx)
				return xerrors.Errorf("archive file storage must be set")
			}
		}

		endpoint, err := r.APIEndpoint()
		if err != nil {
			return xerrors.Errorf("getting api endpoint: %w", err)
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
		finishCh := node.MonitorShutdown(shutdown,
			node.ShutdownHandler{Component: "rpc server", StopFunc: rpcStopper},
			node.ShutdownHandler{Component: "node", StopFunc: stop},
		)
		<-finishCh // fires when shutdown is complete.

		// TODO: properly parse api endpoint (or make it a URL)
		return nil
	},
}

func flagSet(fs ...[]cli.Flag) []cli.Flag {
	var flags []cli.Flag

	for _, f := range fs {
		flags = append(flags, f...)
	}

	return flags
}
