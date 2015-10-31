package main

import (
	"bytes"
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
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
	mock                  = flag.Int("m", 1, "Use mock response when needle is found")
	matches        uint64 = 0
	needles        []string
	mocks          []string
)

const (
	POST_HEAD   = "text/xml;charset=UTF-8"
	SERVER_PATH = "/proxy"
	KILL_PATH   = "/kill"

	NEEDLE_MOCK_SQL = "SELECT needle, mock FROM proxy.mapping"

	MATCH_FOUND_STR   = "Found the needle"
	PROXY_INIT_STR    = "Proxy initializing.."
	NEEDLE_MOCK_STR   = "Loading needles and mocks..."
	PROXY_ACTIVE_STR  = "=========== Proxy activate ==========="
	RECEIVED_MSG_STR  = "<<<<<<< Started request proxy >>>>>>>>"
	MOCK_MSG_STR      = "Returning mock value"
	PROXY_MSG_STR     = "Proxied request: returning actual value"
	DONE_STR          = "<<<<<<<<<<<<<< Finished >>>>>>>>>>>>>>"
	SHUTTING_DOWN_STR = "=========== Shutting down ============"

	DB_USER     = "http"
	DB_PASSWORD = "proxy"
	DB_NAME     = "httpproxy"
)

/**
 * Parse input flags and start HTTP server on *fromPort
 */
func main() {
	flag.Parse()
	log.Println(PROXY_INIT_STR)
	getNeedlesAndMocks()
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
		log.Println(MOCK_MSG_STR)
		responseBody = []byte(mockResponse)
	} else {
		log.Println(PROXY_MSG_STR)
		responseBody = getResponse(requestBody)
	}
	responseChannel <- responseBody
	log.Println(DONE_STR)
}

/**
* Function to determine whether certain needle exists in request
 */
func getMockOnMatch(input []byte) (mockResponse string) {
	for key, needle := range needles {
		if strings.Contains(string(input), needle) {
			atomic.AddUint64(&matches, 1)
			log.Println(MATCH_FOUND_STR, atomic.LoadUint64(&matches))
			if *mock == 1 {
				mockResponse = mocks[key]
			}
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

/**
 * Read needles and mock responses from DB
 */
func getNeedlesAndMocks() {
	log.Println(NEEDLE_MOCK_STR)
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable", DB_USER, DB_PASSWORD, DB_NAME)
	db, err := sql.Open("postgres", dbinfo)
	handleError(err)
	defer db.Close()

	rows, err := db.Query(NEEDLE_MOCK_SQL)
	handleError(err)

	for rows.Next() {
		var needle string
		var mock string
		err = rows.Scan(&needle, &mock)
		handleError(err)
		needles = append(needles, needle)
		mocks = append(mocks, mock)
	}
}
