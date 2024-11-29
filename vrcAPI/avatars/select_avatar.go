package avatars

import (
	"Snorlax/vrcAPI"
	"encoding/json"
	"fmt"
	"io"
)

func SelectAvatar(c *vrcAPI.Client, avatarId string) (*Avatar, error) {
	req, err := c.NewRequest("PUT", vrcAPI.API_ENDPOINT+"/avatars/"+avatarId+"/select", nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("SelectAvatar - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("SelectAvatar - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var avatar Avatar
	if err := json.Unmarshal(bytes, &avatar); err != nil {
		err = fmt.Errorf("SelectAvatar - json.Unmarshal: %v", err)
		return nil, err
	}
	return &avatar, nil
}
