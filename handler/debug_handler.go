package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/yahoojapan/authorization-proxy/service"
)

const (
	// ContentType represents a HTTP header name "Content-Type"
	ContentType = "Content-Type"

	// ApplicationJson represents a HTTP content type "application/json"
	ApplicationJson = "application/json"

	// CharsetUTF8 represents a UTF-8 charset for HTTP response "charset=UTF-8"
	CharsetUTF8 = "charset=UTF-8"
)

type DebugHandler struct {
	authd service.Authorizationd
}

func NewDebugHandler(authd service.Authorizationd) *DebugHandler {
	return &DebugHandler{
		authd: authd,
	}
}

func (dh *DebugHandler) GetPolicyCache(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusOK)
	w.Header().Set(ContentType, fmt.Sprintf("%s;%s", ApplicationJson, CharsetUTF8))
	e := json.NewEncoder(w)
	e.SetIndent("", "\t")
	pc := dh.authd.GetPolicyCache(r.Context())
	return e.Encode(pc)
}
