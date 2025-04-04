package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"sync"
)

var (
	streamKeys     = make(map[string]string)
	streamKeysLock sync.RWMutex
)

func GenerateStreamKey(userID string) (string, error) {

	b := make([]byte, 16)

	_, err := rand.Read(b)

	if err != nil {
		return "", fmt.Errorf("failed to generate stream key: %v", err)
	}

	streamKey := base64.RawURLEncoding.EncodeToString(b)

	streamKeysLock.Lock()

	defer streamKeysLock.Unlock()

	streamKeys[streamKey] = userID //FIXME - Store keys in db and cache for faster access

	return streamKey, nil

}

func ValidateStreamKey(streamKey string) bool {

	streamKeysLock.RLock()

	defer streamKeysLock.RUnlock()

	_, exists := streamKeys[streamKey]

	return exists

}

func RevokeStreamKey(streamKey string) {

	streamKeysLock.Lock()

	defer streamKeysLock.Unlock()

	delete(streamKeys, streamKey)

}

func GetUserForStreamKey(streamKey string) (string, bool) {

	streamKeysLock.RLock()

	defer streamKeysLock.RUnlock()

	userID, exists := streamKeys[streamKey]

	return userID, exists

}

func GetStreamKeyForUser(userID string) (string, bool) {

	streamKeysLock.RLock()

	defer streamKeysLock.RUnlock()

	for key, id := range streamKeys {

		if id == userID {
			return key, true
		}

	}

	return "", false

}
