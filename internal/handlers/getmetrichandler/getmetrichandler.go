package getmetrichandler

import (
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/gorilla/mux"
)

const metrics string = "gauge|counter"

type validMetric struct {
	mtype string
	mname string
}

type ansData struct {
	mvalue string
}

type getMetricHandler struct {
	serv service.Service
}

func NewGetMetricHandler(s service.Service) *getMetricHandler {
	return &getMetricHandler{serv: s}
}

func (h *getMetricHandler) GetMetricHandler(w http.ResponseWriter, r *http.Request) {
	var validMetric *validMetric = new(validMetric)

	getReqData(r, validMetric)

	isValid, status := isValidMetric(r, validMetric)
	if !isValid {
		w.WriteHeader(status)

		return
	} else {
		var answerData *ansData = new(ansData)
		isSetAnsData, status := setAnswerData(validMetric, answerData, h)

		if isSetAnsData {
			w.WriteHeader(status)

			Body := answerData.mvalue
			fmt.Fprintf(w, "%s", Body)

			return
		} else {
			w.WriteHeader(status)

			return
		}
	}
}

func getReqData(r *http.Request, metric *validMetric) {
	metric.mname = mux.Vars(r)["metric_name"]
	metric.mtype = mux.Vars(r)["metric_type"]
}

func isValidMetric(r *http.Request, metric *validMetric) (bool, int) {
	if !validate.IsMethodGet(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string = "^[0-9a-zA-Z/ ]{1,20}$"
	res, _ := validate.IsMatchesTemplate(metric.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + metrics + "$"
	res, _ = validate.IsMatchesTemplate(metric.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

func setAnswerData(metric *validMetric, ansd *ansData, h *getMetricHandler) (bool, int) {
	if metric.mtype == "gauge" {
		setValueByType(metric, ansd, h.serv.GetStringValueGaugeMetric)
	} else if metric.mtype == "counter" {
		setValueByType(metric, ansd, h.serv.GetStringValueCounterMetric)
	}

	return false, http.StatusNotFound
}

func setValueByType(metric *validMetric, ansd *ansData, getFunction func(string) (string, error)) (bool, int) {
	metricStringValue, err := getFunction(metric.mname)

	if err != nil {
		return false, http.StatusNotFound
	} else {
		ansd.mvalue = metricStringValue

		return true, http.StatusOK
	}
}