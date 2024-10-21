package main

import (
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/endpoints"
	"github.com/dmitrovia/collector-metrics/internal/functions/random"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/models"
)

const defPollInterval = 2

const defReportInterval = 10

const metricGaugeCount = 27

type initParams struct {
	url                 string
	PORT                string
	reportInterval      int
	pollInterval        int
	validateAddrPattern string
}

func main() {
	var wg *sync.WaitGroup
	wg = new(sync.WaitGroup)

	var monitor *models.Monitor
	monitor = new(models.Monitor)

	var httpClient *http.Client
	httpClient = new(http.Client)

	var gauges *[]models.Gauge
	gauges = new([]models.Gauge)

	var counters *map[string]models.Counter = new(map[string]models.Counter)

	var params *initParams = new(initParams)
	params.url = "http://"
	params.reportInterval = 10
	params.pollInterval = 2
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err := initialization(params, httpClient, monitor)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	wg.Add(1)

	go collect(monitor, params, wg, gauges, counters)

	wg.Add(1)

	go send(params, wg, httpClient, gauges, counters)
	wg.Wait()
}

func collect(m *models.Monitor, p *initParams, wg *sync.WaitGroup, gs *[]models.Gauge, cs *map[string]models.Counter) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(p.pollInterval) * time.Second):
			setValuesMonitor(m, gs, cs)
		}
	}
}

func send(p *initParams, wg *sync.WaitGroup, httpC *http.Client, gs *[]models.Gauge, cs *map[string]models.Counter) {
	defer wg.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(p.reportInterval) * time.Second):
			go doReqSendMetrics(p.url, httpC, gs, cs)
		}
	}
}

func doReqSendMetrics(url string, httpC *http.Client, gs *[]models.Gauge, cs *map[string]models.Counter) {
	tmp_url := url + "/update/" + "counter/"
	for _, metric := range *cs {
		endpoints.SendMetricEndpoint(tmp_url+metric.Name+"/"+fmt.Sprintf("%v", metric.Value), httpC)
	}

	tmp_url = url + "/update/" + "gauge/"
	for _, metric := range *gs {
		endpoints.SendMetricEndpoint(tmp_url+metric.Name+"/"+strconv.FormatFloat(metric.Value, 'f', -1, 64), httpC)
	}
}

func initialization(params *initParams, httpC *http.Client, m *models.Monitor) error {
	var err error

	*httpC = http.Client{}

	err = parseFlags(params)
	if err != nil {
		return err
	}

	err = getENV(params)
	if err != nil {
		return err
	}

	params.url = params.url + params.PORT
	rand.Seed(time.Now().Unix())
	m.Init()

	return err
}

func getENV(params *initParams) error {
	var err error

	envRunAddr := os.Getenv("ADDRESS")

	if envRunAddr != "" {
		addrIsValid(envRunAddr, params)
	}

	if envReportInterval := os.Getenv("REPORT_INTERVAL"); envReportInterval != "" {
		value, err := strconv.Atoi(envReportInterval)
		if err != nil {
			return err
		} else {
			params.reportInterval = value
		}
	}

	if envPollInterval := os.Getenv("POLL_INTERVAL"); envPollInterval != "" {
		value, err := strconv.Atoi(envPollInterval)
		if err != nil {
			return err
		} else {
			params.pollInterval = value
		}
	}

	return err
}

func addrIsValid(addr string, params *initParams) error {
	res, err := validate.IsMatchesTemplate(addr, params.validateAddrPattern)
	if err != nil {
		return err
	} else {
		if res {
			params.PORT = addr
		} else {
			return errors.New("addr is not valid")
		}
	}

	return err
}

func parseFlags(params *initParams) error {
	var err error

	flag.StringVar(&params.PORT, "a", "localhost:8080", "Port to listen on.")
	flag.IntVar(&params.pollInterval, "p", defPollInterval, "Frequency of sending metrics to the server.")
	flag.IntVar(&params.reportInterval, "r", defReportInterval, "Frequency of polling metrics from the runtime package.")
	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT, params.validateAddrPattern)

	if err != nil {
		return nil
	} else {
		if !res {
			return errors.New("addr is not valid")
		}
	}

	return err
}

func setValuesMonitor(m *models.Monitor, gs *[]models.Gauge, cs *map[string]models.Counter) {
	const minRandomValue float64 = 1.0

	const maxRandomValue float64 = 999.0

	writeFromMemory(m)
	m.PollCount.Value += 1

	tmpCounters := make(map[string]models.Counter, 1)
	tmpCounters["PollCount"] = m.PollCount
	fmt.Println(tmpCounters["PollCount"])

	tmpGauges := make([]models.Gauge, 0, metricGaugeCount)

	m.RandomValue.Value = random.RandomF64(minRandomValue, maxRandomValue)

	tmpGauges = append(tmpGauges, m.Alloc, m.BuckHashSys, m.Frees, m.GCCPUFraction, m.GCSys)
	tmpGauges = append(tmpGauges, m.HeapAlloc, m.HeapIdle, m.HeapInuse, m.HeapObjects, m.HeapReleased)
	tmpGauges = append(tmpGauges, m.HeapSys, m.LastGC, m.Lookups, m.MCacheInuse, m.MCacheInuse)
	tmpGauges = append(tmpGauges, m.MCacheSys, m.MSpanInuse, m.MSpanSys, m.Mallocs, m.NextGC)
	tmpGauges = append(tmpGauges, m.NumForcedGC, m.NumGC, m.OtherSys, m.PauseTotalNs, m.StackInuse)
	tmpGauges = append(tmpGauges, m.StackSys, m.Sys, m.TotalAlloc, m.RandomValue)

	*gs = tmpGauges
	*cs = tmpCounters
}

func writeFromMemory(m *models.Monitor) {
	var rtm runtime.MemStats

	runtime.ReadMemStats(&rtm)

	m.Alloc.Value = float64(rtm.Alloc)
	m.BuckHashSys.Value = float64(rtm.BuckHashSys)
	m.Frees.Value = float64(rtm.Frees)
	m.GCCPUFraction.Value = rtm.GCCPUFraction
	m.GCSys.Value = float64(rtm.GCSys)
	m.HeapAlloc.Value = float64(rtm.HeapAlloc)
	m.HeapIdle.Value = float64(rtm.HeapIdle)
	m.HeapInuse.Value = float64(rtm.HeapInuse)
	m.HeapObjects.Value = float64(rtm.HeapObjects)
	m.HeapReleased.Value = float64(rtm.HeapReleased)
	m.HeapSys.Value = float64(rtm.HeapSys)
	m.LastGC.Value = float64(rtm.LastGC)
	m.Lookups.Value = float64(rtm.Lookups)
	m.MCacheInuse.Value = float64(rtm.MCacheInuse)
	m.MCacheSys.Value = float64(rtm.MCacheSys)
	m.MSpanInuse.Value = float64(rtm.MSpanInuse)
	m.MSpanSys.Value = float64(rtm.MSpanSys)
	m.Mallocs.Value = float64(rtm.Mallocs)
	m.NextGC.Value = float64(rtm.NextGC)
	m.NumForcedGC.Value = float64(rtm.NumForcedGC)
	m.NumGC.Value = float64(rtm.NumGC)
	m.OtherSys.Value = float64(rtm.OtherSys)
	m.PauseTotalNs.Value = float64(rtm.PauseTotalNs)
	m.StackInuse.Value = float64(rtm.StackInuse)
	m.StackSys.Value = float64(rtm.StackSys)
	m.Sys.Value = float64(rtm.Sys)
	m.TotalAlloc.Value = float64(rtm.TotalAlloc)
}
