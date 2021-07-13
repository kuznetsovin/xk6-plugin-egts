package egts

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"sync/atomic"
	"time"

	"github.com/kuznetsovin/egts-protocol/libs/egts"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/lib/metrics"
	"go.k6.io/k6/stats"
)

const egtsPcOk = 0

type EgtsClient struct {
	Conn         net.Conn
	Client       uint32
	actualPID    uint32
	recordNumber uint32
}

func (c *EgtsClient) SendPacket(ctx context.Context, lat, lon float64, sensVal uint32, fuelLvl uint32) error {
	state := lib.GetState(ctx)
	if state == nil {
		return errors.New("state is empty")
	}

	if c.Conn == nil {
		return errors.New("empty connection")
	}
	p := c.createPacket(time.Now().UTC(), lat, lon, sensVal, fuelLvl)
	receivedTime := time.Now().UTC()
	n, err := c.Conn.Write(p)
	if err != nil {
		return err
	}
	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Time:   receivedTime,
		Metric: metrics.DataSent,
		Value:  float64(n),
	})

	if n != len(p) {
		return errors.New("sending not full packet")
	}

	response := make([]byte, 1024)
	if n, err = c.Conn.Read(response); err != nil {
		return err
	}
	now := time.Now().UTC()
	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Time:   now,
		Metric: metrics.DataReceived,
		Value:  float64(n),
	})
	ackPacket := egts.Package{}
	if _, err = ackPacket.Decode(response[:n]); err != nil {
		return err
	}

	ack, ok := ackPacket.ServicesFrameData.(*egts.PtResponse)
	if !ok {
		stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
			Time:   now,
			Metric: EgtsPacketFailed,
			Value:  1.0,
		})
		return errors.New("incorrect ack packet")
	}

	if ack.ProcessingResult != egtsPcOk {
		stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
			Time:   now,
			Metric: EgtsPacketFailed,
			Value:  1.0,
		})
		return fmt.Errorf("incorrect processing result: %d", ack.ProcessingResult)
	}
	if ack.ResponsePacketID != uint16(c.actualPID) {
		stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
			Time:   now,
			Metric: EgtsPacketFailed,
			Value:  1.0,
		})
		return fmt.Errorf("incorrect check packet id: %d != %d", ack.ResponsePacketID, c.actualPID)
	}

	if ack.SDR != nil {
		for _, rec := range *ack.SDR.(*egts.ServiceDataSet) {
			for _, subRec := range rec.RecordDataSet {
				if subRec.SubrecordType == egts.SrRecordResponseType {
					if response, ok := subRec.SubrecordData.(*egts.SrResponse); ok {
						if response.RecordStatus != 0 {
							stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
								Time:   now,
								Metric: EgtsPacketFailed,
								Value:  1.0,
							})
							return fmt.Errorf("incorrect server processing result. Record status: %d", response.RecordStatus)
						}
					}
				}
			}
		}
	}
	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Time:   now,
		Metric: EgtsProcessTime,
		Value:  now.Sub(receivedTime).Seconds(),
	})

	stats.PushIfNotDone(ctx, state.Samples, stats.Sample{
		Time:   now,
		Metric: EgtsPackets,
		Value:  1.0,
	})

	return nil
}

func (c *EgtsClient) createPacket(ts time.Time, lat, lon float64, sensVal uint32, fuelLvl uint32) []byte {
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

	if sensVal > 0 {
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
		rds = append(rds, egts.RecordData{
			SubrecordType:   egts.SrAdSensorsDataType,
			SubrecordLength: sensorData.Length(),
			SubrecordData:   &sensorData,
		})
	}

	if fuelLvl > 0 {
		fuelData := egts.SrLiquidLevelSensor{
			LiquidLevelSensorErrorFlag: "0",
			LiquidLevelSensorValueUnit: "00",
			RawDataFlag:                "0",
			LiquidLevelSensorNumber:    3,
			ModuleAddress:              1,
			LiquidLevelSensorData:      fuelLvl,
		}
		rds = append(rds, egts.RecordData{
			SubrecordType:   egts.SrLiquidLevelSensorType,
			SubrecordLength: fuelData.Length(),
			SubrecordData:   &fuelData,
		})
	}

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
				RecordNumber:             uint16(atomic.AddUint32(&c.recordNumber, 1)),
				SourceServiceOnDevice:    "1",
				RecipientServiceOnDevice: "0",
				Group:                    "0",
				RecordProcessingPriority: "11",
				TimeFieldExists:          "0",
				EventIDFieldExists:       "0",
				ObjectIDFieldExists:      "1",
				ObjectIdentifier:         c.Client,
				SourceServiceType:        2,
				RecipientServiceType:     2,
				RecordDataSet:            rds,
			},
		},
	}
	result, err := p.Encode()
	if err != nil {
		log.Println(err)
	}
	return result
}

func (c *EgtsClient) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}

func NewClient(addr string, clientID uint32) *EgtsClient {
	var err error

	client := &EgtsClient{Client: clientID}
	if addr != "" {
		client.Conn, err = net.Dial("tcp", addr)
		if err != nil {
			log.Fatal(err)
		}

	}

	return client
}
