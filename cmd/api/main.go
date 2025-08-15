package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Kirill-Pinyaev/people-api/internal/app"
	"github.com/Kirill-Pinyaev/people-api/internal/http/router"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	addr := getenv("HTTP_ADDR", "0.0.0.0:8080")
	dsn := getenv("DB_DSN", "postgres://people:people@localhost:5432/people?sslmode=disable")

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		panicf("db open: %v", err)
	}
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)

	if err := pingDB(db); err != nil {
		panicf("db ping: %v", err)
	}

	application := app.New(db, &http.Client{Timeout: 4 * time.Second})
	r := router.New(application)

	srv := &http.Server{
		Addr:    addr,
		Handler: r,
	}

	log.Printf("listening on %s", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		panicf("server: %v", err)
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func pingDB(db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return db.PingContext(ctx)
}

func panicf(format string, args ...any) {
	log.Printf(format, args...)
	panic("fatal")
}
