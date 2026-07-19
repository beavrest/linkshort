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
	path   string
	fileMu sync.Mutex
}

func NewFileStorage(path string) (*FileStorage, error) {
	if path == "" {
		return nil, errors.New("file storage path is empty")
	}

	fs := &FileStorage{
		Memory: NewMemory(),
		path:   path,
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

	return fs.appendRecords([]fileRecord{
		{UUID: fs.Memory.UUID(shortID), ShortURL: shortID, OriginalURL: originalURL},
	})
}

func (fs *FileStorage) SaveBatch(items map[string]string) error {
	if err := fs.Memory.SaveBatch(items); err != nil {
		return err
	}

	recs := make([]fileRecord, 0, len(items))
	for shortID, originalURL := range items {
		recs = append(recs, fileRecord{
			UUID:        fs.Memory.UUID(shortID),
			ShortURL:    shortID,
			OriginalURL: originalURL,
		})
	}
	return fs.appendRecords(recs)
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

		if numericID, err := strconv.ParseInt(r.UUID, 10, 64); err == nil {
			fs.Memory.SetUUID(r.ShortURL, r.UUID, numericID)
		}
	}

	return nil
}

func (fs *FileStorage) appendRecords(recs []fileRecord) error {
	fs.fileMu.Lock()
	defer fs.fileMu.Unlock()

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

	enc := json.NewEncoder(f)
	for _, rec := range recs {
		if err := enc.Encode(rec); err != nil {
			return err
		}
	}
	return nil
}
