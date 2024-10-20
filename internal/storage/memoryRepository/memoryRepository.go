package memoryRepository

import (
	"errors"
	"fmt"

	"github.com/dmitrovia/collector-metrics/internal/models"
)

type MemoryRepository struct {
	gauges   map[string]models.Gauge
	counters map[string]models.Counter
}

func (m *MemoryRepository) Init() {
	m.gauges = make(map[string]models.Gauge)
	m.counters = make(map[string]models.Counter)
}

func (m *MemoryRepository) GetStringValueGaugeMetric(name string) (string, error) {

	val, ok := m.gauges[name]
	if ok {
		return fmt.Sprintf("%.02f", val.Value), nil
	} else {
		return "", errors.New("metric not found")
	}

}

func (m *MemoryRepository) GetStringValueCounterMetric(name string) (string, error) {
	val, ok := m.counters[name]
	if ok {
		return fmt.Sprintf("%v", val.Value), nil
	} else {
		return "", errors.New("metric not found")
	}
}

func (m *MemoryRepository) GetMapStringsAllMetrics() *map[string]string {
	mapMetrics := make(map[string]string)

	for key, value := range m.counters {
		mapMetrics[key] = fmt.Sprintf("%v", value.Value)
	}

	for key, value := range m.gauges {
		mapMetrics[key] = fmt.Sprintf("%.02f", value.Value)
	}

	return &mapMetrics
}

func (m *MemoryRepository) AddGauge(gauge *models.Gauge) {
	m.gauges[gauge.Name] = *gauge
}

func (m *MemoryRepository) AddCounter(counter *models.Counter) {

	val, ok := m.counters[counter.Name]
	if ok {
		var temp *models.Counter = new(models.Counter)
		temp.Name = val.Name
		temp.Value = val.Value + counter.Value
		m.counters[counter.Name] = *temp
	} else {
		m.counters[counter.Name] = *counter
	}
}
