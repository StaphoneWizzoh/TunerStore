package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	p2p "github.com/StaphoneWizzoh/TunerStore/peer2peer"
)

type FileServerOpts struct{
	EncKey []byte
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

func (s *FileServer) stream(msg *Message) error{
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

type MessageStoreFile struct{
	Key string
	Size int64
}

func (s *FileServer) broadcast(msg *Message) error{
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		return err
	}

	for _, peer := range s.peers{
		peer.Send([]byte{p2p.IncomingMessage})
		if err := peer.Send(buf.Bytes()); err != nil{
			return err
		}
	}

	return nil
}

type MessageGetFile struct{
	key string
}

func (s *FileServer) Get(key string)(io.Reader, error){
	if s.store.Has(key){
	fmt.Printf("[%s]serving file (%s) from local disk\n", s.Transport.Addr(), key)

		_, r, err :=s.store.Read(key)
		return r, err
	}

	fmt.Printf("[%s]don't have file (%s) locally, fetching from network...\n", s.Transport.Addr(), key)

	msg := Message{
		Payload: MessageGetFile{
			key: key,
		},
	}

	if err := s.broadcast(&msg); err != nil{
		return nil, err
	}

	// to be removed
	time.Sleep(time.Millisecond * 500)

	for _, peer := range s.peers{
		// first, read the file size so we can limit the amount of  bytes we read from connection
		// of hanging from continuous reading in order to prevent the amount
		var fileSize int64 
		binary.Read(peer, binary.LittleEndian, &fileSize)

		n, err := s.store.WriteDecrypt(s.EncKey, key, io.LimitReader(peer, fileSize))
		// n, err := s.store.Write(key, io.LimitReader(peer, fileSize))
		if err != nil{
			return nil, err
		}
		
		fmt.Printf("[%s] received (%d) bytes over the network from (%s):",s.Transport.Addr(), n, peer.RemoteAddr())

		peer.CloseStream()
	}

	_ ,r, err := s.store.Read(key)
	return r, err
}


func (s *FileServer) Store(key string, r io.Reader) error{
	fileBuffer := new(bytes.Buffer)
	tee := io.TeeReader(r, fileBuffer)

	size, err := s.store.Write(key ,tee)
	if err != nil{
		return err
	}

	msg := Message{
		Payload: MessageStoreFile{
			Key: key,
			Size: size + 16, 
		},
	}

	if err := s.broadcast(&msg); err != nil{
		return err
	}

	// TODO: fix this sleeping
	time.Sleep(5 * time.Millisecond)

	peers := []io.Writer{}
	for _, peer := range s.peers{
		peers = append(peers, peer)
	}
	mu := io.MultiWriter(peers...)
	mu.Write([]byte{p2p.IncomingStream})
	n, err := copyEncrypt(s.EncKey, fileBuffer, mu)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] received and written %d bytes to disk\n",s.Transport.Addr(), n)

	return nil
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
		log.Println("File server stopped due to error or user quit action")
		s.Transport.Close()	
	}()

	for{
		select{
		case rpc := <- s.Transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil{
				log.Println("decoding error: ",err)
			}

			if err := s.handleMessage(rpc.From, &msg); err != nil{
				log.Println("handle message error: ", err)
				
			}
		case <- s.quitCh:
			return
		}
	}
}

func (s *FileServer) handleMessage(from string, msg *Message) error{
	switch v := msg.Payload.(type){
	case MessageStoreFile:
		return s.handleMessageStoreFile(from, v)
	case MessageGetFile:
		return s.handleMessageGetFile(from, v)
	}

	return nil
}

func (s *FileServer) handleMessageStoreFile(from string, msg MessageStoreFile) error{
	peer, ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer (%s) could not be found in the peer map", from)
	}

	n, err := s.store.Write(msg.Key, io.LimitReader(peer, msg.Size))
	if err != nil {
		return err
	}

	log.Printf("[%s] written (%d) bytes to disk\n",s.Transport.Addr(), n)

	peer.CloseStream()

	return nil
}

func (s *FileServer) handleMessageGetFile(from string, msg MessageGetFile)error{
	
	if !s.store.Has(msg.key){
		return fmt.Errorf("[%s] need to serve file (%s) but it does not exist on disk",s.Transport.Addr() , msg.key)
	}

	fmt.Printf("[%s] serving file (%s) over the network\n",s.Transport.Addr(), msg.key)
	
	fileSize, r, err := s.store.Read(msg.key)
	if err != nil {
		return err
	}

	// checking if the reader is a readcloser, if it is then closing it.
	rc, ok := r.(io.ReadCloser)
	if ok{
		fmt.Println("closing readClose")
		defer rc.Close()
	}

	peer, ok := s.peers[from]
	if !ok{
		return fmt.Errorf("peer %s not in map", from)
	}

	// first send the "IncomingStream" byte to the peer
	// and the we can send the file size as an int64
	peer.Send([]byte{p2p.IncomingStream})
	// var fileSize int64 = 32
	binary.Write(peer, binary.LittleEndian, fileSize)

	n ,err := io.Copy(peer, r)
	if err != nil {
		return err
	}

	fmt.Printf("[%s] has written %d bytes over the network to %s\n", s.Transport.Addr(), n, from)

	return nil
}

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

func init(){
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
}