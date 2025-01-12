package auth

import (
	"Snorlax/VRChatAPI"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

type LoginResponse struct {
	SetCookie              string `json:"set-cookie,omitempty"`
	AcceptedTOSVersion     int    `json:"acceptedTOSVersion,omitempty"`
	AcceptedPrivacyVersion int    `json:"acceptedPrivacyVersion,omitempty"`
	AccountDeletionDate    string `json:"accountDeletionDate,omitempty"`
	AccountDeletionLog     []struct {
		Message           string    `json:"message"`
		DeletionScheduled time.Time `json:"deletionScheduled"`
		DateTime          time.Time `json:"dateTime"`
	} `json:"accountDeletionLog,omitempty"`
	ActiveFriends      []string `json:"activeFriends,omitempty"`
	AllowAvatarCopying bool     `json:"allowAvatarCopying,omitempty"`
	Badges             []struct {
		AssignedAt       time.Time `json:"assignedAt"`
		BadgeDescription string    `json:"badgeDescription"`
		BadgeId          string    `json:"badgeId"`
		BadgeImageUrl    string    `json:"badgeImageUrl"`
		BadgeName        string    `json:"badgeName"`
		Hidden           bool      `json:"hidden"`
		Showcased        bool      `json:"showcased"`
		UpdatedAt        time.Time `json:"updatedAt"`
	} `json:"badges,omitempty"`
	Bio                            string    `json:"bio,omitempty"`
	BioLinks                       []string  `json:"bioLinks,omitempty"`
	CurrentAvatar                  string    `json:"currentAvatar,omitempty"`
	CurrentAvatarAssetUrl          string    `json:"currentAvatarAssetUrl,omitempty"`
	CurrentAvatarImageUrl          string    `json:"currentAvatarImageUrl,omitempty"`
	CurrentAvatarThumbnailImageUrl string    `json:"currentAvatarThumbnailImageUrl,omitempty"`
	CurrentAvatarTags              []string  `json:"currentAvatarTags,omitempty"`
	DateJoined                     string    `json:"date_joined,omitempty"`
	DeveloperType                  string    `json:"developerType,omitempty"`
	DisplayName                    string    `json:"displayName,omitempty"`
	EmailVerified                  bool      `json:"emailVerified,omitempty"`
	FallbackAvatar                 string    `json:"fallbackAvatar,omitempty"`
	FriendKey                      string    `json:"friendKey,omitempty"`
	Friends                        []string  `json:"friends,omitempty"`
	HasBirthday                    bool      `json:"hasBirthday,omitempty"`
	HideContentFilterSettings      bool      `json:"hideContentFilterSettings,omitempty"`
	UserLanguage                   string    `json:"userLanguage,omitempty"`
	UserLanguageCode               string    `json:"userLanguageCode,omitempty"`
	HasEmail                       bool      `json:"hasEmail,omitempty"`
	HasLoggedInFromClient          bool      `json:"hasLoggedInFromClient,omitempty"`
	HasPendingEmail                bool      `json:"hasPendingEmail,omitempty"`
	HomeLocation                   string    `json:"homeLocation,omitempty"`
	Id                             string    `json:"id,omitempty"`
	IsBoopingEnabled               bool      `json:"isBoopingEnabled,omitempty"`
	IsFriend                       bool      `json:"isFriend,omitempty"`
	LastActivity                   time.Time `json:"last_activity,omitempty"`
	LastLogin                      time.Time `json:"last_login,omitempty"`
	LastMobile                     time.Time `json:"last_mobile,omitempty"`
	LastPlatform                   string    `json:"last_platform,omitempty"`
	ObfuscatedEmail                string    `json:"obfuscatedEmail,omitempty"`
	ObfuscatedPendingEmail         string    `json:"obfuscatedPendingEmail,omitempty"`
	OculusId                       string    `json:"oculusId,omitempty"`
	GoogleId                       string    `json:"googleId,omitempty"`
	GoogleDetails                  struct {
	} `json:"googleDetails,omitempty"`
	PicoId           string   `json:"picoId,omitempty"`
	ViveId           string   `json:"viveId,omitempty"`
	OfflineFriends   []string `json:"offlineFriends,omitempty"`
	OnlineFriends    []string `json:"onlineFriends,omitempty"`
	PastDisplayNames []struct {
		DisplayName string    `json:"displayName"`
		UpdatedAt   time.Time `json:"updated_at"`
	} `json:"pastDisplayNames,omitempty"`
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
	} `json:"presence,omitempty"`
	ProfilePicOverride          string   `json:"profilePicOverride,omitempty"`
	ProfilePicOverrideThumbnail string   `json:"profilePicOverrideThumbnail,omitempty"`
	Pronouns                    string   `json:"pronouns,omitempty"`
	QueuedInstance              string   `json:"queuedInstance,omitempty"`
	ReceiveMobileInvitations    bool     `json:"receiveMobileInvitations,omitempty"`
	State                       string   `json:"state,omitempty"`
	Status                      string   `json:"status,omitempty"`
	StatusDescription           string   `json:"statusDescription,omitempty"`
	StatusFirstTime             bool     `json:"statusFirstTime,omitempty"`
	StatusHistory               []string `json:"statusHistory,omitempty"`
	SteamDetails                struct {
	} `json:"steamDetails,omitempty"`
	SteamId                  string    `json:"steamId,omitempty"`
	Tags                     []string  `json:"tags,omitempty"`
	TwoFactorAuthEnabled     bool      `json:"twoFactorAuthEnabled,omitempty"`
	TwoFactorAuthEnabledDate time.Time `json:"twoFactorAuthEnabledDate,omitempty"`
	Unsubscribe              bool      `json:"unsubscribe,omitempty"`
	UpdatedAt                time.Time `json:"updated_at,omitempty"`
	UserIcon                 string    `json:"userIcon,omitempty"`
	Error                    struct {
		Message    string `json:"message"`
		StatusCode int    `json:"status_code"`
	} `json:"error,omitempty"`
}

func User(c *VRChatAPI.Client) (*LoginResponse, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"/auth/user", nil)
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
	var currentUserV LoginResponse
	if err := json.Unmarshal(bytes, &currentUserV); err != nil {
		err = fmt.Errorf("User - json.Unmarshal: %v", err)
		return nil, err
	}
	return &currentUserV, nil
}

func Login(c *VRChatAPI.Client, username, password string) (map[string]interface{}, error) {
	req, err := c.NewRequest("GET", VRChatAPI.API_ENDPOINT+"/auth/user", nil)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(username, password)
	meow, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer meow.Body.Close()
	bytes, err := io.ReadAll(meow.Body)
	if err != nil {
		return nil, err
	}
	loginResponseV := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &loginResponseV); err != nil {
		return nil, err
	}
	loginResponseV["Set-Cookie"] = meow.Header.Get("Set-Cookie")
	return loginResponseV, nil
}

func TwoFactorAuthEmailOTP(c *VRChatAPI.Client, code string) (map[string]interface{}, error) {
	req, err := c.NewRequest("POST", VRChatAPI.API_ENDPOINT+"/auth/twofactorauth/emailotp/verify", strings.NewReader(fmt.Sprintf(`{"code": "%s"}`, code)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	bytes, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	twoFactorAuthResponseV := map[string]interface{}{}
	if err := json.Unmarshal(bytes, &twoFactorAuthResponseV); err != nil {
		return nil, err
	}
	return twoFactorAuthResponseV, nil
}
