package app

import (
	"database/sql"
	"net/http"

	"github.com/Kirill-Pinyaev/people-api/internal/external/demographics"
	"github.com/Kirill-Pinyaev/people-api/internal/store"
)

type App struct {
	DB           *sql.DB
	HTTPClient   *http.Client
	Demographics *demographics.Service
	Store        *store.Store
}

func New(db *sql.DB, client *http.Client) *App {
	return &App{
		DB:           db,
		HTTPClient:   client,
		Demographics: demographics.NewService(client),
		Store:        store.New(db),
	}
}
