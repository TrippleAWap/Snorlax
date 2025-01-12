package endpoints

import (
	"Snorlax/RejectDatabase"
	"Snorlax/VRChatAPI"
	"Snorlax/VRChatAPI/auth"
	"Snorlax/VRChatAPI/avatars"
	"Snorlax/cache"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"
)

var (
	CacheV         = cache.New("./cache.db")
	cachePath      string
	GlobalClient   VRChatAPI.Client
	maxThreadCount = 20
	minBatchSize   = 50
)
var cachedIdToAvatar = map[string]string{}
var cachedAvatarIdToAvatar = map[string]avatars.Avatar{}
var GlobalUser *auth.LoginResponse
var cacheIdToCustomModTime = map[string]time.Time{}
var CachedIdToFavorites = map[string]bool{}

func init() {
	appData := os.Getenv("APPDATA")
	cachePath = appData + "\\..\\LocalLow\\VRChat\\VRChat\\Cache-WindowsPlayer\\"
	err := CacheV.Load()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println("Cache loaded")

	fmt.Println(reflect.TypeOf(CacheV.Get("cachedIdToAvatar")))
	if cachedIdToAvatarV := CacheV.Get("cachedIdToAvatar"); cachedIdToAvatarV != nil {
		bytes, err := json.Marshal(cachedIdToAvatarV.(map[string]interface{}))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = json.Unmarshal(bytes, &cachedIdToAvatar); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("CachedIdToAvatar loaded from cache with %d entries\n", len(cachedIdToAvatar))
	}

	if cachedAvatarIdToAvatarV := CacheV.Get("cachedAvatarIdToAvatar"); cachedAvatarIdToAvatarV != nil {
		bytes, err := json.Marshal(cachedAvatarIdToAvatarV.(map[string]interface{}))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = json.Unmarshal(bytes, &cachedAvatarIdToAvatar); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("CachedAvatarIdToAvatar loaded from cache with %d entries\n", len(cachedAvatarIdToAvatar))
	}

	if cacheIdToCustomModTimeV := CacheV.Get("cacheIdToCustomModTime"); cacheIdToCustomModTimeV != nil {
		bytes, err := json.Marshal(cacheIdToCustomModTimeV.(map[string]interface{}))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = json.Unmarshal(bytes, &cacheIdToCustomModTime); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("CacheIdToCustomModTime loaded from cache with %d entries\n", len(cacheIdToCustomModTime))
	}

	if cachedIdToFavoritesV := CacheV.Get("CachedIdToFavorites"); cachedIdToFavoritesV != nil {
		bytes, err := json.Marshal(cachedIdToFavoritesV.(map[string]interface{}))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if err = json.Unmarshal(bytes, &CachedIdToFavorites); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("CachedIdToFavorites loaded from cache with %d entries\n", len(CachedIdToFavorites))
	}
}

var avatarIdRegex = regexp.MustCompile("avtr_[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}")

var ScrapeIdsFromCacheMutex = sync.Mutex{}
var cachedIdToAvatarMutex = sync.Mutex{}

