// go: generate goversioninfo -icon = icon.ico
package main

import (
	"Snorlax/VRChatAPI"
	"Snorlax/VRChatAPI/auth"
	"Snorlax/VRChatAPI/avatars"
	"Snorlax/endpoints"
	"Snorlax/utils"
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

var loggedIn bool

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

const VersionURL = "https://github.com/TrippleAWap/Snorlax/releases/latest"

func getLatestVersion() (string, error) {
	req, _ := http.NewRequest("GET", VersionURL, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to get latest version - %s", res.Status)
	}
	breadcrumbs := strings.Split(res.Request.URL.String(), "/")
	return breadcrumbs[len(breadcrumbs)-1], nil
}

var w webview.WebView
var randomPort int

func main() {
	if !slices.Contains(os.Args, "--launch") {
		if _, err := os.Stat("./update.exe"); os.IsNotExist(err) {
			fmt.Println("No update.exe found, downloading...")
			lastestVersion, err := getLatestVersion()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			file, err := getFile("https://github.com/TrippleAWap/Snorlax/releases/download/" + lastestVersion + "/update.exe")
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
	randomPort = rand.Intn(65535-49152) + 4915
	go endpoints.StartServer(randomPort)
	w = webview.New(slices.Contains(os.Args, "--debug"))
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
		"loginVRChat": func(username, password string) interface{} {
			res, err := auth.Login(&endpoints.GlobalClient, username, password)
			if err != nil {
				return err.Error()
			}
			return res
		},
		"loginVRChatWith2FA": func(authCookie, code string) interface{} {
			//println(authCookie, code)
			*endpoints.GlobalClient.SelectedAccount = authCookie
			res, err := auth.TwoFactorAuthEmailOTP(&endpoints.GlobalClient, code)
			if err != nil {
				*endpoints.GlobalClient.SelectedAccount = ""
				return err.Error()
			}
			return res
		},
		"setCookie": func(cookie string) error {
			*endpoints.GlobalClient.SelectedAccount = cookie
			if err := VRChatAPI.WriteConfig(endpoints.GlobalClient.Config); err != nil {
				return err
			}
			dispatchFunc()
			return nil
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

	w.Dispatch(dispatchFunc)

	w.SetTitle("Snorlax » Loading...")
	w.Run()
}

var defaultConfig = VRChatAPI.Configuration{SelectedAccount: 0, Accounts: []string{""}}

func dispatchFunc() {
	defer utils.PanicHandler()
	configV, err := VRChatAPI.ReadConfig()
	if err != nil {
		if err := VRChatAPI.WriteConfig(&defaultConfig); err != nil {
			panic(err)
		}
		configV = &defaultConfig
	}
	endpoints.GlobalClient = VRChatAPI.Client{
		Config:          configV,
		SelectedAccount: &configV.Accounts[configV.SelectedAccount],
		Client:          http.DefaultClient,
	}
	user, err := auth.User(&endpoints.GlobalClient)
	if err != nil {
		w.SetTitle("Snorlax » Login")
		w.Navigate("http://127.0.0.1:" + strconv.Itoa(randomPort) + "/login")
		return
	}
	for user == nil {
		user, err = auth.User(&endpoints.GlobalClient)
		if err != nil {
			fmt.Println(err)
			time.Sleep(time.Second)
		}
	}
	endpoints.GlobalUser = user
	fmt.Printf("Logged in as %s\n", user.DisplayName)
	w.SetTitle("Snorlax » Home")
	w.Navigate("http://127.0.0.1:" + strconv.Itoa(randomPort) + "/home")
}
