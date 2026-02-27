package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func LoggerInterceptor(baseLogger zerolog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		l := baseLogger.With().
			Str("method", info.FullMethod).
			Str("req_id", uuid.NewString()[:8]).
			Logger()

		start := time.Now()

		resp, err = handler(ctx, req)

		st := status.Code(err)

		l = l.With().
			Str("grpc_code", st.String()).
			Str("duration", time.Since(start).String()).
			Logger()

		if err != nil {
			errLog := l.Error()

			if st == codes.Internal || st == codes.Unknown {
				errLog = errLog.Stack()
			}

			errLog.Err(err).Msg("Ошибка выполнения gRPC запроса")
			return resp, err
		}

		l.Info().Msg("gRPC запрос успешно выполнен")

		return resp, nil
	}
}
