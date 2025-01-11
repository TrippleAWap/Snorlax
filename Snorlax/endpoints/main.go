package endpoints

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strconv"
	"strings"
)

var endpoints = make(map[string]func(w http.ResponseWriter, r *http.Request))

func RegisterEndpoint(endpoint string, handler func(w http.ResponseWriter, r *http.Request)) {
	log.Printf("Register endpoint %s\n", endpoint)
	endpoints[endpoint] = handler
}
func GetEndpoint(endpoint string) *http.Response {
	handler, ok := endpoints[endpoint]
	if !ok {
		return nil
	}
	req, _ := http.NewRequest("GET", endpoint, nil)
	responseWriterV := httptest.NewRecorder()
	handler(responseWriterV, req)
	return responseWriterV.Result()
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

func CombineFilesToHTML(files ...string) (string, error) {
	html := ""
	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			if os.IsNotExist(err) {
				cwd, err := os.Getwd()
				if err != nil {
					return "", err
				}

				fileAbs := path.Join(cwd, file)
				log.Printf("File %s does not exist resolved path %s\n", file, fileAbs)
			}
			return "", err
		}
		switch strings.Split(file, ".")[len(strings.Split(file, "."))-1] {
		case "html":
			html += string(bytes)
		case "css":
			html += "<style>" + string(bytes) + "</style>"
		case "js":
			html += "<script>" + string(bytes) + "</script>"
		default:
			return "", fmt.Errorf("unsupported file type %s", file)
		}
	}
	return html, nil
}
