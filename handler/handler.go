package handler

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/kpango/glg"
	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"
)

// New creates a handler for handling different HTTP requests based on the given services. It also contains a reverse proxy for handling proxy request.
func New(cfg config.Proxy, bp httputil.BufferPool, prov service.Authorizationd) http.Handler {
	scheme := "http"
	if cfg.Scheme != "" {
		scheme = cfg.Scheme
	}

	host := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return &httputil.ReverseProxy{
		BufferPool: bp,
		Director: func(r *http.Request) {
			u := *r.URL
			u.Scheme = scheme
			u.Host = host
			req, err := http.NewRequest(r.Method, u.String(), r.Body)
			if err != nil {
				glg.Fatal(errors.Wrap(err, "NewRequest returned error"))
			}
			req.Header = r.Header
			*r = *req
		},
		Transport: &transport{
			prov:         prov,
			RoundTripper: &http.Transport{},
			cfg:          cfg,
		},
		ErrorHandler: func(rw http.ResponseWriter, r *http.Request, err error) {
			if r != nil && r.Body != nil {
				io.Copy(ioutil.Discard, r.Body)
				r.Body.Close()
			}
			status := http.StatusUnauthorized
			if strings.Index(err.Error(), ErrMsgVerifyRoleToken) < 0 {
				status = http.StatusBadGateway
			}
			// request context canceled
			if errors.Cause(err) == context.Canceled {
				status = http.StatusRequestTimeout
			}
			rw.WriteHeader(status)
		},
	}
}
