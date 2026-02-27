package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CheckIPInterceptor(trustedNet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if trustedNet == nil {
			return handler(ctx, req)
		}

		var ipStr string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				ipStr = values[0]
			}
		}

		if ipStr == "" {
			return nil, status.Error(codes.PermissionDenied, "отсутствует заголовок X-Real-IP")
		}

		ip := net.ParseIP(ipStr)
		if ip == nil {
			return nil, status.Error(codes.PermissionDenied, "некорректный IP-адрес")
		}

		if !trustedNet.Contains(ip) {
			return nil, status.Error(codes.PermissionDenied, "IP не входит в CIDR")
		}

		return handler(ctx, req)
	}
}
