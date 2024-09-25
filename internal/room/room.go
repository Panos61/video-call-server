package room

import (
	"context"
	"encoding/base64"
	"fmt"
	"server/internal/jwtutils"
	"server/internal/redisclient"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Participant struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	IsHost bool   `json:"isHost"`
	Token  string `json:"jwt"`
}

type Room struct {
	ID           string                  `json:"id"`
	Participants map[string]*Participant `json:"participants"`
	HostID       string                  `json:"host_id"`
}

type WSClient struct {
	Conn *websocket.Conn
}

var (
	ctx = context.Background()
	rdb = redisclient.GetRedisClient()
)

func CreateRoom() (*Room, error) {
	roomID := uuid.New().String()

	newRoom := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
	}

	err := rdb.HSet(ctx, "room:"+roomID, map[string]interface{}{"id": roomID}).Err()
	if err != nil {
		return nil, fmt.Errorf("error saving room: %w", err)
	}

	return newRoom, nil
}

func GenerateParticipantID() string {
	// We are shortening the user id by encoding it to base64 making data management in redis easier
	id := uuid.New()
	encodedHostID := base64.URLEncoding.EncodeToString(id[:])
	encodedHostID = encodedHostID[:22]

	return encodedHostID
}

func JoinRoom(roomID, participantID, hostToken, nameInput string, isHost bool) (*Participant, error) {
	hostID, err := GetHost(roomID)
	if err != nil {
		return nil, err
	}

	var participant *Participant

	// if host, update the existing hash
	if isHost {
		_, err := rdb.HSet(ctx, "room:"+roomID+":participant:"+hostID, map[string]interface{}{
			"name": nameInput,
		}).Result()

		if err != nil {
			return nil, err
		}

		participant = &Participant{
			ID:     hostID,
			Name:   nameInput,
			IsHost: true,
			Token:  hostToken,
		}

		return participant, nil
	}

	// If not host, create a new user hash and a room set of participants
	participant = &Participant{
		ID:     GenerateParticipantID(),
		Name:   nameInput,
		IsHost: false,
	}

	pipe := rdb.TxPipeline()
	pipe.SAdd(ctx, "room:"+roomID+":participants", participant.ID)
	pipe.HMSet(ctx, "room:"+roomID+":participant:"+participant.ID, map[string]interface{}{
		"id":     participant.ID,
		"name":   participant.Name,
		"isHost": participant.IsHost,
	})

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	token, err := jwtutils.GenerateJWT(participant.ID, false)
	if err != nil {
		return nil, fmt.Errorf("error generating token for guest: %w", err)
	}

	participant = &Participant{Token: token}

	return participant, nil
}

func SetHostParticipant(roomID string) (*Participant, error) {
	participant := &Participant{
		ID:     GenerateParticipantID(),
		IsHost: true,
	}

	pipe := rdb.TxPipeline()
	pipe.SAdd(ctx, "room:"+roomID+":participants", participant.ID)
	pipe.HMSet(ctx, "room:"+roomID+":participant:"+participant.ID, map[string]interface{}{
		"id":     participant.ID,
		"name":   "",
		"isHost": participant.IsHost,
	})
	pipe.HSet(ctx, "room:"+roomID, "host_id", participant.ID)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("error setting host participant: %w", err)
	}

	return participant, nil
}

func GetHost(roomID string) (string, error) {
	hostID, err := rdb.HGet(ctx, "room:"+roomID, "host_id").Result()
	if err != nil {
		return "", err
	}

	return hostID, err
}

func GetRoom(id string) (*Room, error) {
	roomData, err := rdb.HGetAll(ctx, "room:"+id).Result()
	if err != nil || len(roomData) == 0 {
		return nil, err
	}

	room := &Room{
		ID: roomData["id"],
	}

	return room, nil
}
