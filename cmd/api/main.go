package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"go-auth-api/internal/auth"
	"go-auth-api/internal/database"
	"go-auth-api/internal/handlers"
	"go-auth-api/internal/repository"
	"go-auth-api/internal/service"
)

func main() {
	godotenv.Load()

	// 1. Database & Seed
	db, _ := gorm.Open(postgres.Open(os.Getenv("DB_URL")), &gorm.Config{})
	database.InitDB(db)         // Migrations + Extensions
	database.SeedSuperuser(db)  // Uses .env credentials

	// 2. DI Layers
	repo := repository.NewUserRepository(db)
	svc  := service.NewUserService(repo)
	hdl  := handlers.NewUserHandler(svc)

	// 3. Router
	r := gin.Default()
	r.POST("/login", hdl.Login)
	r.POST("/register", hdl.Register)

	admin := r.Group("/admin").Use(auth.AuthMiddleware(), auth.AuthorizeRole("admin"))
	{
		admin.GET("/stats", hdl.GetStats)
	}
	
	// Add this to your router in main.go
	r.GET("/health", func(c *gin.Context) {
    	// Optionally check DB connection here
    	c.JSON(http.StatusOK, gin.H{"status": "alive"})
	})

	// 4. Server & Shutdown
	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() { srv.ListenAndServe() }()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
	log.Println("Graceful shutdown complete.")
}
