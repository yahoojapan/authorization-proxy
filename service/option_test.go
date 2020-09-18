package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/v4/config"
)

func TestWithServerConfig(t *testing.T) {
	type args struct {
		cfg config.Server
	}
	tests := []struct {
		name      string
		args      args
		checkFunc func(Option) error
	}{
		{
			name: "set succes",
			args: args{
				cfg: config.Server{
					Port: 10000,
				},
			},
			checkFunc: func(o Option) error {
				srv := &server{}
				o(srv)
				if srv.cfg.Port != 10000 {
					return errors.New("value cannot set")
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithServerConfig(tt.args.cfg)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithServiceConfig() error = %v", err)
			}
		})
	}
}

func TestWithServerHandler(t *testing.T) {
	type args struct {
		h http.Handler
	}
	type test struct {
		name      string
		args      args
		checkFunc func(Option) error
	}
	tests := []test{
		func() test {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(999)
			})
			return test{
				name: "set success",
				args: args{
					h: h,
				},
				checkFunc: func(o Option) error {
					srv := &server{}
					o(srv)
					r := &httptest.ResponseRecorder{}
					srv.srvHandler.ServeHTTP(r, nil)
					if r.Code != 999 {
						return errors.New("value cannot set")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithServerHandler(tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithServerHandler() error = %v", err)
			}
		})
	}
}

func TestWithDebugHandler(t *testing.T) {
	type args struct {
		h http.Handler
	}
	type test struct {
		name      string
		args      args
		checkFunc func(Option) error
	}
	tests := []test{
		func() test {
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(999)
			})
			return test{
				name: "set success",
				args: args{
					h: h,
				},
				checkFunc: func(o Option) error {
					srv := &server{}
					o(srv)
					r := &httptest.ResponseRecorder{}
					srv.dsHandler.ServeHTTP(r, nil)
					if r.Code != 999 {
						return errors.New("value cannot set")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithDebugHandler(tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithDebugHandler() error = %v", err)
			}
		})
	}
}
