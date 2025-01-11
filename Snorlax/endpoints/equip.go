package endpoints

import (
	"Snorlax/VRChatAPI/avatars"
	"net/http"
)

func init() {
	RegisterEndpoint("api/avatars/equip", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		if _, err := avatars.SelectAvatar(&GlobalClient, id); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}
