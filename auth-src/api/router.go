package api

import (
	"io/fs"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(apiKey string, handler *Handler, web fs.FS) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/api/auth", handler.AuthUser)
	r.Post("/api/login", handler.Login)
	r.Post("/api/logout", handler.Logout)

	r.Group(func(r chi.Router) {
		r.Use(APIKeyAuth(apiKey))

		r.Get("/api/health", handler.Health)
		r.Get("/api/config", handler.Config)
		r.Get("/api/status", handler.Status)
		r.Get("/api/users", handler.ListUsers)
		r.Post("/api/users", handler.AddUser)
		r.Put("/api/users/{username}", handler.UpdateUser)
		r.Delete("/api/users/{username}", handler.DeleteUser)
		r.Post("/api/users/{username}/kick", handler.KickUser)
	})

	r.Handle("/*", StaticHandler(web))

	return r
}
