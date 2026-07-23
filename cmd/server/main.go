package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muunatic/schub/internal/config"
	"github.com/muunatic/schub/internal/handler"
	"github.com/muunatic/schub/internal/middleware"
	"github.com/muunatic/schub/internal/models"
	"github.com/muunatic/schub/internal/repository"
	"github.com/muunatic/schub/internal/routes"
	"github.com/muunatic/schub/internal/service"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := config.NewLogger()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}

	db, err := config.NewGORM(cfg)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)
	apiKeyRepo := repository.NewAPIKeyRepository(db)

	authService := service.NewAuthService(userRepo, apiKeyRepo, cfg)
	oauthService := service.NewOAuthService(accountRepo, cfg)
	platformService := service.NewPlatformService(accountRepo, postRepo, commentRepo, cfg)

	authMiddleware := middleware.NewAuthMiddleware(db, userRepo, apiKeyRepo, cfg)

	authHandler := handler.NewAuthHandler(authService, cfg)
	oauthHandler := handler.NewOAuthHandler(oauthService, cfg)
	postHandler := handler.NewPostHandler(platformService, cfg)
	commentHandler := handler.NewCommentHandler(platformService, cfg)
	userHandler := handler.NewUserHandler()
	viewHandler := handler.NewViewHandler(userRepo, accountRepo, apiKeyRepo, cfg)

	ginMode := gin.ReleaseMode
	if cfg.Env == "DEVELOPMENT" {
		ginMode = gin.DebugMode
	}
	gin.SetMode(ginMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.LogMiddleware())
	// r.Use(authMiddleware.VerifyMiddleware())

	funcMap := template.FuncMap{
		"toLower": strings.ToLower,
		"hasPlatform": func(accounts []models.Account, platform string) bool {
			for _, acc := range accounts {
				if string(acc.Platform) == platform {
					return true
				}
			}
			return false
		},
		"slice": func(items ...string) []string {
			return items
		},
		"formatDate": func(dateStr string) string {
			t, err := time.Parse(time.RFC3339, dateStr)
			if err != nil {
				return dateStr
			}
			return t.In(time.FixedZone("WIB", 7*60*60)).Format("02 Jan 2006, 15:04")
		},
	}

	tmpl := template.New("")
	tmpl.Funcs(funcMap)
	if err := filepath.Walk("web/templates", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".html") {
			return nil
		}
		relPath := strings.TrimPrefix(path, "web/templates/")
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		tmpl.New(relPath).Parse(string(content))
		return nil
	}); err != nil {
		logger.Fatal("Failed to load templates", zap.Error(err))
	}
	r.SetHTMLTemplate(tmpl)

	r.Static("/static", "web/static")

	routes.SetupRoutes(r, authMiddleware, authHandler, oauthHandler, postHandler, commentHandler, userHandler, viewHandler)

	if cfg.Env == "DEVELOPMENT" || cfg.Env == "PRODUCTION" {
		r.Static("/oas", "./oas")
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      r,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		logger.Info("Starting server", zap.String("port", cfg.Port), zap.String("env", cfg.Env))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}
