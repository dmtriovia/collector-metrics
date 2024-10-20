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
	"github.com/dmitrovia/collector-metrics/internal/functions/random_f"
	"github.com/dmitrovia/collector-metrics/internal/functions/validate_f"
	"github.com/dmitrovia/collector-metrics/internal/models"
)

type initParameters struct {
	url                 string
	PORT                string
	reportInterval      int
	pollInterval        int
	validateAddrPattern string
}

func main() {

	var wg *sync.WaitGroup = new(sync.WaitGroup)
	var monitor *models.Monitor = new(models.Monitor)
	var httpClient *http.Client = new(http.Client)
	var gauges *[]models.Gauge = new([]models.Gauge)
	var counters *map[string]models.Counter = new(map[string]models.Counter)

	var params *initParameters = new(initParameters)
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
	go collectMetrics(monitor, params, wg, gauges, counters)
	wg.Add(1)
	go sendMetrics(params, wg, httpClient, gauges, counters)
	wg.Wait()
}

func collectMetrics(m *models.Monitor, params *initParameters, waitGroup *sync.WaitGroup, inGauges *[]models.Gauge, inCounters *map[string]models.Counter) {

	defer waitGroup.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(params.pollInterval) * time.Second):
			setValuesMonitor(m, inGauges, inCounters)
		}
	}
}

func sendMetrics(params *initParameters, waitGroup *sync.WaitGroup, httpC *http.Client, inGauges *[]models.Gauge, inCounters *map[string]models.Counter) {

	defer waitGroup.Done()

	channelCancel := make(chan os.Signal, 1)
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case <-channelCancel:
			return
		case <-time.After(time.Duration(params.reportInterval) * time.Second):
			go doReqSendMetrics(params.url, httpC, inGauges, inCounters)
		}
	}
}

func doReqSendMetrics(urlServer string, httpC *http.Client, inGauges *[]models.Gauge, inCounters *map[string]models.Counter) {

	tmp_url := urlServer + "/update/" + "counter/"
	for _, metric := range *inCounters {
		endpoints.SendMetricEndpoint(tmp_url+metric.Name+"/"+fmt.Sprintf("%v", metric.Value), httpC)
	}
	tmp_url = urlServer + "/update/" + "gauge/"
	for _, metric := range *inGauges {
		endpoints.SendMetricEndpoint(tmp_url+metric.Name+"/"+fmt.Sprintf("%.02f", metric.Value), httpC)
	}
}

func initialization(params *initParameters, httpC *http.Client, m *models.Monitor) error {

	var err error = nil

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

func getENV(params *initParameters) error {

	var err error = nil

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		res, err := validate_f.IsMatchesTemplate(envRunAddr, params.validateAddrPattern)
		if err != nil {
			return nil
		} else {
			if res {
				params.PORT = envRunAddr
			} else {
				return errors.New("addr is not valid")
			}
		}
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

func parseFlags(params *initParameters) error {

	var err error = nil

	flag.StringVar(&params.PORT, "a", "localhost:8080", "Port to listen on.")
	flag.IntVar(&params.pollInterval, "p", 2, "Frequency of sending metrics to the server.")
	flag.IntVar(&params.reportInterval, "r", 10, "Frequency of polling metrics from the runtime package.")
	flag.Parse()

	res, err := validate_f.IsMatchesTemplate(params.PORT, params.validateAddrPattern)

	if err != nil {
		return nil
	} else {
		if !res {
			return errors.New("addr is not valid")
		}
	}

	return err
}

func setValuesMonitor(m *models.Monitor, inGauges *[]models.Gauge, inCounters *map[string]models.Counter) {

	const minRandomValue float64 = 1.0
	const maxRandomValue float64 = 999.0

	var rtm runtime.MemStats
	runtime.ReadMemStats(&rtm)

	m.PollCount.Value += 1

	tmpCounters := make(map[string]models.Counter, 1)
	tmpCounters["PollCount"] = m.PollCount
	fmt.Println(tmpCounters["PollCount"])

	tmpGauges := make([]models.Gauge, 0, 27)

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
	m.RandomValue.Value = random_f.RandomF64(minRandomValue, maxRandomValue)

	tmpGauges = append(tmpGauges, m.Alloc, m.BuckHashSys, m.Frees, m.GCCPUFraction, m.GCSys, m.HeapAlloc, m.HeapIdle, m.HeapInuse, m.HeapObjects, m.HeapReleased)
	tmpGauges = append(tmpGauges, m.HeapSys, m.LastGC, m.Lookups, m.MCacheInuse, m.MCacheInuse, m.MCacheSys, m.MSpanInuse, m.MSpanSys, m.Mallocs, m.NextGC)
	tmpGauges = append(tmpGauges, m.NumForcedGC, m.NumGC, m.OtherSys, m.PauseTotalNs, m.StackInuse, m.StackSys, m.Sys, m.TotalAlloc, m.RandomValue)

	*inGauges = tmpGauges
	*inCounters = tmpCounters
}
