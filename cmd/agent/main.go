package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const url string = "http://localhost:8080"
const pollInterval int = 2
const reportInterval int = 10
const delay_close int = 10

var wg sync.WaitGroup

type myData struct {
	r   *http.Response
	err error
}

func connect(c context.Context) error {

	defer wg.Done()

	channelCancel := make(chan os.Signal, 1) // we need to reserve to buffer size 1, so the notifier are not blocked
	signal.Notify(channelCancel, os.Interrupt, syscall.SIGTERM)

	data := make(chan myData, 1)
	tr := &http.Transport{}
	//httpClient := &http.Client{Transport: tr}
	req, _ := http.NewRequest("POST", url, nil)

	for {
		/*go func() {
			response, err := httpClient.Do(req)
			if err != nil {
				fmt.Println(err)
				data <- myData{nil, err}
				return
			} else {
				pack := myData{response, err}
				data <- pack
			}
		}()*/

		select {
		case <-channelCancel:
			fmt.Println("Exit")
			return nil
		case <-time.After(time.Duration(pollInterval) * time.Second):
			fmt.Println("Hello in a pollInterval")
		case <-time.After(time.Duration(reportInterval) * time.Second):
			fmt.Println("Hello in a reportInterval")
		case <-c.Done():
			tr.CancelRequest(req)
			<-data
			fmt.Println("The request was canceled!")
			return c.Err()
		case ok := <-data:
			err := ok.err
			resp := ok.r
			if err != nil {
				fmt.Println("Error select:", err)
				return err
			}
			defer resp.Body.Close()

			realHTTPData, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Println("Error select:", err)
				return err
			}
			// Although fmt.Printf() is used here, server processes
			// use the log.Printf() function instead.
			fmt.Printf("Server Response: %s\n", realHTTPData)
		}
	}
	return nil
}

func main() {

	c := context.Background()
	del_close := time.Duration(delay_close) * time.Second
	c, cancel := context.WithTimeout(c, del_close)
	defer cancel()

	wg.Add(1)
	go connect(c)
	wg.Wait()

	fmt.Println("Exiting...")
}
