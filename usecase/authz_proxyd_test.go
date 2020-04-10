package usecase

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"

	"github.com/pkg/errors"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	type test struct {
		name      string
		args      args
		checkFunc func(AuthzProxyDaemon) error
		wantErr   bool
	}
	tests := []test{
		func() test {
			cfg := config.Config{
				Athenz: config.Athenz{
					URL: "athenz.io",
				},
				Authorization: config.Authorization{
					PubKeyRefreshDuration: "10s",
					PubKeySysAuthDomain:   "dummy.sys.auth",
					PubKeyEtagExpTime:     "10s",
					PubKeyEtagFlushDur:    "10s",
					AthenzDomains:         []string{"dummyDom1", "dummyDom2"},
					PolicyExpireMargin:    "10s",
					PolicyRefreshDuration: "10s",
					PolicyEtagExpTime:     "10s",
					PolicyEtagFlushDur:    "10s",
				},
				Server: config.Server{
					HealthzPath: "/dummy",
				},
				Proxy: config.Proxy{
					BufferSize: 512,
				},
			}
			return test{
				name: "new success",
				args: args{
					cfg: cfg,
				},
				checkFunc: func(got AuthzProxyDaemon) error {
					if got == nil {
						return errors.New("got is nil")
					}
					if !reflect.DeepEqual(got.(*authzProxyDaemon).cfg, cfg) {
						return errors.New("got.cfg does not equal")
					}
					if got.(*authzProxyDaemon).athenz == nil {
						return errors.New("got.athenz is nil")
					}
					if got.(*authzProxyDaemon).server == nil {
						return errors.New("got.server is nil")
					}
					return nil
				},
			}
		}(),
		{
			name: "new error",
			args: args{
				cfg: config.Config{
					Authorization: config.Authorization{
						PubKeyRefreshDuration: "dummy",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				if err = tt.checkFunc(got); err != nil {
					t.Errorf("New() error = %v", err)
					return
				}
			}
		})
	}
}

func Test_authzProxyDaemon_Init(t *testing.T) {
	type fields struct {
		athenz service.Authorizationd
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErrStr string
	}{
		{
			name: "init sccuess",
			fields: fields{
				athenz: &service.AuthorizerdMock{
					InitFunc: func(ctx context.Context) error {
						return nil
					},
				},
			},
			args:       args{ctx: context.Background()},
			wantErrStr: "",
		},
		{
			name: "init fail",
			fields: fields{
				athenz: &service.AuthorizerdMock{
					InitFunc: func(ctx context.Context) error {
						return errors.New("authorizerd error")
					},
				},
			},
			args:       args{ctx: context.Background()},
			wantErrStr: "authorizerd error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &authzProxyDaemon{
				athenz: tt.fields.athenz,
			}
			err := g.Init(tt.args.ctx)
			if (err == nil && tt.wantErrStr != "") || (err != nil && err.Error() != tt.wantErrStr) {
				t.Errorf("authzProxyDaemon.Init() error = %v, wantErr %v", err, tt.wantErrStr)
				return
			}
		})
	}
}

