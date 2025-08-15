package router

import (
	"net/http"
	"time"

	"github.com/Kirill-Pinyaev/people-api/internal/app"
	"github.com/Kirill-Pinyaev/people-api/internal/http/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

func New(a *app.App) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(10 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8081", "http://127.0.0.1:8081"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := handlers.New(a)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi.yaml")
	})

	r.Route("/v1", func(r chi.Router) {
		r.Route("/people", func(r chi.Router) {
			r.Get("/", h.PeopleList)
			r.Post("/", h.PeopleCreate)
			r.Get("/{id}", h.PeopleGet)
			r.Patch("/{id}", h.PeopleUpdate)

			r.Get("/surname/{last_name}", h.PeopleBySurname)

			r.Post("/{id}/emails", h.AddEmail)
			r.Get("/{id}/emails", h.ListEmails)
			r.Delete("/{id}/emails/{email_id}", h.DeleteEmail)

			r.Post("/{id}/friends/{friend_id}", h.AddFriend)
			r.Delete("/{id}/friends/{friend_id}", h.RemoveFriend)
			r.Get("/{id}/friends", h.ListFriends)
		})
	})

	return r
}
