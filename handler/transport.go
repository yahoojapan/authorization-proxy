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

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"

	"github.com/kpango/glg"
	"github.com/pkg/errors"
)

type transport struct {
	http.RoundTripper

	prov service.Authorizationd
	cfg  config.Proxy
}

func (t *transport) RoundTrip(r *http.Request) (*http.Response, error) {
	for _, urlPath := range t.cfg.BypassURLPaths {
		if urlPath == r.URL.Path {
			glg.Info("Authorization checking skipped on: " + r.URL.Path)
			return t.RoundTripper.RoundTrip(r)
		}
	}

	if err := t.prov.VerifyRoleToken(r.Context(), r.Header.Get(t.cfg.RoleHeader), r.Method, r.URL.Path); err != nil {
		return nil, errors.Wrap(err, ErrMsgVerifyRoleToken)
	}
	return t.RoundTripper.RoundTrip(r)
}
