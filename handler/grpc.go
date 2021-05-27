package handler

import (
	"context"
	"net"
	"strconv"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func NewGRPC(cfg config.Proxy, prov service.Authorizationd) grpc.StreamHandler {
	target := net.JoinHostPort(cfg.Host, strconv.Itoa(int(cfg.Port)))
	return proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
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
