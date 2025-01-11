package avatars

import (
	"Snorlax/VRChatAPI"
	"encoding/json"
	"fmt"
	"io"
)

func GetAvatar(c *VRChatAPI.Client, avatarId string) (*Avatar, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"/avatars/"+avatarId, nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("GetAvatar - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("GetAvatar - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var avatar Avatar
	if err := json.Unmarshal(bytes, &avatar); err != nil {
		err = fmt.Errorf("GetAvatar - json.Unmarshal: %v - %s", err, string(bytes))
		return nil, err
	}
	return &avatar, nil
}
