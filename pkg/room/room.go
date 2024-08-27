package room

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"server/redisclient"
	"strconv"
	"sync"

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

func CreateRoom() *Room {
	roomID := uuid.New().String()

	newRoom := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
	}

	err := rdb.HSet(ctx, roomID, map[string]interface{}{"id": roomID}).Err()
	if err != nil {
		log.Printf("Error saving room: %s\n", err)
	}

	return newRoom
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

func (r *Room) AddParticipant(p *Participant) {
	r.Participants[strconv.FormatInt(p.ID, 10)] = p
}
