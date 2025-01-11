package worlds

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"time"
)

type Instance struct {
	Active                     bool   `json:"active"`
	CanRequestInvite           bool   `json:"canRequestInvite"`
	Capacity                   int    `json:"capacity"`
	DisplayName                string `json:"displayName"`
	Full                       bool   `json:"full"`
	GameServerVersion          int    `json:"gameServerVersion"`
	ID                         string `json:"id"`
	InstanceId                 string `json:"instanceId"`
	InstancePersistenceEnabled string `json:"instancePersistenceEnabled"`
	Location                   string `json:"location"`
	NUsers                     int    `json:"n_users"`
	Name                       string `json:"name"`
	OwnerId                    string `json:"ownerId"`
	Permanent                  bool   `json:"permanent"`
	PhotonRegion               string `json:"photonRegion"`
	Platforms                  struct {
		Android           int `json:"android"`
		Ios               int `json:"ios"`
		Standalonewindows int `json:"standalonewindows"`
	} `json:"platforms"`
	PlayerPersistenceEnabled string   `json:"playerPersistenceEnabled"`
	Region                   string   `json:"region"`
	SecureName               string   `json:"secureName"`
	ShortName                string   `json:"shortName"`
	Tags                     []string `json:"tags"`
	Type                     string   `json:"type"`
	WorldId                  string   `json:"worldId"`
	Hidden                   string   `json:"hidden"`
	Friends                  string   `json:"friends"`
	Private                  string   `json:"private"`
	QueueEnabled             bool     `json:"queueEnabled"`
	QueueSize                int      `json:"queueSize"`
	RecommendedCapacity      int      `json:"recommendedCapacity"`
	RoleRestricted           bool     `json:"roleRestricted"`
	Strict                   bool     `json:"strict"`
	UserCount                int      `json:"userCount"`
	World                    struct {
		AuthorId            string          `json:"authorId"`
		AuthorName          string          `json:"authorName"`
		Capacity            int             `json:"capacity"`
		RecommendedCapacity int             `json:"recommendedCapacity"`
		CreatedAt           time.Time       `json:"created_at"`
		Description         string          `json:"description"`
		Favorites           int             `json:"favorites"`
		Featured            bool            `json:"featured"`
		Heat                int             `json:"heat"`
		Id                  string          `json:"id"`
		ImageUrl            string          `json:"imageUrl"`
		Instances           [][]interface{} `json:"instances"`
		LabsPublicationDate string          `json:"labsPublicationDate"`
		Name                string          `json:"name"`
		Namespace           string          `json:"namespace"`
		Occupants           int             `json:"occupants"`
		Organization        string          `json:"organization"`
		Popularity          int             `json:"popularity"`
		PreviewYoutubeId    string          `json:"previewYoutubeId"`
		PrivateOccupants    int             `json:"privateOccupants"`
		PublicOccupants     int             `json:"publicOccupants"`
		PublicationDate     string          `json:"publicationDate"`
		ReleaseStatus       string          `json:"releaseStatus"`
		Tags                []string        `json:"tags"`
		ThumbnailImageUrl   string          `json:"thumbnailImageUrl"`
		UnityPackages       []struct {
			Id             string `json:"id"`
			AssetUrl       string `json:"assetUrl"`
			AssetUrlObject struct {
			} `json:"assetUrlObject"`
			AssetVersion        int       `json:"assetVersion"`
			CreatedAt           time.Time `json:"created_at"`
			ImpostorizerVersion string    `json:"impostorizerVersion"`
			PerformanceRating   string    `json:"performanceRating"`
			Platform            string    `json:"platform"`
			PluginUrl           string    `json:"pluginUrl"`
			PluginUrlObject     struct {
			} `json:"pluginUrlObject"`
			UnitySortNumber int64  `json:"unitySortNumber"`
			UnityVersion    string `json:"unityVersion"`
			WorldSignature  string `json:"worldSignature"`
			ImpostorUrl     string `json:"impostorUrl"`
			ScanStatus      string `json:"scanStatus"`
			Variant         string `json:"variant"`
		} `json:"unityPackages"`
		UpdatedAt    time.Time `json:"updated_at"`
		Version      int       `json:"version"`
		Visits       int       `json:"visits"`
		UdonProducts []string  `json:"udonProducts"`
	} `json:"world"`
	Users []struct {
		Bio                            string    `json:"bio"`
		BioLinks                       []string  `json:"bioLinks"`
		CurrentAvatarImageUrl          string    `json:"currentAvatarImageUrl"`
		CurrentAvatarThumbnailImageUrl string    `json:"currentAvatarThumbnailImageUrl"`
		CurrentAvatarTags              []string  `json:"currentAvatarTags"`
		DeveloperType                  string    `json:"developerType"`
		DisplayName                    string    `json:"displayName"`
		FallbackAvatar                 string    `json:"fallbackAvatar"`
		Id                             string    `json:"id"`
		IsFriend                       bool      `json:"isFriend"`
		LastPlatform                   string    `json:"last_platform"`
		LastLogin                      time.Time `json:"last_login"`
		ProfilePicOverride             string    `json:"profilePicOverride"`
		Pronouns                       string    `json:"pronouns"`
		Status                         string    `json:"status"`
		StatusDescription              string    `json:"statusDescription"`
		Tags                           []string  `json:"tags"`
		UserIcon                       string    `json:"userIcon"`
		Location                       string    `json:"location"`
		FriendKey                      string    `json:"friendKey"`
	} `json:"users"`
	GroupAccessType   string    `json:"groupAccessType"`
	HasCapacityForYou bool      `json:"hasCapacityForYou"`
	Nonce             string    `json:"nonce"`
	ClosedAt          time.Time `json:"closedAt"`
	HardClose         bool      `json:"hardClose"`
}

func GetWorldInstance(c *VRChatAPI.Client, worldID string, instanceID int) (*Instance, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"/worlds/"+worldID+"/"+strconv.Itoa(instanceID), nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("GetWorldInstance - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("GetWorldInstance - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var instanceV Instance
	if err := json.Unmarshal(bytes, &instanceV); err != nil {
		err = fmt.Errorf("GetWorldInstance - json.Unmarshal: %v", err)
		return nil, err
	}
	return &instanceV, nil
}
