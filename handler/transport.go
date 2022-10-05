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
	"net/http"
	"strconv"
	"strings"

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
)

type transport struct {
	http.RoundTripper

	prov service.Authorizationd
	cfg  config.Proxy
}

// Based on the following.
// https://github.com/golang/oauth2/blob/bf48bf16ab8d622ce64ec6ce98d2c98f916b6303/transport.go
func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	for _, urlPath := range t.cfg.OriginHealthCheckPaths {
		if urlPath == r.URL.Path {
			glg.Info("Authorization checking skipped on: " + r.URL.Path)
			r.TLS = nil
			return t.RoundTripper.RoundTrip(r)
		}
	}

	reqBodyClosed := false
	if r.Body != nil {
		defer func() {
			if !reqBodyClosed {
				r.Body.Close()
			}
		}()
	}

	p, err := t.prov.Authorize(r, r.Method, r.URL.Path)
	if err != nil {
		glg.Infof("Got unathorizated access: %s %s", r.Method, r.URL.String())
		return nil, errors.Wrap(err, ErrMsgUnverified)
	}

	req2 := cloneRequest(r) // per RoundTripper contract

	req2.Header.Set("X-Athenz-Principal", p.Name())
	req2.Header.Set("X-Athenz-Role", strings.Join(p.Roles(), ","))
	req2.Header.Set("X-Athenz-Domain", p.Domain())
	req2.Header.Set("X-Athenz-Issued-At", strconv.FormatInt(p.IssueTime(), 10))
	req2.Header.Set("X-Athenz-Expires-At", strconv.FormatInt(p.ExpiryTime(), 10))

	if c, ok := p.(authorizerd.OAuthAccessToken); ok {
		req2.Header.Set("X-Athenz-Client-ID", c.ClientID())
	}

	req2.TLS = nil
	// req.Body is assumed to be closed by the base RoundTripper.
	reqBodyClosed = true
	return t.RoundTripper.RoundTrip(req2)
}

// cloneRequest returns a clone of the provided *http.Request.
// The clone is a shallow copy of the struct and its Header map.
func cloneRequest(r *http.Request) *http.Request {
	// shallow copy of the struct
	r2 := new(http.Request)
	*r2 = *r
	// deep copy of the Header
	r2.Header = make(http.Header, len(r.Header))
	for k, s := range r.Header {
		r2.Header[k] = append([]string(nil), s...)
	}
	return r2
}
