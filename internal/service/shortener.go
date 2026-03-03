package service

import "math/rand/v2"

type Store interface {
	Save(shortID, originalURL string) error
	Get(shortID string) (string, bool)
}

type ShortenerService struct {
	store Store
}

func NewShortenerService(store Store) *ShortenerService {
	return &ShortenerService{store: store}
}

func (s *ShortenerService) Shorten(originalURL string) (string, error) {
	id := GenerateID()
	if err := s.store.Save(id, originalURL); err != nil {
		return "", err
	}
	return id, nil
}

func (s *ShortenerService) Expand(id string) (string, bool) {
	return s.store.Get(id)
}

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateID() string {
	id := make([]byte, 6)
	for i := range id {
		id[i] = chars[rand.IntN(len(chars))]
	}
	return string(id)
}
