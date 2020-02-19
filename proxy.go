package gonpm

import (
	"context"
	"crypto/tls"
	"fmt"
	"gonpm/cert"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

// Proxy supports http and https
type Proxy interface {
	http.Handler
	Listen(ctx context.Context) error
}

type server struct {
	addr     string
	unixsock string
}

// NewProxy server
func NewProxy(port int) Proxy {
	addr := fmt.Sprintf(":%d", port)
	unixsock := fmt.Sprintf("/tmp/gonpm_%d.sock", port)
	s := &server{
		addr:     addr,
		unixsock: unixsock,
	}
	return s
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	req, _ := http.NewRequestWithContext(ctx, "GET", r.URL.String(), nil)
	if r.TLS != nil {
		req.URL.Scheme = "https"
		remote := r.TLS.ServerName
		req.Host = remote
		req.URL.Host = remote
	}
	log.Printf("get %s", req.URL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode > 200 {
		s, _ := ioutil.ReadAll(resp.Body)
		message := resp.Status
		if len(s) > 1 {
			message = string(s)
		}
		http.Error(w, message, http.StatusInternalServerError)
		return
	}
	copyHeader(w.Header(), resp.Header)
	io.Copy(w, resp.Body)
}

func (s *server) connect(w http.ResponseWriter, r *http.Request) {
	// log.Printf("https %s", r.Proto)
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// Don't forget to close the connection:
	defer conn.Close()
	msg := r.Proto + " 200 Connection established\r\n\r\n"
	bufrw.Write([]byte(msg))
	bufrw.Flush()
	upstream, err := net.Dial("unix", s.unixsock)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	errc := make(chan error, 1)
	go func() {
		_, err := io.Copy(conn, upstream)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(upstream, conn)
		errc <- err
	}()
	<-errc
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// log.Printf("request %s", r.URL)
	if r.Method == http.MethodGet {
		s.get(w, r)
	} else if r.Method == http.MethodConnect {
		// https here
		s.connect(w, r)
	}
}

func (s *server) Listen(ctx context.Context) error {
	errc := make(chan error, 1)
	go func() {
		log.Printf("http listen on %s", s.addr)
		err := http.ListenAndServe(s.addr, s)
		errc <- err
	}()
	{
		// try to clean up existing socket
		err := os.Remove(s.unixsock)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("remove unixsock %s", err.Error())
			}
		}
		// handle unix socket
		httpsServer := &http.Server{Addr: s.unixsock, Handler: s}
		lns, err := net.Listen("unix", s.unixsock)
		if err != nil {
			return err
		}
		defer lns.Close()
		defer httpsServer.Shutdown(ctx)
		go func() {
			httpsServer.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{
					cert.Dummy,
				},
			}
			log.Printf("https listen on %s", s.unixsock)
			err = httpsServer.ServeTLS(lns, "", "")
			errc <- err
		}()
	}
	select {
	case <-ctx.Done():
		return nil
	case err := <-errc:
		return err
	}
}
