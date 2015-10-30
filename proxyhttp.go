package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
)

var (
	localPort             = flag.String("f", ":12345", "The proxy server's port")
	endpoint              = flag.String("t", "http://wsf.cdyne.com/WeatherWS/Weather.asmx", "The host that the proxy server should forward requests to")
	maxConnections        = flag.Int("c", 25, "The maximum number of concurrent connection")
	needle                = flag.String("n", "94301", "The needle to search from the requests")
	mock                  = flag.String("m", "", "Mock to return when needle is found (without line breaks)")
	matches        uint64 = 0
)

const (
	POST_HEAD         = "text/xml;charset=UTF-8"
	SERVER_PATH       = "/proxy"
	KILL_PATH         = "/kill"
	MATCH_FOUND_STR   = "Found the needle"
	PROXY_ACTIVE_STR  = "========== Activated proxy ==========="
	RECEIVED_MSG_STR  = "<<<<<<< Started request proxy >>>>>>>>"
	MOCK_MSG_STR      = "Returning mock value"
	PROXY_MSG_STR     = "Proxied request: returning actual value"
	DONE_STR          = "<<<<<<<<<<<<<< Finished >>>>>>>>>>>>>>"
	SHUTTING_DOWN_STR = "=========== Shutting down ============"
)

/**
 * Parse input flags and start HTTP server on *fromPort
 */
func main() {
	flag.Parse()
	log.Println(PROXY_ACTIVE_STR)
	http.HandleFunc(SERVER_PATH, ProxyServer)
	http.HandleFunc(KILL_PATH, KillServer)
	err := http.ListenAndServe(*localPort, nil)
	handleError(err)
}

/**
 * HTTP server for proxy
 */
func ProxyServer(outputStream http.ResponseWriter, request *http.Request) {
	responseChannel := make(chan []byte, *maxConnections)
	defer close(responseChannel)
	go proxy(request, responseChannel)
	outputStream.Write(<-responseChannel)
}

/**
 * Actual proxy method
 */
func proxy(request *http.Request, responseChannel chan []byte) {
	log.Println(RECEIVED_MSG_STR)
	var responseBody []byte
	requestBody, err := ioutil.ReadAll(request.Body)
	handleError(err)
	request.Body.Close()
	mockResponse := getMockOnMatch(requestBody)
	if len(mockResponse) > 0 {
		log.Println(PROXY_MSG_STR)
		responseBody = getResponse(requestBody)
	} else {
		log.Println(MOCK_MSG_STR)
		responseBody = []byte(mockResponse)
	}
	responseChannel <- responseBody
	log.Println(DONE_STR)
}

/**
 * Function to determine whether certain needle exists in request
 */
func getMockOnMatch(input []byte) (mockResponse string) {
	if strings.Contains(string(input), *needle) {
		atomic.AddUint64(&matches, 1)
		log.Println(MATCH_FOUND_STR, atomic.LoadUint64(&matches))
		if len(*mock) > 0 {
			mockResponse = *mock
		}
	}
	return mockResponse
}

/**
 * Read the actual response from the endpoint and return it
 */
func getResponse(request []byte) (responseBody []byte) {
	byteReader := bytes.NewBuffer(request)
	response, err := http.Post(*endpoint, POST_HEAD, byteReader)
	handleError(err)
	responseBody, err = ioutil.ReadAll(response.Body)
	handleError(err)
	response.Body.Close()
	return responseBody
}

/**
 * Error handler function to improve readability
 */
func handleError(err error) {
	if err != nil {
		panic(err)
	}
}

/**
 * HTTP server stop servlet
 */
func KillServer(outputStream http.ResponseWriter, request *http.Request) {
	log.Println(SHUTTING_DOWN_STR)
	os.Exit(0)
}
