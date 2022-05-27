package commands

import (
	"context"

	"github.com/bitrainforest/filmeta-hic/core/config"
	"github.com/bitrainforest/filmeta-hic/core/httpservice"
	"github.com/bitrainforest/filmeta-hic/core/log"
	"github.com/bitrainforest/filmeta-hic/core/store"
	"github.com/bitrainforest/pulsar/api/middleware"
	"github.com/bitrainforest/pulsar/api/router"
	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/model"
	"github.com/bitrainforest/pulsar/internal/service/subscriber"

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
			mustLoadConf()

			var (
				opts     []kratos.Option
				httpOpts []http.ServerOption
			)

			//default opts
			opts = append(opts, log.LoggerKratosOption(ServiceName, DefaultVersion))

			// http service
			httpOpts = append(httpOpts, http.Address(fixedEnv.HttpAddr))
			httpServer := httpservice.GetHttpServer(func(engine *gin.Engine) {
				router.Register(engine, fixedEnv)
			}, nil, httpOpts...)

			//  daemon service
			core, err := mustInitSubCore(context.Context)
			assert.CheckErr(err)
			daemon := NewDaemon(context, core)
			opts = append(opts, kratos.Server(httpServer, daemon))
			//go TestMessage(core)

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

func mustLoadConf() {
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

func mustInitSubAllAddressAppId(ctx context.Context) (appIds []string) {
	subAll := dao.NewUserAppSubAllDao()
	list, err := subAll.ListByAllType(ctx, model.DefaultAllType)
	assert.CheckErr(err)
	for _, v := range list {
		appIds = append(appIds, v.AppId)
	}
	return appIds
}
func mustInitSubCore(ctx context.Context) (*subscriber.Core, error) {
	initAppIds := mustInitSubAllAddressAppId(ctx)
	//init nats client
	natsConf := MustLoadNats(conf)
	notify, err := subscriber.NewNotify(natsConf.GetUri())
	if err != nil {
		return nil, err
	}
	// init subscriber
	sub, err := subscriber.NewSub(initAppIds, notify)
	if err != nil {
		return nil, err
	}
	// init core
	core := subscriber.NewCore(sub)
	return core, nil
}

//func TestMessage(core *subscriber.Core) {
//	time.Sleep(3 * time.Second)
//	for {
//		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
//		for i := 0; i < 100; i++ {
//			go func() {
//				tipSet := &types.TipSet{}
//				h, err := multihash.Sum([]byte("TEST"), multihash.SHA3, 4)
//				if err != nil {
//					log.Errorf("multihash.Sum err:%v", err)
//					return
//				}
//				a := cid.NewCidV1(7, h)
//				msg := types.Message{
//					Version: 1,
//					To:      builtin.ReserveAddress,
//					From:    builtin.RootVerifierAddress,
//				}
//				core.MessageApplied(context.Background(), tipSet, a, &msg, nil, true)
//			}()
//		}
//	}
//}
