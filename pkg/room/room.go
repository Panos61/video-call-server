package room

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"server/redisclient"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

type Participant struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	IsHost bool   `json:"isHost"`
}

type Room struct {
	ID              string                            `json:"id"`
	Participants    map[string]*Participant           `json:"participants"`
	PeerConnections map[string]*webrtc.PeerConnection `json:"peer_connections"`
	Mx              sync.Mutex
}

type WSClient struct {
	Conn *websocket.Conn
}

var (
	ctx = context.Background()
	rdb = redisclient.GetRedisClient()
)

func GenerateRoomID() string {
	return fmt.Sprintf("%d", rand.Int63())
}

func CreateRoom() (*Room, error) {
	roomID := uuid.New().String()

	newRoom := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
	}

	err := rdb.HSet(ctx, roomID, map[string]interface{}{"id": roomID}).Err()
	if err != nil {
		return nil, fmt.Errorf("error saving room: %w", err)
	}

	return newRoom, nil
}

func GetRoom(id string) *Room {
	roomData, err := rdb.HGetAll(ctx, id).Result()
	if err != nil || len(roomData) == 0 {
		log.Printf("Cannot get room with id: %s", id)
		return nil
	}

	room := &Room{
		ID: roomData["id"],
	}

	return room
}

func InvitationKeyReverseIndex(invitationKey, roomID string) error {
	// Creates a reverse index mapping invitationKey to roomID
	expires := 1 * time.Minute

	err := rdb.Set(ctx, "invitationKey:"+invitationKey, roomID, expires).Err()
	if err != nil {
		return fmt.Errorf("failed to create invkey reverse index %w", err)
	}

	return nil
}

func AuthorizeInvitationKey(keyInput string) (bool, error) {
	// Checks for any existing room using the invKey reverse index mapped to roomID
	err := rdb.Get(ctx, "invitationKey:"+keyInput).Err()
	if err != nil {
		if err == redis.Nil {
			return false, err
		} else {
			return false, err
		}
	}

	return true, nil
}

func (r *Room) AddParticipant(p *Participant) {
	r.Participants[strconv.FormatInt(p.ID, 10)] = p
}
