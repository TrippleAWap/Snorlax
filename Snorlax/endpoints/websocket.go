package endpoints

import (
	"Snorlax/RejectDatabase"
	"Snorlax/VRChatAPI/avatars"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

var upgrader = websocket.Upgrader{}

type (
	EventType struct {
		EventId string      `json:"eventId"`
		Data    interface{} `json:"data,omitempty"`
	}
	EventTypeAvatars []interface{}
)

var avatarsSent []string
var avatarIds []string
var cacheToAvatarId map[string]string
var AvatarsV []avatars.Avatar

func init() {
	RegisterEndpoint("", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Websocket endpoint")
		// Upgrade upgrades the HTTP server connection to the WebSocket protocol.
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade failed: ", err)
			return
		}

		defer conn.Close()

		for {
			cacheToAvatarId, err = ScrapeIdsFromCache()
			if err != nil {
				fmt.Println(err)
				continue
			}

			for _, id := range cacheToAvatarId {
				if slices.Contains(avatarsSent, id) {
					continue
				}

				avatarsSent = append(avatarsSent, id)
				avatarIds = append(avatarIds, id)
			}

			avatarMap := map[string]avatars.Avatar{}

			fmt.Println("Fetching avatars...")
			avatarMap, err = GetAvatars(avatarIds)
			fmt.Println("Fetched avatars")

			if err != nil {
				_ = conn.WriteJSON(EventType{"error", err.Error()})
				return
			}

		addAvatars:
			for _, avatar := range avatarMap {
				for _, avatarV := range AvatarsV {
					if avatar.Id == avatarV.Id {
						continue addAvatars
					}
				}
				AvatarsV = append(AvatarsV, avatar)
			}

			go func() {
				database, err := RejectDatabase.GetCachedAvatars()
				if err != nil {
					return
				}
				totalAdded := 0
				var uniqueAvatars []avatars.Avatar
			findUnique:
				for _, avatar := range AvatarsV {
					for _, cachedAvatar := range database {
						if cachedAvatar.Id == avatar.Id {
							continue findUnique
						}
					}
					uniqueAvatars = append(uniqueAvatars, avatar)
				}
				for _, avatar := range uniqueAvatars {
					if err := RejectDatabase.AddAvatar(avatar, GlobalUser.DisplayName); err != nil {
						log.Println("add avatar to database failed: ", err)
						continue
					}
					totalAdded++
				}
				log.Printf("Succesfully added %d avatars to the Reject database!\n", totalAdded)
			}()

			slices.SortFunc(AvatarsV, func(a, b avatars.Avatar) int {
				return int(b.CacheTime.Unix() - a.CacheTime.Unix())
			})

			html, visibleCount := RenderAvatars(AvatarsV, Filter)
			cards := strings.SplitAfter(html, `            </div>
        </div>
    </div>`)
			data := []interface{}{strings.Join(cards[:min(len(cards), 1000)], ""), visibleCount, len(AvatarsV)}
			if err := conn.WriteJSON(EventType{"avatars", data}); err != nil {
				log.Println("write json failed: ", err)
				return
			}

			avatarMap = nil
			time.Sleep(5 * time.Second)
			cacheToAvatarId = nil
			avatarIds = nil
		}
	})
}
