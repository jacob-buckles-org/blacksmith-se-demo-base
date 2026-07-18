package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anvil-analytics/anvil/backend/internal/api"
	"github.com/anvil-analytics/anvil/backend/internal/store"
)

func main() {
	addr := os.Getenv("ANVIL_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	var st api.MetricsSource
	if dsn := os.Getenv("DATABASE_URL"); dsn != "" {
		pg, err := store.Connect(ctx, dsn)
		if err != nil {
			log.Fatalf("connect to postgres: %v", err)
		}
		defer pg.Close()
		if err := pg.Migrate(ctx); err != nil {
			log.Fatalf("migrate: %v", err)
		}
		if err := pg.SeedIfEmpty(ctx); err != nil {
			log.Fatalf("seed: %v", err)
		}
		st = pg
		log.Printf("serving metrics from postgres")
	} else {
		st = store.NewMemoryStore()
		log.Printf("DATABASE_URL not set; serving synthesized metrics")
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewRouter(st),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("anvil backend listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
