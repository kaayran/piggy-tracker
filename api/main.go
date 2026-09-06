package main

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed all:webdist
var webFS embed.FS

type server struct {
	db       *pgxpool.Pool
	botToken string
}

func main() {
	if err := run(); err != nil {
		slog.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run() error {
	botToken := os.Getenv("BOT_TOKEN")
	if botToken == "" {
		return errMissingEnv("BOT_TOKEN")
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errMissingEnv("DATABASE_URL")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := migrate(databaseURL); err != nil {
		return err
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return err
	}

	s := &server{db: pool, botToken: botToken}
	slog.Info("listening", "port", port)
	return (&http.Server{
		Addr:              ":" + port,
		Handler:           s.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe()
}

func (s *server) routes() http.Handler {
	api := http.NewServeMux()
	api.HandleFunc("GET /api/me", s.handleGetMe)
	api.HandleFunc("GET /api/transactions", s.handleListTransactions)
	api.HandleFunc("POST /api/transactions", s.handleCreateTransaction)
	api.HandleFunc("DELETE /api/transactions/{id}", s.handleDeleteTransaction)

	mux := http.NewServeMux()
	mux.Handle("/api/", s.authenticate(api))
	mux.Handle("/", spaHandler())
	return mux
}

// spaHandler serves the built frontend, falling back to index.html so client-side routes
// survive a reload.
func spaHandler() http.Handler {
	dist, err := fs.Sub(webFS, "webdist")
	if err != nil {
		panic(err)
	}
	files := http.FileServerFS(dist)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := fs.Stat(dist, path(r.URL.Path)); err != nil {
			r = r.Clone(r.Context())
			r.URL.Path = "/"
		}
		files.ServeHTTP(w, r)
	})
}

func path(p string) string {
	if p == "/" || p == "" {
		return "index.html"
	}
	return p[1:]
}

func migrate(databaseURL string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}

type errMissingEnv string

func (e errMissingEnv) Error() string { return string(e) + " is not set" }
