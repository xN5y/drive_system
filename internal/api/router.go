package api

import (
	"github.com/gin-gonic/gin"
)

// Setup Router
func SetupRouter(handler *Handler, bearerToken string) *gin.Engine {
	router := gin.Default()

	v1 := router.Group("/v1")
	v1.Use(AuthMiddleware(bearerToken))
	{
		v1.POST("/blobs", handler.CreateBlob)
		v1.GET("/blobs/:id", handler.GetBlob)
	}

	return router
}
