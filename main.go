package main

import (
	"bytes"
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
		EncKey: newEncryptionKey(),
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
	s1 := makeServer(":8888", "")
	s2 := makeServer(":80", ":8888")
	s3 := makeServer(":3000", ":8888", ":80")
	go func ()  {log.Fatal(s1.Start())}()
	time.Sleep(time.Millisecond * 500)
	go func ()  {log.Fatal(s2.Start())}()

	time.Sleep(2 * time.Second)

	go s3.Start()
	
	time.Sleep(2 * time.Second)

	for i:=0;i<20;i++{
		key := fmt.Sprintf("Picture_%d.jpg", i)
		data := bytes.NewReader([]byte("a thick data file"))
		s3.Store(key, data)	
		
		if err := s3.store.Delete(key); err != nil{
			log.Fatal(err)
		}

		r, err := s3.Get(key)
		if err != nil{
			log.Fatal(err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
	}

}
