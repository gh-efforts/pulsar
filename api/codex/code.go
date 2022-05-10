package codex

import "github.com/bitrainforest/filmeta-hic/core/errno"

var (
	OK         = errno.NewError(0, "success")
	ErrService = errno.NewError(50000, "Service err:[%v]")
)
