package main

import (
	"handlers"
	"log"
	"middleware"
	"models"
	"net/http"
	"os"
	"os/signal"
	"service"
	"time"

	"github.com/gorilla/mux"
)

var memStorage models.MemStorage
var s *http.Server
var serv *service.MetricService

func main() {

	initialization()

	go func() {
		log.Println("Listening to", handlers.PORT)
		err := s.ListenAndServe()
		if err != nil {
			log.Printf("Error starting server: %s\n", err)
			return
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Println("Quitting after signal:", sig)
	time.Sleep(2 * time.Second)
	s.Shutdown(nil)
}

func initialization() {

	mux := mux.NewRouter()

	memStorage.Init()
	handler := handlers.NewSetMetricHandler(serv, &memStorage)

	postMux := mux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handler.SetMetricHandler)
	postMux.Use(middleware.MiddlewareSetMetric)

	getMux := mux.Methods(http.MethodGet).Subrouter()
	getMux.HandleFunc("/value/{metric_type}/{metric_name}", handler.GetMetricHandler)
	getMux.Use(middleware.MiddlewareSetMetric)

	notAllowed := handlers.NotAllowedHandler{}
	mux.MethodNotAllowedHandler = notAllowed

	mux.NotFoundHandler = http.HandlerFunc(handlers.DefaultHandler)

	s = &http.Server{
		Addr:         handlers.PORT,
		Handler:      mux,
		ErrorLog:     nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

}
