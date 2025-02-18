package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct{
	// The underlying connection of the peer i.e TCP
	net.Conn

	// if we dial a conn => outbound = true 
	// if we accept and retrieve a conn => outbound = false 
	outbound 	bool

	waitGroup *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{
	return &TCPPeer{
		Conn:		conn,
		outbound: 	outbound,
		waitGroup: &sync.WaitGroup{},
	}
}

func (p *TCPPeer) CloseStream(){
	p.waitGroup.Done()
}

func (p *TCPPeer) Send(b []byte) error{
	_, err := p.Conn.Write(b)
	return err
}

type TCPTransportOpts struct{
	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
	OnPeer func(Peer) error
}

type TCPTransport struct{
	TCPTransportOpts
	ListenAddress 		string
	listener 			net.Listener
	rpcCh 				chan RPC

}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcCh: make(chan RPC, 1024),
	}
}

// Addr implements the transport interface returning the address
// of which the transport is accepting connections 
func (t *TCPTransport) Addr()string{
	return t.ListenAddr
}

// consume implements the transport interface,
// wich will return read only channel for reading the incoming messages
// received from another peer in the network.
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcCh
}

// Close implements the transport interface.
func (t *TCPTransport) Close() error{
	return t.listener.Close()
}

// Dial implements the transport interface.
func (t *TCPTransport) Dial(addr string) error{
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}

	go t.handleConn(conn, true)

	return nil
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	log.Printf("TCP transport listening on port: %s\n", t.ListenAddr)

	return nil

}

func (t *TCPTransport) startAcceptLoop(){
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed){
			return
		}
		if err != nil{
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("New incoming connection : %v\n", conn)
		go t.handleConn(conn, false)
	}	
}

func (t *TCPTransport) handleConn(conn net.Conn, outbound bool){
	var err error
	defer func ()  {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
		}()

	peer := NewTCPPeer(conn, outbound)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	for{
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil{
			return
		}

		rpc.From = conn.RemoteAddr().String()

		if rpc.Stream{
			peer.waitGroup.Add(1)
			fmt.Printf("[%s] incoming stream, waiting ...\n", conn.RemoteAddr())
			peer.waitGroup.Wait()
			fmt.Printf("[%s] stream closed, resumiong read loop", conn.RemoteAddr())
			continue
		}

		t.rpcCh <- rpc
	}
}