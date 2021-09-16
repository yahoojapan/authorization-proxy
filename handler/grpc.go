package handler

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/kpango/glg"
	"github.com/mwitkow/grpc-proxy/proxy"
	"golang.org/x/sync/singleflight"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

const (
	gRPC = "grpc"
)

type GRPCHandler struct {
	proxyCfg       config.Proxy
	roleCfg        config.RoleToken
	authorizationd service.Authorizationd
	tlsCfg         *tls.Config
	connMap        sync.Map
	group          singleflight.Group
}

func NewGRPC(opts ...GRPCOption) (grpc.StreamHandler, io.Closer) {
	gh := new(GRPCHandler)
	for _, opt := range append(defaultGRPCOptions, opts...) {
		opt(gh)
	}

	if !strings.EqualFold(gh.proxyCfg.Scheme, gRPC) {
		return nil, nil
	}

	dialOpts := []grpc.DialOption{
		grpc.WithCodec(proxy.Codec()),
		grpc.WithInsecure(),
	}

	target := net.JoinHostPort(gh.proxyCfg.Host, strconv.Itoa(int(gh.proxyCfg.Port)))

	return proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil, status.Errorf(codes.Unauthenticated, ErrGRPCMetadataNotFound)
		}

		rts := md.Get(gh.roleCfg.RoleAuthHeader)
		if len(rts) == 0 {
			return ctx, nil, status.Errorf(codes.Unauthenticated, ErrRoleTokenNotFound)
		}

		p, err := gh.authorizationd.AuthorizeRoleToken(ctx, rts[0], gRPC, fullMethodName)
		if err != nil {
			return ctx, nil, status.Errorf(codes.Unauthenticated, err.Error())
		}

		ctx = metadata.AppendToOutgoingContext(ctx,
			"X-Athenz-Principal", p.Name(),
			"X-Athenz-Role", strings.Join(p.Roles(), ","),
			"X-Athenz-Domain", p.Domain(),
			"X-Athenz-Issued-At", strconv.FormatInt(p.IssueTime(), 10),
			"X-Athenz-Expires-At", strconv.FormatInt(p.ExpiryTime(), 10))

		if c, ok := p.(authorizerd.OAuthAccessToken); ok {
			ctx = metadata.AppendToOutgoingContext(ctx, "X-Athenz-Client-ID", c.ClientID())
		}

		conn, err := gh.dialContext(ctx, target, dialOpts...)
		return ctx, conn, err
	}), gh
}

func (gh *GRPCHandler) Close() error {
	gh.connMap.Range(func(target, v interface{}) bool {
		if conn, ok := v.(*grpc.ClientConn); ok {
			if err := conn.Close(); err != nil {
				glg.Warnf("failed to close connection. target: %s, err: %v", target, err)
			}
			gh.connMap.Delete(target)
		}
		return true
	})
	return nil
}

func (gh *GRPCHandler) dialContext(ctx context.Context, target string, dialOpts ...grpc.DialOption) (conn *grpc.ClientConn, err error) {
	if v, ok := gh.connMap.Load(target); ok {
		if conn, ok = v.(*grpc.ClientConn); ok && isHealthy(conn) {
			return conn, nil
		}
	}

	v, err, _ := gh.group.Do(target, func() (interface{}, error) {
		conn, err := grpc.DialContext(ctx, target, dialOpts...)
		if err != nil {
			return nil, err
		}
		gh.connMap.Store(target, conn)
		return conn, nil
	})
	if err == nil {
		if conn, ok := v.(*grpc.ClientConn); ok {
			return conn, nil
		}
	}
	return grpc.DialContext(ctx, target, dialOpts...)
}

func isHealthy(conn *grpc.ClientConn) bool {
	switch conn.GetState() {
	case connectivity.Ready, connectivity.Idle, connectivity.Connecting:
		return true
	default:
		return false
	}
}
