package room

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"time"
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
	expirationTime := 10 * time.Second

	data, err := rdb.HMSet(ctx, roomID, map[string]interface{}{
		"invitation_key": invKey,
		"expiresIn":      time.Now().Add(expirationTime).Format(time.RFC3339),
	}).Result()

	if err != nil {
		log.Printf("Error setting room invitation key %s\n", err)
		return
	}

	fmt.Printf("data: %#v\n", data)

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
