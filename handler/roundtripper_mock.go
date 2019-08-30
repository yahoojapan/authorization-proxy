package handler

import "net/http"

// RoundTripperMock is a mock of RoundTripper
type RoundTripperMock struct {
	RoundTripFunc func(req *http.Request) (*http.Response, error)
}

// RoundTrip is a mock implementation of RoundTripper.RoundTrip
func (rt *RoundTripperMock) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt.RoundTripFunc(req)
}
