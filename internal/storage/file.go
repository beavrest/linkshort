package storage

import (
	"encoding/json"
	"errors"
	"io"
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
	*Memory

	path string

	mu       sync.Mutex
	uuidByID map[string]string
	nextID   int64
}

func NewFileStorage(path string) (*FileStorage, error) {
	if path == "" {
		return nil, errors.New("file storage path is empty")
	}

	fs := &FileStorage{
		Memory:   NewMemory(),
		path:     path,
		uuidByID: make(map[string]string),
		nextID:   1,
	}

	if err := fs.load(); err != nil {
		return nil, err
	}

	return fs, nil
}

func (fs *FileStorage) Save(shortID, originalURL string) error {
	if err := fs.Memory.Save(shortID, originalURL); err != nil {
		return err
	}

	fs.mu.Lock()
	uuid, ok := fs.uuidByID[shortID]
	if !ok {
		uuid = strconv.FormatInt(fs.nextID, 10)
		fs.uuidByID[shortID] = uuid
		fs.nextID++
	}
	rec := fileRecord{
		UUID:        uuid,
		ShortURL:    shortID,
		OriginalURL: originalURL,
	}
	fs.mu.Unlock()

	return fs.appendRecord(rec)
}

func (fs *FileStorage) load() error {
	f, err := os.Open(fs.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	defer f.Close()

	dec := json.NewDecoder(f)

	var maxID int64

	for {
		var r fileRecord
		if err := dec.Decode(&r); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if err := fs.Memory.Save(r.ShortURL, r.OriginalURL); err != nil {
			return err
		}

		fs.mu.Lock()
		fs.uuidByID[r.ShortURL] = r.UUID
		if id, err := strconv.ParseInt(r.UUID, 10, 64); err == nil && id > maxID {
			maxID = id
		}
		fs.mu.Unlock()
	}

	fs.mu.Lock()
	fs.nextID = maxID + 1
	if fs.nextID < 1 {
		fs.nextID = 1
	}
	fs.mu.Unlock()

	return nil
}

func (fs *FileStorage) appendRecord(rec fileRecord) error {
	dir := filepath.Dir(fs.path)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}

	f, err := os.OpenFile(fs.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(rec)
}
