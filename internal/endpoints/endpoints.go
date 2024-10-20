package endpoints

import (
	"fmt"
	"net/http"
)

func SendMetricEndpoint(endpoint string, httpC *http.Client) {

	const contentTypeSendMetric string = "text/plain"

	req, _ := http.NewRequest("POST", endpoint, nil)
	req.Header.Set("Content-Type", contentTypeSendMetric)
	_, err := httpC.Do(req)
	if err != nil {
		fmt.Println(err)
	}
}
