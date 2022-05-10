//go:build wireinject
// +build wireinject

package internal

import (
	"github.com/bitrainforest/pulsar/internal/dao"
	"github.com/bitrainforest/pulsar/internal/service"
	"github.com/google/wire"
)

var Provider = wire.NewSet(dao.Provider, service.Provider, wire.Struct(new(Services), "*"))

type Services struct {
	UserAppService service.UserAppService
}

func NewServices() Services {
	panic(wire.Build(Provider))
}
