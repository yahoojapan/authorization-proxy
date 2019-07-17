package handler

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strconv"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/infra"
	"github.com/yahoojapan/authorization-proxy/service"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg  config.Proxy
		bp   httputil.BufferPool
		prov service.Authorizationd
	}
	type test struct {
		name       string
		args       args
		checkFunc  func(http.Handler) error
		checkPanic func(interface{}) error
	}
	tests := []test{
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("dummyContent"))
				w.WriteHeader(http.StatusOK)
			}))
			srv := httptest.NewServer(handler)

			return test{
				name: "check request can redirect",
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
						VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
							return nil
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
					if fmt.Sprintf("%v", rw.Body) != "dummyContent" {
						return errors.Errorf("unexpected http response, got: %v, want %v", rw.Body, "dummyContent")
					}
					return nil
				},
			}
		}(),
		func() test {
			handler := http.HandlerFunc(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("dummyContent"))
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
						VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
							return errors.New("deny")
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
				w.Write([]byte("dummyContent"))
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
						VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
							return nil
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
						VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
							return nil
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
				w.Write([]byte("dummyContent"))
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
						VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) error {
							return context.Canceled
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
			defer func() {
				if tt.checkPanic != nil {
					if r := recover(); r != nil {
						if err := tt.checkPanic(r); err != nil {
							t.Errorf("New() error: %v", err)
						}
					}
				}
			}()
			got := New(tt.args.cfg, tt.args.bp, tt.args.prov)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("New() error: %v", err)
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
					err: errors.New(ErrMsgVerifyRoleToken),
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
