package main

import (
	"handlers"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
)

func TestWebhook(t *testing.T) {
	// тип http.HandlerFunc реализует интерфейс http.Handler
	// это поможет передать хендлер тестовому серверу
	handler := http.HandlerFunc(handlers.GetMetricHandler)
	// запускаем тестовый сервер, будет выбран первый свободный порт
	srv := httptest.NewServer(handler)
	// останавливаем сервер после завершения теста
	defer srv.Close()

	// описываем набор данных: метод запроса, ожидаемый код ответа, ожидаемое тело
	testCases := []struct {
		test_number   string
		metric_method string
		metric_type   string
		metric_name   string
		metric_value  string
		expectedCode  int
		expectedBody  string
	}{
		{test_number: "1", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "1.0", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "2", metric_method: "update", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusOK, expectedBody: ""},

		{test_number: "3", metric_method: "no_update", metric_type: "counter", metric_name: "Name", metric_value: "1", expectedCode: http.StatusNotFound, expectedBody: ""},
		{test_number: "4", metric_method: "update", metric_type: "counter_new", metric_name: "Name", metric_value: "1", expectedCode: http.StatusBadRequest, expectedBody: ""},

		{test_number: "5", metric_method: "update", metric_type: "counter", metric_name: "Name", metric_value: "111111111111111111111111111111111111", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{test_number: "6", metric_method: "update", metric_type: "counter", metric_name: "Name", metric_value: "-1", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "7", metric_method: "update", metric_type: "counter", metric_name: "Name", metric_value: "-1.0", expectedCode: http.StatusBadRequest, expectedBody: ""},
		{test_number: "8", metric_method: "update", metric_type: "counter", metric_name: "Name", metric_value: "-1.1", expectedCode: http.StatusBadRequest, expectedBody: ""},

		{test_number: "9", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "111111111111111111111111111111111111.0", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "10", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "-1.0", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "11", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "-1.5", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "12", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "-1", expectedCode: http.StatusOK, expectedBody: ""},
		{test_number: "13", metric_method: "update", metric_type: "gauge", metric_name: "Name", metric_value: "5", expectedCode: http.StatusOK, expectedBody: ""},

		{test_number: "14", metric_method: "update", metric_type: "counter", metric_name: "_Name123_", metric_value: "1", expectedCode: http.StatusNotFound, expectedBody: ""},
	}

	for _, tc := range testCases {
		t.Run(http.MethodPost, func(t *testing.T) {
			// делаем запрос с помощью библиотеки resty к адресу запущенного сервера,
			// который хранится в поле URL соответствующей структуры
			req := resty.New().R()
			req.Method = http.MethodPost
			//req.URL = srv.URL
			req.Header.Set("Content-Type", "text/plain")
			req.URL = "http://localhost:8080/" + tc.metric_method + "/" + tc.metric_type + "/" + tc.metric_name + "/" + tc.metric_value

			resp, err := req.Send()
			assert.NoError(t, err, tc.test_number+": error making HTTP request ")

			assert.Equal(t, tc.expectedCode, resp.StatusCode(), tc.test_number+": Response code didn't match expected")
			// проверяем корректность полученного тела ответа, если мы его ожидаем
			if tc.expectedBody != "" {
				assert.JSONEq(t, tc.expectedBody, string(resp.Body()))
			}
		})
	}
}
