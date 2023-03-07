package dh

import "github.com/monnand/dhkx"

type Peer struct {
	Priv  *dhkx.DHKey
	Group *dhkx.DHGroup
	Pub   *dhkx.DHKey
}

func NewPeer(g *dhkx.DHGroup) *Peer {
	if g == nil {
		g, _ = dhkx.GetGroup(14)
	}
	ret := new(Peer)
	ret.Priv, _ = g.GeneratePrivateKey(nil)
	ret.Group = g
	return ret
}

func (self *Peer) GetPubKey() []byte {
	return self.Priv.Bytes()
}

func (self *Peer) RecvPeerPubKey(pub []byte) {
	pubKey := dhkx.NewPublicKey(pub)
	self.Pub = pubKey
}

// 获得共享密钥
func (self *Peer) GetKey() []byte {
	k, err := self.Group.ComputeKey(self.Pub, self.Priv)
	if err != nil {
		return nil
	}
	return k.Bytes()
}
