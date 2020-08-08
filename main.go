package main

import (
	"fmt"
	"github.com/smf8/sfu-load-test/client"
	"github.com/smf8/sfu-load-test/log"
)

func main() {
	log.SetupLogger()

	c := client.NewClient("name", "publisher", "localhost:50051", 12)

	offer, err := c.Pc.CreateOffer(nil)

	if err != nil {
		panic(err)
	}

	fmt.Println(offer)
}
