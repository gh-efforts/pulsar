package codex

import "github.com/bitrainforest/filmeta-hic/core/errno"

var (
	OK                 = errno.NewError(0, "success")
	ErrService         = errno.NewError(50000, "Service err:[%v]")
	ErrParamIllegal    = errno.NewError(40011, "Parameter is invalid:[%v]")
	ErrUserAppExist    = errno.NewError(40012, "address  exist")
	ErrUserAppNotExist = errno.NewError(40013, "address not exist")
)
