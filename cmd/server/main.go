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

	"github.com/dmitrovia/collector-metrics/internal/functions/validate"
	"github.com/dmitrovia/collector-metrics/internal/handlers/defaulthandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/getmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/notallowedhandler"
	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/middleware"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
)

const rTimeout = 10

const wTimeout = 10

const iTimeout = 30

type initParams struct {
	PORT                string
	validateAddrPattern string
}

func main() {
	var memStorage *memoryrepository.MemoryRepository
	memStorage = new(memoryrepository.MemoryRepository)
	MemoryService := service.NewMemoryService(memStorage)

	var params *initParams
	params = new(initParams)
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	var s *http.Server = new(http.Server)

	err := initiate(params, memStorage, MemoryService, s)
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

func initiate(p *initParams, mr *memoryrepository.MemoryRepository, ms *service.MemoryService, s *http.Server) error {
	var err error

	err = parseFlags(p)
	if err != nil {
		return err
	}

	err = getENV(p)
	if err != nil {
		return err
	}

	mux := mux.NewRouter()

	mr.Init()

	handlerSet := setmetrichandler.NewSetMetricHandler(ms)

	handlerGet := getmetrichandler.NewGetMetricHandler(ms)
	handlerDefault := defaulthandler.NewDefaultHandler(ms)
	handlerNotAllowed := notallowedhandler.NotAllowedHandler{}

	postMux := mux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handlerSet.SetMetricHandler)
	postMux.Use(middleware.MiddlewareSetMetric)

	getMux := mux.Methods(http.MethodGet).Subrouter()
	getMux.HandleFunc("/value/{metric_type}/{metric_name}", handlerGet.GetMetricHandler)
	getMux.Use(middleware.MiddlewareSetMetric)

	mux.MethodNotAllowedHandler = handlerNotAllowed

	mux.NotFoundHandler = http.HandlerFunc(handlerDefault.DefaultHandler)

	*s = http.Server{
		Addr:         p.PORT,
		Handler:      mux,
		ErrorLog:     nil,
		ReadTimeout:  rTimeout * time.Second,
		WriteTimeout: wTimeout * time.Second,
		IdleTimeout:  iTimeout * time.Second,
	}

	return err
}

func parseFlags(params *initParams) error {
	var err error

	flag.StringVar(&params.PORT, "a", "localhost:8080", "Port to listen on.")
	flag.Parse()

	res, err := validate.IsMatchesTemplate(params.PORT, params.validateAddrPattern)

	if !res {
		return errors.New("addr is not valid")
	}

	return err
}

func getENV(params *initParams) error {
	var err error

	envRunAddr := os.Getenv("ADDRESS")

	if envRunAddr != "" {
		addrIsValid(envRunAddr, params)
	}

	return err
}

func addrIsValid(addr string, params *initParams) error {
	res, err := validate.IsMatchesTemplate(addr, params.validateAddrPattern)
	if err != nil {
		return err
	} else {
		if res {
			params.PORT = addr
		} else {
			return errors.New("addr is not valid")
		}
	}

	return err
}
