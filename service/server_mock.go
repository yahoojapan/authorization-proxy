package service

import "context"

type ServerMock struct {
	ListenAndServeFunc func(context.Context) <-chan []error
}

func (sm *ServerMock) ListenAndServe(ctx context.Context) <-chan []error {
	return sm.ListenAndServeFunc(ctx)
}
