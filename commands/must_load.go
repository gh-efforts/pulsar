package commands

import (
	"strconv"

	"github.com/bitrainforest/filmeta-hic/core/assert"
	hicConf "github.com/bitrainforest/filmeta-hic/core/config"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/pkg/errors"
)

type NatsConf struct {
	Addr string `json:"addr"`
	Port int64  `json:"port"`
}

func (conf *NatsConf) isEmpty() bool {
	return conf.Port == 0 || conf.Addr == ""
}

func (conf *NatsConf) GetUri() string {
	return "nats://" + conf.Addr + ":" + strconv.Itoa(int(conf.Port))
}

func MustLoadNats(conf config.Config) NatsConf {
	var (
		natsConf NatsConf
	)
	if err := hicConf.ScanConfValue(conf, "data.nats", &natsConf); err != nil {
		assert.CheckErr(err)
	}
	if natsConf.isEmpty() {
		assert.CheckErr(errors.New("nats conf is empty"))
	}
	return natsConf
}
