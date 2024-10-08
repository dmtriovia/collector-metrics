package service

import (
	"storage"
)

type Service interface {
	GetMetric(string) (string, error)
}

type MetricService struct {
	repo storage.Repository
}

func (s *MetricService) GetMetric(longURL string) (string, error) {

	//repo := &repository{}
	//temp_service := NewGetMetricService(repo)
	//fmt.Println(temp_service)

	//mockRepo := &MockRepository{}
	//service := NewShorterService(mockRepo)

	return "temp_metric", nil
}

func NewGetMetricService(repo storage.Repository) *MetricService {
	return &MetricService{repo: repo}
}

type repository struct{}

func (r *repository) Store(temp1, temp2 string) error {
	//_, err := r.db.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, longURL)
	return nil
}

func (r *repository) Get(temp string) (string, error) {
	//_, err := r.db.Exec("INSERT INTO urls (short_code, long_url) VALUES (?, ?)", shortCode, longURL)
	return "", nil
}
