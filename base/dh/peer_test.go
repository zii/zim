package dh

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"testing"
)

func Test1(_ *testing.T) {
	p1 := NewPeer(nil)
	p2 := NewPeer(nil)

	pub1 := p1.GetPubKey()
	pub2 := p2.GetPubKey()
	fmt.Println("pub1:", len(pub1), base64.StdEncoding.EncodeToString(pub1))
	fmt.Println("pub2:", len(pub2), hex.EncodeToString(pub2))

	p1.RecvPeerPubKey(pub2)
	p2.RecvPeerPubKey(pub1)

	key1 := p1.GetKey()
	key2 := p2.GetKey()
	fmt.Println("key1:", len(key1))
	fmt.Println("key2:", len(key2))
}
