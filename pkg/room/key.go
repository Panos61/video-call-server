package room

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis"
)

func GenerateInvKey(roomID string) string {
	buffer := make([]byte, 12)

	_, err := rand.Read(buffer)
	if err != nil {
		log.Printf("Error generating random bytes: %s\n", err)
		return ""
	}

	invitationKey := base64.URLEncoding.EncodeToString(buffer)
	return invitationKey
}

func (c *WSClient) SetRoomKey(roomID string, invKey string) {
	expirationTime := 20 * time.Second

	err := rdb.HSet(ctx, "room:"+roomID, map[string]interface{}{
		"invitation_key": invKey,
		"expiresIn":      time.Now().Add(expirationTime).Format(time.RFC3339),
	}).Err()

	if err != nil {
		log.Printf("Error setting room invitation key %s\n", err)
		return
	}

	if invKey != "" {
		err := c.Conn.WriteJSON(invKey)
		if err != nil {
			fmt.Printf("Error writing JSON: %s\n", err)
			c.Conn.Close()
			return
		}
	}
}

func GetCurrentKey(roomID string) (string, error) {
	key, err := rdb.HGet(ctx, roomID, "invitation_key").Result()
	fmt.Printf("key current:: %s\n", key)
	if key == "" || err != nil {
		return "", err
	}

	return key, nil
}

func InvitationKeyReverseIndex(invitationKey, roomID string) error {
	// Creates a reverse index mapping invitationKey to roomID
	expires := 20 * time.Second

	err := rdb.Set(ctx, "invitationKey:"+invitationKey, roomID, expires).Err()
	if err != nil {
		return fmt.Errorf("failed to create invkey reverse index %w", err)
	}

	return nil
}

func AuthorizeInvitationKey(keyInput string) (bool, string, error) {
	// Checks for any existing room using the invKey reverse index mapped to roomID
	roomID, err := rdb.Get(ctx, "invitationKey:"+keyInput).Result()
	if err != nil {
		if err == redis.Nil {
			return false, "", err
		} else {
			return false, "", err
		}
	}

	return true, roomID, nil
}
