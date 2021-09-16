package service

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"google.golang.org/grpc"
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

func TestWithRestHandler(t *testing.T) {
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
			got := WithRestHandler(tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithServerHandler() error = %v", err)
			}
		})
	}
}

func TestWithGRPCHandler(t *testing.T) {
	type args struct {
		h grpc.StreamHandler
	}
	type test struct {
		name      string
		args      args
		checkFunc func(Option) error
	}
	tests := []test{
		func() test {
			h := func(srv interface{}, stream grpc.ServerStream) error {
				return nil
			}
			return test{
				name: "set success",
				args: args{
					h: h,
				},
				checkFunc: func(o Option) error {
					srv := &server{}
					o(srv)
					if reflect.ValueOf(srv.grpcHandler).Pointer() != reflect.ValueOf(h).Pointer() {
						return errors.New("value cannot set")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithGRPCHandler(tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithGRPCHandler() error = %v", err)
			}
		})
	}
}

func TestWithGRPCloser(t *testing.T) {
	type args struct {
		c io.Closer
	}
	type test struct {
		name      string
		args      args
		checkFunc func(Option) error
	}
	tests := []test{
		func() test {
			c := &io.PipeReader{}
			return test{
				name: "set success",
				args: args{
					c: c,
				},
				checkFunc: func(o Option) error {
					srv := &server{}
					o(srv)
					if reflect.ValueOf(srv.grpcCloser).Pointer() != reflect.ValueOf(c).Pointer() {
						return errors.New("value cannot set")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithGRPCCloser(tt.args.c)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithGRPCCloser() error = %v", err)
			}
		})
	}
}

func TestWithGRPCServer(t *testing.T) {
	type args struct {
		srv *grpc.Server
	}
	type test struct {
		name      string
		args      args
		checkFunc func(Option) error
	}
	tests := []test{
		func() test {
			gs := &grpc.Server{}
			return test{
				name: "set success",
				args: args{
					srv: gs,
				},
				checkFunc: func(o Option) error {
					srv := &server{}
					o(srv)
					if !reflect.DeepEqual(srv.grpcSrv, gs) {
						return errors.New("value cannot set")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithGRPCServer(tt.args.srv)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithGRPCServer() error = %v", err)
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
