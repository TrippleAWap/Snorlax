package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"unsafe"
)

var user32 = syscall.NewLazyDLL("user32.dll")
var MessageBox = user32.NewProc("MessageBoxW")

func MessageBoxW(title, message string) (int, error) {
	ret, _, err := MessageBox.Call(
		0,
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(message))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))),
		0x00000010,
	)
	if err != nil && err.Error() != "The operation completed successfully." {
		return 0, err
	}
	return int(ret), nil
}

func main() {
	if len(os.Args) < 2 {
		os.Args = append(os.Args, "./Snorlax.exe")
		fmt.Println("No executable provided, using default: Snorlax.exe")
	}

	fmt.Println(os.Args[1])
	defer func() {
		if r := recover(); r != nil {
			stackBuf := make([]byte, 2048)
			stackSize := runtime.Stack(stackBuf, true)

			message := fmt.Sprintf("Panic Occured: %v\n\n%s", r, stackBuf[:stackSize])
			title := "Whoops, something went wrong!"

			if _, err := MessageBoxW(title, message); err != nil {
				panic(err)
			}
		}
	}()
	version, err := getLatestVersion()
	if err != nil {
		panic(err)
	}

	currentVersionHash, err := getCurrentHash()
	if err != nil {
		panic(err)
	}

	bytesV, err := getLatestVersionBytes(version)
	if err != nil {
		panic(err)
	}

	latestVersionHash, err := hashBytes(bytesV)
	if err != nil {
		panic(err)
	}
	procAttr := new(os.ProcAttr)
	procAttr.Files = []*os.File{os.Stdout, os.Stdout, os.Stdout}
	procAttr.Env = os.Environ()
	procAttr.Dir = filepath.Dir(os.Args[1])

	if latestVersionHash == currentVersionHash {
		if _, err := os.StartProcess(os.Args[1], append(os.Args[1:], "--launch"), procAttr); err != nil {
			panic(err)
		}
		os.Exit(0)
	}

	fmt.Println("Installing latest version...")
	if err := os.WriteFile(os.Args[1], bytesV, 0777); err != nil {
		panic(err)
	}

	_, err = os.StartProcess(os.Args[1], []string{"--launch"}, procAttr)
	if err != nil {
		panic(err)
	}
	os.Exit(0)
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

func hashBytes(bytesV []byte) (string, error) {
	hash := sha256.New()

	if _, err := io.Copy(hash, bytes.NewReader(bytesV)); err != nil {
		return "", err
	}

	hashBytes := hash.Sum(nil)

	return fmt.Sprintf("%x", hashBytes), nil
}

func getCurrentHash() (string, error) {
	file, err := os.OpenFile(os.Args[1], os.O_RDONLY, 0777)
	if err != nil {
		if os.IsNotExist(err) {
			return "none", nil
		}
		return "", err
	}
	bytesV, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	fileHash, err := hashBytes(bytesV)
	if err != nil {
		return "", err
	}
	return fileHash, nil
}

func getLatestVersionBytes(version string) ([]byte, error) {
	url := fmt.Sprintf("https://github.com/TrippleAWap/SnorlaxReleases/releases/download/%s/Snorlax.exe", version)
	req, _ := http.NewRequest("GET", url, nil)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("failed to download latest version - %s", res.Status)
	}
	bytesV, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return bytesV, nil
}
