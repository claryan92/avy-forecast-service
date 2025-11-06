package db

import (
	"database/sql"
	"fmt"
	"time"

	"example.com/avalanche/internal/models"
	_ "github.com/lib/pq"
)

type CenterRepository struct {
	DB *sql.DB
}

func NewCenterRepository(db *sql.DB) *CenterRepository {
	return &CenterRepository{DB: db}
}

func ConnectPostgres(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return db, nil
}

func (r *CenterRepository) GetActiveCenters() ([]models.AvalancheCenter, error) {
	rows, err := r.DB.Query(`SELECT center_id, name, base_url FROM avalanche_centers WHERE active = TRUE`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var centers []models.AvalancheCenter
	for rows.Next() {
		var c models.AvalancheCenter
		if err := rows.Scan(&c.ID, &c.Name, &c.URL); err != nil {
			return nil, err
		}
		centers = append(centers, c)
	}
	return centers, rows.Err()
}
