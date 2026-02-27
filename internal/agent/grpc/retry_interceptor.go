package grpc

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const maxAttempts = 4

func NewRetryInterceptor(logger zerolog.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		var lastErr error

		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err == nil {
				return nil
			}
			logger.Error().Err(err).Send()

			lastErr = err

			if errors.Is(err, context.Canceled) {
				return err
			}
			if errors.Is(err, context.DeadlineExceeded) {
				return err
			}

			st, ok := status.FromError(err)
			if !ok {
				return err
			}

			logger.Info().Msg(st.Code().String())
			switch st.Code() {
			case codes.Unavailable,
				codes.DeadlineExceeded,
				codes.ResourceExhausted:
			default:
				return err
			}

			if attempt == maxAttempts {
				break
			}

			delay := time.Second * time.Duration(1<<(attempt-1)) // экспоненциальная задержка

			logger.Error().
				Err(err).
				Str("method", method).
				Msgf("gRPC retry %d/%d через %s", attempt+1, maxAttempts, delay)

			timer := time.NewTimer(delay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return ctx.Err()
			case <-timer.C:
			}
		}

		return lastErr
	}
}
