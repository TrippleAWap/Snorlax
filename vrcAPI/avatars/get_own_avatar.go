package avatars

import (
	"Snorlax/vrcAPI"
	"encoding/json"
	"fmt"
	"io"
)

func GetOwnAvatar(c *vrcAPI.Client, userId string) (*Avatar, error) {
	req, err := c.NewRequest("GET", vrcAPI.API_ENDPOINT+"/users/"+userId+"/avatar", nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("GetOwnAvatar - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("GetOwnAvatar - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var userAvatar Avatar
	if err := json.Unmarshal(bytes, &userAvatar); err != nil {
		err = fmt.Errorf("GetOwnAvatar - json.Unmarshal: %v", err)
		return nil, err
	}
	return &userAvatar, nil
}
