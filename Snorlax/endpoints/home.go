package endpoints

import (
	_ "embed"
	"net/http"
	"strings"
	"time"
)

//go:embed static/index.html
var html string

func init() {
	startTime := time.Now()
	RegisterEndpoint("home", func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "index.html", startTime, strings.NewReader(html))
	})
}
