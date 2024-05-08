package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"time"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	tcpTransportOpts := p2p.TCPTransportOpts{
		ListenAddr: listenAddr,
		HandshakeFunc: p2p.NOPHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}

	tcpTransport := p2p.NewTCPTransport(tcpTransportOpts)

	fileServerOpts := FileServerOpts{
		StorageRoot: listenAddr + "_network",
		PathTransformFunc: CASPathTransformFunc,
		Transport: tcpTransport,
		BootstrapNodes: nodes,
	}

	s :=  NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer

	return s
}

func main(){
	s1 := makeServer(":8000", "")
	s2 := makeServer(":80", ":8000")
	go func ()  {
		log.Fatal(s1.Start())	
	}()

	time.Sleep(1 * time.Second)

	go s2.Start()
	
	time.Sleep(1 * time.Second)

	// data := bytes.NewReader([]byte("a thick data file"))
	// s2.Store("privateData", data)

	r, err := s2.Get("privateData")
	if err != nil{
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))

	select{}
}
