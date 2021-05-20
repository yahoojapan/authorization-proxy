package handler

import (
	"context"

	"github.com/mwitkow/grpc-proxy/proxy"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/yahoojapan/authorization-proxy/v4/config"
	"github.com/yahoojapan/authorization-proxy/v4/service"
)

func NewGRPC(cfg config.Proxy, prov service.Authorizationd) *grpc.Server {
	return grpc.NewServer(
		grpc.CustomCodec(proxy.Codec()),
		grpc.UnknownServiceHandler(proxy.TransparentHandler(func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if ok {
				rs := md.Get("")
				if rs != nil && len(rs) > 0 {
					// Decide on which backend to dial
					if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
						// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.
						conn, err := grpc.DialContext(ctx, "api-service.staging.svc.local", grpc.WithCodec(proxy.Codec()))
						return ctx, conn, err
					} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
						conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithCodec(proxy.Codec()))
						return ctx, conn, err
					}
				}

			}
			return ctx, nil, grpc.Errorf(codes.Unimplemented, ErrGRPCMetadataNotFound)
		})))
}
