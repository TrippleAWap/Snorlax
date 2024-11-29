package endpoints

import (
	"Snorlax/vrcAPI/avatars"
	"encoding/json"
	"fmt"
	"slices"
	"strings"
)

// const CreateCard = async (jsonData) => {
// if (!jsonData.name) {
// console.error('Avatar name is missing', JSON.stringify(jsonData))
// return;
// }
// const platformToDisplay = {
// 'standalonewindows': "Windows",
// 'android': "Android",
// 'ios': "iOS",
// }
// const supportedPlatforms = ['standalonewindows', 'android', 'ios'].filter(p => jsonData.unityPackages?.some(up => up.platform === p))
// const searchBox = document.querySelector('input[type="text"]');
// jsonData.cacheTime = new Date(jsonData.cacheTime);
// // <img
// // 		src="api/avatars/thumbnail?id=${jsonData.id}"
// // 		alt="Avatar Thumbnail"
// // 		className="w-full h-full object-cover"
// // />
// let card = `<div class="w-full sm:w-1/3 md:w-1/5 lg:w-1/7 px-4 mb-8" style="visibility: hidden;">
//
//	       <div class="bg-gray-800 rounded-lg overflow-hidden shadow-lg transition-transform duration-300 hover:scale-105"">
//	           <div class="aspect-[4/3] relative">
//					o~o
//	           </div>
//	           <div class="p-4">
//	               <a>
//	                   <a class="text-lg text-white mb-2 font-bold">${jsonData.name}</a>
//	                   <div class="flex items-center mb-0">
//	                       <div class="w-2 h-2 bg-gray-600 rounded-full mr-2"></div>
//	                       <h2 class="text-base text-gray-400 mb-1">${jsonData.authorName}</h2>
//	                   </div>
//	                   <div class="flex items-center mb-2">
//	                       <div class="w-2 h-2 bg-gray-600 rounded-full mr-2"></div>
//	                       <h2 class="text-sm text-gray-400 mb-1 ">${supportedPlatforms.map(p => platformToDisplay[p]).join(" | ")}</h2>
//	                   </div>
//
// <!--                    bar svg -->
//
//	                   <div class="flex items-center mb-3">
//	                       <div class="w-full bg-gray-600 rounded-full">
//	                           <div class="bg-green-500 h-2 rounded-full" style="width: 0"></div>
//	                       </div>
//	                       <div class="ml-2 text-xs text-gray-400"></div>
//	                   </div>
//	               </a>
//
//					<p class="text-sm text-gray-300 mb-2">${jsonData.cacheTime.toLocaleString('en-US', {
//						hour: 'numeric',
//						minute: 'numeric',
//						hour12: true
//					})} - ${jsonData.cacheTime.toLocaleString('en-US', {
//						month: 'long',
//						day: 'numeric',
//						year: 'numeric'
//					})}</p>
//
//	               <p class="text-gray-400 text-base">${jsonData.description || 'No description'}</p>
//
//	               <button onclick="fetch('/api/avatars/equip?id=${jsonData.id}')" class="bg-gray-600 rounded-lg text-white hover:bg-gray-700 transition-colors duration-300 px-4" style="width: 100%; margin-top: 2rem;">Equip</button>
//	           </div>
//	       </div>
//	   </div>`
//
// avatarCount++;
// document.querySelector('#avatar-count').innerHTML = avatarCount.toString();
// if (searchBox.value !== ” && !card.includes(searchBox.value)) {
// card = card.replace("<div class=\"w-full sm:w-1/2 md:w-1/3 lg:w-1/5 px-4 mb-8\">", "<div class=\"w-full sm:w-1/2 md:w-1/3 lg:w-1/5 px-4 mb-8 hidden\">")
// } else {
// avatarVisibleCount++;
// document.querySelector('#avatar-visible-count').innerHTML = avatarVisibleCount.toString();
// }
//
// document.querySelector('#cards').insertAdjacentHTML('afterbegin', card);
// }
var cardTemplate = `
<div class="w-full sm:w-1/3 md:w-1/5 lg:w-1/7 px-4 mb-8" style="visibility: %s; display: %s;">
        <div class="bg-gray-800 rounded-lg overflow-hidden shadow-lg transition-transform duration-300 hover:scale-105"">
            <div class="aspect-[4/3] relative">
				o~o
            </div>
            <div class="p-4">
                <a>
                    <a class="text-lg text-white mb-2 font-bold">%s</a>
                    <div class="flex items-center mb-0">
                        <div class="w-2 h-2 bg-gray-600 rounded-full mr-2"></div>
                        <h2 class="text-base text-gray-400 mb-1">%s</h2>
                    </div>
                    <div class="flex items-center mb-2">
                        <div class="w-2 h-2 bg-gray-600 rounded-full mr-2"></div>
                        <h2 class="text-sm text-gray-400 mb-1 ">%s</h2>
                    </div>
<!--                    bar svg -->
                    <div class="flex items-center mb-3">
                        <div class="w-full bg-gray-600 rounded-full">
                            <div class="bg-green-500 h-2 rounded-full" style="width: 0"></div>
                        </div>
                        <div class="ml-2 text-xs text-gray-400"></div>
                    </div>
                </a>

				<p class="text-sm text-gray-300 mb-2">%s</p>

                <p class="text-gray-400 text-base">%s</p>

                <button onclick="fetch('/api/avatars/equip?id=%s')" class="bg-gray-600 rounded-lg text-white hover:bg-gray-700 transition-colors duration-300 px-4" style="width: 100%%; margin-top: 2rem;">Equip</button>
            </div>
        </div>
    </div>`

var supportedPlatformToDisplay = map[string]string{
	"standalonewindows": "Windows",
	"android":           "Android",
	"ios":               "iOS",
}

func RenderAvatar(avatar avatars.Avatar, filter string) (bool, string, error) {
	jsonBytes, err := json.Marshal(avatar)
	if err != nil {
		return false, "", err
	}
	visible := strings.Contains(string(jsonBytes), filter)
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
	var visibility string
	var display string
	if visible {
		visibility = "visible"
		display = "block"
	} else {
		visibility = "hidden"
		display = "none"
	}

	cacheTime := avatar.CacheTime.Format("2006-01-02 15:04:05")
	card := fmt.Sprintf(cardTemplate,
		visibility,
		display,
		avatar.Name,
		avatar.AuthorName,
		strings.Join(supportedPlatforms, " | "),
		cacheTime,
		avatar.Description,
		avatar.Id,
	)
	return visible, card, nil
}

var Filter = ""

func RenderAvatars(avatars []avatars.Avatar, filter string) (string, int) {
	fmt.Println("Rendering avatars...")
	var innerHTML string
	visibleAvatars := 0
	for _, avatar := range avatars {
		visible, html, err := RenderAvatar(avatar, filter)
		if err != nil {
			fmt.Println(err)
			continue
		}
		if visible {
			visibleAvatars++
		}
		innerHTML += html
	}
	return innerHTML, visibleAvatars
}
