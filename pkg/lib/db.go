package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func GetDBConnectionPool() (*pgxpool.Pool, error) {
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")

	if dbName == "" {
		dbName = "postgres"
	}

	if dbUser == "" {
		dbUser = "postgres"
	}

	if dbHost == "" {
		dbHost = "postgres"
	}

	postgresDBURL := fmt.Sprintf(
		"postgres://%s:%s@%s:5432/%s", dbUser, dbPassword, dbHost, dbName,
	)

	return Retry(func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), postgresDBURL)
	}, 5*time.Second, 10)
}
