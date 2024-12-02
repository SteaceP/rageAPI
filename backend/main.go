package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SteaceP/coderage/internal/config"
	"github.com/SteaceP/coderage/internal/database"
	"github.com/SteaceP/coderage/internal/handlers"
	"github.com/SteaceP/coderage/pkg/middleware"

	"github.com/gorilla/mux"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Server struct {
	router *mux.Router
	db     *gorm.DB
	logger *zap.Logger
}

func main() {
	// Load configuration
	config.InitConfig()

	// Initialize logger
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Initialize database
	db, err := database.InitDatabase()
	if err != nil {
		logger.Fatal("Database initialization failed", zap.Error(err))
	}

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Fatal("Database migrations failed", zap.Error(err))
	}

	// Create server
	server := &Server{
		router: mux.NewRouter(),
		db:     db,
		logger: logger,
	}

	// Setup routes
	server.setupRoutes()

	// Configure CORS
	corsHandler := middleware.ConfigureCORS().Handler(server.router)

	// HTTP Server configuration
	port := viper.GetString("server.port")
	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      middleware.LoggingMiddleware(logger)(corsHandler),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Graceful server start
	go func() {
		logger.Info("Starting server", zap.String("port", port))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server startup failed", zap.Error(err))
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server gracefully stopped")
}

func (s *Server) setupRoutes() {

	s.router.Use(middleware.Database(s.db))
	// User routes
	s.router.HandleFunc("/users", handlers.CreateUser).Methods("POST")
	s.router.HandleFunc("/users/login", handlers.Login).Methods("POST")
	s.router.HandleFunc("/users/profile", middleware.AuthMiddleware(s.db)(handlers.GetUserProfile)).Methods("GET")

	// Post routes
	s.router.HandleFunc("/posts", handlers.ListPosts).Methods("GET")
	s.router.HandleFunc("/posts", middleware.AuthMiddleware(s.db)(handlers.CreatePost)).Methods("POST")
	s.router.HandleFunc("/posts/{id}", handlers.GetPost).Methods("GET")
	s.router.HandleFunc("/posts/{id}", middleware.AuthMiddleware(s.db)(handlers.UpdatePost)).Methods("PUT")
	s.router.HandleFunc("/posts/{id}", middleware.AuthMiddleware(s.db)(handlers.DeletePost)).Methods("DELETE")

	// Comment routes
	s.router.HandleFunc("/posts/{postId}/comments", middleware.AuthMiddleware(s.db)(handlers.CreateComment)).Methods("POST")
	s.router.HandleFunc("/posts/{postId}/comments", handlers.ListComments).Methods("GET")
}
