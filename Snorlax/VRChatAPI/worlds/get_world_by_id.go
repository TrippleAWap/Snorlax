package worlds

import (
	"Snorlax/VRChatAPI"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
)

func GetWorldByID(c *VRChatAPI.Client, worldID string) (*World, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"/worlds/"+worldID, nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("GetWorldById - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("GetWorldById - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var world World
	if err := json.Unmarshal(bytes, &world); err != nil {
		err = fmt.Errorf("GetWorldById - json.Unmarshal: %v", err)
		return nil, err
	}
	return &world, nil
}

type InstanceEntry struct {
	ID          int
	PlayerCount int
}

func ParseInstances(instances [][]any) []InstanceEntry {
	instancesV := make([]InstanceEntry, len(instances))
	for i, instance := range instances {
		id := strings.Split(instance[0].(string), "~")[0]
		meow, err := strconv.Atoi(id)
		if err != nil {
			log.Fatalf("ParseInstances - strconv.Atoi: %v", err)
		}
		instancesV[i] = InstanceEntry{
			ID:          meow,
			PlayerCount: int(instance[1].(float64)),
		}
	}
	return instancesV
}
