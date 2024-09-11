package lib

import (
	"context"
	"fmt"
	"os"
	"time"

	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func GetDBConnectionString() string {
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")

	if dbName == "" {
		dbName = "postgres"
	}

	if dbUser == "" {
		dbUser = "postgres"
	}

	if dbHost == "" {
		dbHost = "localhost"
	}

	if dbPort == "" {
		dbPort = "5432"
	}

	if dbPassword == "" {
		dbPassword = "postgres"
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName,
	)
}

func GetDBConnectionPool() (*pgxpool.Pool, error) {
	connectionString := GetDBConnectionString()

	return Retry(func() (*pgxpool.Pool, error) {
		return pgxpool.New(context.Background(), connectionString)
	}, 5*time.Second, 10)
}
