package egts

import (
	"context"
	"log"

	"go.k6.io/k6/js/modules"
)

func init() {
	modules.Register("k6/x/egts", new(Egts))
}

type Egts struct{}

func (*Egts) NewClient(addr string, clientID uint32) *EgtsClient {
	return NewClient(addr, clientID)
}

func (*Egts) SendPacket(ctx context.Context, client *EgtsClient, lat, lon float64, sensVal uint32, fuelLvl uint32) {
	if err := client.SendPacket(ctx, lat, lon, sensVal, fuelLvl); err != nil {
		log.Println(err)
	}
}

func (*Egts) CloseConnection(client *EgtsClient) {
	client.Close()
}
