package ping

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	ping "github.com/sparrc/go-ping"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the ping plugin send ping
var (
	Plugin = taskplugin.New("ping", "0.1", exec,
		taskplugin.WithConfig(validConfig, Config{}),
	)
)

// based on Statistics struct from github.com/sparrc/go-ping /w json tags
type pingStats struct {
	PacketsRecv int             `json:"packets_received"`
	PacketsSent int             `json:"packets_sent"`
	PacketLoss  float64         `json:"packet_loss"`
	IPAddr      string          `json:"ip_addr"`
	Rtts        []time.Duration `json:"rtts"`
	MinRtt      time.Duration   `json:"min_rtt"`
	MaxRtt      time.Duration   `json:"max_rtt"`
	AvgRtt      time.Duration   `json:"avg_rtt"`
	StdDevRtt   time.Duration   `json:"std_dev_rtt"`
}

// Config is the configuration needed to send a ping
type Config struct {
	Hostname string `json:"hostname"`
	Count    string `json:"count,omitempty"`
	Interval string `json:"interval_second,omitempty"`
}

func validConfig(config interface{}) error {
	cfg := config.(*Config)

	if cfg.Hostname == "" {
		return errors.New("hostname is missing")
	}

	if cfg.Count != "" {
		if _, err := strconv.ParseUint(cfg.Count, 10, 64); err != nil {
			return fmt.Errorf("can't parse count field %q: %s", cfg.Count, err.Error())
		}
	}

	if cfg.Interval != "" {
		if _, err := strconv.ParseUint(cfg.Interval, 10, 64); err != nil {
			return fmt.Errorf("can't parse interval_second field %q: %s", cfg.Interval, err.Error())
		}
	}

	return nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*Config)

	pinger, err := ping.NewPinger(cfg.Hostname)
	if err != nil {
		return nil, nil, fmt.Errorf("can't initiate ping: %s", err.Error())
	}

	pinger.Count = pingDefault(cfg.Count)
	pinger.Interval = time.Duration(pingDefault(cfg.Interval)) * time.Second

	// Run() is blocking until count is done
	pinger.Run()

	so := pinger.Statistics()

	return &pingStats{
		PacketsRecv: so.PacketsRecv,
		PacketsSent: so.PacketsSent,
		PacketLoss:  so.PacketLoss,
		IPAddr:      so.IPAddr.String(),
		Rtts:        so.Rtts,
		MinRtt:      so.MinRtt,
		MaxRtt:      so.MaxRtt,
		AvgRtt:      so.AvgRtt,
		StdDevRtt:   so.StdDevRtt,
	}, cfg, nil
}

func pingDefault(c string) int {
	if c == "" || c == "0" {
		return 1
	}

	// count or interval is already checked, at validConfig() lvl
	// values must be correct so errors are not evaluated
	count, _ := strconv.Atoi(c)

	return count
}
