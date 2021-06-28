package egts

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/kuznetsovin/egts-protocol/app/egts"
	"github.com/stretchr/testify/assert"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/stats"
)

func Test_createNavPacket(t *testing.T) {
	c := NewClient("", 133552)
	defer c.Close()
	packet := c.createPacket(time.Date(2018, time.July, 5, 20, 8, 53, 0, time.UTC), 55.55389399769574, 37.43236696287812, 1000, 1000)
	assert.Equal(t,
		[]byte{0x1, 0x0, 0x3, 0xb, 0x0, 0x4b, 0x0, 0x1, 0x0, 0x1, 0x3, 0x40, 0x0, 0x1, 0x0, 0x99, 0xb0,
			0x9, 0x2, 0x0, 0x2, 0x2, 0x10, 0x15, 0x0, 0xd5, 0x3f, 0x1, 0x10, 0x6f, 0x1c, 0x5, 0x9e, 0x7a,
			0xb5, 0x3c, 0x35, 0x1, 0xd0, 0x87, 0x2c, 0x1, 0x0, 0x0, 0x0, 0x0, 0x12, 0x1b, 0x0, 0x0, 0x0,
			0xff, 0xe8, 0x3, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0,
			0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x1b, 0x7, 0x0, 0x3, 0x1, 0x0, 0xe8, 0x3, 0x0, 0x0, 0x3a, 0x80},
		packet)

	p := egts.Package{}
	r, e := p.Decode(packet)
	if !assert.NoError(t, e) {
		assert.Equal(t, 0, r)
	}

	packet = c.createPacket(time.Date(2018, time.July, 5, 20, 8, 53, 0, time.UTC), 55.55389399769574, 37.43236696287812, 0, 0)
	assert.Equal(t,
		[]byte{0x1, 0x0, 0x3, 0xb, 0x0, 0x23, 0x0, 0x2, 0x0, 0x1, 0xce, 0x18, 0x0, 0x2, 0x0, 0x99, 0xb0, 0x9,
			0x2, 0x0, 0x2, 0x2, 0x10, 0x15, 0x0, 0xd5, 0x3f, 0x1, 0x10, 0x6f, 0x1c, 0x5, 0x9e, 0x7a, 0xb5,
			0x3c, 0x35, 0x1, 0xd0, 0x87, 0x2c, 0x1, 0x0, 0x0, 0x0, 0x0, 0x18, 0x7b},
		packet)

	p = egts.Package{}
	r, e = p.Decode(packet)
	if !assert.NoError(t, e) {
		assert.Equal(t, 0, r)
	}
}

func TestSendPacket(t *testing.T) {
	addr := ":1111"

	r := egts.Package{
		ProtocolVersion:  1,
		SecurityKeyID:    0,
		Prefix:           "00",
		Route:            "0",
		EncryptionAlg:    "00",
		Compression:      "0",
		Priority:         "11",
		HeaderLength:     11,
		HeaderEncoding:   0,
		FrameDataLength:  3,
		PacketIdentifier: 0,
		PacketType:       egts.PtResponsePacket,
		HeaderCheckSum:   74,
		ServicesFrameData: &egts.PtResponse{
			ResponsePacketID: 1,
			ProcessingResult: egtsPcOk,
		},
	}
	ack, err := r.Encode()
	if !assert.NoError(t, err) {
		return
	}

	stateChannel := make(chan bool, 1)
	mockExportServer(addr, ack, stateChannel)
	<-stateChannel
	c := NewClient(addr, 133552)
	defer c.Close()

	ctx := lib.WithState(context.Background(), &lib.State{
		Samples: make(chan<- stats.SampleContainer, 2),
	})

	assert.NoError(t, c.SendPacket(ctx, 55.55389399769574, 37.43236696287812, 0, 0))

}

func mockExportServer(addr string, response []byte, ch chan bool) net.Listener {
	listener, _ := net.Listen("tcp", addr)
	go func() {
		ch <- true
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			} else {
				defer conn.Close()
				buf := make([]byte, 1024)
				l, err := conn.Read(buf)
				if err != nil {
					return
				}
				if l > 0 {
					if _, err = conn.Write(response); err != nil {
						return
					}
				}

			}
		}
	}()

	return listener
}
