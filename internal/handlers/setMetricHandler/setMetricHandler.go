package setMetricHandler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate_f"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

type setMetricHandler struct {
	serv service.Service
}

const acceptedContentType string = "text/plain"
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

func getReqData(r *http.Request, inMetric *validMetric) {
	inMetric.mtype = mux.Vars(r)["metric_type"]
	inMetric.mname = mux.Vars(r)["metric_name"]
	inMetric.mvalue = mux.Vars(r)["metric_value"]
}

func addMetricToMemStore(h *setMetricHandler, inMetric *validMetric) {
	if inMetric.mtype == "gauge" {
		h.serv.AddGauge(inMetric.mname, inMetric.mvalue_float)

	} else if inMetric.mtype == "counter" {
		h.serv.AddCounter(inMetric.mname, inMetric.mvalue_int)
	}
}

func isValidMetric(r *http.Request, inMetric *validMetric) (bool, int) {

	// не работают локальные автотесты 3 инкремента с данной проверкой
	res, _ := validate_f.IsMatchesTemplate(r.Header.Get("Content-Type"), acceptedContentType)
	if !res {
		return false, http.StatusBadRequest
	}

	if !validate_f.IsMethodPost(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string = "^[0-9a-zA-Z/ ]{1,20}$"
	res, _ = validate_f.IsMatchesTemplate(inMetric.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + metrics + "$"
	res, _ = validate_f.IsMatchesTemplate(inMetric.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	if !isValidMeticValue(inMetric) {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

func isValidMeticValue(inMetric *validMetric) bool {

	if inMetric.mtype == "gauge" {

		if value, err := strconv.ParseFloat(inMetric.mvalue, 64); err == nil {
			inMetric.mvalue_float = value
			return true
		} else {
			return false
		}

	} else if inMetric.mtype == "counter" {

		if value, err := strconv.ParseInt(inMetric.mvalue, 10, 64); err == nil {
			inMetric.mvalue_int = value
			return true
		} else {
			return false
		}

	} else {
		return false
	}
}
