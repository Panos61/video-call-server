package peerconnection

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"server/pkg/media"
	"server/pkg/room"

	"github.com/pion/webrtc/v3"
)

func InitPeerConnection(w http.ResponseWriter, r *http.Request, rm *room.Room) (*webrtc.PeerConnection, error) {
	file, err := os.OpenFile("peerconnection.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	media.ConfigMedia()
	api := webrtc.NewAPI(webrtc.WithMediaEngine(&media.MediaEngine))

	peerConnection, err := api.NewPeerConnection(media.ConfigICEServer)
	if err != nil {
		return nil, err
	}

	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		fmt.Printf("ICE connection state has changed to %s\n", connectionState)
	})

	peerConnection.OnICECandidate(func(ICECandidate *webrtc.ICECandidate) {
		if ICECandidate != nil {
			candidateJSON, err := json.Marshal(ICECandidate.ToJSON())
			if err != nil {
				log.Fatal(err)
				return
			}
			w.Write(candidateJSON)
		}
	})

	// peerConnection.OnTrack(func(tr *webrtc.TrackRemote, r *webrtc.RTPReceiver) {
	// 	for _, p := range rm.PeerConnections {
	// 		if p != peerConnection {
	// 			newTrack, err := webrtc.NewTrackLocalStaticRTP(tr.Codec().RTPCodecCapability, tr.ID(), tr.StreamID())
	// 			if err != nil {
	// 				fmt.Println(err)
	// 				continue
	// 			}

	// 			_, err = p.AddTrack(newTrack)
	// 			fmt.Printf("new track %v", newTrack)
	// 			if err != nil {
	// 				fmt.Println(err)
	// 				continue
	// 			}

	// 			// go func() {
	// 			// 	buf := make([]byte, 1500)
	// 			// 	for {
	// 			// 		i, _, readErr := tr.Read(buf)
	// 			// 		if readErr != nil {
	// 			// 			log.Println(readErr)
	// 			// 			return
	// 			// 		}

	// 			// 		writeErr := newTrack.WriteRTP(buf[:i])
	// 			// 		if writeErr != nil {
	// 			// 			log.Println(writeErr)
	// 			// 			return
	// 			// 		}
	// 			// 	}
	// 			// }()
	// 		}
	// 	}
	// })

	// Read the SDP offer from the request body
	var offer webrtc.SessionDescription
	if err := json.NewDecoder(r.Body).Decode(&offer); err != nil {
		// http.Error(w, "Failed to decode SDP offer", http.StatusBadRequest)
		return nil, err
	}

	// fmt.Printf("offer %v\n", offer)

	// Set the remote description with the received SDP offer
	if err := peerConnection.SetRemoteDescription(offer); err != nil {
		fmt.Printf("remote desc err: %s\n", err)
		http.Error(w, fmt.Sprintf("Failed to set remote description: %v", err), http.StatusInternalServerError)
		return nil, err
	}

	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		http.Error(w, "Failed to create answer", http.StatusInternalServerError)
		return nil, err
	}

	err = peerConnection.SetLocalDescription(answer)
	if err != nil {
		fmt.Println(err)
		http.Error(w, fmt.Sprintf("Failed to set local description: %v", err), http.StatusInternalServerError)
		return nil, err
	}

	answerJSON, err := json.Marshal(answer)
	if err != nil {
		http.Error(w, "Failed to marshal SDP answer", http.StatusInternalServerError)
		return nil, err
	}

	w.Write(answerJSON)

	return peerConnection, nil
}
