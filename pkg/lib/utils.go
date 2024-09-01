package lib

import (
	"fmt"
	"time"
)

type JSON map[string]interface{}

func MakePortString(port string) string {
	return fmt.Sprintf(":%s", port)
}

type Function[T any] func() (T, error)

func Retry[T any](f Function[T], sleepTime time.Duration, maxRetries int) (T, error) {
	var result T
	var err error = nil

	for maxRetries > 0 {
		result, err = f()
		maxRetries -= 1

		if err == nil {
			break
		}

		time.Sleep(sleepTime)
	}

	return result, err
}
