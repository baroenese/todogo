package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/baroenese/todogo/todo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	var configFileName string
	flag.StringVar(&configFileName, "c", "config.yml", "Config file name")
	flag.Parse()
	cfg := defaultConfig()
	cfg.loadFromEnv()
	if len(configFileName) > 0 {
		if err := loadConfigFromFile(configFileName, &cfg); err != nil {
			log.Warn().Str("file", configFileName).Err(err).Msg("canot load config file, use defaults")
		}
	}
	log.Debug().Any("config", cfg).Msg("config loaded")
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	pool, pgErr := pgxpool.New(dbCtx, cfg.DBConfig.ConnStr())
	if pgErr != nil {
		log.Error().Err(pgErr).Msg("unable to connect to database")
	}
	defer pool.Close()
	serviceHandler := service(pool)
	server := &http.Server{Addr: cfg.Listen.Addr(), Handler: serviceHandler}
	go func() {
		log.Info().Str("addr", cfg.Listen.Addr()).Msg("starting server")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("failed to start the server")
		}
	}()
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-signalChan
	log.Info().Str("signal", sig.String()).Msg("shutdown signal received")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()
	go func() {
		<-shutdownCtx.Done()
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Fatal().Msg("graceful shutdown timed out.. forcing exit.")
		}
	}()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("shutdown error")
	}
	log.Info().Msg("closing database connection pool")
	pool.Close()
	log.Info().Msg("server gracefully stopped")
}

func service(newPool *pgxpool.Pool) http.Handler {
	app := chi.NewRouter()
	app.Use(middleware.RequestID)
	app.Use(middleware.RealIP)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	})
	app.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		select {
		case <-time.After(5 * time.Second): // Simulate work
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("completed work.\n"))
		case <-ctx.Done():
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
		}
	})
	todo.SetPool(newPool)
	app.Mount("/api/todo", todo.Router())
	return app
}
