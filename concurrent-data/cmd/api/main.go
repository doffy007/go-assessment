// @title Concurrent Data Processing API
// @version 1.0
// @description Concurrent CSV processor with worker pool and REST interface
// @host localhost:8090
// @BasePath /v1
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"concurrent-data/internal/config"
	"concurrent-data/internal/handler"

	_ "concurrent-data/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	cfg := config.AppConfig

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	pprof.Register(router)

	// CORS config
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8090"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	// Health check
	router.GET("/status", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	ph := &handler.ProcessorHandler{}
	v1 := router.Group("/v1")
	v1.POST("/process", ph.ProcessFiles)
	v1.GET("/progress", ph.GetProgress)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("server listen failed")
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exiting")
}
