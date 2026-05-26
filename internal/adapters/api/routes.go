package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Public Health Route
	r.Get("/health", s.healthCheck)

	// Admin API Router Group
	r.Route("/api/v1/admin", func(r chi.Router) {
		// Public Auth Routes
		r.Post("/auth/login", s.authHandler.Login)
		r.Post("/auth/refresh", s.authHandler.RefreshToken)
		r.Post("/auth/logout", s.authHandler.Logout)

		// Protected Admin Routes Group
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(s.authSvc))

			// Plans Endpoints
			r.Route("/plans", func(r chi.Router) {
				r.Get("/", s.planHandler.List)
				r.Post("/", s.planHandler.Create)
				r.Get("/{id}", s.planHandler.Get)
				r.Put("/{id}", s.planHandler.Update)
				r.Patch("/{id}/status", s.planHandler.ToggleStatus)
			})

			// Subscriptions Endpoints
			r.Route("/subscriptions/requests", func(r chi.Router) {
				r.Get("/", s.subHandler.ListRequests)
				r.Post("/{id}/approve", s.subHandler.Approve)
				r.Post("/{id}/reject", s.subHandler.Reject)
			})

			// Payment Methods Endpoints
			r.Route("/payment-methods", func(r chi.Router) {
				r.Get("/", s.paymentMethodHandler.List)
				r.Post("/", s.paymentMethodHandler.Create)
				r.Put("/{id}", s.paymentMethodHandler.Update)
				r.Delete("/{id}", s.paymentMethodHandler.Delete)
			})
		})
	})

	return r
}
