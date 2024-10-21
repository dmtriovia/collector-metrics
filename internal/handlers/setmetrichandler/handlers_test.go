package setmetrichandler_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/handlers/setmetrichandler"
	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryrepository"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const url string = "http://localhost:8080"

const ok int = http.StatusOK

const nallwd int = http.StatusMethodNotAllowed

const nfnd int = http.StatusNotFound

const bdreq int = http.StatusBadRequest

func SetMetricHandler(t *testing.T) {
	var memStorage *memoryrepository.MemoryRepository = new(memoryrepository.MemoryRepository)
	MemoryService := service.NewMemoryService(memStorage)
	memStorage.Init()

	handler := setmetrichandler.NewSetMetricHandler(MemoryService)

	var tmpstr string = "111111111111111111111111111111111111"

	var tmpstr1 string = "111111111111111111111111111111111111.0"
	testCases := []struct {
		tn     string
		mt     string
		mn     string
		mv     string
		expcod int
		exbody string
		method string
	}{
		{method: "POST", tn: "1", mt: "gauge", mn: "Name", mv: "1.0", expcod: ok, exbody: ""},
		{method: "POST", tn: "2", mt: "counter", mn: "Name", mv: "1", expcod: ok, exbody: ""},
		{method: "POST", tn: "3", mt: "counter", mn: "Name", mv: "1", expcod: ok, exbody: ""},
		{method: "POST", tn: "4", mt: "counter_new", mn: "Name", mv: "1", expcod: bdreq, exbody: ""},
		{method: "POST", tn: "5", mt: "counter", mn: "Name", mv: tmpstr, expcod: bdreq, exbody: ""},
		{method: "POST", tn: "6", mt: "counter", mn: "Name", mv: "-1", expcod: ok, exbody: ""},
		{method: "POST", tn: "7", mt: "counter", mn: "Name", mv: "-1.0", expcod: bdreq, exbody: ""},
		{method: "POST", tn: "8", mt: "counter", mn: "Name", mv: "-1.1", expcod: bdreq, exbody: ""},
		{method: "POST", tn: "9", mt: "gauge", mn: "Name", mv: tmpstr1, expcod: ok, exbody: ""},
		{method: "POST", tn: "10", mt: "gauge", mn: "Name", mv: "-1.0", expcod: ok, exbody: ""},
		{method: "POST", tn: "11", mt: "gauge", mn: "Name", mv: "-1.5", expcod: ok, exbody: ""},
		{method: "POST", tn: "12", mt: "gauge", mn: "Name", mv: "-1", expcod: ok, exbody: ""},
		{method: "POST", tn: "13", mt: "gauge", mn: "Name", mv: "5", expcod: ok, exbody: ""},
		{method: "POST", tn: "14", mt: "counter", mn: "_Name123_", mv: "1", expcod: nfnd, exbody: ""},
		{method: "PATCH", tn: "15", mt: "counter", mn: "Name", mv: "1", expcod: nallwd, exbody: ""},
		{method: "POST", tn: "17", mt: "gauge", mn: "Name", mv: "ASD", expcod: bdreq, exbody: ""},
	}

	for _, tc := range testCases {
		t.Run(http.MethodPost, func(t *testing.T) {
			req, err := http.NewRequest(tc.method, url+"/update/"+tc.mt+"/"+tc.mn+"/"+tc.mv, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", "text/plain")

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/update/{mt}/{mn}/{mv}", handler.SetMetricHandler)
			router.ServeHTTP(rr, req)
			status := rr.Code
			body, _ := io.ReadAll(rr.Body)

			assert.NoError(t, err, tc.tn+": error making HTTP request ")
			assert.Equal(t, tc.expcod, status, tc.tn+": Response code didn't match expected")

			if tc.exbody != "" {
				assert.JSONEq(t, tc.exbody, string(body))
			}
		})
	}
}
