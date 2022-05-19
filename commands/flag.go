package commands

import "github.com/urfave/cli/v2"

var clientAPIFlags struct {
	apiAddr  string
	apiToken string
}

var daemonFlags daemonOpts

type daemonOpts struct {
	repo                string
	bootstrap           bool // TODO: is this necessary - do we want to run bony in this mode?
	config              string
	genesis             string
	archive             bool
	archiveModelStorage string
	archiveFileStorage  string
	importSnapshot      string
}

// clientAPIFlagSet are used by commands that act as clients of a daemon's API
var clientAPIFlagSet = []cli.Flag{
	//repoFlag,
	&cli.StringFlag{
		Name:        "repo",
		Usage:       "Specify path where bony should store chain state.",
		EnvVars:     []string{"PULSAR_REPO"},
		Value:       "~/.pulsar",
		Destination: &daemonFlags.repo,
	},
	&cli.StringFlag{
		Name:        "import-snapshot",
		Usage:       "Import chain state from a given chain export file or url.",
		EnvVars:     []string{"PULSAR_SNAPSHOT"},
		Destination: &daemonFlags.importSnapshot,
	},
	&cli.BoolFlag{
		Name: "bootstrap",
		// TODO: usage description
		EnvVars:     []string{"PULSAR_BOOTSTRAP"},
		Value:       true,
		Destination: &daemonFlags.bootstrap,
		Hidden:      true, // hide until we decide if we want to keep this.
	},
	&cli.StringFlag{
		Name:        "config",
		Usage:       "Specify path of config file to use.",
		EnvVars:     []string{"PULSAR_CONFIG"},
		Destination: &daemonFlags.config,
	},
}

func flagSet(fs ...[]cli.Flag) []cli.Flag {
	var flags []cli.Flag

	for _, f := range fs {
		flags = append(flags, f...)
	}

	return flags
}
