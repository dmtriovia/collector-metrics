package defaultHandler

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/dmitrovia/collector-metrics/internal/service"
)

type defaultHandler struct {
	serv service.Service
}

func NewDefaultHandler(serv service.Service) *defaultHandler {
	return &defaultHandler{serv: serv}
}

type ViewData struct {
	Metrics map[string]string
}

func (h *defaultHandler) DefaultHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotFound)
	data := ViewData{
		Metrics: *h.serv.GetMapStringsAllMetrics(),
	}
	tmpl, err := template.ParseFiles("../../internal/html/allMetricsTemplate.html")
	if err != nil {
		fmt.Println(err)
	} else {
		tmpl.Execute(w, data)
	}
}
