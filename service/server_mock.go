package service

import "context"

// ServerMock is a mock of Server
type ServerMock struct {
	ListenAndServeFunc func(context.Context) <-chan []error
}

// ListenAndServe is a mock implementation of Server.ListenAndServe
func (sm *ServerMock) ListenAndServe(ctx context.Context) <-chan []error {
	return sm.ListenAndServeFunc(ctx)
}
