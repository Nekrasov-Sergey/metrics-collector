package postgres

import (
	"context"
	"database/sql"

	"github.com/pkg/errors"

	"github.com/Nekrasov-Sergey/metrics-collector/internal/types"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/dbutils"
	"github.com/Nekrasov-Sergey/metrics-collector/pkg/errcodes"
)

func (p *Postgres) UpdateMetric(ctx context.Context, metric types.Metric) error {
	const q = `insert into metrics (name, type, delta, value)
values (:name, :type, :delta, :value)
on conflict (name) do update
    set delta = case
                    when excluded.type = 'counter' then metrics.delta + excluded.delta
                    else excluded.delta
        end,
        value = excluded.value
`

	args := map[string]any{
		"name":  metric.Name,
		"type":  metric.MType,
		"delta": metric.Delta,
		"value": metric.Value,
	}
	if err := dbutils.NamedExec(ctx, p.db, q, args); err != nil {
		return errors.Wrapf(err, "не удалось обновить метрику %q", metric.Name)
	}

	return nil
}

func (p *Postgres) GetMetric(ctx context.Context, rowMetric types.Metric) (metric types.Metric, err error) {
	const q = `select name, type, delta, value
from metrics
where name = :name`

	args := map[string]any{
		"name": rowMetric.Name,
	}
	if err = dbutils.NamedGet(ctx, p.db, &metric, q, args); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return metric, errcodes.ErrMetricNotFound
		}
		return metric, errors.Wrapf(err, "не удалось получить метрику %q", metric.Name)
	}

	return metric, nil
}

func (p *Postgres) GetMetrics(ctx context.Context) (metrics []types.Metric, err error) {
	const q = `select name, type, delta, value from metrics order by name`

	if err = dbutils.NamedSelect(ctx, p.db, &metrics, q, map[string]any{}); err != nil {
		return nil, errors.Wrap(err, "не удалось получить метрики")
	}

	return metrics, nil
}

func (p *Postgres) UpdateMetrics(ctx context.Context, metrics []types.Metric) error {
	const updateQuery = `insert into metrics (name, type, delta, value)
values (:name, :type, :delta, :value)
on conflict (name) do update
    set delta = case
                    when excluded.type = 'counter' then metrics.delta + excluded.delta
                    else excluded.delta
        end,
        value = excluded.value
        `

	if err := dbutils.NamedExec(ctx, p.db, updateQuery, metrics); err != nil {
		return errors.Wrap(err, "не удалось обновить метрики")
	}

	return nil
}
