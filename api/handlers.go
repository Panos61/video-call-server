package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"server/pkg/jwtutils"
	"server/pkg/peerconnection"
	"server/pkg/room"
	"strings"
	"time"

	"github.com/pion/webrtc/v3"
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

	response := map[string]interface{}{
		"id":           newRoom.ID,
		"participants": *hostParticipant,
		"token":        jwtToken,
	}

	fmt.Printf("Create-room resp: %v\n", response)

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

func updateInvKeyHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Error upgrading to WS.", http.StatusInternalServerError)
	}

	roomID := r.PathValue("id")

	existingRoom, err := room.GetRoom(roomID)
	if existingRoom == nil || err != nil {
		http.Error(w, "No room existing with this id.", http.StatusBadRequest)
		return
	}

	keyChan := make(chan string)

	go func() {
		for {
			invKey := room.GenerateInvKey(existingRoom.ID)
			time.Sleep(20 * time.Second)
			keyChan <- invKey
		}
	}()

	client := room.WSClient{Conn: conn}

	go func() {
		for invKey := range keyChan {
			client.SetRoomKey(existingRoom.ID, invKey)
			err = room.InvitationKeyReverseIndex(invKey, existingRoom.ID)
			if err != nil {
				return
			}
		}
	}()
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
		// should specify the error here
		// could be internal error
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"isAuthorized": isAuthorized,
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

func signallingHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "No room with the specified id found.", http.StatusBadRequest)
		return
	}

	existingRoom, err := room.GetRoom(id)
	if err != nil {
		return
	}

	peerconnection, err := peerconnection.InitPeerConnection(w, r, existingRoom)
	if peerconnection == nil {
		http.Error(w, "Failed to create peer connection.", http.StatusInternalServerError)
	}

	if err != nil {
		fmt.Println(err)
	}

	// existingRoom.PeerConnections[room.GenerateRoomID()] = peerconnection

	response := struct {
		RoomID          string                            `json:"id"`
		Participants    map[string]*room.Participant      `json:"participants"`
		PeerConnections map[string]*webrtc.PeerConnection `json:"peer_connections"`
	}{
		RoomID:       existingRoom.ID,
		Participants: existingRoom.Participants,
		// PeerConnections: existingRoom.PeerConnections,
	}

	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to encode response"})
	}
}
