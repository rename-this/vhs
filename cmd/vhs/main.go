package main

import (
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

func main() {
	h, err := pcap.OpenLive("eth0", 1024, false, 5*time.Second)
	if err != nil {
		log.Fatal(err)
	}
	defer h.Close()

	var (
		pc = gopacket.NewPacketSource(h, h.LinkType()).Packets()
		tc = time.After(5 * time.Second)
	)

	for {
		select {
		case <-tc:
			return
		case p := <-pc:
			fmt.Println(p)
		}
	}
}
