package scanner

import (
	"context"
	"prometheus-net-discovery/internal/netops/host"
	"prometheus-net-discovery/internal/netops/scanner/scan"
	"time"
)

type Scanner struct {
	// Network name
	Network string `yaml:"network" validate:"required"`

	// Interval of scan
	Interval string `yaml:"interval" validate:"required"`

	// If deep enabled — scanner will scan all hosts of targets — online and offline.
	Deep bool `yaml:"deep" validate:"omitempty,required"`

	// Concurrency of scan
	Concurrency int `yaml:"concurrency" validate:"omitempty,min=0"`

	// Targets of scan (subnets and IPs)
	Targets []string `yaml:"targets" validate:"required,min=1"`

	// Enable discovery hosts by ICMP ping
	Ping bool `yaml:"ping" validate:"omitempty"`

	// Lists of ports to scan
	Ports []string `yaml:"ports" validate:"omitempty"`
}

func (s *Scanner) Scan(ctx context.Context) ([]*host.Host, time.Duration, error) {
	scanOptions := []scan.Option{
		scan.WithTargets(s.Targets),
	}

	if s.Concurrency > 0 {
		scanOptions = append(scanOptions, scan.WithCuncurrency(s.Concurrency))
	}
	if s.Deep {
		scanOptions = append(scanOptions, scan.WithDeep())
	}
	if s.Ping {
		scanOptions = append(scanOptions, scan.WithPing())
	}
	if len(s.Ports) > 0 {
		scanOptions = append(scanOptions, scan.WithPorts(s.Ports))
	}

	scan, err := scan.NewScan(scanOptions...)
	if err != nil {
		return nil, 0, err
	}

	return scan.Scan(ctx)
}
