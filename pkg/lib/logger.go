package lib

import (
	"fmt"

	"go.uber.org/zap"
)

func GetLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, fmt.Errorf("failed to setup logger: %v", err)
	}
	defer logger.Sync()

	return logger, nil
}
