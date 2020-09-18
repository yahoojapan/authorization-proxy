/*
Copyright (C)  2018 Yahoo Japan Corporation Athenz team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/handler"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

// Route contains information and handler of an API endpoint
type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

// NewDebugRoutes returns debug endpoint information. If Dump flag is enabled then the cache dump feature endpoint will be included.
// If Profiling flag is enable then the pprof interface endpoint will be included.
func NewDebugRoutes(cfg config.Debug, a service.Authorizationd) []Route {
	var routes []Route

	if cfg.Dump {
		routes = append(routes, Route{
			"GetPolicyCache",
			[]string{
				http.MethodGet,
			},
			"/debug/cache/policy",
			NewPolicyCacheHandler(a),
		})
	}

	if cfg.Profiling {
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

// NewPolicyCacheHandler returns the handler function to handle get policy cache request.
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
