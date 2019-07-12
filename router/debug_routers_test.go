package router

import (
	"errors"
	"net/http"
	"testing"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"
)

func TestNewDebugRoutes(t *testing.T) {
	type args struct {
		cfg config.Server
		a   service.Authorizationd
	}
	type test struct {
		name      string
		args      args
		checkFunc func(got, want []Route) error
		want      []Route
	}
	tests := []test{
		func() test {
			return test{
				name: "return success",
				args: args{
					cfg: config.Server{},
					a:   nil,
				},
				checkFunc: func(got, want []Route) error {
					if got[0].Name != want[0].Name {
						return errors.New("not match")
					}
					return nil
				},
				want: []Route{
					{
						"GetPolicyCache",
						[]string{
							http.MethodGet,
						},
						"debug/policy-cache",
						NewPolicyCacheHandler(nil),
					},
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDebugRoutes(tt.args.cfg, tt.args.a)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("NewDebugRoutes() error: %v", err)
			}
		})
	}
}
