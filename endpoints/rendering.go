package endpoints

import (
	"Snorlax/vrcAPI/avatars"
	"fmt"
	"log"
	"math"
	"slices"
	"strings"
	"sync"
)

var cardTemplate = `
	<div class="w-full sm:w-1/3 md:w-1/5 lg:w-1/7 px-4 mb-8">
		<div class="mt-1 linear-gradient(to right, rgba(0, 0, 0, 0.24), rgba(15, 15, 15, 0.315)) rounded-lg overflow-hidden shadow-lg transition-transform duration-300 hover:scale-105" style="box-shadow: 0 0 10px 2px #9b77ff, 0 0 10px 4px #9b77ff, 0 0 20px 6px #9b77ff">
			<div class="aspect-[4/3] relative">
				<svg
						viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" fill="%s" style="position: absolute; right: 0; top: 0; aspect-ratio: 1; max-width: 40px; width: 15%%; min-width: 35px;"
						onclick="
						const favorite = this.style.fill !== 'yellow';
						this.style.fill = favorite ? 'yellow' : 'white';
						favoriteAvatar('%s', favorite);
				">
					<path d="M11.245 4.174C11.4765 3.50808 11.5922 3.17513 11.7634 3.08285C11.9115 3.00298 12.0898 3.00298 12.238 3.08285C12.4091 3.17513 12.5248 3.50808 12.7563 4.174L14.2866 8.57639C14.3525 8.76592 14.3854 8.86068 14.4448 8.93125C14.4972 8.99359 14.5641 9.04218 14.6396 9.07278C14.725 9.10743 14.8253 9.10947 15.0259 9.11356L19.6857 9.20852C20.3906 9.22288 20.743 9.23007 20.8837 9.36432C21.0054 9.48051 21.0605 9.65014 21.0303 9.81569C20.9955 10.007 20.7146 10.2199 20.1528 10.6459L16.4387 13.4616C16.2788 13.5829 16.1989 13.6435 16.1501 13.7217C16.107 13.7909 16.0815 13.8695 16.0757 13.9507C16.0692 14.0427 16.0982 14.1387 16.1563 14.3308L17.506 18.7919C17.7101 19.4667 17.8122 19.8041 17.728 19.9793C17.6551 20.131 17.5108 20.2358 17.344 20.2583C17.1513 20.2842 16.862 20.0829 16.2833 19.6802L12.4576 17.0181C12.2929 16.9035 12.2106 16.8462 12.1211 16.8239C12.042 16.8043 11.9593 16.8043 11.8803 16.8239C11.7908 16.8462 11.7084 16.9035 11.5437 17.0181L7.71805 19.6802C7.13937 20.0829 6.85003 20.2842 6.65733 20.2583C6.49056 20.2358 6.34626 20.131 6.27337 19.9793C6.18915 19.8041 6.29123 19.4667 6.49538 18.7919L7.84503 14.3308C7.90313 14.1387 7.93218 14.0427 7.92564 13.9507C7.91986 13.8695 7.89432 13.7909 7.85123 13.7217C7.80246 13.6435 7.72251 13.5829 7.56262 13.4616L3.84858 10.6459C3.28678 10.2199 3.00588 10.007 2.97101 9.81569C2.94082 9.65014 2.99594 9.48051 3.11767 9.36432C3.25831 9.23007 3.61074 9.22289 4.31559 9.20852L8.9754 9.11356C9.176 9.10947 9.27631 9.10743 9.36177 9.07278C9.43726 9.04218 9.50414 8.99359 9.55657 8.93125C9.61593 8.86068 9.64887 8.76592 9.71475 8.57639L11.245 4.174Z" stroke="#000000" stroke-linecap="round" stroke-linejoin="round" stroke-width="0" style="top: 0; right: 0;"></path>
				</svg>
				<img src="api/avatars/thumbnail?id=%s" alt="Avatar Thumbnail" style="object-fit: cover; width: 100%%; height: 100%%;">
			</div>
			<div class="p-4">
				<a class="text-lg text-white mb-2 font-bold">%s</a>
				<div class="flex items-center mb-0">
					<div class="w-2 h-2 bg-zinc-600 rounded-full mr-2"></div>
					<h2 class="text-base text-neutral-400 mb-1">%s</h2>
				</div>
				<div class="flex items-center mb-2">
					<div class="w-2 h-2 bg-zinc-600 rounded-full mr-2"></div>
					<h2 class="text-sm text-neutral-400 mb-1">%s</h2>
				</div>
				<div class="flex items-center mb-3">
					<div class="w-full bg-zinc-600 rounded-full">
						<div class="bg-green-500 h-0.5 rounded-full" style="width: 0"></div>
					</div>
					<div class="ml-2 text-xs text-neutral-400"></div>
				</div>
				<p class="text-sm text-gray-300 mb-2">%s</p>
				<p class="text-neutral-400 text-base">%s</p>
				<button onclick="fetch('/api/avatars/equip?id=%s')" class="bg-gradient-to-r from-[#4949498a] via-[#2e2e2ea6] to-[#2e2e2ea6] text-[#ffffff] font-semibold border border-[rgba(255,255,255,0.2)] rounded-lg text-white hover:bg-zinc-700 transition-colors duration-300 px-4" style="width: 100%%; margin-top: 2rem;">Equip</button>
			</div>
		</div>
	</div>`

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
