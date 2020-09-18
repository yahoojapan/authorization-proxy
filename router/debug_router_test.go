package router

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kpango/glg"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/handler"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func TestNewDebugRouter(t *testing.T) {
	type args struct {
		cfg config.Server
		a   service.Authorizationd
	}
	tests := []struct {
		name      string
		args      args
		checkFunc func(got *http.ServeMux) error
	}{
		{
			name: "new debug router success without routes",
			args: args{
				cfg: config.Server{},
				a:   nil,
			},
			checkFunc: func(got *http.ServeMux) error {
				req := httptest.NewRequest("GET", "http://athenz.io/debug/pprof/", nil)
				h, p := got.Handler(req)
				if reflect.ValueOf(h).Pointer() != reflect.ValueOf(http.NotFoundHandler()).Pointer() {
					return errors.New("HTTP handler is not NotFoundHandler")
				}
				if p != "" {
					return fmt.Errorf("p is not nil: %v", p)
				}
				return nil
			},
		},
		{
			name: "new debug router success with routes",
			args: args{
				cfg: config.Server{
					Debug: config.Debug{
						Profiling: true,
					},
				},
				a: nil,
			},
			checkFunc: func(got *http.ServeMux) error {
				req := httptest.NewRequest("GET", "http://athenz.io/debug/pprof/", nil)
				h, p := got.Handler(req)
				if reflect.ValueOf(h).Pointer() == reflect.ValueOf(http.NotFoundHandler()).Pointer() {
					return errors.New("HTTP handler is NotFoundHandler")
				}
				wantPattern := "/debug/pprof/"
				if p != wantPattern {
					return fmt.Errorf("p got: %v, want: %v", p, wantPattern)
				}
				return nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDebugRouter(tt.args.cfg, tt.args.a)
			if err := tt.checkFunc(got); err != nil {
				t.Errorf("NewDebugRouter() err: %v", err)
			}
		})
	}
}

