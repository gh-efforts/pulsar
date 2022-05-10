package handler

import (
	"context"

	"github.com/bitrainforest/pulsar/api/middleware"
	"github.com/pkg/errors"

	"github.com/bitrainforest/filmeta-hic/core/httpx/response"
	"github.com/bitrainforest/pulsar/api/codex"
	"github.com/bitrainforest/pulsar/api/req"
	"github.com/bitrainforest/pulsar/internal/model"
	"github.com/bitrainforest/pulsar/internal/service"
	"github.com/gin-gonic/gin"
)

type UserAppHandler struct {
	UserAppService service.UserAppService
}

func NewUserAppHandler(userAppService service.UserAppService) UserAppHandler {
	return UserAppHandler{
		UserAppService: userAppService}
}

func (userApp UserAppHandler) Apply(c *gin.Context) response.Response {
	userAppModel := model.NewDefaultApp()
	err := userApp.UserAppService.CreateUserApp(c, &userAppModel)
	if err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}
	var (
		resp req.ApplyResp
	)
	resp.AppId = userAppModel.AppId
	resp.AppSecret = userAppModel.AppSecret
	return codex.OK.WithData(resp)
}

func (userApp UserAppHandler) GetAppWatch(ctx context.Context, appId, address string, fn func(model.UserAppWatch) response.Response) response.Response {
	userWatch, err := userApp.UserAppService.GetAppWatchByAppId(ctx, appId, address)
	if err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}
	if respErr := fn(userWatch); respErr != nil {
		return respErr
	}
	return nil
}

func (userApp UserAppHandler) getAppId(c *gin.Context) (string, error) {
	err := errors.New("app_id not found")
	val, ok := c.Get("appId")
	if !ok {
		return "", err
	}
	applyCx, ok := val.(*middleware.ApplyCx)
	if !ok {
		return "", err
	}
	return applyCx.AppId, nil
}

func (userApp UserAppHandler) AddSub(c *gin.Context) response.Response {
	var (
		param req.AddSubReq
	)
	appId, err := userApp.getAppId(c)
	if err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}
	if err := c.ShouldBind(&param); err != nil {
		return codex.ErrParamIllegal.FormatErrMsg(err)
	}

	if respErr := userApp.GetAppWatch(c, appId,
		param.Address, func(userWatch model.UserAppWatch) response.Response {
			if !userWatch.IsEmpty() {
				return codex.ErrUserAppExist
			}
			return nil
		}); respErr != nil {
		return respErr
	}

	appWatchModel := model.NewDefaultAppWatch()
	appWatchModel.AppId = appId
	appWatchModel.Address = param.Address
	if err := userApp.UserAppService.AddSubAddress(c, appWatchModel); err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}
	return codex.OK
}

func (userApp UserAppHandler) CancelSub(c *gin.Context) response.Response {
	var (
		param req.CancelSubAddressReq
	)
	if err := c.ShouldBind(&param); err != nil {
		return codex.ErrParamIllegal.FormatErrMsg(err)
	}
	appId, err := userApp.getAppId(c)
	if err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}

	if respErr := userApp.GetAppWatch(c, appId,
		param.Address, func(userWatch model.UserAppWatch) response.Response {
			if userWatch.IsEmpty() {
				return codex.ErrUserAppNotExist
			}
			return nil
		}); respErr != nil {
		return respErr
	}
	if err := userApp.UserAppService.CancelSubAddress(c, appId, param.Address); err != nil {
		return codex.ErrService.FormatErrMsg(err)
	}
	return codex.OK
}
