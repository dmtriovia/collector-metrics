package main

import (
	"fmt"
	"handlers"
	"middleware"
	"models"
	"net/http"
	"service"
	"time"
)

var memStorage models.MemStorage
var s *http.Server
var serv *service.MetricService

func main() {

	initialization()

	fmt.Println("Ready to serve at", handlers.PORT)
	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}

func initialization() {

	mux := http.NewServeMux()
	s = &http.Server{
		Addr:         handlers.PORT,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	memStorage.Init()

	handler := handlers.NewAPIHandler(serv, &memStorage)

	mux.Handle("/update/{metric_type}/{metric_name}/{metric_value}", middleware.MiddlewareGetMetric(http.HandlerFunc(handler.SetMetricHandler)))
	mux.Handle("/", http.HandlerFunc(handlers.DefaultHandler))

}
