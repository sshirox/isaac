package usecase

import "github.com/sshirox/isaac/internal/metric"

type Repository interface {
	Upsert(string, string, string) error
	Get(string, string) (string, error)
	GetAllGauges() map[string]float64
	Update(string, string, *float64, *int64) metric.Metrics
	Value(string, string) (metric.Metrics, error)
}

type UseCase struct {
	repo Repository
}

func New(r Repository) *UseCase {
	return &UseCase{
		repo: r,
	}
}

func (uc *UseCase) GetMetric(metricType, name string) (string, error) {
	val, err := uc.repo.Get(metricType, name)
	if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (uc *UseCase) UpsertMetric(metricType, name, value string) error {
	err := uc.repo.Upsert(metricType, name, value)

	if err != nil {
		return err
	}

	return nil
}

func (uc *UseCase) ReceiveMetric(MType, id string) (metric.Metrics, error) {
	m, err := uc.repo.Value(MType, id)
	if err != nil {
		return m, err
	} else {
		return m, nil
	}
}

func (uc *UseCase) UpdateMetric(MType, id string, value *float64, delta *int64) metric.Metrics {
	m := uc.repo.Update(MType, id, value, delta)

	return m
}

func (uc *UseCase) GetAllMetrics() map[string]float64 {
	return uc.repo.GetAllGauges()
}
