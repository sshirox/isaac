package usecase

type Repository interface {
	Upsert(string, string, string) error
	Get(string, string) (string, error)
	GetAllGauges() map[string]string
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

func (uc *UseCase) GetAllMetrics() map[string]string {
	return uc.repo.GetAllGauges()
}
