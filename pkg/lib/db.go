package lib

import (
	"context"
	"os"
	"time"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pgxpool "github.com/jackc/pgx/v5/pgxpool"
)

func GetDBConnectionPool() (*pgxpool.Pool, error) {
	dbName := os.Getenv("POSTGRES_DB")
	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")

	if dbName == "" {
		dbName = "postgres"
	}

	if dbUser == "" {
		dbUser = "postgres"
	}

	ctx := context.Background()

	return Retry(func() (*pgxpool.Pool, error) {
		return pgxpool.NewWithConfig(ctx, &pgxpool.Config{
			ConnConfig: &pgx.ConnConfig{
				Config: pgconn.Config{
					Host:     "localhost",
					Database: dbName,
					User:     dbUser,
					Password: dbPassword,
				},
			},
		})
	}, 5*time.Second, 10)
}
