package egts

import (
	"go.k6.io/k6/stats"
)

var (
	EgtsPackets      = stats.New("egts_packets", stats.Counter)
	EgtsPacketFailed = stats.New("egts_packets_failed", stats.Rate)
	EgtsProcessTime  = stats.New("egts_packets_process_time", stats.Trend, stats.Time)
)