func ScrapeIdsFromCache() (map[string]string, error) {
	ScrapeIdsFromCacheMutex.Lock()
	defer ScrapeIdsFromCacheMutex.Unlock()

	var targetPaths []string
	if err := filepath.WalkDir(cachePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || d.Name() != "__data" {
			return nil
		}

		for cacheId := range cachedIdToAvatar {
			if strings.Contains(path, cacheId) {
				return nil
			}
		}

		targetPaths = append(targetPaths, path)
		return nil
	}); err != nil {
		return nil, err
	}
	wg := sync.WaitGroup{}
	batchSize := max(int(math.Ceil(float64(len(targetPaths))/float64(maxThreadCount))), minBatchSize)

	println(batchSize, len(targetPaths), len(targetPaths)/batchSize, "threads will be spawned")

	for i := 0; i < len(targetPaths); i += batchSize {
		wg.Add(1)
		go func(i int, targetPaths []string) {
			defer wg.Done()

			localCacheIdToAvatar := make(map[string]string)
			for _, path := range targetPaths {
				fileBytes, err := os.ReadFile(path)
				if err != nil {
					fmt.Println(err)
					continue
				}

				avatarId := string(avatarIdRegex.Find(fileBytes))
				cacheId := strings.Split(path, "\\")[8]

				localCacheIdToAvatar[cacheId] = avatarId

				time.Sleep(time.Millisecond * 5)

				fileBytes = nil
			}

			fmt.Printf("Thread %d finished scraping %d files\n", i/batchSize, len(localCacheIdToAvatar))

			cachedIdToAvatarMutex.Lock()
			defer cachedIdToAvatarMutex.Unlock()

			maps.Copy(cachedIdToAvatar, localCacheIdToAvatar)
			time.Sleep(time.Millisecond * 1)
			localCacheIdToAvatar = nil
		}(i, targetPaths[i:min(i+batchSize, len(targetPaths))])
	}
	i := 0
	go func() {
		database, err := RejectDatabase.GetCachedAvatars()
		if err != nil {
			return
		}
		i = 0
		database = nil
		for _, entry := range database {
			if avatarId, ok := cachedIdToAvatar[entry.DataFilePath]; ok {
				if avatarId == entry.Id {
					continue
				}
				entry.DataFilePath = uuid.NewString()
			}

			modTime, err := time.Parse("2006-01-02 15:04:05", entry.Timestamp)
			if err != nil {
				modTime, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", entry.Timestamp)
			}
			if err != nil {
				modTime = time.Now()
			}
			if modTime.After(time.Now()) {
				modTime = time.Now()
			}
			cacheIdToCustomModTime[entry.DataFilePath] = modTime
			cachedIdToAvatar[entry.DataFilePath] = entry.Id
			i++
		}
	}()
	fmt.Printf("ScrapeIdsFromCache - %d entries added from database\n", i)
	wg.Wait()
	fmt.Printf("ScrapeIdsFromCache - %d entries added from cache\n", len(cachedIdToAvatar)-i)
	if err := CacheV.Set("cachedIdToAvatar", cachedIdToAvatar); err != nil {
		panic(err)
	}
	if err := CacheV.Set("cacheIdToCustomModTime", cacheIdToCustomModTime); err != nil {
		panic(err)
	}
	fmt.Printf("ScrapeIdsFromCache - wrote data to cache\n")
	return cachedIdToAvatar, nil
}

var AvatarsMutex = sync.Mutex{}

func GetAvatars(ids []string) (map[string]avatars.Avatar, error) {
	AvatarsMutex.Lock()
	defer AvatarsMutex.Unlock()
	result := make(map[string]avatars.Avatar)
	resultMutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	targetIds := make([]string, 0)
	for _, id := range ids {
		if avatar, ok := cachedAvatarIdToAvatar[id]; ok {
			result[id] = avatar
			continue
		}

		targetIds = append(targetIds, id)
	}
	batchSize := max(int(math.Ceil(float64(len(targetIds))/float64(maxThreadCount))), minBatchSize)

	fmt.Printf("targetIds - %d new - %d cached \n", len(targetIds), len(result))
	fmt.Printf("spawning %d threads with batch size %d\n", len(targetIds)/batchSize, batchSize)
	for i := 0; i < len(targetIds); i += batchSize {
		wg.Add(1)
		go func(i int, ids []string, cachedIdToAvatar map[string]string) {
			defer wg.Done()

			localCachedAvatarIdToAvatar := make(map[string]avatars.Avatar)
			for _, id := range ids {
				var cacheId string
				for cacheIdV, avatarId := range cachedIdToAvatar {
					if avatarId != id {
						continue
					}
					cacheId = cacheIdV
				}

				if len(cacheId) == 0 {
					continue
				}
				avatar, err := GetAvatarFromId(id)
				if err != nil {
					continue
				}
				avatar.CacheId = cacheId
				if _, ok := cacheIdToCustomModTime[cacheId]; ok {
					avatar.CacheTime = cacheIdToCustomModTime[cacheId]
				} else {
					stat, err := os.Stat(cachePath + cacheId)
					if err != nil {
						fmt.Printf("Error getting cache stat for %s: %s\n", cacheId, err)
						continue
					}
					avatar.CacheTime = stat.ModTime()
				}

				localCachedAvatarIdToAvatar[id] = *avatar
			}
			resultMutex.Lock()
			maps.Copy(result, localCachedAvatarIdToAvatar)
			resultMutex.Unlock()
			time.Sleep(time.Millisecond * 1)
			// this should be garbage collected anyway, but without this it leaks memory
			localCachedAvatarIdToAvatar = nil
		}(i, targetIds[i:min(i+batchSize, len(targetIds))], cachedIdToAvatar)
	}
	wg.Wait()
	maps.Copy(cachedAvatarIdToAvatar, result)
	if err := CacheV.Set("cachedAvatarIdToAvatar", cachedAvatarIdToAvatar); err != nil {
		fmt.Printf("Error saving cache: %s\n", err)
		return nil, err
	}
	return result, nil
}

func GetAvatarFromId(id string) (avatar *avatars.Avatar, err error) {
	return avatars.GetAvatar(&GlobalClient, id)
}
