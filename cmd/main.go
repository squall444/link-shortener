package main

import (
	"context"
	"goadv/configs"
	"goadv/internal/auth"
	"goadv/internal/link"
	"goadv/internal/stat"
	"goadv/internal/user"
	"goadv/pkg/db"
	"goadv/pkg/event"
	"goadv/pkg/middleware"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func App() (http.Handler, func()) {
	conf := configs.LoadConfig()
	db := db.NewDb(conf)
	router := http.NewServeMux()
	eventBus := event.NewEventBus()

	//Repositories
	linkRepository := link.NewLinkRepository(db)
	userRepository := user.NewUserRepository(db)
	statRepository := stat.NewStatRepository(db)

	//Serivces
	authService := auth.NewAuthService(userRepository)
	statService := stat.NewStatService(&stat.StatServiceDeps{
		StatRepository: statRepository,
		EventBus:       eventBus,
	})

	//Handlers
	auth.NewAuthHandelr(router, auth.AuthHandlerDeps{
		Config:      conf,
		AuthService: authService,
	})
	link.NewLinkHandler(router, link.LinkHandlerDeps{
		LinkRepository: linkRepository,
		EventBus:       eventBus,
		Config:         conf,
	})
	stat.NewStatHandler(router, stat.StatHandlerDeps{
		StatRepository: statRepository,
		Config:         conf,
	})

	//Middlewares
	stack := middleware.Chain(
		middleware.CORS,
		middleware.Loging,
	)

	ctx, cancel := context.WithCancel(context.Background())
	go statService.AddClick(ctx)

	cleanup := func() {
		cancel()
		if sqlDB, err := db.DB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing database: %v", err)
			} else {
				log.Println("Database connection closed")
			}
		}
		log.Println("Cleanup completed")
	}

	return stack(router), cleanup
}

func main() {
	app, cleanup := App()
	defer cleanup()

	server := &http.Server{
		Addr:    ":8081",
		Handler: app,
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Println("Server is listening on port 8081")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Fatalf("Server error: %v", err)

	case sig := <-quit:
		log.Printf("Received %s signal, starting shutdown", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Graceful shutdown failed: %v", err)
			if err := server.Close(); err != nil {
				log.Fatalf("Force shutdown failed: %v", err)
			}
		}
		log.Println("Server stopped")
	}
}
