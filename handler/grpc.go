package handler

import (
	"context"
	"net"
	"strconv"
	"strings"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	authorizerd "github.com/yahoojapan/athenz-authorizer/v5"
	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func NewGRPC(cfg config.Proxy, roleCfg config.RoleToken, prov service.Authorizationd) grpc.StreamHandler {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	return proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, nil, status.Errorf(codes.Unauthenticated, "gRPC metadata not found")
		}

		rts := md.Get(roleCfg.RoleAuthHeader)
		if len(rts) == 0 {
			return ctx, nil, status.Errorf(codes.Unauthenticated, "roletoken not found")
		}

		p, err := prov.AuthorizeRoleToken(ctx, rts[0], "grpc", fullMethodName) // TODO: action use gRPC?
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

		conn, err := grpc.DialContext(ctx, target, grpc.WithCodec(proxy.Codec()))
		return ctx, conn, err
	})
}

// func NewGRPC(cfg config.Proxy, prov service.Authorizationd) *grpc.Server {
// 	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
// 	return grpc.NewServer(
// 		grpc.CustomCodec(proxy.Codec()),
// 		grpc.UnknownServiceHandler(proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
// 			conn, err := grpc.DialContext(ctx, target, grpc.WithCodec(proxy.Codec()))
// 			return ctx, conn, err
// 			// md, ok := metadata.FromIncomingContext(ctx)
// 			// if ok {
// 			// 	rs := md.Get("")
// 			// 	if rs != nil && len(rs) > 0 {
// 			// 		conn, err := grpc.DialContext(ctx, target, grpc.WithCodec(proxy.Codec()))
// 			// 		return ctx, conn, err
// 			// 	}
// 			//
// 			// }
// 			// return ctx, nil, grpc.Errorf(codes.Unimplemented, ErrGRPCMetadataNotFound)
// 		})))
// }
