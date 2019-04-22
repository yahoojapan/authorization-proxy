package usecase

import (
	"context"
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
	tests := []struct {
		name      string
		args      args
		checkFunc func(AuthorizationDaemon) error
		wantErr   bool
	}{
		{
			name: "new success",
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
						AthenzDomains:             []string{"dummyDom1", "dummyDom2"},
						PolicyExpireMargin:        "10s",
						PolicyRefreshDuration:     "10s",
						PolicyEtagExpTime:         "10s",
						PolicyEtagFlushDur:        "10s",
					},
					Server: config.Server{
						HealthzPath: "/dummy",
					},
					Proxy: config.Proxy{
						BufferSize: 512,
					},
				},
			},
			checkFunc: func(got AuthorizationDaemon) error {
				if got == nil {
					return errors.New("got is nil")
				}
				return nil
			},
		},
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
		checkFunc func(<-chan []error) error
	}
	tests := []test{
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			return test{
				name: "Daemon start success",
				fields: fields{
					athenz: &service.AuthorizedMock{
						StartProviderdFunc: func(context.Context) <-chan error {
							return make(chan error)
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				checkFunc: func(got <-chan []error) error {
					cancel()
					mux := &sync.Mutex{}

					errs := make([][]error, 0)
					go func() {
						select {
						case err := <-got:
							mux.Lock()
							errs = append(errs, err)
							mux.Unlock()
						}
					}()
					time.Sleep(time.Second)

					mux.Lock()
					defer mux.Unlock()
					if len(errs) != 1 || len(errs[0]) != 1 || errs[0][0] != context.Canceled {
						return errors.Errorf("Invalid err, got: %v", errs)
					}
					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			return test{
				name: "Server return fail",
				fields: fields{
					athenz: &service.AuthorizedMock{
						StartProviderdFunc: func(context.Context) <-chan error {
							return make(chan error)
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							go func() {
								ech <- []error{errors.New("dummy")}
							}()
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				checkFunc: func(got <-chan []error) error {
					got1 := <-got
					cancel()
					time.Sleep(time.Second)
					//got2 := <-got
					if got1 == nil || len(got1) != 1 {
						return errors.Errorf("errors is invalid, got: %v", got1)
					}
					if got1[0].Error() != "dummy" {
						return errors.Errorf("got error: %v, want: %v", got1[0], context.Canceled)
					}
					//	if got2 == nil || len(got2) != 1 {
					//		return errors.Errorf("errors 2 is invalid, got: %v", got2)
					//	}
					//	if got2[0] != context.Canceled {
					//		return errors.Errorf("got 2 error: %v, want: %v", got2[0], "dummy")
					//	}
					return nil
				},
			}
		}(),
		func() test {
			ctx, cancel := context.WithCancel(context.Background())
			return test{
				name: "Providerd return fail",
				fields: fields{
					athenz: &service.AuthorizedMock{
						StartProviderdFunc: func(context.Context) <-chan error {
							ech := make(chan error)
							go func() {
								ech <- errors.New("dummy")
							}()
							return ech
						},
					},
					server: &service.ServerMock{
						ListenAndServeFunc: func(ctx context.Context) <-chan []error {
							ech := make(chan []error)
							return ech
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				checkFunc: func(got <-chan []error) error {
					time.Sleep(time.Millisecond * 200)
					cancel()
					errs := <-got
					if errs == nil || len(errs) != 2 {
						return errors.Errorf("errors is invalid, got: %v", errs)
					}
					if errs[0].Error() != "1 times appeared: dummy" {
						return errors.Errorf("got error: %v, want: %v", errs[0], "1 times appeared: dummy")
					}
					if errs[1] != context.Canceled {
						return errors.Errorf("got error: %v, want: %v", errs[1], context.Canceled)
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
			if err := tt.checkFunc(got); err != nil {
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
						AthenzDomains:             []string{"dummyDom1", "dummyDom2"},
						PolicyExpireMargin:        "10s",
						PolicyRefreshDuration:     "10s",
						PolicyEtagExpTime:         "10s",
						PolicyEtagFlushDur:        "10s",
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
