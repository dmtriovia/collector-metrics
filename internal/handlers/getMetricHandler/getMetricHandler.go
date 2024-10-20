package getMetricHandler

import (
	"fmt"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate_f"
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

func NewGetMetricHandler(serv service.Service) *getMetricHandler {
	return &getMetricHandler{serv: serv}
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

func getReqData(r *http.Request, inMetric *validMetric) {
	inMetric.mname = mux.Vars(r)["metric_name"]
	inMetric.mtype = mux.Vars(r)["metric_type"]
}

func isValidMetric(r *http.Request, inMetric *validMetric) (bool, int) {

	if !validate_f.IsMethodGet(r.Method) {
		return false, http.StatusMethodNotAllowed
	}

	var pattern string = "^[0-9a-zA-Z/ ]{1,20}$"
	res, _ := validate_f.IsMatchesTemplate(inMetric.mname, pattern)

	if !res {
		return false, http.StatusNotFound
	}

	pattern = "^" + metrics + "$"
	res, _ = validate_f.IsMatchesTemplate(inMetric.mtype, pattern)

	if !res {
		return false, http.StatusBadRequest
	}

	return true, http.StatusOK
}

func setAnswerData(inMetric *validMetric, inAnswerData *ansData, h *getMetricHandler) (bool, int) {

	if inMetric.mtype == "gauge" {
		metricStringValue, err := h.serv.GetStringValueGaugeMetric(inMetric.mname)

		if err != nil {
			return false, http.StatusNotFound
		} else {
			inAnswerData.mvalue = metricStringValue
			return true, http.StatusOK
		}
	} else if inMetric.mtype == "counter" {
		metricStringValue, err := h.serv.GetStringValueCounterMetric(inMetric.mname)
		if err != nil {
			return false, http.StatusNotFound
		} else {
			inAnswerData.mvalue = metricStringValue
			return true, http.StatusOK
		}
	} else {
		return false, http.StatusNotFound
	}

}
