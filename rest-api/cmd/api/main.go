// @title REST API Example
// @version 1.0
// @description Example REST API with JWT Authentication
// @host localhost:8080
// @BasePath /v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"rest-api/internal/api"
	"rest-api/internal/auth"
	"rest-api/internal/comment"
	"rest-api/internal/config"
	"rest-api/internal/post"
	"rest-api/internal/user"
	"rest-api/internal/user/role"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	_ "rest-api/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	config.ConfigApps()

	router := gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/status"},
	}))
	router.Use(gin.Recovery())
	pprof.Register(router)

	// CORS config
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Swagger endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/status", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"success": true})
	})

	v1 := router.Group("/v1")

	public := v1.Group("/")
	{
		public.POST("/register", auth.Register)
		public.POST("/signin", auth.SignIn)
		public.GET("/posts/:postID", post.GetPost)
		public.GET("/posts", post.GetPosts)
		public.GET("/comments/:commentID", comment.GetComment)
		public.GET("/comments", comment.GetComments)
		public.GET("/comments/post/:postID", comment.GetCommentsByPost)
	}

	users := v1.Group("/users")
	users.Use(api.SetupParamsFactory)
	{
		users.PUT("/", user.UpdateUser)
		users.GET("/me", user.GetMe)
	}

	admin := v1.Group("/admin")
	admin.Use(api.SetupParamsFactory, api.MustHaveRoles(role.SuperAdmin))
	{
		admin.POST("/users/:userId/assign", role.AssignRole)
		admin.POST("/users/:userId/unassign", role.UnassignRole)
		admin.GET("/users/:userId", user.GetUser)
		admin.GET("/users", user.GetUsers)
	}

	writter := v1.Group("/writter")
	writter.Use(api.SetupParamsFactory, api.MustHaveRoles(role.Writter))
	{
		writter.POST("/posts", post.CreatePost)
		writter.PUT("/posts/:postID", post.UpdatePost)
		writter.DELETE("/posts/:postID", post.DeletePost)
	}

	comments := v1.Group("/comments")
	comments.Use(api.SetupParamsFactory)
	{
		comments.POST("/posts/:postID", comment.CreateComment)
		comments.PUT("/:commentID", comment.UpdateComment)
		comments.DELETE("/:commentID", comment.DeleteComment)
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.AppConfig.Server.Port),
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
	log.Info().Msg("shutting down server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("server forced to shutdown")
	}

	log.Info().Msg("server exiting")
}
