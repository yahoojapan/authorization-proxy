package handler

import "net/http"

type RoundTripperMock struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

func (rt *RoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.RoundTripFunc(req)
}
