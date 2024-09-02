package lib

import (
	"fmt"
	"time"
)

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

func Includes[T comparable](list []T, elem T) bool {
	for i := 0; i < len(list); i++ {
		if list[i] == elem {
			return true
		}
	}

	return false
}
