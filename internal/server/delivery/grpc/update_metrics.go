package grpc

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Nekrasov-Sergey/metrics-collector/internal/proto"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/network"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/utils"
)

func (s *MetricsServer) UpdateMetrics(ctx context.Context, in *pb.UpdateMetricsRequest) (*pb.UpdateMetricsResponse, error) {
	protoMetrics := in.GetMetrics()
	metrics := make([]types.Metric, 0, len(protoMetrics))

	for _, protoMetric := range protoMetrics {
		metric := types.Metric{
			Name:  types.MetricName(protoMetric.GetId()),
			MType: types.MetricType(strings.ToLower(protoMetric.GetType().String())),
		}

		if metric.Name == "" {
			return nil, status.Error(codes.NotFound, "отсутствует имя метрики")
		}

		switch metric.MType {
		case types.Gauge:
			if !protoMetric.HasValue() {
				return nil, status.Error(codes.InvalidArgument, "значение метрики Gauge не задано")
			}
			metric.Value = utils.Ptr(protoMetric.GetValue())
		case types.Counter:
			if !protoMetric.HasDelta() {
				return nil, status.Error(codes.InvalidArgument, "значение метрики Counter не задано")
			}
			metric.Delta = utils.Ptr(protoMetric.GetDelta())
		default:
			return nil, status.Errorf(codes.InvalidArgument, "некорректный тип метрики: %s", protoMetric.GetType().String())
		}

		metrics = append(metrics, metric)
	}

	if err := s.service.UpdateMetrics(ctx, metrics); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	metricNames := make([]types.MetricName, 0, len(metrics))
	for _, metric := range metrics {
		metricNames = append(metricNames, metric.Name)
	}

	event := &types.AuditEvent{
		TS:        time.Now().Unix(),
		Metrics:   metricNames,
		IPAddress: network.GetRemoteIP(ctx),
	}

	s.audit.Info(ctx, event)

	return &pb.UpdateMetricsResponse{}, nil
}
