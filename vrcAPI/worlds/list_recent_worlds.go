package worlds

import (
	"Snorlax/vrcAPI"
	"encoding/json"
	"fmt"
	"io"
)

func ListRecentWorlds(c *vrcAPI.Client) ([]World, error) {
	req, err := c.NewRequest("GET", vrcAPI.API_ENDPOINT+"/worlds/recent", nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("ListRecentWorlds - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("ListRecentWorlds - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var worlds []World
	if err := json.Unmarshal(bytes, &worlds); err != nil {
		err = fmt.Errorf("ListRecentWorlds - json.Unmarshal: %v", err)
		return nil, err
	}
	return worlds, nil
}
