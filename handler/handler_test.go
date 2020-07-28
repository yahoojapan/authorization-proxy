package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	authorizerd "github.com/yahoojapan/athenz-authorizer/v4"
	"github.com/yahoojapan/athenz-authorizer/v4/access"
	"github.com/yahoojapan/athenz-authorizer/v4/role"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/v3/config"
	"github.com/yahoojapan/authorization-proxy/v3/infra"
	"github.com/yahoojapan/authorization-proxy/v3/service"
)

func TestNew(t *testing.T) {
	rt := &role.Token{
		Domain:        "domain",
		Roles:         []string{"role1", "role2", "role3"},
		Principal:     "principal",
		IntTimeStamp:  1595908257,
		ExpiryTime:    time.Now(),
		IntExpiryTime: 1595908265,
		KeyID:         "",
		Signature:     "",
		UnsignedToken: "",
	}
	bc := access.BaseClaim{jwt.StandardClaims{
		Audience:  "domain",
		ExpiresAt: 1595908265,
		IssuedAt:  1595908257,
		Subject:   "principal",
	}}
	at := &access.OAuth2AccessTokenClaim{
		AuthTime:       0,
		Version:        0,
		ClientID:       "client_id",
		UserID:         "",
		ProxyPrincipal: "",
		Scope:          []string{"role1", "role2", "role3"},
		Confirm:        nil,
		BaseClaim:      bc,
	}
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
				name: "Check that the request with roletoken headers is redirected",
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
							return rt, nil
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
					key, want = "X-Athenz-Principal", rt.GetName()
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Role", strings.Join(rt.GetRoles(), ",")
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Domain", rt.GetDomain()
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Issued-At", strconv.FormatInt(rt.GetIssueTime(), 10)
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Expires-At", strconv.FormatInt(rt.GetExpiryTime(), 10)
					if err := f(key, want); err != nil {
						return err
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
				name: "Check that the request with accesstoken headers is redirected",
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
							return at, nil
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
					key, want = "X-Athenz-Principal", rt.GetName()
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Role", strings.Join(rt.GetRoles(), ",")
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Domain", rt.GetDomain()
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Issued-At", strconv.FormatInt(rt.GetIssueTime(), 10)
					if err := f(key, want); err != nil {
						return err
					}
					key, want = "X-Athenz-Expires-At", strconv.FormatInt(rt.GetExpiryTime(), 10)
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
							return rt, nil
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
							return rt, nil
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
						return errors.Errorf("unexpected status code, got: %v, want: %v", rw.Code, http.StatusOK)
					}
					return nil
				},
			}
		}(),
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

func TestReverseProxyFatal(t *testing.T) {
	type DummyPrincipal struct {
		role.Token
	}
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
	tests := []test{
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
				name: "check fatal, new request failed",
				args: args{
					cfg: config.Proxy{
						Host: "invalid_URL_@@@",
						Port: func() uint16 {
							a, _ := strconv.ParseInt(strings.Split(srv.URL, ":")[2], 0, 64)
							return uint16(a)
						}(),
					},
					bp: infra.NewBuffer(64),
					prov: &service.AuthorizerdMock{
						VerifyFunc: func(r *http.Request, act, res string) (authorizerd.Principal, error) {
							return &DummyPrincipal{}, nil
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
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// test http.NewRequest() fatal in httputil.ReverseProxy.Director with another process (cannot include in test coverage)
			if os.Getenv("RUN_TEST_REVERSE_PROXY_FATAL") == "1" {
				got := New(tt.args.cfg, tt.args.bp, tt.args.prov)
				if err := tt.checkFunc(got); err != nil {
					t.Errorf("New() error: %v", err)
				}
				return
			}
			cmd := exec.Command(os.Args[0], "-test.run=TestReverseProxyFatal")
			cmd.Env = append(os.Environ(), "RUN_TEST_REVERSE_PROXY_FATAL=1")
			err := cmd.Run()
			if e, ok := err.(*exec.ExitError); !ok || e.ExitCode() != 1 {
				t.Errorf("process ran with err: %v, want exit status 1", err)
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
