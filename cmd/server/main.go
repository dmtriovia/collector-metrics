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

var errParseFlags = errors.New("addr is not valid")

type initParams struct {
	PORT                string
	validateAddrPattern string
}

func main() {
	var memStorage *memoryrepository.MemoryRepository

	var params *initParams

	var server *http.Server

	memStorage = new(memoryrepository.MemoryRepository)
	MemoryService := service.NewMemoryService(memStorage)
	server = new(http.Server)

	params = new(initParams)
	params.validateAddrPattern = "^[a-zA-Z/ ]{1,100}:[0-9]{1,10}$"

	err := initiate(params, memStorage, MemoryService, server)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	go func() {
		log.Println("Listening to", ":"+params.PORT)

		err := server.ListenAndServe()
		if err != nil {
			log.Printf("Error starting server: %s\n", err)

			return
		}
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	sig := <-sigs
	log.Println("Quitting after signal:", sig)

	err = server.Shutdown(context.TODO())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func initiate(par *initParams, mrep *memoryrepository.MemoryRepository, mser *service.MemoryService, server *http.Server) error {
	var err error

	err = parseFlags(par)
	if err != nil {
		return err
	}

	err = getENV(par)
	if err != nil {
		return err
	}

	mux := mux.NewRouter()

	mrep.Init()

	handlerSet := setmetrichandler.NewSetMetricHandler(mser)

	handlerGet := getmetrichandler.NewGetMetricHandler(mser)
	handlerDefault := defaulthandler.NewDefaultHandler(mser)
	handlerNotAllowed := notallowedhandler.NotAllowedHandler{}

	postMux := mux.Methods(http.MethodPost).Subrouter()
	postMux.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handlerSet.SetMetricHandler)
	postMux.Use(middleware.SetMetric)

	getMux := mux.Methods(http.MethodGet).Subrouter()
	getMux.HandleFunc("/value/{metric_type}/{metric_name}", handlerGet.GetMetricHandler)

	mux.MethodNotAllowedHandler = handlerNotAllowed

	mux.NotFoundHandler = http.HandlerFunc(handlerDefault.DefaultHandler)

	*server = http.Server{
		Addr:         par.PORT,
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
		return errParseFlags
	}

	return fmt.Errorf("parseFlags: %w", err)
}

func getENV(params *initParams) error {
	var err error

	envRunAddr := os.Getenv("ADDRESS")

	if envRunAddr != "" {
		err = addrIsValid(envRunAddr, params)
		if err != nil {
			return err
		}
	}

	return err
}

func addrIsValid(addr string, params *initParams) error {
	res, err := validate.IsMatchesTemplate(addr, params.validateAddrPattern)
	if err == nil {
		if res {
			params.PORT = addr
		} else {
			return errParseFlags
		}
	}

	return fmt.Errorf("addrIsValid: %w", err)
}
