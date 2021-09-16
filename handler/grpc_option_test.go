package handler

import (
	"crypto/tls"
	"reflect"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func TestWithProxyConfig(t *testing.T) {
	type args struct {
		cfg config.Proxy
	}
	type test struct {
		name      string
		args      args
		checkFunc func(GRPCOption) error
	}
	tests := []test{
		func() test {
			cfg := config.Proxy{
				Host: "http://test_server.com",
			}
			return test{
				name: "set success",
				args: args{
					cfg: cfg,
				},
				checkFunc: func(o GRPCOption) error {
					h := &GRPCHandler{}
					o(h)
					if !reflect.DeepEqual(h.proxyCfg, cfg) {
						return errors.New("config not match")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithProxyConfig(tt.args.cfg)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithProxyConfig() error = %v", err)
			}
		})
	}
}

func TestWithRoleTokenConfig(t *testing.T) {
	type args struct {
		cfg config.RoleToken
	}
	type test struct {
		name      string
		args      args
		checkFunc func(GRPCOption) error
	}
	tests := []test{
		func() test {
			cfg := config.RoleToken{
				Enable: true,
			}
			return test{
				name: "set success",
				args: args{
					cfg: cfg,
				},
				checkFunc: func(o GRPCOption) error {
					h := &GRPCHandler{}
					o(h)
					if !reflect.DeepEqual(h.roleCfg, cfg) {
						return errors.New("config not match")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithRoleTokenConfig(tt.args.cfg)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithRoleTokenConfig() error = %v", err)
			}
		})
	}
}

func TestWithAuthorizationd(t *testing.T) {
	type args struct {
		a service.Authorizationd
	}
	type test struct {
		name      string
		args      args
		checkFunc func(GRPCOption) error
	}
	tests := []test{
		func() test {
			a := &service.AuthorizerdMock{}
			return test{
				name: "set success",
				args: args{
					a: a,
				},
				checkFunc: func(o GRPCOption) error {
					h := &GRPCHandler{}
					o(h)
					if !reflect.DeepEqual(h.authorizationd, a) {
						return errors.New("authorizationd not match")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithAuthorizationd(tt.args.a)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithAuthorizationd() error = %v", err)
			}
		})
	}
}

func TestWithTLSConfig(t *testing.T) {
	type args struct {
		cfg *tls.Config
	}
	type test struct {
		name      string
		args      args
		checkFunc func(GRPCOption) error
	}
	tests := []test{
		func() test {
			cfg := &tls.Config{}
			return test{
				name: "set success",
				args: args{
					cfg: cfg,
				},
				checkFunc: func(o GRPCOption) error {
					h := &GRPCHandler{}
					o(h)
					if !reflect.DeepEqual(h.tlsCfg, cfg) {
						return errors.New("config not match")
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WithTLSConfig(tt.args.cfg)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("WithTLSConfig() error = %v", err)
			}
		})
	}
}
