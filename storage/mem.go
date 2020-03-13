package storage

import (
	"bytes"
	"context"
	"io/ioutil"
)

type memStorage struct {
	m map[string][]byte
}

// NewMem storage on disk at the given folder
func NewMem(param string) (Driver, error) {
	return &memStorage{m: map[string][]byte{}}, nil
}

func (s *memStorage) Writer(ctx context.Context, key string) (w Writer, err error) {
	m := &memwriter{}
	m.save = func() {
		s.m[key] = m.data
	}
	return m, nil
}

func (s *memStorage) Reader(ctx context.Context, key string) (r Reader, err error) {
	data, ok := s.m[key]
	if ok {
		return ioutil.NopCloser(bytes.NewBuffer(data)), nil
	}
	return nil, nil
}

type memwriter struct {
	aborted bool
	save    func()
	data    []byte
}

func (w *memwriter) Write(data []byte) (written int, err error) {
	w.data = append(w.data, data...)
	written = len(data)
	return
}

func (w *memwriter) Abort(err error) error {
	w.aborted = true
	return w.Close()
}

func (w *memwriter) Close() (err error) {
	if w.aborted {
		return nil
	}
	w.save()
	return
}
