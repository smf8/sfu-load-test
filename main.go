package main

import (
	"github.com/sirupsen/logrus"
	"github.com/smf8/sfu-load-test/publisher"

	"github.com/smf8/sfu-load-test/client"
	"github.com/smf8/sfu-load-test/log"
)

const address = "localhost:50051"

func main() {
	log.SetupLogger()

	c := client.NewClient("name", "publisher", address, 12)

	p, err := publisher.NewPublisher("output.webm", c)
	if err != nil {
		logrus.Fatalf("could not create publisher %v", err)
	}

	p.Publish()

	c.Connect()
}
