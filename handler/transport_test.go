package handler

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"testing"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"
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
					VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
						return errors.New("dummy error")
					},
				},
				cfg: config.Proxy{
					RoleHeader: "",
				},
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
					VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
						return nil
					},
				},
				cfg: config.Proxy{
					RoleHeader: "",
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

func Test_transportWithBypass_RoundTrip(t *testing.T) {
	bypassRT := &RoundTripperMock{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 200,
			}, nil
		},
	}
	wrappedRT := &RoundTripperMock{
		RoundTripFunc: func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 401,
			}, nil
		},
	}

	type fields struct {
		bypassRoundTripper http.RoundTripper
		roundTripper       http.RoundTripper
		cfg                config.Proxy
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
			name: "len(cfg.BypassURLPaths) == 0, bypass NONE",
			fields: fields{
				bypassRoundTripper: bypassRT,
				roundTripper:       wrappedRT,
				cfg:                config.Proxy{},
			},
			args: args{
				r: func() *http.Request {
					r, _ := http.NewRequest("GET", "http://athenz.io", nil)
					return r
				}(),
			},
			want: &http.Response{
				StatusCode: 401,
			},
			wantErr: false,
		},
		{
			name: "BypassURLPaths match, bypass",
			fields: fields{
				bypassRoundTripper: bypassRT,
				roundTripper:       wrappedRT,
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
			name: "BypassURLPaths NOT match, used wrapped round tripper",
			fields: fields{
				bypassRoundTripper: bypassRT,
				roundTripper:       wrappedRT,
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
			want: &http.Response{
				StatusCode: 401,
			},
			wantErr: false,
		},
		{
			name: "BypassURLPaths match ANY, bypass",
			fields: fields{
				bypassRoundTripper: bypassRT,
				roundTripper:       wrappedRT,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &transportWithBypass{
				bypassRoundTripper: tt.fields.bypassRoundTripper,
				roundTripper:       tt.fields.roundTripper,
				cfg:                tt.fields.cfg,
			}
			got, err := transport.RoundTrip(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("transportWithBypass.RoundTrip() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transportWithBypass.RoundTrip() = %v, want %v", got, tt.want)
			}
		})
	}
}
