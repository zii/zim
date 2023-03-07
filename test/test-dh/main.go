package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"

	"zim.cn/base"

	"github.com/monnand/dhkx"
)

/*
找素数(prime)的原根(generator)的方法:
1. https://blog.csdn.net/weixin_44932880/article/details/106555427
如果g是原根,一定有 g^x mod p != 1
x是p−1最大约数, 那么g就一定是原根

2. https://www.dounaite.com/article/62566b61ae87fd3f795dca96.html

3. 查素数原根表(较小)
http://blog.leanote.com/post/rockdu/TX06

4. Oakley Default Group(较专业)
https://www.rfc-editor.org/rfc/rfc2409.html#page-21
*/

/*
web端加密方案:

https://www.infoq.cn/article/o3agjb1nr16vxg7iw3ik
*/

func test1() {
	b := new(big.Int).Exp(big.NewInt(29), big.NewInt(30), big.NewInt(71))
	fmt.Println("b:", b.Int64(), b.Bytes(), b.String())
	fmt.Println("expmod:", expmod(17, 99, 99))
}

func expmod(g, k, m int) int {
	a := 1
	for k > 0 {
		if k%2 == 1 {
			a = a * g % m
		}
		g = g * g % m
		k >>= 1
	}
	return a
}

type peer struct {
	priv  *dhkx.DHKey
	group *dhkx.DHGroup
	pub   *dhkx.DHKey
}

func newPeer(g *dhkx.DHGroup) *peer {
	ret := new(peer)
	ret.priv, _ = g.GeneratePrivateKey(nil)
	ret.group = g
	return ret
}

func (self *peer) getPubKey() []byte {
	return self.priv.Bytes()
}

func (self *peer) recvPeerPubKey(pub []byte) {
	pubKey := dhkx.NewPublicKey(pub)
	self.pub = pubKey
}

func (self *peer) getKey() []byte {
	k, err := self.group.ComputeKey(self.pub, self.priv)
	if err != nil {
		return nil
	}
	return k.Bytes()
}

func testdh() {
	g, err := dhkx.GetGroup(14)
	base.Raise(err)
	fmt.Println("g:", g.G(), g.P())
	p1 := newPeer(g)
	p2 := newPeer(g)

	pub1 := p1.getPubKey()
	pub2 := p2.getPubKey()
	fmt.Println("pub1:", len(pub1), base64.StdEncoding.EncodeToString(pub1))
	fmt.Println("pub2:", len(pub2), hex.EncodeToString(pub2))

	p1.recvPeerPubKey(pub2)
	p2.recvPeerPubKey(pub1)

	key1 := p1.getKey()
	key2 := p2.getKey()
	fmt.Println("key1:", len(key1))
	fmt.Println("key2:", len(key2))
}

func main() {
	test1()
	//testdh()
}
