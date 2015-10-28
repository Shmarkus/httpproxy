package main

import (
	"net/http"
	"io/ioutil"
	"strings"
	"log"
	"flag"
)

var fromHost = flag.String("f", ":12345", "The proxy server's port")
var toHost = flag.String("t", "http://wsf.cdyne.com/WeatherWS/Weather.asmx", "The host that the proxy server should forward requests to")
var maxConnections = flag.Int("c", 25, "The maximum number of concurrent connection")
var needle = flag.String("n", "94301", "The needle to search from the requests")

const POST_TYPE = "text/xml;charset=UTF-8"

/**
 * App listens on TCP 12345. everything posted on this port will be submitted to cdyne weather webservice
 */
func ProxyServer(w http.ResponseWriter, request *http.Request) {
	responseChannel := make(chan []byte, *maxConnections)
	defer close(responseChannel)
	go worker(request, responseChannel)
	w.Write(<-responseChannel)
}

func worker(request *http.Request, responseChannel chan []byte) {
	requestBody, _ := ioutil.ReadAll(request.Body)
	request.Body.Close()
	if strings.Contains(string(requestBody), *needle) {
		log.Println("Contains match!")
	}
	response, _:= http.Post(*toHost, POST_TYPE, strings.NewReader(string(requestBody)))//TODO: bytes.NewBuffer(rbytes[:read]))
	responseBody, _ := ioutil.ReadAll(response.Body)
	response.Body.Close()
	responseChannel <- responseBody
}

func generator () {
	id := make(chan int64)
	go func() {
		var counter int64 = 0
		for {
			id <- counter
			counter += 1
		}
	}()
}

func main() {
	flag.Parse()
	http.HandleFunc("/", ProxyServer)
	http.ListenAndServe(*fromHost, nil)
}