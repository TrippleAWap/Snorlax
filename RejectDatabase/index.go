package RejectDatabase

import (
	"Snorlax/vrcAPI/avatars"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	API_ENDPOINT = "https://0119f0b05a10.ngrok.app"
)

type CachedAvatar struct {
	DataFilePath string `json:"data_file_path"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	Username     string `json:"username"`
}

func init() {
	go func() {
		for {
			_, err := GetCachedAvatars()
			if err != nil {
				fmt.Println("Failed to get cached avatars:", err)
			}
			time.Sleep(time.Minute*5 + time.Second)
		}
	}()
}

var cachedDatabase []CachedAvatar
var lastCached time.Time

func GetCachedAvatars() ([]CachedAvatar, error) {
	if time.Since(lastCached) < time.Minute*5 {
		return cachedDatabase, nil
	}
	req, err := http.NewRequest("GET", API_ENDPOINT+"/avatars", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bytesV, err := io.ReadAll(res.Body)

	var cachedAvatars []CachedAvatar
	err = json.NewDecoder(bytes.NewReader(bytesV)).Decode(&cachedAvatars)
	if err != nil {
		return nil, fmt.Errorf("GetCachedAvatars: failed to decode response: %w | %s", err, string(bytesV))
	}
	lastCached = time.Now()
	cachedDatabase = cachedAvatars
	return cachedAvatars, nil
}

func AddAvatar(avatar avatars.Avatar, username string) error {
	for _, cachedAvatar := range cachedDatabase {
		if cachedAvatar.Id == avatar.Id {
			return fmt.Errorf("AddAvatar: avatar already exists in cache: %s", cachedAvatar.Id)
		}
	}
	entry := CachedAvatar{
		DataFilePath: avatar.CacheId,
		Id:           avatar.Id,
		Name:         avatar.Name,
		Status:       "active",
		Timestamp:    avatar.CacheTime.String(),
		Username:     username,
	}
	cachedDatabase = append(cachedDatabase, entry)
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", API_ENDPOINT+"/avatars", bytes.NewReader(data))
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusCreated {
		return fmt.Errorf("AddAvatar: failed to add avatar, status: %s", res.Status)
	}
	return nil
}
