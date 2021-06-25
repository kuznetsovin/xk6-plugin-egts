package egts

import (
	"sync/atomic"
	"time"

	"github.com/kuznetsovin/egts-protocol/app/egts"
)

var actualPID uint32
var RecordNumber uint32

func createNavPacket(client uint32, ts time.Time, lat, lon float64) []byte {
	posData := egts.SrPosData{
		NavigationTime:      ts,
		Latitude:            lat,
		Longitude:           lon,
		ALTE:                "0",
		LOHS:                "0",
		LAHS:                "0",
		MV:                  "0",
		BB:                  "0",
		CS:                  "0",
		FIX:                 "0",
		VLD:                 "1",
		DirectionHighestBit: 1,
		AltitudeSign:        0,
		Speed:               200,
		Direction:           172,
		Odometer:            []byte{0x01, 0x00, 0x00},
		DigitalInputs:       0,
		Source:              0,
	}

	rds := egts.RecordDataSet{
		egts.RecordData{
			SubrecordType:   egts.SrPosDataType,
			SubrecordLength: posData.Length(),
			SubrecordData:   &posData,
		},
	}

	return createPacketWithRDS(client, rds)
}

func createPacketWithRDS(client uint32, rds egts.RecordDataSet) []byte {
	p := egts.Package{
		ProtocolVersion:  1,
		SecurityKeyID:    0,
		Prefix:           "00",
		Route:            "0",
		EncryptionAlg:    "00",
		Compression:      "0",
		Priority:         "11",
		HeaderLength:     11,
		HeaderEncoding:   0,
		PacketIdentifier: uint16(atomic.AddUint32(&actualPID, 1)),
		PacketType:       1,
		ServicesFrameData: &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordLength:             rds.Length(),
				RecordNumber:             uint16(atomic.AddUint32(&RecordNumber, 1)),
				SourceServiceOnDevice:    "1",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "11",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "0",
				ObjectIDFieldExists:      "1",
				ObjectIdentifier:         client,
				SourceServiceType:        2,
				RecipientServiceType:     2,
				RecordDataSet:            rds,
			},
		},
	}
	result, _ := p.Encode()
	return result
}
