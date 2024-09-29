package usecase

import "github.com/sshirox/isaac/internal/model"

type Repository interface {
	Upsert(string, string, *float64, *int64) model.Metric
	Get(string, string) (model.Metric, error)
	GetAllGauges() map[string]float64
}

type UseCase struct {
	repo Repository
}

func New(r Repository) *UseCase {
	return &UseCase{
		repo: r,
	}
}

func (uc *UseCase) GetMetric(MType, id string) (model.Metric, error) {
	m, err := uc.repo.Get(MType, id)
	if err != nil {
		return m, err
	} else {
		return m, nil
	}
}

func (uc *UseCase) UpsertMetric(MType, id string, value *float64, delta *int64) model.Metric {
	m := uc.repo.Upsert(MType, id, value, delta)

	return m
}

func (uc *UseCase) GetAllMetrics() map[string]float64 {
	return uc.repo.GetAllGauges()
}
