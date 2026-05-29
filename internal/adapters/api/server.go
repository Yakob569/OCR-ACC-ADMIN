package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cashflow/admin-service/internal/adapters/handlers"
	"github.com/cashflow/admin-service/internal/core/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	port                 string
	planHandler          *handlers.PlanHandler
	authHandler          *handlers.AuthHandler
	subHandler           *handlers.SubscriptionHandler
	paymentMethodHandler *handlers.PaymentMethodHandler
	merchantHandler      *handlers.MerchantHandler
	tokenAnalyticsHandler *handlers.TokenAnalyticsHandler
	dashboardHandler     *handlers.DashboardHandler
	authSvc              ports.AuthService
	adminUsername        string
	adminPassword        string
	db                   *pgxpool.Pool
	httpServer           *http.Server
}

func NewServer(port string, planHandler *handlers.PlanHandler, authHandler *handlers.AuthHandler, subHandler *handlers.SubscriptionHandler, paymentMethodHandler *handlers.PaymentMethodHandler, merchantHandler *handlers.MerchantHandler, tokenAnalyticsHandler *handlers.TokenAnalyticsHandler, dashboardHandler *handlers.DashboardHandler, authSvc ports.AuthService, adminUser, adminPass string, db *pgxpool.Pool) *Server {
	return &Server{
		port:                  port,
		planHandler:           planHandler,
		authHandler:           authHandler,
		subHandler:            subHandler,
		paymentMethodHandler:  paymentMethodHandler,
		merchantHandler:       merchantHandler,
		tokenAnalyticsHandler: tokenAnalyticsHandler,
		dashboardHandler:      dashboardHandler,
		authSvc:               authSvc,
		adminUsername:        adminUser,
		adminPassword:        adminPass,
		db:                   db,
	}
}

func (s *Server) Start(ctx context.Context) error {
	handler := s.RegisterRoutes()

	s.httpServer = &http.Server{
		Addr:    ":" + s.port,
		Handler: handler,
	}

	errChan := make(chan error, 1)

	go func() {
		log.Printf("🚀 Admin Service running on :%s (Hexagonal Architecture)", s.port)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Println("Shutdown signal received")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("error during server shutdown: %w", err)
		}
		return nil
	}
}

func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	log.Printf("➡️  [HealthCheck] Received request from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")

	if s.db == nil {
		log.Println("❌ [HealthCheck] Database connection pool is nil")
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"unavailable"}`)
		return
	}

	if err := s.db.Ping(r.Context()); err != nil {
		log.Printf("❌ [HealthCheck] Database ping failed: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"unhealthy","database":"disconnected"}`)
		return
	}
	log.Println("✅ [HealthCheck] Database connected successfully")
	fmt.Fprintf(w, `{"status":"healthy","database":"connected"}`)
	log.Println("⬅️  [HealthCheck] Responded with healthy status")
}
