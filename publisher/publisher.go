package publisher

import (
	"context"
	"encoding/json"
	"fmt"
	sfu "github.com/pion/ion-sfu/cmd/server/grpc/proto"
	"github.com/pion/webrtc/v3"
	"github.com/sirupsen/logrus"
	"github.com/smf8/sfu-load-test/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
)

type Publisher interface {
	Publish()
}

type ClientPublisher struct {
	client client.Client
}

func NewPublisher(filePath string, client client.Client) (*ClientPublisher, error) {
	err := client.AddVideo(filePath)
	if err != nil {
		return nil, err
	}

	return &ClientPublisher{client: client}, nil
}

//nolint:funlen
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

	offer, err := cp.client.Pc.CreateOffer(nil)
	if err != nil {
		logrus.Errorf("Error creating local SD for %s: %v aborting", cp.client.Name, err)
		return
	}

	err = cp.client.Pc.SetLocalDescription(offer)
	if err != nil {
		logrus.Errorf("Error setting local SD for %s: %v aborting", cp.client.Name, err)
		return
	}

	ctx := context.Background()
	client, err := cp.client.C.Signal(ctx)

	if err != nil {
		logrus.Errorf("Error publishing stream for %s: %v aborting", cp.client.Name, err)
		return
	}
	cp.client.Pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			// gathering done
			return
		}

		bytes, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Fatalf("OnIceCandidate error %s", err)
		}

		err = client.Send(&sfu.SignalRequest{
			Payload: &sfu.SignalRequest_Trickle{
				Trickle: &sfu.Trickle{
					Init: string(bytes),
				},
			},
		})
		if err != nil {
			log.Fatalf("OnIceCandidate error %s", err)
		}
	})

	err = client.Send(&sfu.SignalRequest{
		Payload: &sfu.SignalRequest_Join{
			Join: &sfu.JoinRequest{
				Sid: cp.client.Sid,
				Offer: &sfu.SessionDescription{
					Type: offer.Type.String(),
					Sdp:  []byte(cp.client.Pc.LocalDescription().SDP),
				},
			},
		},
	})

	for {
		reply, err := client.Recv()

		if err == io.EOF {
			// WebRTC Transport closed
			fmt.Println("WebRTC Transport Closed")
		}

		if err != nil {
			log.Fatalf("Error receving publish response: %v", err)
		}

		switch payload := reply.Payload.(type) {
		case *sfu.SignalReply_Join:
			fmt.Printf("Got answer from sfu. Starting streaming for pid %v!\n", payload.Join.Pid)
			// Set the remote SessionDescription
			if err = cp.client.Pc.SetRemoteDescription(webrtc.SessionDescription{
				Type: webrtc.SDPTypeAnswer,
				SDP:  string(payload.Join.Answer.Sdp),
			}); err != nil {
				panic(err)
			}

		case *sfu.SignalReply_Trickle:
			var candidate webrtc.ICECandidateInit
			err := json.Unmarshal([]byte(payload.Trickle.Init), &candidate)
			if err != nil {
				logrus.Errorf("error parsing ice candidate: %v", err)
			}

			if err := cp.client.Pc.AddICECandidate(candidate); err != nil {
				logrus.Errorf("%w", status.Errorf(codes.Internal, "error adding ice candidate"))
			}
		}
	}
}
