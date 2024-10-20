package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/dmitrovia/collector-metrics/internal/functions/validate_f"
	"github.com/dmitrovia/collector-metrics/internal/handlers/defaultHandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getMetricHandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/notAllowedHandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setMetricHandler"
	"github.com/dmitrovia/collector-metrics/internal/middleware"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryRepository"

	"github.com/gorilla/mux"
)

type initParameters struct {
	PORT                string
	validateAddrPattern string
}

func main() {

	var memStorage *memoryRepository.MemoryRepository = new(memoryRepository.MemoryRepository)

	var s *http.Server = new(http.Server)
	MemoryService := service.NewMemoryService(memStorage)
	var params *initParameters = new(initParameters)
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err := initialization(params, memStorage, MemoryService, s)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		log.Println("Listening to", ":"+params.PORT)
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
	s.Shutdown(context.TODO())
}

func initialization(params *initParameters, inMemStorage *memoryRepository.MemoryRepository, inService *service.MemoryService, inServer *http.Server) error {

	var err error = nil

	err = parseFlags(params)
	if err != nil {
		return err
	}
	err = getENV(params)
	if err != nil {
		return err
	}

	mux := mux.NewRouter()

	inMemStorage.Init()
	handlerSet := setMetricHandler.NewSetMetricHandler(inService)
	handlerGet := getMetricHandler.NewGetMetricHandler(inService)
	handlerDefault := defaultHandler.NewDefaultHandler(inService)
	handlerNotAllowed := notAllowedHandler.NotAllowedHandler{}

	postMux := mux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handlerSet.SetMetricHandler)
	postMux.Use(middleware.MiddlewareSetMetric)

	getMux := mux.Methods(http.MethodGet).Subrouter()
	getMux.HandleFunc("/value/{metric_type}/{metric_name}", handlerGet.GetMetricHandler)
	getMux.Use(middleware.MiddlewareSetMetric)

	mux.MethodNotAllowedHandler = handlerNotAllowed

	mux.NotFoundHandler = http.HandlerFunc(handlerDefault.DefaultHandler)

	*inServer = http.Server{
		Addr:         params.PORT,
		Handler:      mux,
		ErrorLog:     nil,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	return err
}

func parseFlags(params *initParameters) error {

	var err error = nil
	flag.StringVar(&params.PORT, "a", "localhost:8080", "Port to listen on.")
	flag.Parse()

	res, err := validate_f.IsMatchesTemplate(params.PORT, params.validateAddrPattern)
	if !res {
		return errors.New("addr is not valid")
	}
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
	return err
}
