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
	"time"

	"github.com/kpango/glg"
	"github.com/pkg/errors"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
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
				glg.Error(errors.Wrap(err, "NewRequest returned error"))
				r.URL.Scheme = scheme
				return
			}
			req.Header = r.Header
			req.TLS = r.TLS
			if cfg.PreserveHost {
				req.Host = r.Host
				glg.Debugf("proxy.PreserveHost enabled, forward host header: %s\n", req.Host)
			}
			if cfg.ForceContentLength {
				req.ContentLength = r.ContentLength
				req.TransferEncoding = r.TransferEncoding
				glg.Debugf("proxy.ForceContentLength enabled, forward content-length header: %d\n", req.ContentLength)
			}

			*r = *req
		},
		Transport: &transport{
			prov:         prov,
			RoundTripper: transportFromCfg(cfg.Transport),
			cfg:          cfg,
		},
		ErrorHandler: handleError,
	}
}

func transportFromCfg(cfg config.Transport) *http.Transport {
	isZero := func(v interface{}) bool {
		switch v.(type) {
		case int:
			return v == 0
		case int64:
			return v == 0
		case time.Duration:
			return v == time.Duration(0)
		case bool:
			return v == false
		default:
			glg.Fatal("Undefined type on proxy transport config")
			return false
		}
	}

	t := &http.Transport{}
	if !isZero(cfg.TLSHandshakeTimeout) {
		t.TLSHandshakeTimeout = cfg.TLSHandshakeTimeout
	}
	if !isZero(cfg.DisableKeepAlives) {
		t.DisableKeepAlives = cfg.DisableKeepAlives
	}
	if !isZero(cfg.DisableCompression) {
		t.DisableCompression = cfg.DisableCompression
	}
	if !isZero(cfg.MaxIdleConns) {
		t.MaxIdleConns = cfg.MaxIdleConns
	}
	if !isZero(cfg.MaxIdleConnsPerHost) {
		t.MaxIdleConnsPerHost = cfg.MaxIdleConnsPerHost
	}
	if !isZero(cfg.MaxConnsPerHost) {
		t.MaxConnsPerHost = cfg.MaxConnsPerHost
	}
	if !isZero(cfg.IdleConnTimeout) {
		t.IdleConnTimeout = cfg.IdleConnTimeout
	}
	if !isZero(cfg.ResponseHeaderTimeout) {
		t.ResponseHeaderTimeout = cfg.ResponseHeaderTimeout
	}
	if !isZero(cfg.ExpectContinueTimeout) {
		t.ExpectContinueTimeout = cfg.ExpectContinueTimeout
	}
	if !isZero(cfg.MaxResponseHeaderBytes) {
		t.MaxResponseHeaderBytes = cfg.MaxResponseHeaderBytes
	}
	if !isZero(cfg.WriteBufferSize) {
		t.WriteBufferSize = cfg.WriteBufferSize
	}
	if !isZero(cfg.ReadBufferSize) {
		t.ReadBufferSize = cfg.ReadBufferSize
	}
	if !isZero(cfg.ForceAttemptHTTP2) {
		t.ForceAttemptHTTP2 = cfg.ForceAttemptHTTP2
	}

	glg.Debugf("proxy transport: %+v\n", t)
	return t
}

func handleError(rw http.ResponseWriter, r *http.Request, err error) {
	if r != nil && r.Body != nil {
		io.Copy(ioutil.Discard, r.Body)
		r.Body.Close()
	}
	status := http.StatusUnauthorized
	if !strings.Contains(err.Error(), ErrMsgUnverified) {
		glg.Warn("handleError: " + err.Error())
		status = http.StatusBadGateway
	}
	// request context canceled
	if errors.Cause(err) == context.Canceled {
		status = http.StatusRequestTimeout
	}
	rw.WriteHeader(status)
}
