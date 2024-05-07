package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

type FileServerOpts struct{
	StorageRoot string
	PathTransformFunc PathTransformFunc
	Transport p2p.Transport
	BootstrapNodes []string
}

type FileServer struct{
	FileServerOpts

	peerLock sync.Mutex
	peers map[string]p2p.Peer
	store *Store
	quitCh chan struct{}
}

func NewFileServer(opts FileServerOpts) *FileServer{
	storeOpts := StoreOpts{
		Root: opts.StorageRoot,
		PathTransformFunc: opts.PathTransformFunc,
	}

	return &FileServer{
		FileServerOpts: opts,
		store: NewStore(storeOpts),
		quitCh: make(chan struct{}),
		peers: make(map[string]p2p.Peer),
	}
}

func (s *FileServer) broadcast(msg *Message) error{
	peers := []io.Writer{}

	for _, peer := range s.peers{
		peers = append(peers, peer)
	}

	multiWriter := io.MultiWriter(peers...) 
	
	return gob.NewEncoder(multiWriter).Encode(msg)
}

type Message struct{
	Payload any
}

func (s *FileServer) StoreData(key string, r io.Reader) error{
	// buf := new(bytes.Buffer)
	// tee := io.TeeReader(r, buf)

	buf := new(bytes.Buffer)
	msg := Message{
		Payload: []byte("storage_key"),
	}

	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers{
		if err := peer.Send(buf.Bytes()); err != nil{
			return err
		}
	}

	payload := []byte("THIS IS A LARGE FILE")
	for _, peer := range s.peers{
		if err := peer.Send(payload); err != nil{
			return err
		}
	}

	return nil
	
	// if err := s.store.Write(key ,tee); err != nil{
	// 	return err
	// }

	// p := &DataMessage{
	// 	Key: key,
	// 	Data: buf.Bytes(),
	// }

	// return s.broadcast(&Message{
	// 	From: "todo",
	// 	Payload: p,
	// })
	
}

func (s *FileServer) Stop(){
	close(s.quitCh)
}

func (s *FileServer) OnPeer(p p2p.Peer) error{
	s.peerLock.Lock()
	defer s.peerLock.Unlock()

	s.peers[p.RemoteAddr().String()] = p

	log.Printf("connected with remote %s", p.RemoteAddr())

	return nil
}

func (s *FileServer) loop(){
	defer func ()  {
		log.Println("File server stopped due to user quit action")
		s.Transport.Close()	
	}()

	for{
		select{
		case rpc := <- s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil{
				log.Println(err)
			}

			// if err := s.handleMessage(&m); err != nil{
			// 	log.Println(err)
			// }

			fmt.Printf("received %s:", string(msg.Payload.([]byte)))


		case <- s.quitCh:
			return
		}
	}
}

// func (s *FileServer) handleMessage(msg *Message) error{
// 	switch v := msg.Payload.(type){
// 	case *DataMessage:
// 		fmt.Printf("received data %+v\n", v)
// 	}

// 	return nil
// }

func (s *FileServer) bootstrapNetwork() error{
	for _, addr:= range s.BootstrapNodes{
		if len(addr) == 0 {
			continue
		}

		go func (addr string)  {
			fmt.Println("attempting to connect with remote:", addr)
			if err := s.Transport.Dial(addr); err != nil {
				log.Println("dial error:", err)
			}
		}(addr)
	}

	return nil
}

func (s *FileServer) Start() error{
	if err := s.Transport.ListenAndAccept(); err != nil {
		return err
	}

	s.bootstrapNetwork()

	s.loop()

	return nil
}