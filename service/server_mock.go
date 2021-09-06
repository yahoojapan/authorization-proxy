package service

import (
	"context"
	"errors"
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

type ResponseWriter struct {
}

func (rw *ResponseWriter) Header() http.Header {
	return http.Header{}
}

func (rw *ResponseWriter) Write(buf []byte) (int, error) {
	return len(buf), errors.New("Test error")
}

// WriteHeader implements http.ResponseWriter.
func (rw *ResponseWriter) WriteHeader(code int) {
}
