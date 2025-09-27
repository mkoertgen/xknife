package pkg

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/michimani/gotwi/resources"
)

type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	TTL       time.Duration `json:"ttl"`
}

type FileCache struct {
	cacheDir string
}

func NewFileCache(cacheDir string) *FileCache {
	os.MkdirAll(cacheDir, 0755)
	return &FileCache{cacheDir: cacheDir}
}

func (c *FileCache) getCacheKey(key string) string {
	hash := md5.Sum([]byte(key))
	return fmt.Sprintf("%x", hash)
}

func (c *FileCache) getCacheFile(key string) string {
	return filepath.Join(c.cacheDir, c.getCacheKey(key)+".json")
}

func (c *FileCache) Get(key string, result interface{}) bool {
	file := c.getCacheFile(key)
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false
	}

	// Check if expired
	if time.Since(entry.Timestamp) > entry.TTL {
		os.Remove(file) // Clean up expired entry
		return false
	}

	// Unmarshal the cached data
	dataBytes, err := json.Marshal(entry.Data)
	if err != nil {
		return false
	}

	return json.Unmarshal(dataBytes, result) == nil
}

func (c *FileCache) Set(key string, data interface{}, ttl time.Duration) error {
	entry := CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		TTL:       ttl,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	file := c.getCacheFile(key)
	return ioutil.WriteFile(file, jsonData, 0644)
}

// UserCache provides specific caching for user data
type UserCache struct {
	cache *FileCache
}

func NewUserCache(cacheDir string) *UserCache {
	return &UserCache{
		cache: NewFileCache(filepath.Join(cacheDir, "users")),
	}
}

func (uc *UserCache) GetUser(username string) (*resources.User, bool) {
	var user resources.User
	if uc.cache.Get("user:"+username, &user) {
		return &user, true
	}
	return nil, false
}

func (uc *UserCache) SetUser(username string, user *resources.User) error {
	// Cache user data for 1 hour
	return uc.cache.Set("user:"+username, user, time.Hour)
}

func (uc *UserCache) GetFollowers(userId string, pageSize int) ([]resources.User, bool) {
	key := fmt.Sprintf("followers:%s:%d", userId, pageSize)
	var followers []resources.User
	if uc.cache.Get(key, &followers) {
		return followers, true
	}
	return nil, false
}

func (uc *UserCache) SetFollowers(userId string, pageSize int, followers []resources.User) error {
	key := fmt.Sprintf("followers:%s:%d", userId, pageSize)
	// Cache followers for 30 minutes (they change frequently)
	return uc.cache.Set(key, followers, 30*time.Minute)
}