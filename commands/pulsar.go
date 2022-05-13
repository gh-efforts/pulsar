package commands

import (
	"github.com/bitrainforest/pulsar/api/middleware"
	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/service/subscriber"

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

var (
	conf     kratosConfig.Config
	fixedEnv envx.FixedEnv
)

func init() {
	fixedEnv = envx.GetEnvs()
}

var (
	PulsarCommand = &cli.Command{
		Name:    "http",
		Aliases: []string{"n"},
		Flags: flagSet(
			clientAPIFlagSet),
		Action: func(context *cli.Context) error {
			// load conf
			MustLoadConf()

			var (
				opts     []kratos.Option
				httpOpts []http.ServerOption
			)

			//default opts
			opts = append(opts, log.LoggerKratosOption(ServiceName, DefaultVersion))

			// http service
			httpOpts = append(httpOpts, http.Address(fixedEnv.HttpAddr))
			httpServer := httpservice.GetHttpServer(func(engine *gin.Engine) {
				router.Register(engine)
			}, nil, httpOpts...)

			//  daemon service
			// todo add nats uri
			core, err := subscriber.NewCore("",
				subscriber.WithUserAppWatchDao(dao.NewUserAppWatchDao()))
			assert.CheckErr(err)
			daemon := NewDaemon(context, core)

			opts = append(opts, kratos.Server(httpServer, daemon))

			// init kratos core
			app := kratos.New(opts...)

			// run
			if err := app.Run(); err != nil {
				log.Errorf("app run err:%v", err)
			}
			return nil
		},
	}
)

func MustLoadConf() {
	// log
	log.SetUp(ServiceName, log.LevelInfo)
	var (
		err error
	)
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
}
