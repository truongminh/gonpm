package storage

import (
	"context"
	"gonpm"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type fsStorage struct {
	folder string
}

func NewFs(folder string) gonpm.Storage {
	return &fsStorage{folder}
}

func (fs *fsStorage) rel(key string) string {
	if len(key) < 4 {
		return key
	}
	// add sub dirs
	dir1 := key[:2]
	dir2 := key[2:4]
	return filepath.Join(fs.folder, dir1, dir2, key)
}

// Put safely write data to disk
func (fs *fsStorage) Put(ctx context.Context, key string, data []byte) (err error) {
	f, err := ioutil.TempFile(os.TempDir(), "config-")
	if err != nil {
		return
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
		// do not return here
		// continue to close
	}
	// change err if err == nil
	if err1 := f.Close(); err == nil {
		err = err1
	}
	if err != nil {
		return
	}
	filename := fs.rel(key)
	err = os.Rename(f.Name(), filename)
	return
}

func (fs *fsStorage) Get(ctx context.Context, key string) (data []byte, err error) {
	filename := fs.rel(key)
	return ioutil.ReadFile(filename)
}
