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
	"github.com/spf13/viper"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/database"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/handlers"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository/user"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository/team"
	"github.com/evakaiing/PR-Reviewer-Assignment-Service/internal/repository/pr"
)

// TODO: (task-2) add validator
// TODO: (task-3) add middleware for logg

func initConfig() error {
    viper.AddConfigPath("configs")
    viper.SetConfigName("config")
    viper.SetConfigType("yml")
    viper.AutomaticEnv()
    
    return viper.ReadInConfig()
}

func main() {
	if err := initConfig(); err != nil {
        log.Fatalf("failed to initialize configs: %v", err.Error())
    }

	logger := *slog.New(slog.NewTextHandler(os.Stdout, nil))

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatalf("cannot connect to DB: %v", err)
	}
	defer db.Close()

	userRepo := user.NewRepository(db)
	teamRepo := team.NewRepository(db)
	prRepo := pr.NewRepository(db)
	userHandler := handlers.NewUserHandler(logger, userRepo)
	teamHandler := handlers.NewTeamHandler(logger, teamRepo)
	prHandler := handlers.NewPullRequestHandler(logger, prRepo)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("POST /users/setIsActive", userHandler.SetIsActive)
	mux.HandleFunc("GET /users/getReview", userHandler.GetReview)

	mux.HandleFunc("POST /team/add", teamHandler.Add)
	mux.HandleFunc("GET /team/get", teamHandler.Get)

	mux.HandleFunc("POST /pullRequest/create", prHandler.Create)
	mux.HandleFunc("POST /pullRequest/merge", prHandler.Merge)
	mux.HandleFunc("POST /pullRequest/reassign", prHandler.Reassign)

    srv := &http.Server{
        Addr:         ":" + viper.GetString("server.port"),
        Handler:      mux,
        ReadTimeout:  viper.GetDuration("server.read_timeout"),
        WriteTimeout: viper.GetDuration("server.write_timeout"),
        IdleTimeout:  viper.GetDuration("server.idle_timeout"),
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
