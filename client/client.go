package client

import (
	"fmt"
	sfu_client "github.com/pion/ion-sfu/cmd/server/grpc/proto"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/smf8/producer"
	"github.com/smf8/producer/ivf"
	"github.com/smf8/producer/webm"
	"google.golang.org/grpc"
	"path/filepath"
)

const Publisher = "publisher"
const Subscriber = "subscriber"

type Client struct {
	Name       string
	Sid        uint32
	Pc         *webrtc.PeerConnection
	cType      string
	AudioTrack *webrtc.Track
	VideoTrack *webrtc.Track
	conn       *grpc.ClientConn
	C          sfu_client.SFUClient
	Media      producer.IFileProducer
}

func NewClient(name, cType, address string, sid uint32) *Client {
	logrus.Debugln("Creating a new client")

	var err error
	client := new(Client)
	client.Name = name
	client.Sid = sid

	client.conn, err = grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logrus.Errorf("did not connect: %v", err)

		return nil
	}

	client.C = sfu_client.NewSFUClient(client.conn)

	conf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	client.Pc, err = webrtc.NewPeerConnection(conf)
	if err != nil {
		logrus.Errorf("error creating peer connection %w", err)

		return nil
	}

	if cType == Publisher || cType == Subscriber {
		client.cType = cType
	} else {
		logrus.Errorf("invalid client type, it must be either a subscriber or a publisher")

		return nil
	}

	return client
}

func (c *Client) AddVideo(fileAddress string) error {
	if c.cType != Publisher {
		return fmt.Errorf("invalid client type for publisher")
	}
	ext := filepath.Ext(fileAddress)
	if ext == ".webm" {
		c.Media = webm.NewMFileProducer(fileAddress, 0, producer.TrackSelect{
			Audio: true,
			Video: true,
		})
	} else if ext == ".ivf" {
		c.Media = ivf.NewIVFProducer(fileAddress, 1)
	} else {
		return fmt.Errorf("invalid video file extention")
	}

	return nil
}