func Test_routing(t *testing.T) {
	type args struct {
		m []string
		t time.Duration
		h handler.Func
	}
	type test struct {
		name      string
		args      args
		checkFunc func(http.Handler) error
	}
	tests := []test{
		func() test {
			testStr := "testhoge"
			want := testStr
			wantStatusCode := http.StatusOK

			return test{
				name: "Check whether Handler can handle request: single HTTP method",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: time.Second * 10,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						_, err := rw.Write([]byte(testStr))
						if err != nil {
							return err
						}
						return nil
					},
				},
				checkFunc: func(server http.Handler) error {
					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()
					server.ServeHTTP(record, request)
					response := record.Result()

					defer response.Body.Close()

					byteArray, _ := ioutil.ReadAll(response.Body)
					got := string(byteArray)
					gotStatusCode := response.StatusCode

					if got != want || gotStatusCode != wantStatusCode {
						return fmt.Errorf("Handler could not handle the request: request: %v  got response: %v  want: %v  got statuscode: %d  want statuscode: %d", request, got, want, gotStatusCode, wantStatusCode)
					}

					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := testStr
			wantStatusCode := http.StatusOK

			return test{
				name: "Check whether Handler can handle request: multiple HTTP methods",
				args: args{
					m: []string{
						http.MethodGet,
						http.MethodPost,
					},
					t: time.Second * 10,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						_, err := rw.Write([]byte(testStr))
						if err != nil {
							return err
						}
						return nil
					},
				},
				checkFunc: func(server http.Handler) error {
					methods := []string{
						http.MethodGet,
						http.MethodPost,
					}
					for _, method := range methods {
						request := httptest.NewRequest(method, "/", nil)
						record := httptest.NewRecorder()
						server.ServeHTTP(record, request)
						response := record.Result()

						defer response.Body.Close()

						byteArray, _ := ioutil.ReadAll(response.Body)
						got := string(byteArray)
						gotStatusCode := response.StatusCode

						if got != want || gotStatusCode != wantStatusCode {
							return fmt.Errorf("Handler could not handle the request: request: %v  got response: %v  want: %v  got statuscode: %d  want statuscode: %d", request, got, want, gotStatusCode, wantStatusCode)
						}
					}
					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := "Error: " + testStr + "\t" + http.StatusText(http.StatusInternalServerError) + "\n"
			wantStatusCode := http.StatusInternalServerError

			return test{
				name: "Check whether Handler returns 'Internal Server Error' status when error occurs",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: time.Second * 10,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						return fmt.Errorf(testStr)
					},
				},
				checkFunc: func(server http.Handler) error {

					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()
					server.ServeHTTP(record, request)
					response := record.Result()

					defer response.Body.Close()

					byteArray, _ := ioutil.ReadAll(response.Body)
					got := string(byteArray)
					gotStatusCode := response.StatusCode

					if got != want || gotStatusCode != wantStatusCode {
						return fmt.Errorf("Handler could not handle the request: request: %v  got response: %v  want: %v  got statuscode: %d  want statuscode: %d", request, got, want, gotStatusCode, wantStatusCode)
					}

					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := "Method: GET" + "\t" + http.StatusText(http.StatusMethodNotAllowed) + "\n"
			wantStatusCode := http.StatusMethodNotAllowed

			return test{
				name: "Check whether Handler returns 'Method Not Allowed' when requested invalid HTTP method: no matches in the list",
				args: args{
					m: []string{
						http.MethodHead,
					},
					t: time.Second * 10,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						return fmt.Errorf(testStr)
					},
				},
				checkFunc: func(server http.Handler) error {

					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()
					server.ServeHTTP(record, request)
					response := record.Result()

					defer response.Body.Close()

					byteArray, _ := ioutil.ReadAll(response.Body)
					got := string(byteArray)
					gotStatusCode := response.StatusCode

					if got != want || gotStatusCode != wantStatusCode {
						return fmt.Errorf("Handler could not handle the request: request: %v  got response: %v  want: %v  got statuscode: %d  want statuscode: %d", request, got, want, gotStatusCode, wantStatusCode)
					}

					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := "Method: GET" + "\t" + http.StatusText(http.StatusMethodNotAllowed) + "\n"
			wantStatusCode := http.StatusMethodNotAllowed

			return test{
				name: "Check whether Handler returns 'Method Not Allowed' when requested invalid HTTP method: the list is empty",
				args: args{
					t: time.Second * 10,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						return fmt.Errorf(testStr)
					},
				},
				checkFunc: func(server http.Handler) error {

					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()
					server.ServeHTTP(record, request)
					response := record.Result()

					defer response.Body.Close()

					byteArray, _ := ioutil.ReadAll(response.Body)
					got := string(byteArray)
					gotStatusCode := response.StatusCode

					if got != want || gotStatusCode != wantStatusCode {
						return fmt.Errorf("Handler could not handle the request: request: %v  got response: %v  want: %v  got statuscode: %d  want statuscode: %d", request, got, want, gotStatusCode, wantStatusCode)
					}

					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := "Handler Time Out:"

			timeoutSec := time.Second * 1
			waitSec := time.Second * 10

			return test{
				name: "Check whether logging when timeout",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: timeoutSec,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						time.Sleep(waitSec)
						_, err := rw.Write([]byte(testStr))
						if err != nil {
							return err
						}
						return nil
					},
				},
				checkFunc: func(server http.Handler) error {
					// set stdlog output destination
					logBuffer := new(bytes.Buffer)
					glg.Get().SetMode(glg.WRITER).SetWriter(logBuffer)

					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()
					server.ServeHTTP(record, request)
					response := record.Result()

					defer response.Body.Close()

					// check error message
					got := logBuffer.String()
					if !strings.Contains(got, want) {
						return fmt.Errorf("Handler could not write error log: request: %v  got: %v  want: %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() test {
			testStr := "testhoge"
			want := "Handler Time Out:"

			timeoutSec := time.Second * 1
			waitSec := time.Second * 10

			return test{
				name: "Check whether Handler can handle the request when parent context closed",
				args: args{
					m: []string{
						http.MethodGet,
					},
					t: waitSec,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						time.Sleep(waitSec)
						_, err := rw.Write([]byte(testStr))
						if err != nil {
							return err
						}
						return nil
					},
				},
				checkFunc: func(server http.Handler) error {
					// set stdlog output destination
					logBuffer := new(bytes.Buffer)
					glg.Get().SetMode(glg.WRITER).SetWriter(logBuffer)

					request := httptest.NewRequest(http.MethodGet, "/", nil)
					record := httptest.NewRecorder()

					ctx, cancel := context.WithCancel(context.Background())
					go func() {
						time.Sleep(timeoutSec)
						cancel()
					}()

					server.ServeHTTP(record, request.WithContext(ctx))

					// check error message
					got := logBuffer.String()
					if !strings.Contains(got, want) {
						return fmt.Errorf("Handler could not write error log: request: %v  got: %v  want: %v", request, got, want)
					}

					return nil
				},
			}
		}(),
		func() test {
			timeoutSec := time.Second * 1
			want := io.ErrClosedPipe

			return test{
				name: "Check whether Handler can handle unexpected HTTP request and can write error log",
				args: args{
					m: []string{},
					t: timeoutSec,
					h: func(rw http.ResponseWriter, r *http.Request) error {
						return nil
					},
				},
				checkFunc: func(server http.Handler) (testErr error) {
					// set stdlog output destination
					pipeReader, pipeWriter := io.Pipe()
					glg.Get().SetMode(glg.WRITER).SetWriter(pipeWriter)
					pipeWriter.Close()
					pipeReader.Close()

					// prepare closed pipe for request
					requestPipeReader, requestPipeWriter := io.Pipe()
					requestPipeWriter.Close()
					requestPipeReader.Close()

					request := httptest.NewRequest(http.MethodGet, "/", requestPipeReader)
					record := httptest.NewRecorder()

					defer func() {
						got := recover()
						if got != want {
							testErr = fmt.Errorf("error occurred: got: %v  want: %v", got, want)
						}
					}()

					server.ServeHTTP(record, request)

					return nil
				},
			}
		}(),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := routing(tt.args.m, tt.args.t, tt.args.h)
			if err := tt.checkFunc(got); err != nil {
				t.Error(err)
			}
		})
	}
}
