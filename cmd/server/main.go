package main

import (
	"fmt"
	"handlers"
	"middleware"
	"net/http"
	"time"
)

func main() {

	mux := http.NewServeMux()
	s := &http.Server{
		Addr:         handlers.PORT,
		Handler:      mux,
		IdleTimeout:  10 * time.Second,
		ReadTimeout:  time.Second,
		WriteTimeout: time.Second,
	}

	//mux.Handle("/update/gauge/{filename:[a-zA-Z0-9][a-zA-Z0-9\\.]*[a-zA-Z0-9]}", middleware.MiddlewareGetMetric(http.HandlerFunc(handlers.MetricGaugeHandler)))
	//mux.Handle("/update/counter/{filename:[a-zA-Z0-9][a-zA-Z0-9\\.]*[a-zA-Z0-9]}", middleware.MiddlewareGetMetric(http.HandlerFunc(handlers.MetricCounterHandler)))
	mux.Handle("/update/{metric_type}/{metric_name}/{metric_value}", middleware.MiddlewareGetMetric(http.HandlerFunc(handlers.GetMetricHandler)))
	mux.Handle("/", http.HandlerFunc(handlers.DefaultHandler))

	fmt.Println("Ready to serve at", handlers.PORT)
	err := s.ListenAndServe()
	if err != nil {
		fmt.Println(err)
		return
	}
}
