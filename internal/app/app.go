package app

import (
	"log"
	"net/http"
	"os"
	"time"

	"example.com/avalanche/internal/clients"
	"example.com/avalanche/internal/db"
	"example.com/avalanche/internal/handlers"
	"example.com/avalanche/internal/models"
	"example.com/avalanche/internal/notifier"
	"example.com/avalanche/internal/services"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type App struct {
	DB      *gorm.DB
	Service *services.ForecastService
	Repo    *db.CenterRepository
	Handler *handlers.ForecastHandler
	Router  *http.ServeMux
}

func New() (*App, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@db:5432/avalanche?sslmode=disable"
	}

	dbConn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Auto-migrate database schema if enabled
	if os.Getenv("ENABLE_AUTOMIGRATE") == "true" {
		if err := dbConn.AutoMigrate(
			&models.AvalancheCenter{},
			&models.Forecast{},
			&models.Subscription{},
			&models.ForecastCache{},
		); err != nil {
			return nil, err
		}
	}

	repo := db.NewCenterRepository(dbConn)

	httpClient := &clients.HTTPClient{
		Client: &http.Client{Timeout: 10 * time.Second},
	}

	apiClient := clients.NewAvalancheAPIClient("https://api.avalanche.org/v2/public", httpClient)

	service := services.NewForecast(apiClient)

	handler := handlers.NewForecastHandlerWithRepo(service, repo)

	subRepo := db.NewSubscriptionRepository(dbConn)

	// Email sender for subscriptions
	emailSender := notifier.NewSendGridEmailClient()

	// Create SubscriptionService with all dependencies
	subService := services.NewSubscriptionService(subRepo, repo, service, emailSender)

	// Create SubscriptionHandler with the service
	subHandler := handlers.NewSubscriptionHandler(subService)

	app := &App{
		DB:      dbConn,
		Service: service,
		Repo:    repo,
		Handler: handler,
		Router:  http.NewServeMux(),
	}

	app.setupRoutes(subHandler)

	return app, nil
}

// New method
func (a *App) setupRoutes(subHandler *handlers.SubscriptionHandler) {
	// Subscription routes
	a.Router.HandleFunc("/api/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			subHandler.GetSubscriptions(w, r)
		case http.MethodPost:
			subHandler.CreateSubscription(w, r)
		case http.MethodDelete:
			subHandler.DeleteSubscription(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Forecast routes
	a.Router.HandleFunc("/api/forecast", a.Handler.GetForecast)

	// Health check
	a.Router.HandleFunc("/api/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
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
