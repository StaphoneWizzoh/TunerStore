package p2p

import (
	"fmt"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct{
	// conn is the underlying connection of the peer
	conn 		net.Conn

	// if we dial a conn => outbound = true 
	// if we accept and retrieve a conn => outbound = false 
	outbound 	bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{
	return &TCPPeer{
		conn:		conn,
		outbound: 	outbound,
	}
}

// Close implements the peer interface
func (p *TCPPeer) Close() error{
	return p.conn.Close()
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
		rpcCh: make(chan RPC),
	}
}

// consume implements the transport interface,
// wich will return read only channel for reading the incoming messages
// received from another peer in the network
func (t *TCPTransport) Consume() <- chan RPC{
	return t.rpcCh
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	t.listener, err = net.Listen("tcp", t.ListenAddress)
	if err != nil {
		return err
	}

	go t.startAcceptLoop()

	return nil

}

func (t *TCPTransport) startAcceptLoop(){
	for {
		conn, err := t.listener.Accept()
		if err != nil{
			fmt.Printf("TCP accept error: %s\n", err)
		}

		fmt.Printf("New incoming connection : %v\n", conn)
		go t.handleConn(conn)
	}	
}

func (t *TCPTransport) handleConn(conn net.Conn){
	var err error
	defer func ()  {
		fmt.Printf("dropping peer connection: %s", err)
		conn.Close()
		}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	// Read loop
	rpc := RPC{}
	for{
		if err := t.Decoder.Decode(conn, &rpc); err != nil{
			fmt.Printf("TCP error: %s", err)
			continue
		}

		rpc.From = conn.RemoteAddr()
		t.rpcCh <- rpc
	}
}