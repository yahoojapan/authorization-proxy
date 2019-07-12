package router

import (
	"errors"
	"net/http"
	"reflect"
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
					for i, gotValue := range got {
						wantValue := want[i]
						if gotValue.Name != wantValue.Name {
							return errors.New("name not match")
						}
						if !reflect.DeepEqual(gotValue.Methods, wantValue.Methods) {
							return errors.New("methods not match")
						}
						if gotValue.Pattern != wantValue.Pattern {
							return errors.New("pattern not match")
						}
						if reflect.ValueOf(gotValue.HandlerFunc).Pointer() != reflect.ValueOf(wantValue.HandlerFunc).Pointer() {
							return errors.New("handler not match")
						}
					}
					return nil
				},
				want: []Route{
					{
						"GetPolicyCache",
						[]string{
							http.MethodGet,
						},
						"/debug/cache/policy",
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
