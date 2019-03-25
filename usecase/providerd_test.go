package usecase

import (
	"context"
	"reflect"
	"testing"

	"github.com/yahoojapan/authorization-proxy/config"
	"github.com/yahoojapan/authorization-proxy/service"
)

func TestNew(t *testing.T) {
	type args struct {
		cfg config.Config
	}
	tests := []struct {
		name    string
		args    args
		want    AuthorizationDaemon
		wantErr bool
	}{
		//		{
		//			name: "new success",
		//			args: args{
		//				cfg: config.Config{},
		//			},
		//		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
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
		checkFunc func(chan []error) error
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
						ListenAndServeFunc: func(context.Context) chan []error {
							return make(chan []error)
						},
					},
				},
				args: args{
					ctx: ctx,
				},
				checkFunc: func(chan []error) error {
					cancel()

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
						AthenzConfRefreshDuration: "10s",
						AthenzConfSysAuthDomain:   "dummy.sys.auth",
						AthenzConfEtagExpTime:     "10s",
						AthenzConfEtagFlushDur:    "10s",
						AthenzDomains:             []string{"dummyDom1", "dummyDom2"},
						PolicyExpireMargin:        "10s",
						PolicyRefreshDuration:     "10s",
						PolicyEtagExpTime:         "10s",
						PolicyEtagFlushDur:        "10s",
					},
				},
			},
			checkFunc: func(got service.Authorizationd) error {
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
			if cerr := tt.checkFunc(got); cerr != nil {
				t.Errorf("newAuthorizationd() error = %v", cerr)
				return
			}
		})
	}
}
