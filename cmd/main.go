package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/database"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/handlers"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository/user"
)

func main() {
	logger := *slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("cannot connect to DB: %v", err)
	}
	defer db.Close()

	userRepo := user.NewRepository(db)
	userHandler := &handlers.UserHandler{
		Logger:   logger,
		UserRepo: userRepo,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello!"))
	})
	mux.HandleFunc("/users/setIsActive", userHandler.SetIsActive)
	mux.HandleFunc("/users/getReview", userHandler.GetReview)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	<-ctx.Done()
	log.Println("timeout of 5 seconds.")
	log.Println("Server exiting")
}
