package dao

import "github.com/google/wire"

var Provider = wire.NewSet(
	NewUserAppDao, NewUserAppWatchDao,
)
