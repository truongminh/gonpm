package storage

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type fsStorage struct {
	folder string
	limit  uint64
}

var defaultFsLimit = uint64(4 * aGIGABYTE)

// NewFS storage on disk
// uri format: file://<folder>?limit=4GB
// support leading current folder [./] and parent folder [../]
func NewFS(uri string) (d Driver, err error) {
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	// change leading dot and dot-dot characters
	if strings.HasPrefix(uri, "fs://./") {
		uri = strings.Replace(uri, "fs://./", "fs://"+cwd+"/", 1)
	} else if strings.HasPrefix(uri, "fs://../") {
		parent := filepath.Join(cwd, "..")
		uri = strings.Replace(uri, "fs://../", "fs://"+parent+"/", 1)
	}
	p, err := url.Parse(uri)
	if err != nil {
		return
	}
	// log.Printf("cwd %s path %s query %+v", cwd, p.Path, p.Query())
	folder := p.Path
	limit, err := ToBytes(p.Query().Get("limit"))
	if err != nil {
		return
	}
	if limit == 0 {
		limit = defaultFsLimit
	}
	return &fsStorage{folder: folder, limit: limit}, nil
}

func (fs *fsStorage) rel(key string) (folder string, filename string) {
	if len(key) < 4 {
		folder = fs.folder
	} else {
		// add sub dirs
		dir1 := key[:2]
		dir2 := key[2:4]
		folder = filepath.Join(fs.folder, dir1, dir2)
	}
	filename = filepath.Join(folder, key)
	return
}

func (fs *fsStorage) Writer(ctx context.Context, key string) (w Writer, err error) {
	folder, filename := fs.rel(key)
	return newWriter(folder, filename)
}

func (fs *fsStorage) Reader(ctx context.Context, key string) (r Reader, err error) {
	_, filename := fs.rel(key)
	r, err = os.Open(filename)
	return
}

type fswriter struct {
	folder   string
	filename string
	aborted  bool
	*os.File
	sync.Mutex
}

func newWriter(folder string, filename string) (w *fswriter, err error) {
	f, err := ioutil.TempFile(os.TempDir(), "gonpm-fs-")
	if err != nil {
		return
	}
	w = &fswriter{folder: folder, filename: filename, File: f}
	return
}

func (w *fswriter) Abort(err error) error {
	w.aborted = true
	return w.Close()
}

func (w *fswriter) Close() (err error) {
	w.Lock()
	defer w.Unlock()
	if w.aborted {
		err = w.File.Close()
		if err != nil {
			return
		}
		os.Remove(w.Name())
		return nil
	}
	defer w.File.Close()
	err = os.MkdirAll(w.folder, 0755)
	if err != nil {
		log.Printf("[fscache] create folder %s error %s", w.folder, err.Error())
		return
	}
	// move file cross-device
	// os.Rename does not work
	// copy the file to a tmp place in the dest device
	_, err = w.File.Seek(0, io.SeekStart)
	if err != nil {
		return
	}
	tmpFilename := fmt.Sprintf("%s.%d.tmp", w.filename, time.Now().Unix())
	tmpFile, err := os.Create(tmpFilename)
	if err != nil {
		return
	}
	_, err = io.Copy(tmpFile, w.File)
	if err != nil {
		return
	}
	err = tmpFile.Close()
	if err != nil {
		return
	}
	err = os.Rename(tmpFilename, w.filename)
	if err != nil {
		log.Printf("[fscache] save %s error %s", w.filename, err.Error())
	}
	return
}
