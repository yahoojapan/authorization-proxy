package usecase

import (
	"context"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	type test struct {
		name      string
		args      args
		checkFunc func(AuthorizationDaemon) error
		wantErr   bool
	}
	tests := []test{
		func() test {
			cfg := config.Config{
				Athenz: config.Athenz{
					URL: "athenz.com",
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
				checkFunc: func(got AuthorizationDaemon) error {
					if got == nil {
						return errors.New("got is nil")
					}
					if !reflect.DeepEqual(got.(*providerDaemon).cfg, cfg) {
						return errors.New("got.cfg does not equal")
					}
					if got.(*providerDaemon).athenz == nil {
						return errors.New("got.athenz is nil")
					}
					if got.(*providerDaemon).server == nil {
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

func Test_providerDaemon_Start(t *testing.T) {
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
						StartFunc: func(context.Context) <-chan error {
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
									// prevent race with Authorizerd.Start() in select
									// also, simulate graceful shutdown
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
					errors.WithMessage(context.Canceled, "providerd: 1 times appeared"),
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
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrs[0], wantErrs) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrs, [][]error{wantErrs})
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
						StartFunc: func(context.Context) <-chan error {
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
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrs[0], wantErrs) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrs[0], wantErrs)
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
				name: "Provider daemon fails, multiple times",
				fields: fields{
					athenz: &service.AuthorizerdMock{
						StartFunc: func(context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								defer close(ech)

								// simulate fails
								ech <- errors.WithMessage(dummyErr, "provider daemon fails")
								ech <- errors.WithMessage(dummyErr, "provider daemon fails")
								ech <- errors.WithMessage(dummyErr, "provider daemon fails")

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
									// prevent race with Authorizerd.Start() in select
									// also, simulate graceful shutdown
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
					errors.WithMessage(errors.Cause(errors.WithMessage(dummyErr, "provider daemon fails")), "providerd: 3 times appeared"),
					errors.WithMessage(context.Canceled, "providerd: 1 times appeared"),
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
					if len(gotErrs) != 1 || !reflect.DeepEqual(gotErrs[0], wantErrs) {
						return errors.Errorf("Invalid err, got: %v, want: %v", gotErrs, [][]error{wantErrs})
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := &providerDaemon{
				cfg:    tt.fields.cfg,
				athenz: tt.fields.athenz,
				server: tt.fields.server,
			}
			got := g.Start(tt.args.ctx)
			if err := tt.checkFunc(got, tt.wantErrs); err != nil {
				t.Errorf("providerDaemon.Start() error: %v", err)
			}
		})
	}
}

func Test_newAuthorizationd(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	tests := []struct {
		name      string
		args      args
		checkFunc func(service.Authorizationd) error
		wantErr   bool
	}{
		{
			name: "test success new Authorization",
			args: args{
				cfg: config.Config{
					Athenz: config.Athenz{
						URL: "athenz.com",
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
				},
			},
			checkFunc: func(got service.Authorizationd) error {
				if got == nil {
					return errors.New("got: nil")
				}
				return nil
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := newAuthorizationd(tt.args.cfg)
			if err != nil && !tt.wantErr {
				t.Errorf("newAuthorizationd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err = tt.checkFunc(got); err != nil {
				t.Errorf("newAuthorizationd() error = %v", err)
				return
			}
		})
	}
}
