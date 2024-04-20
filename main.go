package main

import (
	"log"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

func main(){
	tcpOpts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		Decoder: p2p.GOBDecoder{},
		HandshakeFunc: p2p.NOPHandshakeFunc,
	}
	tr := p2p.NewTCPTransport(tcpOpts)

	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}

	select{}
}
