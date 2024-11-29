package worlds

import "time"

type World struct {
	AuthorId            string    `json:"authorId"`
	AuthorName          string    `json:"authorName"`
	Capacity            int       `json:"capacity"`
	CreatedAt           time.Time `json:"created_at"`
	Description         string    `json:"description"`
	Favorites           int       `json:"favorites"`
	Featured            bool      `json:"featured"`
	Heat                int       `json:"heat"`
	ID                  string    `json:"id"`
	ImageUrl            string    `json:"imageUrl"`
	LabsPublicationDate time.Time `json:"labsPublicationDate"`
	Name                string    `json:"name"`
	Occupants           int       `json:"occupants"`
	Organization        string    `json:"organization"`
	Popularity          int       `json:"popularity"`
	PreviewYoutubeId    *string   `json:"previewYoutubeId"`
	PublicationDate     time.Time `json:"publicationDate"`
	RecommendedCapacity int       `json:"recommendedCapacity"`
	ReleaseStatus       string    `json:"releaseStatus"`
	Tags                []string  `json:"tags"`
	ThumbnailImageUrl   string    `json:"thumbnailImageUrl"`
	UdonProducts        []string  `json:"udonProducts"`
	Instances           [][]any   `json:"instances,omitempty"`
	UnityPackages       []struct {
		AssetUrl       string `json:"assetUrl"`
		AssetUrlObject struct {
		} `json:"assetUrlObject,omitempty"`
		AssetVersion    int       `json:"assetVersion"`
		CreatedAt       time.Time `json:"created_at"`
		Id              string    `json:"id"`
		Platform        string    `json:"platform"`
		PluginUrl       string    `json:"pluginUrl,omitempty"`
		PluginUrlObject struct {
		} `json:"pluginUrlObject,omitempty"`
		UnitySortNumber int64  `json:"unitySortNumber"`
		UnityVersion    string `json:"unityVersion"`
		WorldSignature  string `json:"worldSignature,omitempty"`
	} `json:"unityPackages"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	Visits    int       `json:"visits"`
}
