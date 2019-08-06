package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/service"
)

type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

// NewDebugRoutes returns debug endpoint information. If EnableDump flag is enabled then the cache dump feature endpoint will be included.
// If EnableProfiling flag is enable then the pprof interface endpoint will be included.
func NewDebugRoutes(cfg config.DebugServer, a service.Authorizationd) []Route {
	var routes []Route

	if cfg.EnableDump {
		routes = append(routes, Route{
			"GetPolicyCache",
			[]string{
				http.MethodGet,
			},
			"/debug/cache/policy",
			NewPolicyCacheHandler(a),
		})
	}

	if cfg.EnableProfiling {
		routes = append(routes, []Route{
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
		}...)
	}

	return routes
}

// NewPolicyCacheHandler returns the handler function to handle get policy cache request..
func NewPolicyCacheHandler(authd service.Authorizationd) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", fmt.Sprintf("%s;%s", "application/json", "charset=UTF-8"))
		e := json.NewEncoder(w)
		e.SetIndent("", "\t")
		pc := authd.GetPolicyCache(r.Context())
		return e.Encode(pc)
	}
}

func toHandler(f http.HandlerFunc) func(w http.ResponseWriter, r *http.Request) error {
	return func(w http.ResponseWriter, r *http.Request) error {
		f(w, r)
		return nil
	}
}
