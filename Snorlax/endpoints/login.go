package endpoints

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

//go:embed static/login.html
var loginHTML string

func init() {
	startTime := time.Now()
	RegisterEndpoint("login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", startTime, strings.NewReader(loginHTML))
	})
}
