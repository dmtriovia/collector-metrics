package models

type gauge struct {
	name  string
	value float64
}

type counter struct {
	name  string
	value int64
}

type MemStorage struct {
	gauges   []gauge
	counters map[string]counter
}
