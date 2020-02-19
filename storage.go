package gonpm

import (
	"context"
	"crypto/md5"
	"encoding/hex"
)

// Storage npm responses
type Storage interface {
	Put(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}

// StorageKey get the key for a string
// The key can be used across different Storage
func StorageKey(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}
