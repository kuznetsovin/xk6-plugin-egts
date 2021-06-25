package egts

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/kuznetsovin/egts-protocol/app/egts"
)

var actualPID uint32
var RecordNumber uint32

type EgtsClient struct {
	client       uint32
	actualPID    uint32
	RecordNumber uint32
}

func (c *EgtsClient) createNavPacket(ts time.Time, lat, lon float64) []byte {
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

	return c.createPacketWithRDS(rds)
}

func (c *EgtsClient) createNavPacketWithSensor(ts time.Time, lat, lon float64, sensVal uint32) []byte {
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
		Speed:               100,
		Direction:           172,
		Odometer:            []byte{0x00, 0x00, 0x00},
		DigitalInputs:       0,
		Source:              0,
	}

	sensorData := egts.SrAdSensorsData{
		DigitalInputsOctetExists1: "0",
		DigitalInputsOctetExists2: "0",
		DigitalInputsOctetExists3: "0",
		DigitalInputsOctetExists4: "0",
		DigitalInputsOctetExists5: "0",
		DigitalInputsOctetExists6: "0",
		DigitalInputsOctetExists7: "0",
		DigitalInputsOctetExists8: "0",
		AnalogSensorFieldExists1:  "1",
		AnalogSensorFieldExists2:  "1",
		AnalogSensorFieldExists3:  "1",
		AnalogSensorFieldExists4:  "1",
		AnalogSensorFieldExists5:  "1",
		AnalogSensorFieldExists6:  "1",
		AnalogSensorFieldExists7:  "1",
		AnalogSensorFieldExists8:  "1",
		AnalogSensor1:             sensVal,
	}

	rds := egts.RecordDataSet{
		egts.RecordData{
			SubrecordType:   egts.SrPosDataType,
			SubrecordLength: posData.Length(),
			SubrecordData:   &posData,
		},
		egts.RecordData{
			SubrecordType:   egts.SrAdSensorsDataType,
			SubrecordLength: sensorData.Length(),
			SubrecordData:   &sensorData,
		},
	}

	return c.createPacketWithRDS(rds)
}

func (c *EgtsClient) createNavPacketWithFuel(ts time.Time, lat, lon float64, fuelLvl uint32) []byte {
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
		Speed:               100,
		Direction:           172,
		Odometer:            []byte{0x00, 0x00, 0x00},
		DigitalInputs:       0,
		Source:              0,
	}

	fuelData := egts.SrLiquidLevelSensor{
		LiquidLevelSensorErrorFlag: "0",
		LiquidLevelSensorValueUnit: "00",
		RawDataFlag:                "0",
		LiquidLevelSensorNumber:    3,
		ModuleAddress:              1,
		LiquidLevelSensorData:      fuelLvl,
	}

	rds := egts.RecordDataSet{
		egts.RecordData{
			SubrecordType:   egts.SrPosDataType,
			SubrecordLength: posData.Length(),
			SubrecordData:   &posData,
		},
		egts.RecordData{
			SubrecordType:   egts.SrLiquidLevelSensorType,
			SubrecordLength: fuelData.Length(),
			SubrecordData:   &fuelData,
		},
	}

	return c.createPacketWithRDS(rds)
}

func (c *EgtsClient) createPacketWithRDS(rds egts.RecordDataSet) []byte {
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
		PacketIdentifier: uint16(atomic.AddUint32(&c.actualPID, 1)),
		PacketType:       1,
		ServicesFrameData: &egts.ServiceDataSet{
			egts.ServiceDataRecord{
				RecordLength:             rds.Length(),
				RecordNumber:             uint16(atomic.AddUint32(&c.RecordNumber, 1)),
				SourceServiceOnDevice:    "1",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "11",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "0",
				ObjectIDFieldExists:      "1",
				ObjectIdentifier:         c.client,
				SourceServiceType:        2,
				RecipientServiceType:     2,
				RecordDataSet:            rds,
			},
		},
	}
	result, err := p.Encode()
	if err != nil {
		fmt.Println(err)
	}
	return result
}

func NewClient(clientID uint32) *EgtsClient {
	return &EgtsClient{client: clientID}
}
