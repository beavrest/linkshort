package repository

type Storer interface {
	Save(shortID, originalURL string)
	Get(shortID string) (string, bool)
}
