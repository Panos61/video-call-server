// Package signal contains helpers to exchange the SDP session
package signalling

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/pion/webrtc/v3"
)

// Allows compressing offer/answer to bypass terminal input limits.
// const compress = false

// MustReadStdin blocks until input is received from stdin
func MustReadStdin() string {
	r := bufio.NewReader(os.Stdin)

	var in string
	for {
		var err error
		in, err = r.ReadString('\n')
		if err != io.EOF {
			if err != nil {
				panic(err)
			}
		}
		in = strings.TrimSpace(in)
		if len(in) > 0 {
			break
		}
	}

	fmt.Println("")

	return in
}

// Encode encodes the input in base64
// It can optionally zip the input before encoding
func Encode(sd webrtc.SessionDescription) string {
	sdp, err := json.Marshal(sd)
	if err != nil {
		log.Fatalf("Failed to marshal SDP: %v", err)
	}

	return base64.StdEncoding.EncodeToString(sdp)
}

// Decode decodes a base64 SessionDescription
func Decode(encoded string, sd *webrtc.SessionDescription) {
	sdp, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		log.Fatalf("Failed to decode base64 SDP: %v", err)
	}

	err = json.Unmarshal(sdp, sd)
	if err != nil {
		log.Fatalf("Failed to unmarshal SDP: %v", err)
	}
}

func DecodeMessage(data []byte) (SignalMessage, error) {
	var message SignalMessage
	err := json.Unmarshal(data, &message)
	if err != nil {
		return SignalMessage{}, err
	}

	return message, nil
}

// func zip(in []byte) []byte {
// 	var b bytes.Buffer
// 	gz := gzip.NewWriter(&b)
// 	_, err := gz.Write(in)
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = gz.Flush()
// 	if err != nil {
// 		panic(err)
// 	}
// 	err = gz.Close()
// 	if err != nil {
// 		panic(err)
// 	}
// 	return b.Bytes()
// }

// func unzip(in []byte) []byte {
// 	var b bytes.Buffer
// 	_, err := b.Write(in)
// 	if err != nil {
// 		panic(err)
// 	}
// 	r, err := gzip.NewReader(&b)
// 	if err != nil {
// 		panic(err)
// 	}
// 	res, err := io.ReadAll(r)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return res
// }
