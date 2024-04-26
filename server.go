package main

import (
	"fmt"
	"log"

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
	}
}

func (s *FileServer) Stop(){
	close(s.quitCh)
}

func (s *FileServer) loop(){
	defer func ()  {
		log.Println("File server stopped due to user quit action")
		s.Transport.Close()	
	}()

	for{
		select{
		case msg := <- s.Transport.Consume():
			fmt.Println(msg)
		case <- s.quitCh:
			return
		}
	}
}

func (s *FileServer) bootstrapNetwork() error{
	for _, addr:= range s.BootstrapNodes{
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