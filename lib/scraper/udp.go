package scraper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"net/url"
	"time"
)

type udpTransport struct {
	c  *net.UDPConn
	tx map[uint32]chan []byte
}

func newUDPTransport() *udpTransport {
	a, err := net.ResolveUDPAddr("udp", "0.0.0.0:25010")
	if err != nil {
		panic(err)
	}

	c, err := net.ListenUDP("udp", a)
	if err != nil {
		panic(err)
	}

	fmt.Printf("listening on %s\n", c.LocalAddr().String())

	return &udpTransport{
		c:  c,
		tx: make(map[uint32]chan []byte),
	}
}

func (t *udpTransport) run() {
	for {
		fmt.Printf("run()\n")

		b := make([]byte, 65535)

		n, a, err := t.c.ReadFrom(b)
		if err != nil {
			panic(err)
		}

		fmt.Printf("received: %d %q\n", n, a.String())

		if n < 8 {
			continue
		}

		var h responseHeader
		if err := binary.Read(bytes.NewReader(b), binary.BigEndian, &h); err != nil {
			panic(err)
		}

		fmt.Printf("incoming: %#v\n", h)

		if c, ok := t.tx[h.TxID]; !ok {
			fmt.Printf("unrecognised transaction id in response: %d\n", h.TxID)
		} else {
			c <- b
		}
	}
}

func (t *udpTransport) req(a *net.UDPAddr, out outgoing, in incoming) error {
	c := make(chan []byte, 1)

	txid := rand.Uint32()

	out.setTxid(txid)

	t.tx[txid] = c
	defer func() {
		delete(t.tx, txid)
	}()

	if _, err := t.c.WriteTo(out.build(), a); err != nil {
		return err
	}

	fmt.Printf("wrote packet to %s\n", a.String())

	select {
	case r := <-c:
		return in.parse(r)
	case <-time.After(time.Second * 5):
		return fmt.Errorf("timed out waiting for response")
	}
}

func (t *udpTransport) newUDPTracker(u *url.URL) (Backend, error) {
	return newUDPTracker(t, u)
}

type udpTracker struct {
	t   *udpTransport
	u   *url.URL
	a   *net.UDPAddr
	cid uint64
	err error
}

func newUDPTracker(x *udpTransport, u *url.URL) (Backend, error) {
	t := &udpTracker{
		t:   x,
		u:   u,
		cid: uint64(0x41727101980),
	}

	raddr, err := net.ResolveUDPAddr("udp", u.Host)
	if err != nil {
		t.err = err
		return t, nil
	}

	t.a = raddr

	var (
		creq connectionRequest
		cres connectionResponse
	)

	if err := t.req(&creq, &cres); err != nil {
		t.err = err
		return t, nil
	}

	if cres.Action != 0 {
		t.err = fmt.Errorf("action is not set to connect (0), instead got %d", cres.Action)
		return t, nil
	}

	t.cid = cres.CID

	fmt.Printf("all done! cid is %d\n", t.cid)

	return t, nil
}

func (t *udpTracker) BatchSize() int {
	return 50
}

func (t *udpTracker) String() string {
	return t.u.String()
}

func (t *udpTracker) Scrape(hashes []Hash) (map[Hash]Scrape, error) {
	return nil, ErrUnimplemented
}

func (t *udpTracker) req(out outgoing, in incoming) error {
	out.setCid(t.cid)

	return t.t.req(t.a, out, in)
}

type outgoing interface {
	setCid(cid uint64)
	setTxid(txid uint32)
	build() []byte
}

type incoming interface {
	parse(b []byte) error
}

type requestHeader struct {
	cid    uint64
	action uint32
	txid   uint32
}

func (h *requestHeader) setCid(cid uint64) {
	h.cid = cid
}

func (h *requestHeader) setTxid(txid uint32) {
	h.txid = txid
}

type connectionRequest struct {
	requestHeader
}

func (c *connectionRequest) build() []byte {
	b := bytes.NewBuffer(nil)

	if err := binary.Write(b, binary.BigEndian, c); err != nil {
		panic(err)
	}

	return b.Bytes()
}

type responseHeader struct {
	Action uint32
	TxID   uint32
}

type connectionResponse struct {
	responseHeader
	CID uint64
}

func (c *connectionResponse) parse(b []byte) error {
	return binary.Read(bytes.NewReader(b), binary.BigEndian, c)
}
