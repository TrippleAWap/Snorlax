package endpoints

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

var endpoints = make(map[string]func(w http.ResponseWriter, r *http.Request))

func RegisterEndpoint(endpoint string, handler func(w http.ResponseWriter, r *http.Request)) {
	log.Printf("Register endpoint %s\n", endpoint)
	endpoints[endpoint] = handler
}

func StartServer(port int) {
	fmt.Printf("Started server on 127.0.0.1:%d\n", port)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		endpoint := strings.ToLower(strings.TrimSpace(r.URL.Path))
		endpoint = strings.TrimSuffix(endpoint, "/")
		endpoint = strings.TrimPrefix(endpoint, "/")
		log.Printf("Endpoint \"%s\" called with the method %s by %s\n", endpoint, r.Method, r.RemoteAddr)
		if handler, ok := endpoints[endpoint]; ok {
			defer handler(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(port), nil))
}
