package subscriber

import (
	"log"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/smf8/sfu-load-test/client"
)

const PLICycle = 3 * time.Second

type Subscriber interface {
	Subscribe()
}

type ClientSubscriber struct {
	client *client.Client
}

func NewClientSubscriber(client *client.Client) *ClientSubscriber {
	// a media engine with default codecs was added implicitly
	// Allow us to receive 1 audio track, and 1 video track
	if _, err := client.Pc.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		panic(err)
	}

	if _, err := client.Pc.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio, webrtc.RtpTransceiverInit{
		Direction: webrtc.RTPTransceiverDirectionRecvonly,
	}); err != nil {
		panic(err)
	}

	return &ClientSubscriber{client}
}

func (cs ClientSubscriber) Subscribe() {
	cs.client.Pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logrus.Printf("Client %s Connection State has changed %s \n", cs.client.Name, connectionState.String())

		if connectionState == webrtc.ICEConnectionStateClosed {
			logrus.Println("Finished receiving media")
		}
	})

	cs.client.Pc.OnTrack(func(track *webrtc.Track, receiver *webrtc.RTPReceiver) {
		go func() {
			ticker := time.NewTicker(PLICycle)
			for range ticker.C {
				errSend := cs.client.Pc.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: track.SSRC()}})
				if errSend != nil {
					log.Println(errSend)
				}
			}
		}()

		cs.handleTrack(track)
	})
}

func (cs *ClientSubscriber) handleTrack(tr *webrtc.Track) {
	logrus.Println("handling track", tr.Kind())
}
