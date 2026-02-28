package storage

import "sync"

type Memory struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{data: make(map[string]string)}
}

func (m *Memory) Save(shortID, originalURL string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[shortID] = originalURL
}

func (m *Memory) Get(shortID string) (string, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	url, ok := m.data[shortID]
	return url, ok
}
