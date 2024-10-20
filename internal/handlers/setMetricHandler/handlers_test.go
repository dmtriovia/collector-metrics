package setMetricHandler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dmitrovia/collector-metrics/internal/service"
	"github.com/dmitrovia/collector-metrics/internal/storage/memoryRepository"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const url string = "http://localhost:8080"

func TestSetMetricHandler(t *testing.T) {

	var memStorage *memoryRepository.MemoryRepository = new(memoryRepository.MemoryRepository)
	MemoryService := service.NewMemoryService(memStorage)

	memStorage.Init()
	handler := NewSetMetricHandler(MemoryService)

	testCases := []struct {
		test_number   string
		metric_method string
		metric_type   string
		metric_name   string
		metric_value  string
		expectedCode  int
		expectedBody  string
		method        string
		contentType   string
	}{

		{contentType: "text/plain", method: "POST", test_number: "1", metric_type: "gauge", metric_name: "Name", metric_value: "1.0", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "2", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusOK, expectedBody: ""},

		{contentType: "text/plain", method: "POST", test_number: "3", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "4", metric_type: "counter_new", metric_name: "Name", metric_value: "1", expectedCode: http.StatusBadRequest, expectedBody: ""},

		{contentType: "text/plain", method: "POST", test_number: "5", metric_type: "counter", metric_name: "Name", metric_value: "111111111111111111111111111111111111", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "6", metric_type: "counter", metric_name: "Name", metric_value: "-1", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "7", metric_type: "counter", metric_name: "Name", metric_value: "-1.0", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "8", metric_type: "counter", metric_name: "Name", metric_value: "-1.1", expectedCode: http.StatusBadRequest, expectedBody: ""},

		{contentType: "text/plain", method: "POST", test_number: "9", metric_type: "gauge", metric_name: "Name", metric_value: "111111111111111111111111111111111111.0", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "10", metric_type: "gauge", metric_name: "Name", metric_value: "-1.0", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "11", metric_type: "gauge", metric_name: "Name", metric_value: "-1.5", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "12", metric_type: "gauge", metric_name: "Name", metric_value: "-1", expectedCode: http.StatusOK, expectedBody: ""},
		{contentType: "text/plain", method: "POST", test_number: "13", metric_type: "gauge", metric_name: "Name", metric_value: "5", expectedCode: http.StatusOK, expectedBody: ""},

		{contentType: "text/plain", method: "POST", test_number: "14", metric_type: "counter", metric_name: "_Name123_", metric_value: "1", expectedCode: http.StatusNotFound, expectedBody: ""},
		{contentType: "text/plain", method: "PATCH", test_number: "15", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusMethodNotAllowed, expectedBody: ""},

		//{contentType: "application/json", method: "POST", test_number: "16", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusBadRequest, expectedBody: ""},

		{contentType: "text/plain", method: "POST", test_number: "17", metric_type: "gauge", metric_name: "Name", metric_value: "ASD", expectedCode: http.StatusBadRequest, expectedBody: ""},
	}

	for _, tc := range testCases {
		t.Run(http.MethodPost, func(t *testing.T) {

			req, err := http.NewRequest(tc.method, url+"/update/"+tc.metric_type+"/"+tc.metric_name+"/"+tc.metric_value, nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Content-Type", tc.contentType)
			//req.SetPathValue("metric_type", tc.metric_type)
			//req.SetPathValue("metric_name", tc.metric_name)
			//req.SetPathValue("metric_value", tc.metric_value)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.HandleFunc("/update/{metric_type}/{metric_name}/{metric_value}", handler.SetMetricHandler)
			router.ServeHTTP(rr, req)

			status := rr.Code

			body, _ := io.ReadAll(rr.Body)
			assert.NoError(t, err, tc.test_number+": error making HTTP request ")

			assert.Equal(t, tc.expectedCode, status, tc.test_number+": Response code didn't match expected")

			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(body))
			}
		})
	}
}
