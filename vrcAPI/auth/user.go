package auth

import (
	"Snorlax/vrcAPI"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type CurrentUser struct {
	AcceptedTOSVersion     int    `json:"acceptedTOSVersion"`
	AcceptedPrivacyVersion int    `json:"acceptedPrivacyVersion"`
	AccountDeletionDate    string `json:"accountDeletionDate"`
	AccountDeletionLog     []struct {
		Message           string    `json:"message"`
		DeletionScheduled time.Time `json:"deletionScheduled"`
		DateTime          time.Time `json:"dateTime"`
	} `json:"accountDeletionLog"`
	ActiveFriends      []string `json:"activeFriends"`
	AllowAvatarCopying bool     `json:"allowAvatarCopying"`
	Badges             []struct {
		AssignedAt       time.Time `json:"assignedAt"`
		BadgeDescription string    `json:"badgeDescription"`
		BadgeId          string    `json:"badgeId"`
		BadgeImageUrl    string    `json:"badgeImageUrl"`
		BadgeName        string    `json:"badgeName"`
		Hidden           bool      `json:"hidden"`
		Showcased        bool      `json:"showcased"`
		UpdatedAt        time.Time `json:"updatedAt"`
	} `json:"badges"`
	Bio                            string    `json:"bio"`
	BioLinks                       []string  `json:"bioLinks"`
	CurrentAvatar                  string    `json:"currentAvatar"`
	CurrentAvatarAssetUrl          string    `json:"currentAvatarAssetUrl"`
	CurrentAvatarImageUrl          string    `json:"currentAvatarImageUrl"`
	CurrentAvatarThumbnailImageUrl string    `json:"currentAvatarThumbnailImageUrl"`
	CurrentAvatarTags              []string  `json:"currentAvatarTags"`
	DateJoined                     string    `json:"date_joined"`
	DeveloperType                  string    `json:"developerType"`
	DisplayName                    string    `json:"displayName"`
	EmailVerified                  bool      `json:"emailVerified"`
	FallbackAvatar                 string    `json:"fallbackAvatar"`
	FriendKey                      string    `json:"friendKey"`
	Friends                        []string  `json:"friends"`
	HasBirthday                    bool      `json:"hasBirthday"`
	HideContentFilterSettings      bool      `json:"hideContentFilterSettings"`
	UserLanguage                   string    `json:"userLanguage"`
	UserLanguageCode               string    `json:"userLanguageCode"`
	HasEmail                       bool      `json:"hasEmail"`
	HasLoggedInFromClient          bool      `json:"hasLoggedInFromClient"`
	HasPendingEmail                bool      `json:"hasPendingEmail"`
	HomeLocation                   string    `json:"homeLocation"`
	Id                             string    `json:"id"`
	IsBoopingEnabled               bool      `json:"isBoopingEnabled"`
	IsFriend                       bool      `json:"isFriend"`
	LastActivity                   time.Time `json:"last_activity"`
	LastLogin                      time.Time `json:"last_login"`
	LastMobile                     time.Time `json:"last_mobile"`
	LastPlatform                   string    `json:"last_platform"`
	ObfuscatedEmail                string    `json:"obfuscatedEmail"`
	ObfuscatedPendingEmail         string    `json:"obfuscatedPendingEmail"`
	OculusId                       string    `json:"oculusId"`
	GoogleId                       string    `json:"googleId"`
	GoogleDetails                  struct {
	} `json:"googleDetails"`
	PicoId           string   `json:"picoId"`
	ViveId           string   `json:"viveId"`
	OfflineFriends   []string `json:"offlineFriends"`
	OnlineFriends    []string `json:"onlineFriends"`
	PastDisplayNames []struct {
		DisplayName string    `json:"displayName"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"pastDisplayNames"`
	Presence struct {
		AvatarThumbnail     string   `json:"avatarThumbnail"`
		CurrentAvatarTags   string   `json:"currentAvatarTags"`
		DisplayName         string   `json:"displayName"`
		Groups              []string `json:"groups"`
		Id                  string   `json:"id"`
		Instance            string   `json:"instance"`
		InstanceType        string   `json:"instanceType"`
		IsRejoining         string   `json:"isRejoining"`
		Platform            string   `json:"platform"`
		ProfilePicOverride  string   `json:"profilePicOverride"`
		Status              string   `json:"status"`
		TravelingToInstance string   `json:"travelingToInstance"`
		TravelingToWorld    string   `json:"travelingToWorld"`
		UserIcon            string   `json:"userIcon"`
		World               string   `json:"world"`
	} `json:"presence"`
	ProfilePicOverride          string   `json:"profilePicOverride"`
	ProfilePicOverrideThumbnail string   `json:"profilePicOverrideThumbnail"`
	Pronouns                    string   `json:"pronouns"`
	QueuedInstance              string   `json:"queuedInstance"`
	ReceiveMobileInvitations    bool     `json:"receiveMobileInvitations"`
	State                       string   `json:"state"`
	Status                      string   `json:"status"`
	StatusDescription           string   `json:"statusDescription"`
	StatusFirstTime             bool     `json:"statusFirstTime"`
	StatusHistory               []string `json:"statusHistory"`
	SteamDetails                struct {
	} `json:"steamDetails"`
	SteamId                  string    `json:"steamId"`
	Tags                     []string  `json:"tags"`
	TwoFactorAuthEnabled     bool      `json:"twoFactorAuthEnabled"`
	TwoFactorAuthEnabledDate time.Time `json:"twoFactorAuthEnabledDate"`
	Unsubscribe              bool      `json:"unsubscribe"`
	UpdatedAt                time.Time `json:"updated_at"`
	UserIcon                 string    `json:"userIcon"`
}

func User(c *vrcAPI.Client) (*CurrentUser, error) {
	req, err := c.NewRequest("GET", vrcAPI.API_ENDPOINT+"/auth/user", nil)
	meow, err := c.DoWDefaults(req)
	if err != nil {
		err = fmt.Errorf("User - %v", err)
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		err = fmt.Errorf("User - io.ReadAll: %v", err)
		return nil, err
	}
	// Unmarshal the response into a slice of World
	var currentUserV CurrentUser
	if err := json.Unmarshal(bytes, &currentUserV); err != nil {
		err = fmt.Errorf("User - json.Unmarshal: %v", err)
		return nil, err
	}
	return &currentUserV, nil
}
