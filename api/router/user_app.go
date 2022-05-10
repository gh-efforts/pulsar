package router

import (
	"github.com/bitrainforest/filmeta-hic/core/assert"
	"github.com/bitrainforest/filmeta-hic/core/httpx/response"
	"github.com/bitrainforest/pulsar/api/handler"
	"github.com/bitrainforest/pulsar/api/middleware"
	"github.com/bitrainforest/pulsar/internal"
	"github.com/gin-gonic/gin"
)

func RegisterUserApp(e *gin.RouterGroup) {
	userAppHandler := handler.NewUserAppHandler(internal.NewServices().UserAppService)
	g := e.Group("/user-app")
	g.POST("", response.Json(userAppHandler.Apply))

	authMiddleware, err := middleware.NewAuthMiddleware(
		middleware.NewJWTMiddleware(internal.NewServices().UserAppService))
	assert.CheckErr(err)

	g.POST("/token/get", authMiddleware.LoginHandler)
	g.GET("/token/refresh", authMiddleware.RefreshHandler)

	w := g.Group("/sub")
	w.Use(authMiddleware.MiddlewareFunc())
	w.POST("", response.Json(userAppHandler.AddSub))
	w.DELETE("", response.Json(userAppHandler.CancelSub))
}
