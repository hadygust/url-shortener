package env

import (
	"errors"
	"os"
)

func LoadEnvFallback(key string, fallback string) string {
	res, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return res
}

func LoadEnv(key string) (string, error) {
	res, ok := os.LookupEnv(key)

	if !ok {
		return "", errors.New("Key does not exists")
	}

	return res, nil
}
