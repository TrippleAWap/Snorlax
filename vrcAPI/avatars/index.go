package avatars

import (
	"encoding/json"
	"fmt"
	"time"
)

type Avatar struct {
	AssetUrl       string `json:"assetUrl"`
	AssetUrlObject struct {
	} `json:"assetUrlObject"`
	AuthorId          string    `json:"authorId"`
	AuthorName        string    `json:"authorName"`
	CreatedAt         time.Time `json:"created_at"`
	Description       string    `json:"description"`
	Featured          bool      `json:"featured"`
	Id                string    `json:"id"`
	ImageUrl          string    `json:"imageUrl"`
	Name              string    `json:"name"`
	ReleaseStatus     string    `json:"releaseStatus"`
	Tags              []string  `json:"tags"`
	ThumbnailImageUrl string    `json:"thumbnailImageUrl"`
	UnityPackageUrl   string    `json:"unityPackageUrl"`
	UnityPackages     []struct {
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
	UpdatedAt time.Time `json:"updated_at"`
	Version   int       `json:"version"`
	CacheId   string    `json:"cacheId,omitempty"`
	CacheTime time.Time `json:"cacheTime,omitempty"`
}

func GetAvatarExample() *Avatar {
	data := `{"authorId":"usr_7545492c-b925-4890-bf53-69d0eb608499","authorName":"Nanners JR","created_at":"2024-10-04T14:32:34.335Z","description":"Lil GnagMonkey","featured":false,"id":"avtr_f7651863-4507-42b4-827f-edc5e94e1bb8","imageUrl":"https://api.vrchat.cloud/api/1/file/file_3f20fc42-3951-4891-9824-78230f66ed6c/1/file","name":"Lil GnagMonkey","releaseStatus":"public","tags":[],"thumbnailImageUrl":"https://api.vrchat.cloud/api/1/image/file_3f20fc42-3951-4891-9824-78230f66ed6c/1/256","unityPackageUrl":"","unityPackageUrlObject":{},"unityPackages":[{"assetVersion":1,"created_at":"2024-10-04T14:32:34.335Z","id":"unp_35db6495-7b5f-40c1-9bc8-c668f2ffd11a","performanceRating":"VeryPoor","platform":"standalonewindows","scanStatus":"passed","unityVersion":"2019.4.31f1","variant":"security"},{"assetVersion":1,"created_at":"2024-10-15T14:45:52.763Z","id":"unp_826178fc-8131-480d-89fe-105bd5e67c25","impostorizerVersion":"1.1.1","platform":"standalonewindows","unityVersion":"2022.3.22f1","variant":"impostor"},{"assetVersion":1,"created_at":"2024-10-15T14:45:53.221Z","id":"unp_8090213a-a635-4cb6-8b79-4a164e280f03","impostorizerVersion":"1.1.1","platform":"android","unityVersion":"2022.3.22f1","variant":"impostor"},{"assetVersion":1,"created_at":"2024-10-15T14:45:54.009Z","id":"unp_8e2f3954-e56c-48c3-b043-922ee36d44c4","impostorizerVersion":"1.1.1","platform":"ios","unityVersion":"2022.3.22f1","variant":"impostor"}],"updated_at":"2024-10-06T13:31:25.921Z","version":10}`
	avatar := &Avatar{}
	err := json.Unmarshal([]byte(data), avatar)
	if err != nil {
		fmt.Println(err)
	}
	return avatar
}
