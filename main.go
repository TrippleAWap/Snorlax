// go: generate goversioninfo -icon = icon.ico
package main

import (
	"Snorlax/endpoints"
	"Snorlax/utils"
	"Snorlax/vrcAPI"
	"Snorlax/vrcAPI/auth"
	"Snorlax/vrcAPI/avatars"
	"fmt"
	webview "github.com/webview/webview_go"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

func getFile(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP status code %d", resp.StatusCode)
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

func SpawnProcess(exeRelativePath string, args ...string) error {
	attr := &os.ProcAttr{
		Dir:   os.Getenv("CWD"),
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
	}
	_, err := os.StartProcess(exeRelativePath, args, attr)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	if !slices.Contains(os.Args, "--launch") {
		if _, err := os.Stat("./update.exe"); os.IsNotExist(err) {
			fmt.Println("No update.exe found, downloading...")
			file, err := getFile("https://github.com/TrippleAWap/SnorlaxReleases/blob/main/update.exe?raw=true")
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			err = os.WriteFile("update.exe", file, 0777)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
		}
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		updatePath := path.Join(cwd, "update.exe")
		fmt.Println("Running", updatePath)

		err = SpawnProcess("C:\\Windows\\System32\\cmd.exe", "/c", updatePath, os.Args[0])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	pathV, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(pathV)
	defer utils.PanicHandler()
	//PrismicDatabase.GetDatabase()
	randomPort := rand.Intn(65535-49152) + 4915
	go endpoints.StartServer(randomPort)
	w := webview.New(slices.Contains(os.Args, "--debug"))
	w.SetSize(400, 200, webview.HintMin)
	defer w.Destroy()
	w.Window()

	for k, v := range map[string]interface{}{
		"updateFilter": func(filter string) []interface{} {
			endpoints.Filter = strings.ToLower(filter)
			slices.SortFunc(endpoints.AvatarsV, func(a, b avatars.Avatar) int {
				return int(b.CacheTime.Unix() - a.CacheTime.Unix())
			})

			innerHtml, visibleAvatars := endpoints.RenderAvatars(endpoints.AvatarsV, filter)
			cards := strings.SplitAfter(innerHtml, `            </div>
        </div>
    </div>`)
			return []interface{}{strings.Join(cards[:min(len(cards), 1000)], ""), visibleAvatars, len(endpoints.AvatarsV)}
		},
		"updateFavoritesOnly": func(b bool) {
			endpoints.FavoritesOnly = b
		},
		"favoriteAvatar": func(avatarId string, favorite bool) {
			if favorite {
				endpoints.CachedIdToFavorites[avatarId] = true
			} else {
				delete(endpoints.CachedIdToFavorites, avatarId)
			}
			if err := endpoints.CacheV.Set("CachedIdToFavorites", endpoints.CachedIdToFavorites); err != nil {
				panic(err)
			}
		},
	} {
		if err := w.Bind(k, v); err != nil {
			panic(err)
		}
	}

	w.Dispatch(func() {
		defer utils.PanicHandler()
		configV, err := vrcAPI.ReadConfig()
		if err != nil {
			panic(err)
		}
		endpoints.GlobalClient = vrcAPI.Client{
			Config: configV,
			Client: http.DefaultClient,
		}
		user, err := auth.User(&endpoints.GlobalClient)
		if err != nil {
			w.Navigate("https://vrchat.com/home/login")
			w.Init(``)
		}
		for user == nil {
			time.Sleep(time.Second)
			user, err = auth.User(&endpoints.GlobalClient)
			if err != nil {
				panic(err)
			}
		}
		endpoints.GlobalUser = user
		fmt.Printf("Logged in as %s\n", user.DisplayName)
		w.Navigate("http://127.0.0.1:" + strconv.Itoa(randomPort) + "/home")
	})

	w.SetTitle("Snorlax » github.com/TrippleAWap")
	w.Run()
}
