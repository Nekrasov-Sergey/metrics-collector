package grpc

import (
	"context"
	"net"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"

	pb "github.com/Nekrasov-Sergey/metrics-collector/internal/proto"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/network"
)

type Service interface {
	UpdateMetrics(ctx context.Context, metrics []types.Metric) error
}

type Audit interface {
	Info(ctx context.Context, event *types.AuditEvent)
}

type options struct {
	gRPCAddress   string
	trustedSubnet string
}

type Option func(*options)

func WithGRPCAddress(gRPCAddress string) Option {
	return func(o *options) {
		o.gRPCAddress = gRPCAddress
	}
}

func WithTrustedSubnet(trustedSubnet string) Option {
	return func(o *options) {
		o.trustedSubnet = trustedSubnet
	}
}

type MetricsServer struct {
	pb.UnimplementedMetricsServer
	address string
	server  *grpc.Server
	service Service
	audit   Audit
	logger  zerolog.Logger
}

func New(service Service, audit Audit, logger zerolog.Logger, opts ...Option) (*MetricsServer, error) {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	trustedNet, err := network.ParseCIDR(o.trustedSubnet)
	if err != nil {
		return nil, err
	}

	return &MetricsServer{
		address: o.gRPCAddress,
		server:  grpc.NewServer(grpc.ChainUnaryInterceptor(LoggerInterceptor(logger), CheckIPInterceptor(trustedNet))),
		service: service,
		audit:   audit,
		logger:  logger,
	}, nil
}

func (s *MetricsServer) Run() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return errors.Wrap(err, "не удалось открыть tcp-сокет для gRPC-сервера")
	}

	pb.RegisterMetricsServer(s.server, s)

	s.logger.Info().Msgf("gRPC-сервер запущен на %s", s.address)

	if err := s.server.Serve(listener); err != nil {
		return errors.Wrap(err, "gRPC-сервер завершился с ошибкой")
	}

	return nil
}

func (s *MetricsServer) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		s.logger.Info().Msg("Запущен graceful shutdown gRPC-сервера")
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	case <-done:
		s.logger.Info().Msg("gRPC-сервер корректно остановлен")
		return nil
	}
}
