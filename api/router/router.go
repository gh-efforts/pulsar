package router

import (
	"github.com/bitrainforest/filmeta-hic/core/envx"
	"github.com/bitrainforest/filmeta-hic/core/httpx/response"
	"github.com/bitrainforest/pulsar/api/codex"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
)

// Register http router
func Register(e *gin.Engine, env envx.FixedEnv) {
	// base
	e.GET("/test", response.Json(func(ctx *gin.Context) response.Response {
		return codex.OK.WithData("1")
	}))
	// todo more router
	g := e.Group("/api/v1")
	RegisterUserApp(g)

	if env.GinMODE == "debug" {
		pprof.Register(e)
	}

}