func Test_authzProxyDaemon_Start(t *testing.T) {
	type fields struct {
		cfg    config.Config
		athenz service.Authorizationd
		server service.Server
	}
	type args struct {
		ctx context.Context
	}
	type test struct {
		name      string
		fields    fields
		args      args
		wantErrs  []error
		checkFunc func(<-chan []error, []error) error
	}
	getLessErrorFunc := func(errs []error) func(i, j int) bool {
		return func(i, j int) bool {
			return errs[i].Error() < errs[j].Error()
		}
	}
	tests := []test{
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			return test{
				name: "Daemon start success",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(ctx context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									ech <- ctx.Err()
									return
								}
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									// simulate graceful shutdown
									time.Sleep(1 * time.Millisecond)

									ech <- []error{ctx.Err()}
									return
								}
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				wantErrs: []error{
					errors.WithMessage(context.Canceled, "authorizerd: 1 times appeared"),
					context.Canceled,
				},
				checkFunc: func(got <-chan []error, wantErrs []error) error {
					cancel()
					mux := &sync.Mutex{}

					gotErrs := make([][]error, 0)
					mux.Lock()
					go func() {
						defer mux.Unlock()
						select {
						case err, ok := <-got:
							if !ok {
								return
							}
							gotErrs = append(gotErrs, err)
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()

					// check only send errors once and the errors are expected ignoring order
					sort.Slice(gotErrs[0], getLessErrorFunc(gotErrs[0]))
					sort.Slice(wantErrs, getLessErrorFunc(wantErrs))
					gotErrsStr := fmt.Sprintf("%v", gotErrs[0])
					wantErrsStr := fmt.Sprintf("%v", wantErrs)
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrsStr, wantErrsStr) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrsStr, wantErrsStr)
					}

					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			dummyErr := errors.New("dummy")
			return test{
				name: "Server fails",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(ctx context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									ech <- ctx.Err()
									return
								}
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								defer close(ech)
								ech <- []error{errors.WithMessage(dummyErr, "server fails")}
								return
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				wantErrs: []error{
					errors.WithMessage(context.Canceled, "authorizerd: 1 times appeared"),
					errors.WithMessage(dummyErr, "server fails"),
				},
				checkFunc: func(got <-chan []error, wantErrs []error) error {
					mux := &sync.Mutex{}

					gotErrs := make([][]error, 0)
					mux.Lock()
					go func() {
						defer mux.Unlock()
						select {
						case err, ok := <-got:
							if !ok {
								return
							}
							gotErrs = append(gotErrs, err)
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()

					// check only send errors once and the errors are expected ignoring order
					sort.Slice(gotErrs[0], getLessErrorFunc(gotErrs[0]))
					sort.Slice(wantErrs, getLessErrorFunc(wantErrs))
					gotErrsStr := fmt.Sprintf("%v", gotErrs[0])
					wantErrsStr := fmt.Sprintf("%v", wantErrs)
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrsStr, wantErrsStr) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrsStr, wantErrsStr)
					}

					cancel()
					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			dummyErr := errors.New("dummy")
			return test{
				name: "Authorizer daemon fails, multiple times",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(ctx context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)

								// simulate fails
								ech <- errors.WithMessage(dummyErr, "authorizer daemon fails")
								ech <- errors.WithMessage(dummyErr, "authorizer daemon fails")
								ech <- errors.WithMessage(dummyErr, "authorizer daemon fails")

								// return only if context cancel
								select {
								case <-ctx.Done():
									ech <- ctx.Err()
									return
								}
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									// simulate graceful shutdown
									time.Sleep(1 * time.Millisecond)

									ech <- []error{ctx.Err()}
									return
								}
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				wantErrs: []error{
					errors.WithMessage(errors.Cause(errors.WithMessage(dummyErr, "authorizer daemon fails")), "authorizerd: 3 times appeared"),
					errors.WithMessage(context.Canceled, "authorizerd: 1 times appeared"),
					context.Canceled,
				},
				checkFunc: func(got <-chan []error, wantErrs []error) error {
					cancel()
					mux := &sync.Mutex{}

					gotErrs := make([][]error, 0)
					mux.Lock()
					go func() {
						defer mux.Unlock()
						select {
						case err, ok := <-got:
							if !ok {
								return
							}
							gotErrs = append(gotErrs, err)
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()

					// check only send errors once and the errors are expected ignoring order
					sort.Slice(gotErrs[0], getLessErrorFunc(gotErrs[0]))
					sort.Slice(wantErrs, getLessErrorFunc(wantErrs))
					gotErrsStr := fmt.Sprintf("%v", gotErrs[0])
					wantErrsStr := fmt.Sprintf("%v", wantErrs)
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrsStr, wantErrsStr) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrsStr, wantErrsStr)
					}

					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			return test{
				name: "Daemon start end successfully, server shutdown without error",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(ctx context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)
								// return only if context cancel
								select {
								case <-ctx.Done():
									ech <- ctx.Err()
									return
								}
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									// simulate graceful shutdown
									time.Sleep(1 * time.Millisecond)

									ech <- []error{}
									return
								}
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				wantErrs: []error{
					errors.WithMessage(context.Canceled, "authorizerd: 1 times appeared"),
					errors.New(""),
				},
				checkFunc: func(got <-chan []error, wantErrs []error) error {
					cancel()
					mux := &sync.Mutex{}

					gotErrs := make([][]error, 0)
					mux.Lock()
					go func() {
						defer mux.Unlock()
						select {
						case err, ok := <-got:
							if !ok {
								return
							}
							gotErrs = append(gotErrs, err)
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()

					// check only send errors once and the errors are expected ignoring order
					sort.Slice(gotErrs[0], getLessErrorFunc(gotErrs[0]))
					sort.Slice(wantErrs, getLessErrorFunc(wantErrs))
					gotErrsStr := fmt.Sprintf("%v", gotErrs[0])
					wantErrsStr := fmt.Sprintf("%v", wantErrs)
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrsStr, wantErrsStr) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrsStr, wantErrsStr)
					}

					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			dummyErr := errors.New("dummy")
			return test{
				name: "Daemon start end successfully, server shutdown >1 errors",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(ctx context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)
								// return only if context cancel
								select {
								case <-ctx.Done():
									ech <- ctx.Err()
									return
								}
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								defer close(ech)
								select {
								case <-ctx.Done():
									// simulate graceful shutdown
									time.Sleep(1 * time.Millisecond)

									ech <- []error{dummyErr, ctx.Err()}
									return
								}
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				wantErrs: []error{
					errors.WithMessage(context.Canceled, "authorizerd: 1 times appeared"),
					errors.Wrap(dummyErr, context.Canceled.Error()),
				},
				checkFunc: func(got <-chan []error, wantErrs []error) error {
					cancel()
					mux := &sync.Mutex{}

					gotErrs := make([][]error, 0)
					mux.Lock()
					go func() {
						defer mux.Unlock()
						select {
						case err, ok := <-got:
							if !ok {
								return
							}
							gotErrs = append(gotErrs, err)
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()

					// check only send errors once and the errors are expected ignoring order
					sort.Slice(gotErrs[0], getLessErrorFunc(gotErrs[0]))
					sort.Slice(wantErrs, getLessErrorFunc(wantErrs))
					gotErrsStr := fmt.Sprintf("%v", gotErrs[0])
					wantErrsStr := fmt.Sprintf("%v", wantErrs)
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrsStr, wantErrsStr) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrsStr, wantErrsStr)
					}

					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &authzProxyDaemon{
				cfg:    tt.fields.cfg,
				athenz: tt.fields.athenz,
				server: tt.fields.server,
			}
			got := g.Start(tt.args.ctx)
			if err := tt.checkFunc(got, tt.wantErrs); err != nil {
				t.Errorf("authzProxyDaemon.Start() error: %v", err)
			}
		})
	}
}

