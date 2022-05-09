package router

import (
	"github.com/bitrainforest/filmeta-hic/core/httpx/response"
	"github.com/bitrainforest/filmeta-hic/core/log"
	"github.com/bitrainforest/pulsar/api/codex"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

// Register http router
func Register(e *gin.Engine) {
	// base
	e.GET("/test", response.Json(func(ctx *gin.Context) response.Response {
		log.Errorf("there %v err:%v", "pulsar", errors.New("is router wrong"))
		return codex.OK.WithData("1")
	}))
	// todo more router
}
