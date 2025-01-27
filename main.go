package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/baroenese/todogo/database"
	"github.com/baroenese/todogo/todo"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg := loadConfig()
	log.Debug().Any("config", cfg).Msg("config loaded")
	globalDBPool, err := initDB(cfg.DBConfig.ConnStr())
	if err != nil {
		log.Error().
			Err(err).
			Msg("unable to connect to database")
		os.Exit(1)
	}
	defer globalDBPool.Close()
	// Start koneksion todo
	if err := database.SetPool(globalDBPool); err != nil {
		log.Error().Err(err).Msg("failed to set pool")
		os.Exit(1)
	}
	if database.IsPoolInitialized() {
		log.Info().Msg("database pool is initialized!")
	}
	server := &http.Server{Addr: cfg.Listen.Addr(), Handler: service()}
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
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Error().Err(err).Msg("shutdown error")
		}
	}()
	go func() {
		<-shutdownCtx.Done()
		log.Debug().Msgf("Active connections before shutdown: %d", globalDBPool.Stat().TotalConns())
		if shutdownCtx.Err() == context.DeadlineExceeded {
			log.Error().Msg("graceful shutdown timed out.. forcing exit.")
		}
	}()
	wg.Wait()
	log.Info().Msg("closing database connection pool")
	globalDBPool.Close()
	log.Info().Msg("server gracefully stopped")
}

func initDB(connStr string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}
	config.MinConns = 10
	config.MaxConns = 20
	config.HealthCheckPeriod = 1 * time.Minute
	dbCtx, dbCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer dbCancel()
	pool, err := pgxpool.NewWithConfig(dbCtx, config)
	if err != nil {
		log.Error().Err(err).Msg("unable to connect to database")
		return nil, err
	}
	log.Debug().Msgf("Active connections: %d", pool.Stat().TotalConns())
	return pool, nil
}

var cfgCache *config

func loadConfig() config {
	if cfgCache != nil {
		return *cfgCache
	}
	var configFileName string
	flag.StringVar(&configFileName, "c", "config.yml", "Config file name")
	flag.Parse()
	cfg := defaultConfig()
	cfg.loadFromEnv()
	if configFileName != "" {
		if err := loadConfigFromFile(configFileName, &cfg); err != nil {
			log.Error().
				Str("file", configFileName).
				Err(err).
				Msg("canot load config file, use defaults")
		}
	}
	cfgCache = &cfg
	return cfg
}

func service() http.Handler {
	app := chi.NewRouter()
	app.Use(middleware.RequestID)
	app.Use(middleware.RealIP)
	app.Use(middleware.Recoverer)
	app.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Info().
				Str("method", r.Method).
				Str("url", r.URL.String()).
				Str("remote", r.RemoteAddr).
				Dur("duration", time.Since(start)).
				Msg("request handled")
		})
	})
	app.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "max-age=3600")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello world!"))
	})
	app.Get("/slow", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-time.After(20 * time.Second):
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("completed work.\n"))
		case <-ctx.Done():
			http.Error(w, "request cancelled", http.StatusRequestTimeout)
		}
	})
	app.Mount("/api/todo", todo.Router())
	return app
}
