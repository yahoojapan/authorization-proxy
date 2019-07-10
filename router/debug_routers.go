package router

import (
	"net/http"

	"github.com/yahoojapan/authorization-proxy/handler"
)

type Route struct {
	Name        string
	Methods     []string
	Pattern     string
	HandlerFunc handler.Func
}

func NewDebugRoutes(h *handler.DebugHandler) []Route {
	return []Route{
		{
			"GetPolicyCache",
			[]string{
				http.MethodGet,
			},
			"/policyCache",
			h.GetPolicyCache,
		},
	}
}
