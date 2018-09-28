package slimgo

import (
	"net/http"
)

type ResponseWriter interface {
	http.ResponseWriter
	Written() bool
}

// NewResponseWriter creates a ResponseWriter that wraps an http.ResponseWriter
func NewResponseWriter(res http.ResponseWriter) ResponseWriter {
	return &responseWriter{
		res:  res,
		code: 0,
	}
}

type responseWriter struct {
	res  http.ResponseWriter
	code int
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
	r.res.WriteHeader(code)
}

func (r *responseWriter) Written() bool {
	return r.code != 0
}

func (r *responseWriter) Write(b []byte) (int, error) {
	if r.code == 0 {
		r.code = http.StatusOK
	}
	return r.res.Write(b)
}

func (r *responseWriter) Header() http.Header {
	return r.res.Header()
}
