package invite

import (
	"Snorlax/VRChatAPI"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type SentNotification struct {
	Id             string `json:"id"`
	ReceiverUserId string `json:"receiverUserId"`
	SenderUserId   string `json:"senderUserId"`
	Type           string `json:"type"`
	Message        string `json:"message"`
	Details        struct {
	} `json:"details"`
	CreatedAt time.Time `json:"created_at"`
}

func InviteMyselfToInstance(c *VRChatAPI.Client, worldID, instanceID string) (*SentNotification, error) {
	req, err := c.NewRequest("POST", VRChatAPI.API_ENDPOINT+"/invite/myself/to/"+worldID+":"+instanceID, nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("InviteMyselfToInstance - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("InviteMyselfToInstance - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var sentNotificationV SentNotification
	if err := json.Unmarshal(bytes, &sentNotificationV); err != nil {
		err = fmt.Errorf("InviteMyselfToInstance - json.Unmarshal: %v", err)
		return nil, err
	}
	return &sentNotificationV, nil
}
