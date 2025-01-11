package avatars

import (
	"Snorlax/VRChatAPI"
	"encoding/json"
	"fmt"
	"io"
)

type SortOrder string

const (
	SortOrderPopular             SortOrder = "popularity"
	SortOrderHeat                SortOrder = "heat"
	SortOrderTrust               SortOrder = "trust"
	SortOrderShuffle             SortOrder = "shuffle"
	SortOrderFavorites           SortOrder = "favorites"
	SortOrderRandom              SortOrder = "random"
	SortOrderReportScore         SortOrder = "reportScore"
	SortOrderReportCount         SortOrder = "reportCount"
	SortOrderPublicationDate     SortOrder = "publicationDate"
	SortOrderLastPublicationDate SortOrder = "lastPublicationDate"
	SortOrderCreated             SortOrder = "created"
	SortOrderCreatedAt           SortOrder = "createdAt"
	SortOrderUpdated             SortOrder = "updated"
	SortOrderUpdatedAt           SortOrder = "updatedAt"
	SortOrderOrder               SortOrder = "order"
	SortOrderRelevance           SortOrder = "relevance"
	SortOrderMagic               SortOrder = "magic"
	SortOrderName                SortOrder = "name"
)

type SearchAvatarsParams struct {
	// Filters on featured results.
	featured bool
	// The sortOrder of the results.
	sortOrder SortOrder
	// Set to me for searching own avatars.
	user string
	// Filter by userId
	userId string
}

// todo: dont need this as of now, but could be useful later.

// Search and list avatars by query filters. You can only search your own or featured avatars. It is not possible as a normal user to search other peoples avatars.
func SearchAvatars(c *VRChatAPI.Client, params SearchAvatarsParams) (*Avatar, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"", nil)
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
