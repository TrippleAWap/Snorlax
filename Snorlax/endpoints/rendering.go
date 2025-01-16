package endpoints

import (
	"Snorlax/VRChatAPI/avatars"
	_ "embed"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"
)

//go:embed static/card.html
var cardTemplate string

var supportedPlatformToDisplay = map[string]string{
	"standalonewindows": "Windows",
	"android":           "Android",
	"ios":               "iOS",
}

func RenderAvatar(avatar avatars.Avatar, filter string, favoritesOnly bool) (bool, string, error) {
	if favoritesOnly && !CachedIdToFavorites[avatar.Id] {
		return false, "", nil
	}
	filter = strings.ToLower(filter)

	var supportedPlatforms []string

	for _, up := range avatar.UnityPackages {
		display, ok := supportedPlatformToDisplay[up.Platform]
		if !ok {
			display = up.Platform
		}
		if slices.Contains(supportedPlatforms, display) {
			continue
		}
		supportedPlatforms = append(supportedPlatforms, display)
	}

	platformsString := strings.Join(supportedPlatforms, " | ")

	cacheTime := avatar.CacheTime.Format("01/02/06 15:04:05")

	searchedString := avatar.Id + "\uFFFF" + avatar.Name + "\uFFFF" + avatar.AuthorName + "\uFFFF" + platformsString + "\uFFFF" + cacheTime + "\uFFFF" + avatar.Description + "\uFFFF" + avatar.Id
	starColor := "white"
	if CachedIdToFavorites[avatar.Id] {
		starColor = "gold"
		searchedString += "\uFFFF" + "favorite"
	}

	if !strings.Contains(strings.ToLower(searchedString), filter) {
		return false, "", nil
	}
	card := fmt.Sprintf(cardTemplate,
		starColor,
		avatar.Id,
		avatar.Id,
		avatar.Name,
		avatar.AuthorName,
		platformsString,
		cacheTime,
		avatar.Description,
		avatar.Id,
	)
	return true, card, nil
}

var Filter = ""
var FavoritesOnly = false

func RenderAvatars(avatarsV []avatars.Avatar, filter string) (string, int) {
	log.Println("Rendering avatars...")
	var innerHTML []string
	var innerHTMLMutex sync.Mutex
	visibleAvatars := 0
	batchSize := 100
	wg := sync.WaitGroup{}
	fmt.Printf("Rendering %d avatars with %d threads\n", len(avatarsV), int(math.Ceil(float64(len(avatarsV))/float64(batchSize))))
	threadId := 0
	for i := 0; i < len(avatarsV); i += batchSize {
		wg.Add(1)
		innerHTML = append(innerHTML, "")
		go func(threadId int, avatarsV []avatars.Avatar, filter string, favoritesOnly bool) {
			defer wg.Done()
			var innerHTMLV string

			for _, avatar := range avatarsV {
				visible, html, err := RenderAvatar(avatar, filter, favoritesOnly)
				if err != nil {
					fmt.Println(err)
					continue
				}
				if visible {
					visibleAvatars++
				}
				innerHTMLV += html
			}

			innerHTMLMutex.Lock()
			defer innerHTMLMutex.Unlock()
			innerHTML[threadId] = innerHTMLV
		}(threadId, avatarsV[i:min(i+batchSize, len(avatarsV))], filter, FavoritesOnly)
		threadId++
	}
	wg.Wait()
	log.Printf("Rendered %d avatars with %d visible avatars\n", len(avatarsV), visibleAvatars)
	return strings.Join(innerHTML, ""), visibleAvatars
}
