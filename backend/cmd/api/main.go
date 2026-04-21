package main

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	"lets-sub-it/backend/internal/config"
	"lets-sub-it/backend/internal/db"
	apihttp "lets-sub-it/backend/internal/http"
	"lets-sub-it/backend/internal/jobs"
)

func main() {
	cfg := config.Load()
	sqlDB, err := db.Open(cfg.DatabasePath)
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()

	repo, err := jobs.NewSQLiteRepository(sqlDB)
	if err != nil {
		log.Fatal(err)
	}

	service := jobs.NewService(repo, time.Now)
	router := apihttp.NewRouter(service)
	router.StaticFS("/assets", gin.Dir(cfg.StorageDir, false))

	log.Fatal(router.Run(cfg.BackendAddr))
}
