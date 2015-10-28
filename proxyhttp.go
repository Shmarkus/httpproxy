package main

import (
	"net/http"
	"io/ioutil"
	"strings"
)
const POST_TYPE = "text/xml;charset=UTF-8"
const URI = "http://wsf.cdyne.com/WeatherWS/Weather.asmx"

/**
 * App listens on TCP 12345. everything posted on this port will be submitted to cdyne weather webservice
 */
func ProxyServer(w http.ResponseWriter, request *http.Request) {
	output := make(chan []byte, 1)
	go worker(request, output)
	w.Write(<-output)
}

func worker(request *http.Request, output chan []byte) {
	requestBody, _ := ioutil.ReadAll(request.Body)
	defer request.Body.Close()
	response, _:= http.Post(URI, POST_TYPE, strings.NewReader(string(requestBody)))//TODO: bytes.NewBuffer(rbytes[:read]))
	responseBody, _ := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	output <- responseBody
}

func main() {
	http.HandleFunc("/", ProxyServer)
	http.ListenAndServe(":12345", nil)
}