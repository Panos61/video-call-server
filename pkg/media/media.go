package media

import (
	"log"
	"os"

	"github.com/pion/mediadevices"
	"github.com/pion/mediadevices/pkg/codec/opus"
	"github.com/pion/webrtc/v3"
)

var ConfigICEServer = webrtc.Configuration{
	ICEServers: []webrtc.ICEServer{
		{
			URLs: []string{"stun:stun.l.google.com:19302"},
		},
	},
}

var (
	MediaEngine webrtc.MediaEngine
)

func ConfigMedia() {
	file, err := os.OpenFile("media.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()
	log.SetOutput(file)

	// x264, err := x264.NewParams()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	opus, err := opus.NewParams()
	if err != nil {
		log.Fatal(err)
	}

	codecSelector := mediadevices.NewCodecSelector(
		// mediadevices.WithVideoEncoders(&x264),
		mediadevices.WithAudioEncoders(&opus),
	)

	codecSelector.Populate(&MediaEngine)
}
