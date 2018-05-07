package slimgo

import (
	"bufio"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
)

// ResponseWriter is a wrapper around http.ResponseWriter that provides extra information about
// the response. It is recommended that middleware handlers use this construct to wrap a responsewriter
// if the functionality calls for it.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	Written() bool
	Close()
}

var (
	gzipWriterPool = &sync.Pool{
		New: func() interface{} {
			r, _ := gzip.NewWriterLevel(nil, gzip.BestCompression)
			return r
		},
	}

	flateWriterPool = &sync.Pool{
		New: func() interface{} {
			r, _ := flate.NewWriter(nil, flate.BestCompression)
			return r
		},
	}
)

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(res http.ResponseWriter, req *http.Request, enableCompress bool) ResponseWriter {
	r := &responseWriter{
		res:            res,
		req:            req,
		code:           0,
		enableCompress: enableCompress,
	}

	if r.enableCompress && len(r.req.Header.Get("Accept-Encoding")) > 0 {
		acceptEncodings := strings.Split(r.req.Header.Get("Accept-Encoding"), ",")
		for _, encoding := range acceptEncodings {
			encoding = strings.TrimSpace(encoding)
			if encoding == "gzip" {
				gzipWriter := gzipWriterPool.Get().(*gzip.Writer)
				gzipWriter.Reset(r.res)
				r.compressWriter = gzipWriter
				r.res.Header().Set("Content-Encoding", "gzip")
				break
			}
			if encoding == "deflate" {
				flateWriter := flateWriterPool.Get().(*flate.Writer)
				flateWriter.Reset(r.res)
				r.compressWriter = flateWriter
				r.res.Header().Set("Content-Encoding", "deflate")
				break
			}
		}
	}
	return r
}

type responseWriter struct {
	res            http.ResponseWriter
	req            *http.Request
	compressWriter io.WriteCloser
	enableCompress bool
	code           int
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
	r.res.WriteHeader(code)
}

func (r *responseWriter) Written() bool {
	return r.code != 0
}

func (r *responseWriter) Write(b []byte) (int, error) {
	// if _, ok := r.Header()["Content-Type"]; !ok {
	// 	r.Header().Set("Content-Type", http.DetectContentType(b))
	// }

	if r.compressWriter != nil {
		r.res.Header().Del("Content-Length")
		return r.compressWriter.Write(b)
	}

	return r.res.Write(b)
}

func (r *responseWriter) Close() {
	if r.compressWriter != nil {
		r.compressWriter.Close()
		// put back to pool
		switch r.compressWriter.(type) {
		case *gzip.Writer:
			gzipWriterPool.Put(r.compressWriter)
		case *flate.Writer:
			flateWriterPool.Put(r.compressWriter)
		}
	}
}

func (r *responseWriter) Header() http.Header {
	return r.res.Header()
}

func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := r.res.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("the ResponseWriter doesn't support the Hijacker interface")
	}
	return hijacker.Hijack()
}

func (r *responseWriter) CloseNotify() <-chan bool {
	return r.res.(http.CloseNotifier).CloseNotify()
}

func (r *responseWriter) Flush() {
	if flusher, ok := r.res.(http.Flusher); ok {
		flusher.Flush()
	}
}
