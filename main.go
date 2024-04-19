package main

import (
	"log"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

func main(){
	tr := p2p.NewTCPTransport(":3000")

	if err := tr.ListenAndAccept(); err != nil{
		log.Fatal(err)
	}

	select{}
}
