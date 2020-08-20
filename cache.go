package npmcp

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"npmcp/storage"
	"time"
)

type cache struct {
	storage.Driver
}

func (c *cache) key(url string) string {
	h := md5.New()
	h.Write([]byte(url))
	return hex.EncodeToString(h.Sum(nil))
}

func (c *cache) CopyAndCache(
	ctx context.Context, dst io.Writer, src io.Reader, url string,
) (written int, err error) {
	key := c.key(url)
	cache, err := c.Writer(ctx, key)
	if err != nil {
		return
	}
	bufSize := 64 << 10           // 64 KB
	ch := make(chan []byte, 1028) // full 64MB
	defer func() {
		ch <- nil
	}()
	var cerr error
	go func() {
		cwritten := 0
		for {
			data := <-ch
			if data == nil {
				break
			}
			// put into cache
			ncw, ecw := cache.Write(data)
			if ecw != nil {
				cerr = ecw
			} else if len(data) != ncw {
				cerr = io.ErrShortWrite
			}
			cwritten += ncw
			if cerr != nil {
				close(ch)
				break
			}
		}
		if cerr != nil {
			log.Printf("[cache] cache url %s error %s", key, cerr.Error())
			cache.Abort(cerr)
			return
		}
		if err != nil {
			log.Printf("[cache] download url %s error %s", key, err.Error())
			cache.Abort(err)
			return
		}
		if err2 := cache.Close(); err2 != nil {
			log.Printf("[cache] save url %s error %s", key, err2.Error())
		} else {
			if written != cwritten {
				cerr = fmt.Errorf("url %s key=%s written=%d bytes of %d bytes", url, key, cwritten, written)
				log.Printf("[cache] error %s", cerr.Error())
				cache.Abort(cerr)
				return
			}
			log.Printf("[cache] added url %s key=%s size=%d bytes", url, key, cwritten)
		}
	}()
	put := func(data []byte) {
		// ignore cache error
		if cerr == nil {
			// use pool to minimize allocation
			b2 := make([]byte, len(data))
			copy(b2, data)
			// Q: channel is full due to slow cache Write?
			if len(ch) >= cap(ch) {
				// channel is full, wait for a while
				time.Sleep(time.Second)
				if len(ch) >= cap(ch) {
					// still full? cache write is not fast enough
					// should not wait the cache?
					cerr = errors.New("cache write too slow")
					return
				}
			}
			ch <- b2
		}
	}
	// copy src to dst
	// put buf to cache during the meantime
	// on cache's first error, no more put to cache
	buf := make([]byte, bufSize)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += nw
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
			put(buf[:nr])
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func (c *cache) Copy(
	ctx context.Context, dst io.Writer, url string,
) (written int64, err error) {
	key := c.key(url)
	src, err := c.Reader(ctx, key)
	if err != nil {
		return
	}
	if src == nil {
		// log.Printf("[cache] missed %s", url)
		err = fmt.Errorf("[cache] url %s key %s not found", url, key)
		return
	}
	defer src.Close()
	// log.Printf("[cache] hit %s key=%s", url, key)
	written, err = io.Copy(dst, src)
	return
}
