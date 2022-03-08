package router

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/pprof"
	"reflect"
	"testing"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func TestNewDebugRoutes(t *testing.T) {
	type args struct {
		cfg config.Debug
		a   service.Authorizationd
	}
	type test struct {
		name      string
		args      args
		checkFunc func(got, want []Route) error
		want      []Route
	}
	tests := []test{
		{
			name: "return all enable success",
			args: args{
				cfg: config.Debug{
					Dump:      true,
					Profiling: true,
				},
				a: nil,
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
					// toHandler() should return a new function with different pointer
					// if reflect.ValueOf(gotValue.HandlerFunc).Pointer() != reflect.ValueOf(wantValue.HandlerFunc).Pointer() {
					// 	return errors.New(gotValue.Name + " handler not match")
					// }
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
				{
					"Debug pprof",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/",
					toHandler(pprof.Index),
				},
				{
					"Debug cmdline",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/cmdline",
					toHandler(pprof.Cmdline),
				},
				{
					"Debug profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/profile",
					toHandler(pprof.Profile),
				},
				{
					"Debug symbol profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/symbol",
					toHandler(pprof.Symbol),
				},
				{
					"Debug trace profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/trace",
					toHandler(pprof.Trace),
				},
				{
					"Debug heap profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/heap",
					toHandler(pprof.Handler("heap").ServeHTTP),
				},
				{
					"Debug goroutine profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/goroutine",
					toHandler(pprof.Handler("goroutine").ServeHTTP),
				},
				{
					"Debug thread profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/threadcreate",
					toHandler(pprof.Handler("threadcreate").ServeHTTP),
				},
				{
					"Debug block profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/block",
					toHandler(pprof.Handler("block").ServeHTTP),
				},
			},
		},
		{
			name: "return enable dump only success",
			args: args{
				cfg: config.Debug{
					Dump:      true,
					Profiling: false,
				},
				a: nil,
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
						return errors.New(gotValue.Name + " handler not match")
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
		},
		{
			name: "return enable profiling success",
			args: args{
				cfg: config.Debug{
					Dump:      false,
					Profiling: true,
				},
				a: nil,
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
					// toHandler() should return a new function with different pointer
					// if reflect.ValueOf(gotValue.HandlerFunc).Pointer() != reflect.ValueOf(wantValue.HandlerFunc).Pointer() {
					// 	return errors.New(gotValue.Name + " handler not match")
					// }
				}
				return nil
			},
			want: []Route{
				{
					"Debug pprof",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/",
					toHandler(pprof.Index),
				},
				{
					"Debug cmdline",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/cmdline",
					toHandler(pprof.Cmdline),
				},
				{
					"Debug profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/profile",
					toHandler(pprof.Profile),
				},
				{
					"Debug symbol profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/symbol",
					toHandler(pprof.Symbol),
				},
				{
					"Debug trace profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/trace",
					toHandler(pprof.Trace),
				},
				{
					"Debug heap profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/heap",
					toHandler(pprof.Handler("heap").ServeHTTP),
				},
				{
					"Debug goroutine profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/goroutine",
					toHandler(pprof.Handler("goroutine").ServeHTTP),
				},
				{
					"Debug thread profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/threadcreate",
					toHandler(pprof.Handler("threadcreate").ServeHTTP),
				},
				{
					"Debug block profile",
					[]string{
						http.MethodGet,
					},
					"/debug/pprof/block",
					toHandler(pprof.Handler("block").ServeHTTP),
				},
			},
		},
		{
			name: "disable all and return success",
			args: args{
				cfg: config.Debug{
					Dump:      false,
					Profiling: false,
				},
				a: nil,
			},
			checkFunc: func(got, want []Route) error {
				if len(got) != len(want) {
					return fmt.Errorf("got: %v, want: %v", got, want)
				}
				return nil
			},
			want: nil,
		},
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
