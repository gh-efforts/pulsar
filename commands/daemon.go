package commands

import (
	"context"
	"io/ioutil"
	"path/filepath"

	"github.com/filecoin-project/go-paramfetch"
	lotusbuild "github.com/filecoin-project/lotus/build"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/lotus/lib/peermgr"
	"github.com/filecoin-project/lotus/lib/ulimit"
	"github.com/filecoin-project/lotus/node"
	lotusmodules "github.com/filecoin-project/lotus/node/modules"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/go-kratos/kratos/v2/log"
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
	Value:   "~/.bony",
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
	Description: "sub-commands allow you to start a bony daemon process.",
	Action: func(c *cli.Context) error {
		lotuslog.SetupLogLevels()

		if c.Bool("manage-fdlimit") {
			if _, _, err := ulimit.ManageFdLimit(); err != nil {
				log.Errorf("setting file descriptor limit: %s", err)
			}
		}

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

		r.SetConfigPath(daemonFlags.config)

		err = r.Init(repo.FullNode)
		if err != nil && err != repo.ErrRepoExists {
			return xerrors.Errorf("repo init error: %w", err)
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), lotusbuild.ParametersJSON(), lotusbuild.SrsJSON(), 0); err != nil {
			return xerrors.Errorf("fetching proof parameters: %w", err)
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

		liteModeDeps := node.Options()
		stop, err := node.New(ctx,

			node.Base(),
			node.Repo(r),

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

			if err != nil {
				stop(ctx)
				return err
			}
		}

		_, err = r.APIEndpoint()
		return err
	},
}
