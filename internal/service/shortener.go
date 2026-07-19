package service

import (
	"errors"
	"math/rand/v2"
)

var ErrURLExists = errors.New("original url already exists")

type Store interface {
	Save(shortID, originalURL string) (string, error)
	Get(shortID string) (string, bool)
	SaveBatch(items map[string]string) error
}

type ShortenerService struct {
	store Store
}

func NewShortenerService(store Store) *ShortenerService {
	return &ShortenerService{store: store}
}

func (s *ShortenerService) Shorten(originalURL string) (string, error) {
	id := Generate()
	shortID, err := s.store.Save(id, originalURL)
	if err != nil && !errors.Is(err, ErrURLExists) {
		return "", err
	}
	return shortID, err
}

func (s *ShortenerService) Expand(id string) (string, bool) {
	return s.store.Get(id)
}

func (s *ShortenerService) ShortenBatch(originalURLs []string) ([]string, error) {
	shortIDs := make([]string, len(originalURLs))
	items := make(map[string]string, len(originalURLs))
	for i, u := range originalURLs {
		id := Generate()
		shortIDs[i] = id
		items[id] = u
	}
	if err := s.store.SaveBatch(items); err != nil {
		return nil, err
	}
	return shortIDs, nil
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func Generate() string {
	id := make([]byte, 6)
	for i := range id {
		id[i] = chars[rand.IntN(len(chars))]
	}
	return string(id)
}
