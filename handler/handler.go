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

// Func represents the a handle function type
type Func func(http.ResponseWriter, *http.Request) error

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
		ErrorHandler: handleError,
	}
}

func handleError(rw http.ResponseWriter, r *http.Request, err error) {
	if r != nil && r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}
	status := http.StatusUnauthorized
	if !strings.Contains(err.Error(), ErrMsgVerifyRoleToken) {
		glg.Debug("handleError: " + err.Error())
		status = http.StatusBadGateway
	}
	// request context canceled
	if errors.Cause(err) == context.Canceled {
		status = http.StatusRequestTimeout
	}
	rw.WriteHeader(status)
}
