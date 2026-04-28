package cache

import "time"

type Cache interface {
	Get(string) (any, error)
	Set(string, any, time.Duration) error
	Delete(string) error
}
