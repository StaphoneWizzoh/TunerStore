package main

import (
	"fmt"
	"log"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

func main(){
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		Decoder: p2p.DefaultDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}

	tr := p2p.NewTCPTransport(tcpOpts)

	go func ()  {
		for{
			msg := <-tr.Consume()
			fmt.Printf("Message: %v\n", msg)
		}
	}()

	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}

	select{}
}
