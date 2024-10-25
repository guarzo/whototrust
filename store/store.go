package store

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/gambtho/whototrust/model"
	"sync"
)

var Store *HomeDataStore

// HomeDataStore stores HomeData in memory
type HomeDataStore struct {
	sync.RWMutex
	store map[int64]model.HomeData
	ETag  string
}

// NewHomeDataStore creates a new HomeDataStore
func NewHomeDataStore() *HomeDataStore {
	return &HomeDataStore{
		store: make(map[int64]model.HomeData),
		ETag:  "",
	}
}

func (s *HomeDataStore) Set(id int64, homeData model.HomeData) (string, error) {
	s.Lock()
	defer s.Unlock()
	s.store[id] = homeData

	etag, err := GenerateETag(homeData)
	if err != nil {
		return "", err
	}
	s.ETag = etag
	return etag, nil
}

func (s *HomeDataStore) Get(id int64) (model.HomeData, string, bool) {
	s.RLock()
	defer s.RUnlock()
	homeData, ok := s.store[id]
	return homeData, s.ETag, ok
}

// Delete removes an identity from the store
func (s *HomeDataStore) Delete(id int64) {
	s.Lock()
	defer s.Unlock()
	delete(s.store, id)
}

func GenerateETag(homeData model.HomeData) (string, error) {
	data, err := json.Marshal(homeData)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

func init() {
	Store = NewHomeDataStore()
}
