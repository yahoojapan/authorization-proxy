package router

import (
	"errors"
	"net/http"
	"testing"

	"github.com/yahoojapan/authorization-proxy/handler"
)

func TestNewDebugRoutes(t *testing.T) {
	type args struct {
		h *handler.DebugHandler
	}
	type test struct {
		name      string
		args      args
		checkFunc func(got, want []Route) error
		want      []Route
	}
	tests := []test{
		func() test {
			h := handler.NewDebugHandler(nil)
			return test{
				name: "return success",
				args: args{
					h: h,
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
						"/policyCache",
						h.GetPolicyCache,
					},
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDebugRoutes(tt.args.h)
			if err := tt.checkFunc(got, tt.want); err != nil {
				t.Errorf("NewDebugRoutes() error: %v", err)
			}
		})
	}
}
