package storage

import (
	"context"
	"errors"
	"io"
	"strings"
)

// Writer interface for storage
type Writer interface {
	io.WriteCloser
	Abort(err error) error
}

// Reader interface for storage
type Reader interface {
	io.ReadCloser
}

// Driver for data storage
type Driver interface {
	Writer(ctx context.Context, key string) (Writer, error)
	Reader(ctx context.Context, key string) (Reader, error)
}

type openner func(string) (Driver, error)

var openners = map[string]openner{}

// Register storage driver
func Register(scheme string, d openner) {
	openners[scheme] = d
}

// Open a new storage
func Open(uri string) (d Driver, err error) {
	if uri == "mem" {
		fn := openners["mem"]
		return fn(uri)
	}
	for scheme, fn := range openners {
		if strings.HasPrefix(uri, scheme+"://") {
			return fn(uri)
		}
	}
	err = errors.New("unknown scheme " + uri)
	return
}

func init() {
	Register("mem", NewMem)
	Register("fs", NewFS)
}
