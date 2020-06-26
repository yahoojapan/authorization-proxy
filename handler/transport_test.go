package handler

import (
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/yahoojapan/authorization-proxy/v2/config"
	"github.com/yahoojapan/authorization-proxy/v2/service"
)

func Test_transport_RoundTrip(t *testing.T) {
	type fields struct {
		RoundTripper http.RoundTripper
		prov         service.Authorizationd
		cfg          config.Proxy
	}
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *http.Response
		wantErr bool
	}{
		{
			name: "verify role token failed",
			fields: fields{
				RoundTripper: nil,
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return errors.New("dummy error")
					},
				},
				cfg: config.Proxy{},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io", nil)
					return r
				}(),
			},
			wantErr: true,
		},
		{
			name: "verify role token success",
			fields: fields{
				RoundTripper: &RoundTripperMock{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 999,
						}, nil
					},
				},
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return nil
					},
				},
				cfg: config.Proxy{},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io", nil)
					return r
				}(),
			},
			want: &http.Response{
				StatusCode: 999,
			},
			wantErr: false,
		},
		{
			name: "verify role token success (empty bypass URLs)",
			fields: fields{
				RoundTripper: &RoundTripperMock{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 999,
						}, nil
					},
				},
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return nil
					},
				},
				cfg: config.Proxy{
					BypassURLPaths: []string{},
				},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io", nil)
					return r
				}(),
			},
			want: &http.Response{
				StatusCode: 999,
			},
			wantErr: false,
		},
		{
			name: "BypassURLPaths match, bypass role token verification",
			fields: fields{
				RoundTripper: &RoundTripperMock{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
						}, nil
					},
				},
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return errors.New("role token error")
					},
				},
				cfg: config.Proxy{
					BypassURLPaths: []string{
						"/healthz",
					},
				},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io/healthz", nil)
					return r
				}(),
			},
			want: &http.Response{
				StatusCode: 200,
			},
			wantErr: false,
		},
		{
			name: "BypassURLPaths ANY match, bypass role token verification",
			fields: fields{
				RoundTripper: &RoundTripperMock{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 200,
						}, nil
					},
				},
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return errors.New("role token error")
					},
				},
				cfg: config.Proxy{
					BypassURLPaths: []string{
						"/healthz",
						"/healthz/",
					},
				},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io/healthz/", nil)
					return r
				}(),
			},
			want: &http.Response{
				StatusCode: 200,
			},
			wantErr: false,
		},
		{
			name: "BypassURLPaths NONE match, verify role token",
			fields: fields{
				RoundTripper: nil,
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return errors.New("role token error")
					},
				},
				cfg: config.Proxy{
					BypassURLPaths: []string{
						"/healthz",
					},
				},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io/healthz/", nil)
					return r
				}(),
			},
			wantErr: true,
		},
		{
			name: "BypassURLPaths NOT set, verify role token",
			fields: fields{
				RoundTripper: nil,
				prov: &service.AuthorizerdMock{
					VerifyFunc: func(r *http.Request, act, res string) error {
						return errors.New("role token error")
					},
				},
				cfg: config.Proxy{},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io/healthz", nil)
					return r
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &transport{
				RoundTripper: tt.fields.RoundTripper,
				prov:         tt.fields.prov,
				cfg:          tt.fields.cfg,
			}
			got, err := tr.RoundTrip(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("transport.RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transport.RoundTrip() = %v, want %v", got, tt.want)
			}
		})
	}
}
