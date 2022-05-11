package commands

import (
	"github.com/bitrainforest/pulsar/api/middleware"

	"github.com/bitrainforest/pulsar/api/router"

	"github.com/bitrainforest/filmeta-hic/core/httpservice"

	"github.com/bitrainforest/filmeta-hic/core/log"

	"github.com/bitrainforest/filmeta-hic/core/store"

	"github.com/bitrainforest/filmeta-hic/core/config"

	"github.com/bitrainforest/filmeta-hic/core/envx"

	"github.com/pkg/errors"

	"github.com/bitrainforest/filmeta-hic/core/assert"
	mongox2 "github.com/bitrainforest/filmeta-hic/core/store/mongox"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2"
	kconf "github.com/go-kratos/kratos/v2/config"
	kratosConfig "github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/urfave/cli/v2"
)

const (
	ServiceName    = "pulsar"
	DefaultVersion = "v1"
)

var PulsarCommand = &cli.Command{
	Name:    "http",
	Aliases: []string{"n"},
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
	Before: BeforeDaemon,
	Action: func(context *cli.Context) error {
		// log
		log.SetUp(ServiceName)

		var (
			opts     []kratos.Option
			httpOpts []http.ServerOption
			conf     kratosConfig.Config
			err      error
		)
		fixedEnv := envx.GetEnvs()

		if conf, err = config.LoadConfigAndInitData(ServiceName, func() (schema config.Schema, host string) {
			return config.Etcd, fixedEnv.ConfigETCD
		}, func() (schema config.Schema, host string) {
			return config.File, fixedEnv.ConfigPath
		}); err != nil {
			assert.CheckErr(err)
		}

		// must init mongo
		store.MustLoadMongoDB(conf, func(cfg kconf.Config) (*mongox2.Conf, error) {
			v := cfg.Value("data.mongo.uri")
			mongoUri, err := v.String()
			if err != nil {
				assert.CheckErr(err)
			}
			if mongoUri == "" {
				assert.CheckErr(errors.New("mongoUri must not be empty"))
			}
			return &mongox2.Conf{Uri: mongoUri}, nil
		})
		// load jwt.secret
		middleware.MustLoadSecret(conf)
		// mustLoadRedis
		store.MustLoadRedis(conf)

		httpOpts = append(httpOpts, http.Address(fixedEnv.HttpAddr))
		// set up http service

		httpServer := httpservice.GetHttpServer(func(engine *gin.Engine) {
			router.Register(engine)
		}, nil, httpOpts...)

		opts = append(opts, log.LoggerKratosOption(ServiceName, DefaultVersion))
		opts = append(opts, kratos.Server(httpServer))

		app := kratos.New(opts...)
		if err := app.Run(); err != nil {
			log.Infof("app run err:%v", err)
		}
		<-finishCh
		return nil
	},
}
