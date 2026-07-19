package storage

import (
	"strconv"
	"sync"

	"github.com/beavrest/linkshort/internal/service"
)

type Memory struct {
	data     map[string]string
	uuidByID map[string]string
	nextID   int64
	mu       sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		data:     make(map[string]string),
		uuidByID: make(map[string]string),
		nextID:   1,
	}
}

func (m *Memory) Save(shortID, originalURL string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for existingID, u := range m.data {
		if u == originalURL {
			return existingID, service.ErrURLExists
		}
	}
	m.data[shortID] = originalURL
	return shortID, nil
}

func (m *Memory) SaveBatch(items map[string]string) (map[string]string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	result := make(map[string]string, len(items))
	for shortID, originalURL := range items {
		existingID := ""
		for id, u := range m.data {
			if u == originalURL {
				existingID = id
				break
			}
		}
		if existingID != "" {
			result[originalURL] = existingID
			continue
		}
		m.data[shortID] = originalURL
		result[originalURL] = shortID
	}
	return result, nil
}

func (m *Memory) Get(shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[shortID]
	return url, ok
}

func (m *Memory) UUID(shortID string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	if uuid, ok := m.uuidByID[shortID]; ok {
		return uuid
	}
	uuid := strconv.FormatInt(m.nextID, 10)
	m.uuidByID[shortID] = uuid
	m.nextID++
	return uuid
}

func (m *Memory) SetUUID(shortID, uuid string, numericID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.uuidByID[shortID] = uuid
	if numericID+1 > m.nextID {
		m.nextID = numericID + 1
	}
}
