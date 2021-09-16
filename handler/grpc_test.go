package handler

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"reflect"
	"sync"
	"testing"

	"github.com/mwitkow/grpc-proxy/proxy"
	"github.com/pkg/errors"
	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestNewGRPC(t *testing.T) {
	type args struct {
		opts []GRPCOption
	}
	type test struct {
		name       string
		args       args
		beforeFunc func()
		checkFunc  func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error
		afterFunc  func()
		want       grpc.StreamHandler
		want1      io.Closer
	}
	checkGRPCSrvRunning := func(addr, roleTok string) error {
		conn, err := grpc.Dial(addr, grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := context.Background()
		if roleTok != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "role-header", roleTok)
		}

		return conn.Invoke(ctx, "/method/", new(emptypb.Empty), new(emptypb.Empty))
	}
	defaultCheckFunc := func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error {
		if !reflect.DeepEqual(got, want) {
			return errors.Errorf("NewGRPC() got = %v, want %v", got, want)
		}
		if !reflect.DeepEqual(got1, want1) {
			return errors.Errorf("NewGRPC() got = %v, want %v", got1, want1)
		}
		return nil
	}

	tests := []test{
		{
			name: "return nil when scheme is not gRPC",
			args: args{
				opts: []GRPCOption{
					WithProxyConfig(config.Proxy{
						Scheme: "http",
					}),
				},
			},
			want:  nil,
			want1: nil,
		},
		func() test {
			targetExecuted := false

			grpcHandler := func(srv interface{}, stream grpc.ServerStream) error {
				targetExecuted = true
				return stream.SendMsg(new(emptypb.Empty))
			}
			grpcSrv := grpc.NewServer(
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler),
			)
			l, err := net.Listen("tcp", ":9997")
			if err != nil {
				t.Fatal(err)
			}

			var proxySrv *grpc.Server
			pl, err := net.Listen("tcp", ":9998")
			if err != nil {
				t.Fatal(err)
			}

			return test{
				name: "return handler when scheme is gRPC",
				args: args{
					opts: []GRPCOption{
						WithProxyConfig(config.Proxy{
							Scheme: "grpc",
							Host:   "127.0.0.1",
							Port:   9997,
						}),
						WithRoleTokenConfig(config.RoleToken{
							Enable:         true,
							RoleAuthHeader: "role-header",
						}),
						WithAuthorizationd(&service.AuthorizerdMock{
							VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) (authorizerd.Principal, error) {
								return &PrincipalMock{
									NameFunc: func() string {
										return "name"
									},
									RolesFunc: func() []string {
										return []string{"role"}
									},
									DomainFunc: func() string {
										return "domain"
									},
									IssueTimeFunc: func() int64 {
										return 15991044
									},
									ExpiryTimeFunc: func() int64 {
										return 16000000
									},
								}, nil
							},
						}),
					},
				},
				beforeFunc: func() {
					go func() {
						if err := grpcSrv.Serve(l); err != nil {
							t.Fatal(err)
						}
					}()
				},
				checkFunc: func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error {
					proxySrv = grpc.NewServer(
						grpc.CustomCodec(proxy.Codec()),
						grpc.UnknownServiceHandler(got),
					)

					// start proxy server
					go func() {
						if err := proxySrv.Serve(pl); err != nil {
							t.Fatal(err)
						}
					}()

					if err := checkGRPCSrvRunning("127.0.0.1:9998", "roletok"); err != nil {
						return err
					}
					if !targetExecuted {
						return errors.New("target server is no executed")
					}

					return nil
				},
				afterFunc: func() {
					grpcSrv.Stop()
					l.Close()
					proxySrv.Stop()
					pl.Close()
				},
			}
		}(),
		func() test {
			targetExecuted := false

			grpcHandler := func(srv interface{}, stream grpc.ServerStream) error {
				targetExecuted = true
				return stream.SendMsg(new(emptypb.Empty))
			}
			grpcSrv := grpc.NewServer(
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler),
			)
			l, err := net.Listen("tcp", ":9996")
			if err != nil {
				t.Fatal(err)
			}

			var proxySrv *grpc.Server
			pl, err := net.Listen("tcp", ":9999")
			if err != nil {
				t.Fatal(err)
			}

			return test{
				name: "return error when scheme is gRPC but role token not appended",
				args: args{
					opts: []GRPCOption{
						WithProxyConfig(config.Proxy{
							Scheme: "grpc",
							Host:   "127.0.0.1",
							Port:   9996,
						}),
						WithRoleTokenConfig(config.RoleToken{
							Enable:         true,
							RoleAuthHeader: "role-header",
						}),
					},
				},
				beforeFunc: func() {
					go func() {
						if err := grpcSrv.Serve(l); err != nil {
							t.Fatal(err)
						}
					}()
				},
				checkFunc: func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error {
					proxySrv = grpc.NewServer(
						grpc.CustomCodec(proxy.Codec()),
						grpc.UnknownServiceHandler(got),
					)

					// start proxy server
					go func() {
						if err := proxySrv.Serve(pl); err != nil {
							t.Fatal(err)
						}
					}()

					if err := checkGRPCSrvRunning("127.0.0.1:9999", ""); !errors.Is(err, status.Errorf(codes.Unauthenticated, ErrRoleTokenNotFound)) {
						return errors.Errorf("unexpected err, got: %s", err)
					}
					if targetExecuted {
						return errors.New("target server is executed")
					}

					return nil
				},
				afterFunc: func() {
					grpcSrv.Stop()
					l.Close()
					proxySrv.Stop()
					pl.Close()
				},
			}
		}(),
		func() test {
			targetExecuted := false

			grpcHandler := func(srv interface{}, stream grpc.ServerStream) error {
				targetExecuted = true
				return stream.SendMsg(new(emptypb.Empty))
			}
			grpcSrv := grpc.NewServer(
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler),
			)
			l, err := net.Listen("tcp", ":9995")
			if err != nil {
				t.Fatal(err)
			}

			var proxySrv *grpc.Server
			pl, err := net.Listen("tcp", ":9994")
			if err != nil {
				t.Fatal(err)
			}

			return test{
				name: "return error when role token is invalid",
				args: args{
					opts: []GRPCOption{
						WithProxyConfig(config.Proxy{
							Scheme: "grpc",
							Host:   "127.0.0.1",
							Port:   9995,
						}),
						WithRoleTokenConfig(config.RoleToken{
							Enable:         true,
							RoleAuthHeader: "role-header",
						}),
						WithAuthorizationd(&service.AuthorizerdMock{
							VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) (authorizerd.Principal, error) {
								return nil, errors.New("unauthenticated")
							},
						}),
					},
				},
				beforeFunc: func() {
					go func() {
						if err := grpcSrv.Serve(l); err != nil {
							t.Fatal(err)
						}
					}()
				},
				checkFunc: func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error {
					proxySrv = grpc.NewServer(
						grpc.CustomCodec(proxy.Codec()),
						grpc.UnknownServiceHandler(got),
					)

					// start proxy server
					go func() {
						if err := proxySrv.Serve(pl); err != nil {
							t.Fatal(err)
						}
					}()

					if err := checkGRPCSrvRunning("127.0.0.1:9994", "roletok"); errors.Is(err, errors.New("rpc error: code = Unauthenticated desc = unauthenticated")) {
						return errors.Errorf("unexpected err, got: %s", err)
					}
					if targetExecuted {
						return errors.New("target server is executed")
					}

					return nil
				},
				afterFunc: func() {
					grpcSrv.Stop()
					l.Close()
					proxySrv.Stop()
					pl.Close()
				},
			}
		}(),
		func() test {
			targetExecuted := false

			grpcHandler := func(srv interface{}, stream grpc.ServerStream) error {
				targetExecuted = true
				return stream.SendMsg(new(emptypb.Empty))
			}
			grpcSrv := grpc.NewServer(
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(grpcHandler),
			)
			l, err := net.Listen("tcp", ":9993")
			if err != nil {
				t.Fatal(err)
			}

			var proxySrv *grpc.Server
			pl, err := net.Listen("tcp", ":9992")
			if err != nil {
				t.Fatal(err)
			}

			return test{
				name: "return handler when scheme is gRPC and authorized token is OAuthAccessToken",
				args: args{
					opts: []GRPCOption{
						WithProxyConfig(config.Proxy{
							Scheme: "grpc",
							Host:   "127.0.0.1",
							Port:   9993,
						}),
						WithRoleTokenConfig(config.RoleToken{
							Enable:         true,
							RoleAuthHeader: "role-header",
						}),
						WithAuthorizationd(&service.AuthorizerdMock{
							VerifyRoleTokenFunc: func(ctx context.Context, tok, act, res string) (authorizerd.Principal, error) {
								return &OAuthAccessTokenMock{
									PrincipalMock: PrincipalMock{
										NameFunc: func() string {
											return "name"
										},
										RolesFunc: func() []string {
											return []string{"role"}
										},
										DomainFunc: func() string {
											return "domain"
										},
										IssueTimeFunc: func() int64 {
											return 15991044
										},
										ExpiryTimeFunc: func() int64 {
											return 16000000
										},
									},
									ClientIDFunc: func() string {
										return "clientID"
									},
								}, nil
							},
						}),
					},
				},
				beforeFunc: func() {
					go func() {
						if err := grpcSrv.Serve(l); err != nil {
							t.Fatal(err)
						}
					}()
				},
				checkFunc: func(got grpc.StreamHandler, got1 io.Closer, want grpc.StreamHandler, want1 io.Closer) error {
					proxySrv = grpc.NewServer(
						grpc.CustomCodec(proxy.Codec()),
						grpc.UnknownServiceHandler(got),
					)

					// start proxy server
					go func() {
						if err := proxySrv.Serve(pl); err != nil {
							t.Fatal(err)
						}
					}()

					if err := checkGRPCSrvRunning("127.0.0.1:9992", "roletok"); err != nil {
						return err
					}
					if !targetExecuted {
						return errors.New("target server is no executed")
					}

					return nil
				},
				afterFunc: func() {
					grpcSrv.Stop()
					l.Close()
					proxySrv.Stop()
					pl.Close()
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.beforeFunc != nil {
				tt.beforeFunc()
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}
			checkFunc := defaultCheckFunc
			if tt.checkFunc != nil {
				checkFunc = tt.checkFunc
			}
			got, got1 := NewGRPC(tt.args.opts...)

			if err := checkFunc(got, got1, tt.want, tt.want1); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestGRPCHandler_Close(t *testing.T) {
	type fields struct {
		proxyCfg       config.Proxy
		roleCfg        config.RoleToken
		authorizationd service.Authorizationd
		tlsCfg         *tls.Config
		connMap        sync.Map
		group          singleflight.Group
	}
	type test struct {
		name    string
		fields  fields
		wantErr bool
	}
	tests := []test{
		{
			name: "close success when map is empty",
			fields: fields{
				connMap: sync.Map{},
			},
			wantErr: false,
		},
		func() test {
			connMap := sync.Map{}
			conn, err := grpc.Dial("127.0.0.1", grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}
			conn2, err := grpc.Dial("127.0.0.2", grpc.WithCodec(proxy.Codec()), grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}
			connMap.Store("127.0.0.1", conn)
			connMap.Store("127.0.0.2", conn2)

			return test{
				name: "close all connection in connMap",
				fields: fields{
					connMap: connMap,
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gh := &GRPCHandler{
				proxyCfg:       tt.fields.proxyCfg,
				roleCfg:        tt.fields.roleCfg,
				authorizationd: tt.fields.authorizationd,
				tlsCfg:         tt.fields.tlsCfg,
				connMap:        tt.fields.connMap,
				group:          tt.fields.group,
			}
			if err := gh.Close(); (err != nil) != tt.wantErr {
				t.Errorf("GRPCHandler.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGRPCHandler_dialContext(t *testing.T) {
	type fields struct {
		proxyCfg       config.Proxy
		roleCfg        config.RoleToken
		authorizationd service.Authorizationd
		tlsCfg         *tls.Config
		connMap        sync.Map
		group          singleflight.Group
	}
	type args struct {
		ctx      context.Context
		target   string
		dialOpts []grpc.DialOption
	}
	type test struct {
		name      string
		fields    fields
		args      args
		wantConn  *grpc.ClientConn
		wantErr   error
		checkFunc func(gotConn, wantConn *grpc.ClientConn, gotErr, wantErr error) error
		afterFunc func()
	}
	defaultCheckFunc := func(gotConn, wantConn *grpc.ClientConn, gotErr, wantErr error) error {
		if !errors.Is(gotErr, wantErr) {
			return errors.Errorf("GRPCHandler.dialContext() error = %v, wantErr %v", gotErr, wantErr)
		}
		if !reflect.DeepEqual(gotConn, wantConn) {
			return errors.Errorf("GRPCHandler.dialContext() = %v, want %v", gotConn, wantConn)
		}
		return nil
	}
	tests := []test{
		func() test {
			target := "127.0.0.1:80"
			conn, err := grpc.Dial(target, grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}

			connMap := sync.Map{}
			connMap.Store(target, conn)

			return test{
				name: "return cached healthy connection",
				fields: fields{
					connMap: connMap,
				},
				args: args{
					ctx:    context.Background(),
					target: target,
					dialOpts: []grpc.DialOption{
						grpc.WithInsecure(),
					},
				},
				checkFunc: func(gotConn, wantConn *grpc.ClientConn, gotErr, wantErr error) error {
					if !errors.Is(gotErr, wantErr) {
						return errors.Errorf("GRPCHandler.dialContext() error = %v, wantErr %v", gotErr, wantErr)
					}

					if gotConn.Target() != target {
						return errors.Errorf("invalid target, got: %s", gotConn.Target())
					}
					if s := gotConn.GetState(); s != connectivity.Idle && s != connectivity.Ready && s != connectivity.Connecting {
						return errors.Errorf("connection not ready, state: %s", gotConn.GetState())
					}
					return nil
				},
				afterFunc: func() {
					conn.Close()
				},
			}
		}(),
		func() test {
			target := "127.0.0.1:83"
			conn, err := grpc.Dial(target, grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}
			conn.Close()

			connMap := sync.Map{}
			connMap.Store(target, conn)

			return test{
				name: "return new connection when the cached connection is not healthy",
				fields: fields{
					connMap: connMap,
				},
				args: args{
					ctx:    context.Background(),
					target: target,
					dialOpts: []grpc.DialOption{
						grpc.WithInsecure(),
					},
				},
				checkFunc: func(gotConn, wantConn *grpc.ClientConn, gotErr, wantErr error) error {
					if !errors.Is(gotErr, wantErr) {
						return errors.Errorf("GRPCHandler.dialContext() error = %v, wantErr %v", gotErr, wantErr)
					}

					if gotConn.Target() != target {
						return errors.Errorf("invalid target, got: %s", gotConn.Target())
					}
					if s := gotConn.GetState(); s != connectivity.Idle && s != connectivity.Ready && s != connectivity.Connecting {
						return errors.Errorf("connection not ready, state: %s", gotConn.GetState())
					}
					return nil
				},
				afterFunc: func() {
					conn.Close()
				},
			}
		}(),
		func() test {
			target := "127.0.0.1:85"
			connMap := sync.Map{}

			return test{
				name: "return new connection with empty cache",
				fields: fields{
					connMap: connMap,
				},
				args: args{
					ctx:    context.Background(),
					target: target,
					dialOpts: []grpc.DialOption{
						grpc.WithInsecure(),
					},
				},
				checkFunc: func(gotConn, wantConn *grpc.ClientConn, gotErr, wantErr error) error {
					if !errors.Is(gotErr, wantErr) {
						return errors.Errorf("GRPCHandler.dialContext() error = %v, wantErr %v", gotErr, wantErr)
					}

					if gotConn.Target() != target {
						return errors.Errorf("invalid target, got: %s", gotConn.Target())
					}
					if s := gotConn.GetState(); s != connectivity.Idle && s != connectivity.Ready && s != connectivity.Connecting {
						return errors.Errorf("connection not ready, state: %s", gotConn.GetState())
					}
					return nil
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checkFunc := defaultCheckFunc
			if tt.checkFunc != nil {
				checkFunc = tt.checkFunc
			}
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			gh := &GRPCHandler{
				proxyCfg:       tt.fields.proxyCfg,
				roleCfg:        tt.fields.roleCfg,
				authorizationd: tt.fields.authorizationd,
				tlsCfg:         tt.fields.tlsCfg,
				connMap:        tt.fields.connMap,
				group:          tt.fields.group,
			}
			gotConn, err := gh.dialContext(tt.args.ctx, tt.args.target, tt.args.dialOpts...)
			if err := checkFunc(gotConn, tt.wantConn, err, tt.wantErr); err != nil {
				t.Error(err)
			}
		})
	}
}

func Test_isHealthy(t *testing.T) {
	type args struct {
		conn *grpc.ClientConn
	}
	type test struct {
		name      string
		args      args
		want      bool
		afterFunc func()
	}
	tests := []test{
		func() test {
			conn, err := grpc.Dial("127.0.0.1:92", grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}

			return test{
				name: "return true when connection state is ready",
				args: args{
					conn: conn,
				},
				want: true,
				afterFunc: func() {
					conn.Close()
				},
			}
		}(),
		func() test {
			conn, err := grpc.Dial("127.0.0.1:93", grpc.WithInsecure())
			if err != nil {
				t.Error(err)
			}
			conn.Close()

			return test{
				name: "return false when connection state is closed",
				args: args{
					conn: conn,
				},
				want: false,
				afterFunc: func() {
					conn.Close()
				},
			}
		}(),
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.afterFunc != nil {
				defer tt.afterFunc()
			}

			if got := isHealthy(tt.args.conn); got != tt.want {
				t.Errorf("isHealthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
