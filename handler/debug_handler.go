package handler

import (
	"fmt"
	"net/http"

	"github.com/yahoojapan/authorization-proxy/service"
)

const (
	// ContentType represents a HTTP header name "Content-Type"
	ContentType = "Content-Type"

	// ApplicationJson represents a HTTP content type "application/json"
	ApplicationJSON = "application/json"

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
	/*
		e := json.NewEncoder(w)
		e.SetIndent("", "\t")
		return e.Encode(dh.authd.GetPolicyCache(r.Context()))
	*/
	_, err := fmt.Fprintf(w, "%+s", dh.authd.GetPolicyCache(r.Context()))
	return err
}
