package vrcAPI

import (
	"encoding/json"
	"errors"
	"os"
	"time"
)

type Configuration struct {
	// AuthCookie is the name of the cookie used for authentication.
	AuthCookie    string `json:"auth_cookie"`
	TwoFactorAuth string `json:"twoFactorAuth"`
	// BatchingSize is the maximum number of events per thread.
	BatchingSize int `json:"batching_size"`
}

var (
	lastUpdated time.Time
	lastConfig  *Configuration
	configPath  = "./config.json"
)

func ReadConfig() (*Configuration, error) {
	if lastConfig != nil {
		fstat, err := os.Lstat(configPath)
		if err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		if lastUpdated == fstat.ModTime() {
			return lastConfig, nil
		}
	}
	file, err := os.Open(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			defaultConfig := Configuration{
				AuthCookie:    "auth_cookie",
				TwoFactorAuth: "",
				BatchingSize:  10,
			}

			bytes, err := json.MarshalIndent(defaultConfig, "", "  ")
			if err != nil {
				return nil, err
			}
			if err := os.WriteFile(configPath, bytes, 0644); err != nil {
				return nil, err
			}
			return &defaultConfig, nil
		}
		return nil, err
	}
	var config Configuration
	if err := json.NewDecoder(file).Decode(&config); err != nil {
		return nil, err
	}
	fstat, err := os.Lstat(configPath)
	if err != nil {
		return nil, err
	}
	lastUpdated = fstat.ModTime()
	lastConfig = &config
	return &config, nil
}
