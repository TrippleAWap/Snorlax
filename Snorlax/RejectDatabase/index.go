package RejectDatabase

import (
	"Snorlax/VRChatAPI/avatars"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	ApiEndpoint = "https://0119f0b05a10.ngrok.app"
)

type CachedAvatar struct {
	DataFilePath string `json:"data_file_path"`
	Id           string `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Timestamp    string `json:"timestamp"`
	Username     string `json:"username"`
}

var cachedDatabase []CachedAvatar

func GetCachedAvatars() []CachedAvatar {
	return cachedDatabase
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
	req, err := http.NewRequest("POST", ApiEndpoint+"/avatars", bytes.NewReader(data))
	if err != nil {
		return err
	}
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
