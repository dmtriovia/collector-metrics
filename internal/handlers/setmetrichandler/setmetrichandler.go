package setmetrichandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

type setMetricHandler struct {
	serv service.Service
}

const metrics string = "gauge|counter"

type validMetric struct {
	mtype        string
	mname        string
	mvalue       string
	mvalue_float float64
	mvalue_int   int64
}

func NewSetMetricHandler(serv service.Service) *setMetricHandler {
	return &setMetricHandler{serv: serv}
}

func (h *setMetricHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
}

func (h *setMetricHandler) SetMetricHandler(w http.ResponseWriter, r *http.Request) {
	var validMetric *validMetric = new(validMetric)

	getReqData(r, validMetric)

	isValid, status := isValidMetric(r, validMetric)
	if !isValid {
		w.WriteHeader(status)

		return
	} else {
		addMetricToMemStore(h, validMetric)
		w.WriteHeader(status)

		Body := "OK\n"
		fmt.Fprintf(w, "%s", Body)

		return
	}
}

func getReqData(r *http.Request, m *validMetric) {
	m.mtype = mux.Vars(r)["metric_type"]
	m.mname = mux.Vars(r)["metric_name"]
	m.mvalue = mux.Vars(r)["metric_value"]
}

func addMetricToMemStore(h *setMetricHandler, m *validMetric) {
	if m.mtype == "gauge" {
		h.serv.AddGauge(m.mname, m.mvalue_float)
	} else if m.mtype == "counter" {
		h.serv.AddCounter(m.mname, m.mvalue_int)
	}
}

func isValidMetric(r *http.Request, m *validMetric) (bool, int) {
	if !validate.IsMethodPost(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string = "^[0-9a-zA-Z/ ]{1,20}$"
	res, _ := validate.IsMatchesTemplate(m.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + metrics + "$"
	res, _ = validate.IsMatchesTemplate(m.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	if !isValidMeticValue(m) {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

func isValidMeticValue(m *validMetric) bool {
	if m.mtype == "gauge" {
		return isValidGaugeValue(m)
	} else if m.mtype == "counter" {
		return isValidCounterValue(m)
	}

	return false
}

func isValidGaugeValue(m *validMetric) bool {
	value, err := strconv.ParseFloat(m.mvalue, 64)
	if err == nil {
		m.mvalue_float = value

		return true
	} else {
		return false
	}
}

func isValidCounterValue(m *validMetric) bool {
	value, err := strconv.ParseInt(m.mvalue, 10, 64)
	if err == nil {
		m.mvalue_int = value

		return true
	} else {
		return false
	}
}
