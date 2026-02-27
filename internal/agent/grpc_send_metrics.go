package agent

import (
	"context"
	"errors"

	"google.golang.org/protobuf/proto"

	pb "github.com/Nekrasov-Sergey/metrics-collector/internal/proto"
	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
)

func (a *Agent) gRPCSendMetrics(ctx context.Context, metrics []types.Metric, w int) {
	grpcMetrics := make([]*pb.Metric, 0, len(metrics))
	for _, metric := range metrics {
		mType, err := toProtoMType(metric.MType)
		if err != nil {
			a.logger.Error().Err(err).Int("воркер", w).Msg("Не удалось преобразовать тип метрики")
			continue
		}

		grpcMetrics = append(grpcMetrics, pb.Metric_builder{
			Id:    proto.String(string(metric.Name)),
			Type:  mType.Enum(),
			Delta: metric.Delta,
			Value: metric.Value,
		}.Build())
	}

	_, err := a.grpcClient.UpdateMetrics(ctx, pb.UpdateMetricsRequest_builder{
		Metrics: grpcMetrics,
	}.Build())
	if err != nil {
		a.logger.Error().Err(err).Msg("Ошибка при отправке метрик по gRPC")
		return
	}

	a.logger.Info().Int("воркер", w).Msgf("Отправлены метрики на gRPC-сервер %s", a.config.GRPCAddr)
}

func toProtoMType(t types.MetricType) (pb.Metric_MType, error) {
	switch t {
	case types.Gauge:
		return pb.Metric_GAUGE, nil
	case types.Counter:
		return pb.Metric_COUNTER, nil
	default:
		return 0, errors.New("неизвестный тип метрики")
	}
}
