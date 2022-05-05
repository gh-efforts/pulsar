package commands

import (
	paramfetch "github.com/filecoin-project/go-paramfetch"
	lotusbuild "github.com/filecoin-project/lotus/build"
	lcli "github.com/filecoin-project/lotus/cli"
	"github.com/filecoin-project/lotus/lib/lotuslog"
	"github.com/filecoin-project/lotus/node/repo"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/mitchellh/go-homedir"
	"github.com/urfave/cli/v2"
	"golang.org/x/xerrors"

	"github.com/bitrainforest/pulsar/config"
)

var initFlags struct {
	repo           string
	config         string
	importSnapshot string
}

var InitCmd = &cli.Command{
	Name:  "init",
	Usage: "Initialise a bony repository.",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "repo",
			Usage:       "Specify path where bony should store chain state.",
			EnvVars:     []string{"BONY_REPO"},
			Value:       "~/.pulsar",
			Destination: &initFlags.repo,
		},
		&cli.StringFlag{
			Name:        "config",
			Usage:       "Specify path of config file to use.",
			EnvVars:     []string{"BONY_CONFIG"},
			Destination: &initFlags.config,
		},
	},
	Action: func(c *cli.Context) error {
		lotuslog.SetupLogLevels()
		{
			dir, err := homedir.Expand(initFlags.repo)
			if err != nil {
				log.Warnw("could not expand repo location", "error", err)
			} else {
				log.Infof("lotus repo: %s", dir)
			}
		}

		r, err := repo.NewFS(initFlags.repo)
		if err != nil {
			return xerrors.Errorf("opening fs repo: %w", err)
		}

		if initFlags.config != "" {
			if err := config.EnsureExists(initFlags.config); err != nil {
				return xerrors.Errorf("ensuring config is present at %q: %w", initFlags.config, err)
			}
			r.SetConfigPath(initFlags.config)
		}

		err = r.Init(repo.FullNode)
		if err != nil && err != repo.ErrRepoExists {
			return xerrors.Errorf("repo init error: %w", err)
		}

		if err := paramfetch.GetParams(lcli.ReqContext(c), lotusbuild.ParametersJSON(), lotusbuild.SrsJSON(), 0); err != nil {
			return xerrors.Errorf("fetching proof parameters: %w", err)
		}

		return nil
	},
}
