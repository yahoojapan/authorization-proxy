package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/infra"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg  config.Proxy
		bp   httputil.BufferPool
		prov service.Authorizationd
	}
	type test struct {
		name      string
		args      args
		checkFunc func(http.Handler) error
	}
	pm := PrincipalMock{
		NameFunc: func() string {
			return "rt_principal"
		},
		RolesFunc: func() []string {
			return []string{"rt_role1", "rt_role2", "rt_role3"}
		},
		DomainFunc: func() string {
			return "rt_domain"
		},
		IssueTimeFunc: func() int64 {
			return 1595908257
		},
		ExpiryTimeFunc: func() int64 {
			return 1595908265
		},
	}
	oatm := OAuthAccessTokenMock{
		PrincipalMock: PrincipalMock{
			NameFunc: func() string {
				return "at_principal"
			},
			RolesFunc: func() []string {
				return []string{"at_role1", "at_role2", "at_role3"}
			},
			DomainFunc: func() string {
				return "at_domain"
			},
			IssueTimeFunc: func() int64 {
				return 1595908267
			},
			ExpiryTimeFunc: func() int64 {
				return 1595908275
			},
		},
		ClientIDFunc: func() string {
			return "client_id"
		},
	}
	tests := []test{
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				header := map[string]string{
					"X-Athenz-Principal":  r.Header.Get("X-Athenz-Principal"),
					"X-Athenz-Role":       r.Header.Get("X-Athenz-Role"),
					"X-Athenz-Domain":     r.Header.Get("X-Athenz-Domain"),
					"X-Athenz-Issued-At":  r.Header.Get("X-Athenz-Issued-At"),
					"X-Athenz-Expires-At": r.Header.Get("X-Athenz-Expires-At"),
				}

				body, err1 := json.Marshal(header)
				if err1 != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}

				_, err2 := w.Write(body)
				if err2 != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "Check that the request with role token headers is redirected",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &pm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusOK {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusOK)
					}
					header := make(map[string]string)
					json.Unmarshal(rw.Body.Bytes(), &header)

					f := func(key string, want string) error {
						if header[key] != want {
							return errors.Errorf("unexpected header %v, got: %v, want %v", key, header[key], want)
						}
						return nil
					}

					var key, want string
					key, want = "X-Athenz-Principal", "rt_principal"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Role", "rt_role1,rt_role2,rt_role3"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Domain", "rt_domain"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Issued-At", "1595908257"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Expires-At", "1595908265"
					if err := f(key, want); err != nil {
						return err
					}

					key, want = "X-Athenz-Client-ID", "nil"
					if _, ok := header[key]; ok {
						return errors.Errorf("unexpected header %v, got: %v, want %v", key, header[key], want)
					}

					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				header := map[string]string{
					"X-Athenz-Principal":  r.Header.Get("X-Athenz-Principal"),
					"X-Athenz-Role":       r.Header.Get("X-Athenz-Role"),
					"X-Athenz-Domain":     r.Header.Get("X-Athenz-Domain"),
					"X-Athenz-Issued-At":  r.Header.Get("X-Athenz-Issued-At"),
					"X-Athenz-Expires-At": r.Header.Get("X-Athenz-Expires-At"),
					"X-Athenz-Client-ID":  r.Header.Get("X-Athenz-Client-ID"),
				}

				body, err1 := json.Marshal(header)
				if err1 != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}

				_, err2 := w.Write(body)
				if err2 != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "Check that the request with access token headers is redirected",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &oatm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusOK {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusOK)
					}
					header := make(map[string]string)
					json.Unmarshal(rw.Body.Bytes(), &header)

					f := func(key string, want string) error {
						if header[key] != want {
							return errors.Errorf("unexpected header %v, got: %v, want %v", key, header[key], want)
						}
						return nil
					}

					var key, want string
					key, want = "X-Athenz-Principal", "at_principal"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Role", "at_role1,at_role2,at_role3"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Domain", "at_domain"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Issued-At", "1595908267"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Expires-At", "1595908275"
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Client-ID", "client_id"
					if err := f(key, want); err != nil {
						return err
					}

					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("dummyContent"))
				if err != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check request unauthorized",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return nil, errors.New("deny")
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusUnauthorized {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusUnauthorized)
					}
					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("dummyContent"))
				if err != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check request can redirect to configured scheme",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
						Scheme: "http",
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &pm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "https://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusOK {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusOK)
					}
					if fmt.Sprintf("%v", rw.Body) != "dummyContent" {
						return errors.Errorf("unexpected http response, got: %v, want %v", rw.Body, "dummyContent")
					}
					return nil
				},
			}
		}(),
		func() test {
			return test{
				name: "check request destination cannot reach",
				args: args{
					cfg: config.Proxy{
						Host: "dummyHost",
						Port: 9999,
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &pm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusBadGateway {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusBadGateway)
					}
					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := w.Write([]byte("dummyContent"))
				if err != nil {
					w.WriteHeader(http.StatusNotImplemented)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check context done",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return nil, context.Canceled
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusRequestTimeout {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusRequestTimeout)
					}
					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check error on new request",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &pm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					r.Method = "invalid_method()"
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusBadGateway {
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusBadGateway)
					}
					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// check host header
				if r.Host != "remote.host.460" {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check custom host header is used",
				args: args{
					cfg: config.Proxy{
						Host: strings.Split(strings.Replace(srv.URL, "http://", "", 1), ":")[0],
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
						Request: config.Request{
							Host: "remote.host.460",
						},
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &pm, nil
						},
					},
				},
				checkFunc: func(h http.Handler) error {
					rw := httptest.NewRecorder()
					r := httptest.NewRequest("GET", "http://dummy.com", nil)
					h.ServeHTTP(rw, r)
					if rw.Code != http.StatusOK {
						return errors.Errorf("unexpected status code (invalid host header), got: %v, want: %v", rw.Code, http.StatusOK)
					}
					return nil
				},
			}
		}(),
		{
			name: "check custom transport is used",
			args: args{
				cfg: config.Proxy{
					Transport: config.Transport{
						MaxIdleConnsPerHost: 442,
					},
				},
			},
			checkFunc: func(h http.Handler) error {
				got := h.(*httputil.ReverseProxy).Transport.(*transport).RoundTripper.(*http.Transport).MaxIdleConnsPerHost
				want := 442
				if got != want {
					return errors.Errorf("unexpected MaxConnsPerHost in custom transport, got: %v, want: %v", got, want)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.cfg, tt.args.bp, tt.args.prov)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("New() error: %v", err)
			}
		})
	}
}

