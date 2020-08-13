package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	sfuclient "github.com/pion/ion-sfu/cmd/server/grpc/proto"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/smf8/producer"
	"github.com/smf8/producer/ivf"
	"github.com/smf8/producer/webm"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const Publisher = "publisher"
const Subscriber = "subscriber"

type Client struct {
	Name       string
	Sid        string
	Pc         *webrtc.PeerConnection
	cType      string
	AudioTrack *webrtc.Track
	VideoTrack *webrtc.Track
	conn       *grpc.ClientConn
	C          sfuclient.SFUClient
	Media      producer.IFileProducer
}

func NewClient(name, cType, address string, sid string) *Client {
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

	client.C = sfuclient.NewSFUClient(client.conn)

	conf := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	client.Pc, err = webrtc.NewPeerConnection(conf)
	if err != nil {
		logrus.Errorf("error creating peer connection %v", err)

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

//nolint:funlen
func (c *Client) Connect() {
	offer, err := c.Pc.CreateOffer(nil)
	if err != nil {
		logrus.Errorf("Error creating local SD for %s: %v aborting", c.Name, err)
		return
	}

	err = c.Pc.SetLocalDescription(offer)
	if err != nil {
		logrus.Errorf("Error setting local SD for %s: %v aborting", c.Name, err)
		return
	}

	ctx := context.Background()
	client, err := c.C.Signal(ctx)

	if err != nil {
		logrus.Errorf("Error publishing stream for %s: %v aborting", c.Name, err)
		return
	}

	c.Pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			// gathering done
			return
		}

		bytes, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			logrus.Errorf("OnIceCandidate error %s", err)
		}

		err = client.Send(&sfuclient.SignalRequest{
			Payload: &sfuclient.SignalRequest_Trickle{
				Trickle: &sfuclient.Trickle{
					Init: string(bytes),
				},
			},
		})
		if err != nil {
			logrus.Errorf("OnIceCandidate error %s", err)
		}
	})

	err = client.Send(&sfuclient.SignalRequest{
		Payload: &sfuclient.SignalRequest_Join{
			Join: &sfuclient.JoinRequest{
				Sid: c.Sid,
				Offer: &sfuclient.SessionDescription{
					Type: offer.Type.String(),
					Sdp:  []byte(c.Pc.LocalDescription().SDP),
				},
			},
		},
	})

	if err != nil {
		logrus.Errorf("%s failed sending join request %v", c.Name, err)
	}

	for {
		reply, err := client.Recv()

		if err == io.EOF {
			// WebRTC Transport closed
			logrus.Println("WebRTC Transport Closed")
		}

		if err != nil {
			logrus.Errorf("Error receiving publish response: %v", err)
		}

		switch payload := reply.Payload.(type) {
		case *sfuclient.SignalReply_Join:
			fmt.Printf("Got answer from sfu. Starting streaming for pid %v!\n", payload.Join.Pid)
			// Set the remote SessionDescription
			if err = c.Pc.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  string(payload.Join.Answer.Sdp),
			}); err != nil {
				panic(err)
			}

		case *sfuclient.SignalReply_Trickle:
			var candidate webrtc.ICECandidateInit
			err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)

			if err != nil {
				logrus.Errorf("error parsing ice candidate: %v", err)
			}

			if err := c.Pc.AddICECandidate(candidate); err != nil {
				logrus.Errorf("%v", status.Errorf(codes.Internal, "error adding ice candidate %v", err))
			}
		}
	}
}

func (c *Client) AddVideo(fileAddress string) error {
	if c.cType != Publisher {
		return fmt.Errorf("invalid client type for publisher")
	}

	ext := filepath.Ext(fileAddress)
	switch ext {
	case ".webm":
		c.Media = webm.NewMFileProducer(fileAddress, 0, producer.TrackSelect{
			Audio: true,
			Video: true,
		})
	case ".ivf":
		c.Media = ivf.NewIVFProducer(fileAddress, 1)
	default:
		return fmt.Errorf("invalid video file extension")
	}

	return nil
}
