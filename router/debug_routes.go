package router

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/handler"
	"github.com/yahoojapan/authorization-proxy/service"
)

// Route represents aa API endpoint and its handler function
type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

// NewDebugRoutes returns an array of supporting Route
func NewDebugRoutes(cfg config.Server, a service.Authorizationd) []Route {
	return []Route{
		{
			"GetPolicyCache",
			[]string{
				http.MethodGet,
			},
			"/debug/cache/policy",
			NewPolicyCacheHandler(a),
		},
	}
}

// NewPolicyCacheHandler returns a handler function for getting policy cache
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
