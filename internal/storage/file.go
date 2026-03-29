package storage

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
)

type fileRecord struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type FileStorage struct {
	path     string
	mu       sync.RWMutex
	data     map[string]string
	uuidByID map[string]string
	nextID   int64
}

func NewFileStorage(path string) (*FileStorage, error) {
	fs := &FileStorage{
		path:     path,
		data:     make(map[string]string),
		uuidByID: make(map[string]string),
		nextID:   1,
	}
	if err := fs.load(); err != nil {
		return nil, err
	}
	return fs, nil
}

func (fs *FileStorage) Save(shortID, originalURL string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	fs.data[shortID] = originalURL

	if _, exists := fs.uuidByID[shortID]; !exists {
		fs.uuidByID[shortID] = strconv.FormatInt(fs.nextID, 10)
		fs.nextID++
	}

	return fs.flushLocked()
}

func (fs *FileStorage) Get(shortID string) (string, bool) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	url, ok := fs.data[shortID]
	return url, ok
}

func (fs *FileStorage) load() error {
	b, err := os.ReadFile(fs.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(b) == 0 {
		return nil
	}

	var recs []fileRecord
	if err := json.Unmarshal(b, &recs); err != nil {
		return err
	}

	var maxID int64
	for _, r := range recs {
		fs.data[r.ShortURL] = r.OriginalURL
		fs.uuidByID[r.ShortURL] = r.UUID

		if id, err := strconv.ParseInt(r.UUID, 10, 64); err == nil && id > maxID {
			maxID = id
		}
	}

	fs.nextID = maxID + 1
	if fs.nextID < 1 {
		fs.nextID = 1
	}
	return nil
}

func (fs *FileStorage) flushLocked() error {
	recs := make([]fileRecord, 0, len(fs.data))

	for shortID, original := range fs.data {
		uuid := fs.uuidByID[shortID]
		recs = append(recs, fileRecord{
			UUID:        uuid,
			ShortURL:    shortID,
			OriginalURL: original,
		})
	}

	b, err := json.MarshalIndent(recs, "", "   ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(fs.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	tmp := fs.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, fs.path)
}