func Test_transportFromCfg(t *testing.T) {
	type args struct {
		cfg config.Transport
	}
	tests := []struct {
		name string
		args args
		want *http.Transport
	}{
		{
			name: "transport from default",
			args: args{
				cfg: config.Transport{},
			},
			want: &http.Transport{},
		},
		{
			name: "transport from custom values",
			args: args{
				cfg: config.Transport{
					TLSHandshakeTimeout:    468 * time.Second,
					DisableKeepAlives:      true,
					DisableCompression:     true,
					MaxIdleConns:           471,
					MaxIdleConnsPerHost:    472,
					MaxConnsPerHost:        473,
					IdleConnTimeout:        474 * time.Second,
					ResponseHeaderTimeout:  475 * time.Second,
					ExpectContinueTimeout:  476 * time.Second,
					MaxResponseHeaderBytes: 477,
					WriteBufferSize:        478,
					ReadBufferSize:         479,
					ForceAttemptHTTP2:      true,
				},
			},
			want: &http.Transport{
				TLSHandshakeTimeout:    468 * time.Second,
				DisableKeepAlives:      true,
				DisableCompression:     true,
				MaxIdleConns:           471,
				MaxIdleConnsPerHost:    472,
				MaxConnsPerHost:        473,
				IdleConnTimeout:        474 * time.Second,
				ResponseHeaderTimeout:  475 * time.Second,
				ExpectContinueTimeout:  476 * time.Second,
				MaxResponseHeaderBytes: 477,
				WriteBufferSize:        478,
				ReadBufferSize:         479,
				ForceAttemptHTTP2:      true,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := transportFromCfg(tt.args.cfg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transportFromCfg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleError(t *testing.T) {
	type args struct {
		rw  http.ResponseWriter
		r   *http.Request
		err error
	}
	type test struct {
		name      string
		args      args
		checkFunc func() error
	}
	tests := []test{
		func() test {
			rw := httptest.NewRecorder()
			return test{
				name: "handleError status return bad gateway",
				args: args{
					rw:  rw,
					r:   httptest.NewRequest("GET", "http://127.0.0.1", bytes.NewBufferString("test")),
					err: errors.New("other error"),
				},
				checkFunc: func() error {
					if rw.Code != http.StatusBadGateway {
						return errors.Errorf("invalid status code: %v", rw.Code)
					}
					return nil
				},
			}
		}(),
		func() test {
			rw := httptest.NewRecorder()
			return test{
				name: "handleError status return verify role token",
				args: args{
					rw:  rw,
					r:   httptest.NewRequest("GET", "http://127.0.0.1", bytes.NewBufferString("test")),
					err: errors.New(ErrMsgUnverified),
				},
				checkFunc: func() error {
					if rw.Code != http.StatusUnauthorized {
						return errors.Errorf("invalid status code: %v", rw.Code)
					}
					return nil
				},
			}
		}(),
		func() test {
			rw := httptest.NewRecorder()
			return test{
				name: "handleError status return request timeout",
				args: args{
					rw:  rw,
					r:   httptest.NewRequest("GET", "http://127.0.0.1", bytes.NewBufferString("test")),
					err: context.Canceled,
				},
				checkFunc: func() error {
					if rw.Code != http.StatusRequestTimeout {
						return errors.Errorf("invalid status code: %v", rw.Code)
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handleError(tt.args.rw, tt.args.r, tt.args.err)
			if err := tt.checkFunc(); err != nil {
				t.Errorf("handleError error: %v", err)
			}
		})
	}
}
