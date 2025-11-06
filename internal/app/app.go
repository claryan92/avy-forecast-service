package app

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"example.com/avalanche/internal/clients"
	"example.com/avalanche/internal/db"
	"example.com/avalanche/internal/handlers"
	"example.com/avalanche/internal/services"

	_ "github.com/lib/pq"
)

type App struct {
	DB      *sql.DB
	Service *services.ForecastService
	Repo    *db.CenterRepository
	Handler *handlers.ForecastHandler
	Router  *http.ServeMux
}

// New initializes the application and its dependencies.
func New() (*App, error) {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@db:5432/avalanche?sslmode=disable"
	}

	pg, err := db.ConnectPostgres(connStr)
	if err != nil {
		return nil, err
	}

	repo := db.NewCenterRepository(pg)

	httpClient := &clients.HTTPClient{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	apiClient := clients.NewAvalancheAPIClient("https://api.avalanche.org/v2/public", httpClient)

	service := services.NewForecast(apiClient)

	handler := handlers.NewForecastHandlerWithRepo(service, repo)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/forecast", handler.GetForecast)
	mux.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &App{
		DB:      pg,
		Service: service,
		Repo:    repo,
		Handler: handler,
		Router:  mux,
	}, nil
}

func (a *App) Run() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := ":" + port
	log.Printf("Server running on %s", addr)
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
