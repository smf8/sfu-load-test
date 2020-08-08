package publisher

import (
	"log"

	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/smf8/sfu-load-test/client"
)

type Publisher interface {
	Publish()
}

type ClientPublisher struct {
	client *client.Client
}

func NewPublisher(filePath string, client *client.Client) (*ClientPublisher, error) {
	err := client.AddVideo(filePath)
	if err != nil {
		return nil, err
	}

	return &ClientPublisher{client: client}, nil
}

func (cp *ClientPublisher) Publish() {
	log.Printf("Publishing stream for client: %s", cp.client.Name)

	if cp.client.Media.AudioTrack() != nil {
		if _, err := cp.client.Pc.AddTrack(cp.client.Media.AudioTrack()); err != nil {
			logrus.Errorf("Error adding audio track for %s: %v aborting", cp.client.Name, err)
			return
		}
	}

	if cp.client.Media.VideoTrack() != nil {
		if _, err := cp.client.Pc.AddTrack(cp.client.Media.VideoTrack()); err != nil {
			logrus.Errorf("Error adding video track for %s: %v aborting", cp.client.Name, err)
			return
		}
	}

	cp.client.Pc.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		logrus.Printf("Client %v publisher State has changed %s \n", cp.client.Name, connectionState.String())
		// start sending video when connection is complete
		if connectionState == webrtc.ICEConnectionStateConnected {
			cp.client.Media.Start()
		}
	})
}
