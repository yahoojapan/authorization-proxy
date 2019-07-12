package router

import (
	"encoding/json"
	"fmt"
	"net/http"

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

func NewDebugRoutes(cfg config.Server, a service.Authorizationd) []Route {
	return []Route{
		{
			"GetPolicyCache",
			[]string{
				http.MethodGet,
			},
			"/debug/policy-cache",
			NewPolicyCacheHandler(a),
		},
	}
}

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
