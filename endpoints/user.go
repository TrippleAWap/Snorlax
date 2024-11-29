package endpoints

import (
	"encoding/json"
	"net/http"
)

func init() {
	RegisterEndpoint("api/user", func(w http.ResponseWriter, r *http.Request) {
		bytes, err := json.Marshal(GlobalUser)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			panic(err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write(bytes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			panic(err)
			return
		}
	})
}
