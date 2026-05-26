package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/cashflow/admin-service/internal/adapters/api"
	"github.com/cashflow/admin-service/internal/adapters/auth"
	"github.com/cashflow/admin-service/internal/adapters/handlers"
	"github.com/cashflow/admin-service/internal/adapters/repositories"
	"github.com/cashflow/admin-service/internal/config"
	"github.com/cashflow/admin-service/internal/core/services"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	// 1. Initialize Database
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	dbManager := repositories.NewDatabaseManager(ctx, cfg.DatabaseURL, cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBPort, cfg.DBName)
	defer func() {
		log.Println("Closing database connection...")
		dbManager.Close()
	}()

	if dbManager.Pool == nil {
		log.Fatal("database connection is required to start the service")
	}

	// 2. Wire Hexagonal Architecture
	authAdapter := auth.NewJWTAuthAdapter(cfg.JWTSecret)
	authHandler := handlers.NewAuthHandler(authAdapter, cfg.AdminUsername, cfg.AdminPassword)
	planRepo := repositories.NewPricingPlanRepository(dbManager.Pool)
	planSvc := services.NewPricingPlanService(planRepo)
	planHandler := handlers.NewPlanHandler(planSvc)
	subRepo := repositories.NewSubscriptionRepository(dbManager.Pool)
	subSvc := services.NewSubscriptionService(subRepo, planRepo)
	subHandler := handlers.NewSubscriptionHandler(subSvc)
	paymentMethodRepo := repositories.NewPaymentMethodRepository(dbManager.Pool)
	paymentMethodSvc := services.NewPaymentMethodService(paymentMethodRepo)
	paymentMethodHandler := handlers.NewPaymentMethodHandler(paymentMethodSvc)

	// 3. Initialize Server
	server := api.NewServer(cfg.Port, planHandler, authHandler, subHandler, paymentMethodHandler, authAdapter, cfg.AdminUsername, cfg.AdminPassword, dbManager.Pool)

	// 4. Start Server and handle graceful shutdown internally
	if err := server.Start(ctx); err != nil {
		log.Fatalf("Server stopped with error: %v", err)
	}

	log.Println("✅ Admin Service stopped gracefully")
}
