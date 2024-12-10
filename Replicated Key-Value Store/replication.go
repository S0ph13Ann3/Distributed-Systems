package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

//imported libraries

// /Our view store global variable
// var view = make(map[string]string)
var servers []string = strings.Split(os.Getenv("VIEW"), ",")

// PUT request at /view with JSON body {"socket-address":"<IP:PORT>"}
func PutReplication(key string, value string, version int) {
	for _, address := range servers {

		if address == ipAddr {
			continue
		}

		//creating url
		httpStr := "http://"
		endpoint := "/replica"
		url := httpStr + address + endpoint

		//creating body
		body := make(map[string]interface{})
		body["method"] = "PUT"
		body["key"] = key
		body["value"] = value
		body["version"] = version

		fmt.Println("replicating : ", key, " to ", version)
		fmt.Println("body: ", body)

		bodyData, _ := json.Marshal(body)
		formattedBody := bytes.NewBuffer(bodyData)

		//making request
		request, _ := http.NewRequest("PUT", url, formattedBody)

		client := &http.Client{
			Timeout: 750 * time.Millisecond,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		response, badResponse := client.Do(request)

		if badResponse == nil {
			response.Body.Close()
		}

	}
}

// DELETE request at /view with JSON body {"socket-address":"<IP:PORT>"}
func DeleteReplication(key string, version int) {
	for _, address := range servers {

		if address == ipAddr {
			continue
		}

		// Creating url
		httpStr := "http://"
		endpoint := "/replica"
		url := httpStr + address + endpoint

		// Creating body
		body := make(map[string]interface{})
		body["method"] = "DELETE"
		body["key"] = key
		body["version"] = version
		bodyData, _ := json.Marshal(body)
		formattedBody := bytes.NewBuffer(bodyData)

		// Making request
		request, _ := http.NewRequest("DELETE", url, formattedBody)
		request.Header.Set("Content-Type", "application/json")
		client := &http.Client{
			Timeout: 750 * time.Millisecond,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		response, badResponse := client.Do(request)

		if badResponse == nil {
			response.Body.Close()
		}

	}
}
