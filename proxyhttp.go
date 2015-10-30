package main

import( "net/http"
	"io/ioutil"
	"strings"
	"log"
	"flag"
	"sync/atomic"
	"bytes"
	"os"
)

var (
	fromPort = flag.String("f", ":12345", "The proxy server's port")
	toHost = flag.String("t", "http://wsf.cdyne.com/WeatherWS/Weather.asmx", "The host that the proxy server should forward requests to")
	maxConnections = flag.Int("c", 25, "The maximum number of concurrent connection")
	needle = flag.String("n", "94301", "The needle to search from the requests")
	mock = flag.String("m", "", "Mock to return when needle is found (without line breaks)")
	matches uint64 = 0
)

const (
	POST_TYPE = "text/xml;charset=UTF-8"
	SERVER_PATH = "/"
	MATCH_FOUND_STR = "Found the needle"
	PROXY_ACTIVE_STR = "Proxy active..."
	RECEIEVED_MSG_STR = "Started request proxy >>>>>>>>>>>>>>>>"
	MOCK_MSG_STR = "...using mock value"
	PROXY_MSG_STR = "...using actual value"
	PROXY_STR = "Proxyng request"
	REVERSE_PROXY_STR = "Reverse-proxyng response..."
	DONE_STR = "<<<<<<<<<<<<<<<<<Finished"
	SHUTTING_DOWN_STR = "Shutting proxy down"
)

/**
 * HTTP server for proxy
 */
func ProxyServer(outputStream http.ResponseWriter, request *http.Request) {
	responseChannel := make(chan []byte, *maxConnections)						//create a channel for responses
	defer close(responseChannel)												//eventually close channel when done
	go func() {																	//async function for proxying
		var responseBody []byte
		log.Println(RECEIEVED_MSG_STR)
		match := false															//flag whether needle is found
		requestBody, err := ioutil.ReadAll(request.Body)						//read request that needs to be proxied
		handleErr(err)
		request.Body.Close()
		if strings.Contains(string(requestBody), *needle) {						//check for needle in request
			match = true														//set flag
			atomic.AddUint64(&matches, 1)										//increment thread-safe counter
			log.Println(MATCH_FOUND_STR, atomic.LoadUint64(&matches))
		}
		log.Println(REVERSE_PROXY_STR)
		if match && *mock != "" {												//if match is found and mock is set
			log.Println(MOCK_MSG_STR)
			responseBody = []byte(*mock)										//then return mock response
		} else {																//when no needle is set or match found
			log.Println(PROXY_STR)
			response, err := http.Post(*toHost, POST_TYPE, bytes.NewBuffer(requestBody)) //return actual response
			handleErr(err)

			log.Println(PROXY_MSG_STR)
			responseBody, err = ioutil.ReadAll(response.Body)					//read actual response body
			handleErr(err)
			response.Body.Close()
		}
		responseChannel <- responseBody											//write response to response channel
		log.Println(DONE_STR)
	}()
	outputStream.Write(<-responseChannel)										//write channel content to original system
}

/**
 * HTTP server stop servlet
 */
func KillServer (outputStream http.ResponseWriter, request *http.Request) {
	log.Println(SHUTTING_DOWN_STR)
	os.Exit(0)
}

/**
 * Parse input flags and start HTTP server on *fromPort
 */
func main() {
	flag.Parse()
	log.Println(PROXY_ACTIVE_STR)
	http.HandleFunc("/proxy", ProxyServer)
	http.HandleFunc("/die", KillServer)
	err := http.ListenAndServe(*fromPort, nil)
	handleErr(err)
}

/**
 * Error handler
 */
func handleErr(err error) {
	if err != nil {
		panic(err)
	}
}