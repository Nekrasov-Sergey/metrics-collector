package grpc

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Nekrasov-Sergey/metrics-collector/internal/proto"
)

type options struct {
	gRPCAddress string
	localIP     string
}

type Option func(*options)

func WithGRPCAddress(gRPCAddress string) Option {
	return func(o *options) {
		o.gRPCAddress = gRPCAddress
	}
}

func WithLocalIP(localIP string) Option {
	return func(o *options) {
		o.localIP = localIP
	}
}

type MetricsClient struct {
	Client pb.MetricsClient
	conn   *grpc.ClientConn
}

func New(logger zerolog.Logger, opts ...Option) (*MetricsClient, error) {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	conn, err := grpc.NewClient(
		o.gRPCAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(NewRetryInterceptor(logger), NewSetIPInterceptor(o.localIP)),
	)
	if err != nil {
		return nil, errors.Wrapf(err, "ошибка подключения к gRPC серверу по адресу %s", o.gRPCAddress)
	}

	return &MetricsClient{
		Client: pb.NewMetricsClient(conn),
		conn:   conn,
	}, nil
}

func (c *MetricsClient) Close() error {
	if err := c.conn.Close(); err != nil {
		return errors.Wrap(err, "ошибка закрытия gRPC соединения")
	}
	return nil
}
