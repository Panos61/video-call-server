package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"server/pkg/jwtutils"
	"server/pkg/room"
	"strings"
	"time"
)

func createRoomHandler(w http.ResponseWriter, r *http.Request) {
	newRoom, err := room.CreateRoom()
	if err != nil {
		http.Error(w, "Failed to create room", 500)
	}

	hostParticipant, err := room.SetHostParticipant(newRoom.ID)
	if err != nil {
		http.Error(w, "Failed to set host participant", 500)
		return
	}

	jwtToken, err := jwtutils.GenerateJWT(hostParticipant.ID, true)
	if err != nil {
		http.Error(w, "Failed to generate jwt", 500)
		return
	}

	hostParticipant.Token = jwtToken

	response := map[string]interface{}{
		"id":           newRoom.ID,
		"participants": *hostParticipant,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func joinRoomHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "No room with the specified id found.", http.StatusBadRequest)
		return
	}

	existingRoom, err := room.GetRoom(id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var participant room.Participant

	err = json.Unmarshal(body, &participant)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var claims *jwtutils.Claims
	var participantID string
	var isHost bool

	tokenStr := r.Header.Get("Authorization")
	tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")

	if tokenStr != "" && tokenStr != "null" {
		claims, err = jwtutils.ValidateToken(tokenStr)
		if err != nil {
			http.Error(w, "error validating token", 500)
		}

		participantID = claims.ParticipantID
		isHost = claims.IsHost
	} else {
		isHost = false
	}

	_, err = room.JoinRoom(existingRoom.ID, participantID, tokenStr, participant.Name, isHost)
	if err != nil {
		http.Error(w, "error joining room", 500)
		return
	}

	response := struct {
		RoomID       string                       `json:"room_id"`
		Participants map[string]*room.Participant `json:"participants"`
	}{
		RoomID:       existingRoom.ID,
		Participants: existingRoom.Participants,
	}

	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
	}
}

func setInvKeyHandler(w http.ResponseWriter, r *http.Request) {
	roomID := r.PathValue("id")

	existingRoom, err := room.GetRoom(roomID)
	if existingRoom == nil || err != nil {
		http.Error(w, "No room existing with this id.", http.StatusBadRequest)
		return
	}

	invKey := room.GenerateInvKey(existingRoom.ID)
	err = room.SetRoomKey(existingRoom.ID, invKey)
	if err != nil {
		http.Error(w, "failed to set invKey to this room", 500)
		return
	}

	err = room.InvitationKeyReverseIndex(invKey, existingRoom.ID)
	if err != nil {
		http.Error(w, "Failed to create reverse index for new invitation key.", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"invitation_key": invKey,
	})
}

// Server-sent event handler to check for expired invitation key
func sseKeyUpdateHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	roomID := r.PathValue("id")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Listen for client disconnect
	notify := r.Context().Done()

	for {
		select {
		case <-notify:
			log.Printf("Client disconnected from room: %s", roomID)
			return

		default:
			isExpired, err := room.IsKeyExpired(roomID)
			if err != nil {
				fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
				flusher.Flush()
				return
			}

			if isExpired {
				newKey := room.GenerateInvKey(roomID)
				err := room.SetRoomKey(roomID, newKey)
				if err != nil {
					fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
					flusher.Flush()
					return
				}

				fmt.Fprintf(w, "event: update\ndata: %s\n\n", newKey)
				flusher.Flush()

				err = room.InvitationKeyReverseIndex(newKey, roomID)
				if err != nil {
					fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
					flusher.Flush()
					return
				}
			}

			time.Sleep(30 * time.Second)
		}
	}
}

func authorizeInvitationHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var requestBody struct {
		KeyInput string `json:"keyInput"`
	}

	err = json.Unmarshal(body, &requestBody)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	isAuthorized, roomID, err := room.AuthorizeInvitationKey(requestBody.KeyInput)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isAuthorized": false,
			"roomID":       "",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"isAuthorized": isAuthorized,
		"roomID":       roomID,
	})
}
