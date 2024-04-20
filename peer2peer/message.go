package p2p

import "net"

// Message represents any artbitrary data that is being sent over each
// transport between two nodes in the network
type RPC struct{
	From net.Addr
	Payload []byte
}