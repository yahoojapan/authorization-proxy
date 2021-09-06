package service

import (
	"context"
	"net/http"
)

// ServerMock is a mock of Server
type ServerMock struct {
	ListenAndServeFunc func(context.Context) <-chan []error
}

// ListenAndServe is a mock implementation of Server.ListenAndServe
func (sm *ServerMock) ListenAndServe(ctx context.Context) <-chan []error {
	return sm.ListenAndServeFunc(ctx)
}

// ResponseWriterMock is a mock of ResponseWriter
type ResponseWriterMock struct {
	header    http.Header
	writeFunc func(buf []byte) (int, error)
	code      int
}

// Header is a mock implementation of ResponseWriter.Header
func (rw *ResponseWriterMock) Header() http.Header {
	return rw.header
}

// Write is a mock implementation of ResponseWriter.Write
func (rw *ResponseWriterMock) Write(buf []byte) (int, error) {
	return rw.writeFunc(buf)
}

// WriteHeader is a mock implementation of ResponseWriter.WriteHeader
func (rw *ResponseWriterMock) WriteHeader(code int) {
	rw.code = code
}
