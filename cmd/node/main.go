package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"github.com/hashicorp/serf/serf"
)

func main() {
	_, cancel := context.WithCancel(context.Background())

	// start the serf instance
	eventCh := make(chan serf.Event, 64)
	serfConfig := serf.DefaultConfig()
	serflib, err := startSerfInstance(serfConfig, eventCh)
	if err != nil {
		fmt.Printf("error creating serflib\n")
		serflib = nil
	}

	// join other cluster
	otherGroup := []string{"127.0.0.2", "127.0.0.3"}
	_, err = serflib.Join(otherGroup, false)
	if err != nil {
		fmt.Printf("join existing cluster failed\n")
	}

	go processReceivedEvent(eventCh)

	// broadcast messages
	var b strings.Builder
	b.WriteString("ping all from ")
	b.WriteString(serfConfig.NodeName)
	err = serflib.UserEvent(b.String(), []byte("tests"), false)
	if err != nil {
		fmt.Printf("fail to send events")
	}

	// handle ctrl-C
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)
	select {
	case <-signalCh:
		cancel()
		serflib.Shutdown()
	}
}

func startSerfInstance(serfConfig *serf.Config, eventCh chan serf.Event) (*serf.Serf, error) {
	serfConfig.EventCh = eventCh
	bindAddr := os.Getenv("MY_POD_IP")
	serfConfig.NodeName = os.Getenv("MY_POD_NAME")
	serfConfig.MemberlistConfig.BindAddr = bindAddr
	return serf.Create(serfConfig)
}

func processReceivedEvent(eventCh chan serf.Event) {
	for {
		select {
		case event := <-eventCh:
			fmt.Printf("SUCCESS!! received user Event: %s\n", event.String())
		}
	}
}
