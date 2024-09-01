package schedulerapi

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type APIContext struct {
	db     *pgxpool.Pool
	logger *zap.Logger
	ctx    context.Context
}

func NewAPIContext(db *pgxpool.Pool, logger *zap.Logger) *APIContext {
	ctx := context.Background()

	return &APIContext{
		db,
		logger,
		ctx,
	}
}
