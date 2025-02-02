package endpoints

import (
	"Snorlax/VRChatAPI"
	"fmt"
	"io"
	"net/http"
)

func init() {
	RegisterEndpoint("api/avatars/thumbnail", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		avatarData := cachedAvatarIdToAvatar[id]

		req, err := http.NewRequest("GET", avatarData.ThumbnailImageUrl, nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for k, v := range VRChatAPI.DefaultHeaders {
			req.Header.Set(k, v)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Printf("Failed to get thumbnail %s: %s\n", avatarData.ThumbnailImageUrl, err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			http.Error(w, fmt.Sprintf("Failed to get thumbnail %s: %s", avatarData.ThumbnailImageUrl, res.Status), http.StatusInternalServerError)
			return
		}

		bytes, err := io.ReadAll(res.Body)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to read thumbnail %s: %s", avatarData.ThumbnailImageUrl, err.Error()), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=2592000")

		if _, err := w.Write(bytes); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
