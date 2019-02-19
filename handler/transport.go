package handler

import (
	"net/http"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"

	"github.com/pkg/errors"
)

type transport struct {
	http.RoundTripper

	prov service.Authorizationd
	cfg  config.Proxy
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	//TODO check RoleToken Here
	if err := t.prov.VerifyRoleToken(r.Context(), r.Header.Get(t.cfg.RoleHeader), r.Method, r.URL.Path); err != nil {
		return nil, errors.Wrap(err, "VerifyRoleToken returned error in RoundTrip")
	}

	return t.RoundTripper.RoundTrip(r)
}
