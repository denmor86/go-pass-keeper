package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	headerAuthorize = "authorization"
)

type TokenStorage interface {
	GetToken(ctx context.Context) (string, error)
}

func AuthInterceptor(
	storage TokenStorage,
) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		token, err := storage.GetToken(ctx)
		if err != nil || len(token) == 0 {
			return invoker(ctx, method, req, reply, cc, opts...)
		}

		authorize := "bearer " + token
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headerAuthorize, authorize))

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AuthStreamInterceptor(
	storage TokenStorage,
) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		token, err := storage.GetToken(ctx)
		if err != nil || len(token) == 0 {
			return streamer(ctx, desc, cc, method, opts...)
		}

		authorize := "bearer " + token
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs(headerAuthorize, authorize))

		return streamer(ctx, desc, cc, method, opts...)
	}
}
