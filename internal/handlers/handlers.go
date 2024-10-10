package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"service"
	"strconv"
)

const PORT string = ":8080"
const metrics string = "gauge|counter"
const acceptedContentType string = "text/plain"

type metricHandler struct {
	serv service.Service
}

func newMetricHandler(serv service.Service) *metricHandler {
	return &metricHandler{serv: serv}
}

func DefaultHandler(w http.ResponseWriter, r *http.Request) {

	/*if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}*/

	w.WriteHeader(http.StatusNotFound)
}

func GetMetricHandler(w http.ResponseWriter, r *http.Request) {

	/*use in the future*/

	/*s := &service.MetricService{}
	handler := newMetricHandler(s)
	temp, err := handler.serv.GetMetric("any")
	fmt.Println(temp, err)*/

	if !isValidContentType(r.Header.Get("Content-Type")) { // в middleware ?
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isMethodPost(r.Method) { // в middleware ?
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var mType string = r.PathValue("metric_type")
	var mName string = r.PathValue("metric_name")
	var mValue string = r.PathValue("metric_value")

	if !isValidMetricName(mName) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if !isValidMetricType(mType) || !isValidMeticValue(mValue, mType) {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	Body := "OK\n"
	fmt.Fprintf(w, "%s", Body)

}

func isMethodPost(method string) bool {
	if method == http.MethodPost {
		return true
	} else {
		return false
	}
}

func isValidContentType(contentType string) bool {

	var pattern string = acceptedContentType

	res, err := MatchString(pattern, contentType)
	if err == nil && res == true {
		return true
	} else {
		return false
	}

}

func MatchString(pattern string, s string) (matched bool, err error) {
	re, err := regexp.Compile(pattern)

	if err == nil {
		return re.MatchString(s), nil
	} else {
		return false, err
	}

}

func isValidMetricType(metricType string) bool {

	var pattern string = "^" + metrics + "$"

	res, err := MatchString(pattern, metricType)
	if err == nil && res == true {
		return true
	} else {
		return false
	}

}

func isValidMetricName(metricName string) bool {

	var pattern string = "^[a-zA-Z/ ]{1,20}$"

	res, err := MatchString(pattern, metricName)
	if err == nil && res == true {
		return true
	} else {
		return false
	}
}

func isValidMeticValue(metricValue string, metricType string) bool {

	if metricType == "gauge" {

		if _, err := strconv.ParseFloat(metricValue, 64); err == nil {
			return true
		} else {
			return false
		}

	} else if metricType == "counter" {

		if _, err := strconv.ParseInt(metricValue, 10, 64); err == nil {
			return true
		} else {
			return false
		}

	} else {
		return false
	}
}
