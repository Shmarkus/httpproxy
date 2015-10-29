package main

import( "net/http"; "io/ioutil"; "strings"; "log"; "flag"; "sync/atomic"; "bytes")

var fromHost = flag.String("f", ":12345", "The proxy server's port")
var toHost = flag.String("t", "http://wsf.cdyne.com/WeatherWS/Weather.asmx", "The host that the proxy server should forward requests to")
var maxConnections = flag.Int("c", 25, "The maximum number of concurrent connection")
var needle = flag.String("n", "94301", "The needle to search from the requests")
var matches uint64 = 0

const POST_TYPE = "text/xml;charset=UTF-8"
const MATCH_FOUND_STR = "Match found! #"

func ProxyServer(w http.ResponseWriter, request *http.Request) {
	responseChannel := make(chan []byte, *maxConnections)
	defer close(responseChannel)
	go func() {
		requestBody, _ := ioutil.ReadAll(request.Body)
		request.Body.Close()
		if strings.Contains(string(requestBody), *needle) {
			atomic.AddUint64(&matches, 1)
			log.Println(MATCH_FOUND_STR, atomic.LoadUint64(&matches))
		}
		response, _:= http.Post(*toHost, POST_TYPE, bytes.NewBuffer(requestBody))
		responseBody, _ := ioutil.ReadAll(response.Body)
		response.Body.Close()
		responseChannel <- responseBody
	}()
	w.Write(<-responseChannel)
}

func main() {
	flag.Parse()
	http.HandleFunc("/", ProxyServer)
	http.ListenAndServe(*fromHost, nil)
}
