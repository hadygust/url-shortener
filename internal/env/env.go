package env

import "os"

func LoadEnv(key string, fallback string) string {
	res, ok := os.LookupEnv(key)

	if !ok {
		return fallback
	}

	return res
}
