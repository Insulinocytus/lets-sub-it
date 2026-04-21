package http

import (
	"github.com/gin-gonic/gin"

	"lets-sub-it/backend/internal/jobs"
)

func NewRouter(jobService *jobs.Service) *gin.Engine {
	handler := &Handler{jobs: jobService}

	router := gin.New()
	router.Use(gin.Recovery())
	router.POST("/api/jobs", handler.createJob)
	router.GET("/api/jobs/:id", handler.getJob)
	router.GET("/api/videos/:videoId/subtitles", handler.getAssetByVideoID)

	return router
}
