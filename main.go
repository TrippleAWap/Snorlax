// go: generate goversioninfo -icon = icon.ico
package main

import (
	"Snorlax/endpoints"
	"Snorlax/vrcAPI"
	"Snorlax/vrcAPI/auth"
	"fmt"
	webview "github.com/webview/webview_go"
	"math/rand"
	"net/http"
	"os"
	"runtime"
	"slices"
	"strconv"
	"syscall"
	"unsafe"
)

type SHFILEINFOA struct {
	hIcon         uintptr
	iIcon         int32
	dwAttributes  int16
	szDisplayName [syscall.MAX_PATH]byte
	szTypeName    [80]byte
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			// spawn message box with error message
			user32 := syscall.NewLazyDLL("user32.dll")
			MessageBox := user32.NewProc("MessageBoxW")

			stackBuf := make([]byte, 2048)
			stackSize := runtime.Stack(stackBuf, true)

			message := fmt.Sprintf("Panic Occured: %v\n\n%s", r, stackBuf[:stackSize])
			title := "Whoops, something went wrong!"

			titlePtr, _ := syscall.UTF16PtrFromString(title)
			messagePtr, _ := syscall.UTF16PtrFromString(message)

			_, _, err := MessageBox.Call(
				0,
				uintptr(unsafe.Pointer(messagePtr)),
				uintptr(unsafe.Pointer(titlePtr)),
				0x00000010,
			)
			if err != nil && err.Error() != "The operation completed successfully." {
				panic(err)
			}
		}
	}()
	randomPort := rand.Intn(65535-49152) + 4915
	go endpoints.StartServer(randomPort)
	w := webview.New(slices.Contains(os.Args, "--debug"))
	w.SetSize(400, 200, webview.HintMin)
	defer w.Destroy()
	w.Window()

	for k, v := range map[string]interface{}{
		"updateFilter": func(filter string) []interface{} {
			endpoints.Filter = filter
			innerHtml, visibleAvatars := endpoints.RenderAvatars(endpoints.AvatarsV, filter)
			return []interface{}{innerHtml, visibleAvatars, len(endpoints.AvatarsV)}
		},
	} {
		if err := w.Bind(k, v); err != nil {
			panic(err)
		}
	}

	w.Dispatch(func() {
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
			panic(err)
		}
		endpoints.GlobalUser = user
		fmt.Printf("Logged in as %s\n", user.DisplayName)
		w.Navigate("http://127.0.0.1:" + strconv.Itoa(randomPort) + "/home")
	})

	w.SetTitle("Snorlax » github.com/TrippleAWap » Release v1.0.0")
	w.Run()
}
