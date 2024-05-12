package p2p

import (
	"encoding/gob"
	"io"
)

type Decoder interface{
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct{}

func (dec GOBDecoder) Decode(r io.Reader, msg *RPC)error{
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct{}

func (dec DefaultDecoder) Decode(r io.Reader, msg *RPC)error{

	peerBuf := make([]byte, 1)
	if _, err := r.Read(peerBuf); err != nil{
		return err
	}

	// In case of a stream, we are not decoding what is being sent 
	// over a network. We are just  setting stream true so we can
	// handle that in our logic
	stream := peerBuf[0] == IncomingStream
	if stream{
		msg.Stream = true
		return nil
	}

	buf := make([]byte, 1028)
	n, err := r.Read(buf)

	if err != nil {
		return err
	}

	msg.Payload = buf[:n]
	return nil
}