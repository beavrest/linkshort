package shortener

import "math/rand/v2"

const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateID() string {
	id := make([]byte, 6)
	for i := range id {
		id[i] = chars[rand.IntN(len(chars))]
	}
	return string(id)
}