// this requires integration test
func Test_newAuthzD(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	tests := []struct {
		name       string
		args       args
		want       bool
		wantErrStr string
	}{
		{
			name: "test new Authorization fail",
			args: args{
				cfg: config.Config{
					Authorization: config.Authorization{
						PubKeyRefreshDuration: "invalid_duration",
					},
				},
			},
			want:       false,
			wantErrStr: "error create pubkeyd: invalid refresh duration: time: invalid duration invalid_duration",
		},
		{
			name: "test success new Authorization",
			args: args{
				cfg: config.Config{
					Athenz: config.Athenz{
						URL: "athenz.io",
					},
					Proxy: config.Proxy{
						RoleHeader: "Athenz-Role-Auth",
					},
					Authorization: config.Authorization{
						PubKeyRefreshDuration: "10s",
						PubKeySysAuthDomain:   "dummy.sys.auth",
						PubKeyEtagExpTime:     "10s",
						PubKeyEtagFlushDur:    "10s",
						AthenzDomains:         []string{"dummyDom1", "dummyDom2"},
						PolicyExpireMargin:    "10s",
						PolicyRefreshDuration: "10s",
						PolicyEtagExpTime:     "10s",
						PolicyEtagFlushDur:    "10s",
						Role: config.Role{
							Enable: true,
						},
					},
				},
			},
			want: true,
		},
		{
			name: "test success access token enable/disable",
			args: args{
				cfg: config.Config{
					Authorization: config.Authorization{
						PubKeyRefreshDuration: "10s",
						PubKeyEtagExpTime:     "10s",
						PubKeyEtagFlushDur:    "10s",
						PolicyExpireMargin:    "10s",
						PolicyRefreshDuration: "10s",
						PolicyEtagExpTime:     "10s",
						PolicyEtagFlushDur:    "10s",
						Access: config.Access{
							Enable:               true,
							VerifyCertThumbprint: false,
							CertBackdateDur:      "10s",
							CertOffsetDur:        "10s",
						},
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAuthzD(tt.args.cfg)

			if (err == nil && tt.wantErrStr != "") || (err != nil && err.Error() != tt.wantErrStr) {
				t.Errorf("newAuthzD() error = %v, wantErr %v", err, tt.wantErrStr)
				return
			}
			if (got != nil) != tt.want {
				t.Errorf("newAuthzD()= %v, want= %v", got, tt.want)
				return
			}
		})
	}
}
