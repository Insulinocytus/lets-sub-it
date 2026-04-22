package http

import (
	"strings"

	"github.com/gin-gonic/gin"

	"lets-sub-it/backend/internal/jobs"
)

func NewRouter(jobService *jobs.Service) *gin.Engine {
	handler := &Handler{jobs: jobService}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/assets/") {
			origin := c.GetHeader("Origin")
			if origin == "" {
				origin = "*"
			}
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Range")
			c.Writer.Header().Set("Vary", "Origin")
			if c.Request.Method == "OPTIONS" {
				c.AbortWithStatus(204)
				return
			}
		}

		c.Next()
	})
	router.POST("/api/jobs", handler.createJob)
	router.GET("/api/jobs/:id", handler.getJob)
	router.GET("/api/jobs/:id/subtitles", handler.getAssetByJobID)
	router.GET("/api/videos/:videoId/subtitles", handler.getAssetByVideoID)

	return router
}
