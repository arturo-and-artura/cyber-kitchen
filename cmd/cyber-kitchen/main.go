package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/arturo-and-artura/cyber-kitchen/internal/agent"
	"github.com/arturo-and-artura/cyber-kitchen/internal/api"
	"github.com/arturo-and-artura/cyber-kitchen/internal/domain"
	"github.com/arturo-and-artura/cyber-kitchen/internal/seed"
	"github.com/arturo-and-artura/cyber-kitchen/internal/store"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "serve" {
		fmt.Fprintln(os.Stderr, "usage: cyber-kitchen serve [flags]")
		os.Exit(2)
	}

	flags := flag.NewFlagSet("serve", flag.ExitOnError)
	listenAddress := flags.String("listen", envOr("CYBER_KITCHEN_LISTEN", ":8080"), "HTTP listen address")
	corsOrigins := flags.String("cors-origins", envOr("CYBER_KITCHEN_CORS_ORIGINS", "http://localhost:5173"), "comma-separated allowed CORS origins")
	databaseURL := flags.String("database-url", envOr("CYBER_KITCHEN_DATABASE_URL", "postgres://cyber_kitchen:cyber_kitchen@localhost:5432/cyber_kitchen?sslmode=disable"), "PostgreSQL connection URL")
	_ = flags.Parse(os.Args[2:])

	initial := seed.InitialState()
	postgresStore, err := store.NewPostgres(*databaseURL, initial)
	if err != nil {
		log.Fatalf("open kitchen database: %v", err)
	}
	defer func() {
		if err := postgresStore.Close(); err != nil {
			log.Printf("close kitchen database: %v", err)
		}
	}()
	service := domain.NewService(postgresStore, time.Now, func() string {
		return fmt.Sprintf("history-%d", time.Now().UnixNano())
	})
	kitchenAgent := agent.New(nil, initial.Meals)
	handler := api.New(service, api.Config{AllowedOrigins: splitCommaList(*corsOrigins)}, kitchenAgent)
	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-shutdownContext.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("HTTP shutdown: %v", err)
		}
	}()

	log.Printf("Cyber Kitchen listening on %s", *listenAddress)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func envOr(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func splitCommaList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return strings.Split(value, ",")
}
