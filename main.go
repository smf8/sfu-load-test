package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/smf8/sfu-load-test/client"
	"github.com/smf8/sfu-load-test/log"
	"github.com/smf8/sfu-load-test/publisher"
	"github.com/smf8/sfu-load-test/subscriber"
)

func main() {
	log.SetupLogger()

	server := flag.String("server", "localhost:50051", "sfu's server and port")
	subs := flag.Int("sub", 0, "number of subscribers")
	pubs := flag.Int("pub", 0, "number of publishers")
	sid := flag.Uint("sid", 1, "session ID to join in SFU")
	filepath := flag.String("file", "", "video file to publish")

	flag.Parse()

	fmt.Println(*filepath)
	if *subs == 0 && *pubs == 0 {
		logrus.Fatalf("you must specify one of -pub or -sub to a non-zero number.")
	}

	subscribers := make([]subscriber.Subscriber, *subs)
	publishers := make([]publisher.Publisher, *pubs)
	clients := make([]*client.Client, 0)

	for sub := range subscribers {
		cl := client.NewClient(fmt.Sprintf("subscriber_%d", sub), client.Subscriber, *server, uint32(*sid))
		s := subscriber.NewClientSubscriber(cl)
		clients = append(clients, cl)

		s.Subscribe()
		subscribers[sub] = s
	}

	for pub := range publishers {
		cl := client.NewClient(fmt.Sprintf("publisher_%d", pub), client.Publisher, *server, uint32(*sid))
		clients = append(clients, cl)

		p, err := publisher.NewPublisher(*filepath, cl)
		if err != nil {
			logrus.Errorf("failed to create publisher %v", err)
		}

		p.Publish()
		publishers[pub] = p
	}

	for cl := range clients {
		go clients[cl].Connect()
		fmt.Println("sdfd")
		<-time.After(time.Second)
	}

	select {}
}
