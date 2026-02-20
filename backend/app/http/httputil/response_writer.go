package httputil

import (
	"bufio"
	"net"
	"net/http"
)

// ResponseWriter wraps http.ResponseWriter to capture status code and response size
type ResponseWriter struct {
	http.ResponseWriter
	status      int
	size        int
	wroteHeader bool
}

// NewResponseWriter creates a new ResponseWriter wrapper
func NewResponseWriter(w http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{
		ResponseWriter: w,
		status:         http.StatusOK,
	}
}

// WriteHeader captures the status code
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	if rw.wroteHeader {
		return
	}
	rw.status = statusCode
	rw.wroteHeader = true
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write captures the size and writes the data
func (rw *ResponseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	n, err := rw.ResponseWriter.Write(b)
	rw.size += n
	return n, err
}

// Status returns the captured status code
func (rw *ResponseWriter) Status() int {
	return rw.status
}

// Size returns the captured response size
func (rw *ResponseWriter) Size() int {
	return rw.size
}

// Written returns whether the response has been written
func (rw *ResponseWriter) Written() bool {
	return rw.wroteHeader
}

// Hijack implements the http.Hijacker interface
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Flush implements the http.Flusher interface
func (rw *ResponseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}
