package commands

import (
	"fmt"

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
	Action: func(context *cli.Context) error {
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
		// log
		log.SetUp(ServiceName)
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
			return fmt.Errorf("app run err:%v", err)
		}
		return nil
	},
}
