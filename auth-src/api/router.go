package api

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(apiKey string, handler *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/auth", handler.AuthUser)

	r.Group(func(r chi.Router) {
		r.Use(APIKeyAuth(apiKey))

		r.Get("/api/health", handler.Health)
		r.Get("/api/users", handler.ListUsers)
		r.Post("/api/users", handler.AddUser)
		r.Delete("/api/users/{username}", handler.DeleteUser)
	})

	return r
}
