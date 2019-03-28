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
				prov: &service.AuthorizedMock{
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
			name: "Verify roletoken success",
			fields: fields{
				RoundTripper: &RoundTripperMock{
					RoundTripFunc: func(req *http.Request) (*http.Response, error) {
						return &http.Response{
							StatusCode: 999,
						}, nil
					},
				},
				prov: &service.AuthorizedMock{
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
