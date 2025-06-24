package discovery

import (
	"context"
	"fmt"
	"prometheus-net-discovery/internal/config"
	"prometheus-net-discovery/internal/netops/scanner"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type Discovery struct {
	InstanceID string             `yaml:"instanceId"`
	Scanners   []*scanner.Scanner `yaml:"scanners"`
	Reports    map[string]*Report
	Metrics    *Metrics
}

func New(c *config.Config) *Discovery {
	m := NewMetrics(c.Global.InstanceID)

	return &Discovery{
		InstanceID: c.Global.InstanceID,
		Scanners:   c.Scanners,
		Reports:    make(map[string]*Report),
		Metrics:    m,
	}
}

func (d *Discovery) Run(ctx context.Context) error {
	d.MetricsRegister()

	return d.runScanners(ctx)
}

func (d *Discovery) runScanners(ctx context.Context) error {
	grp, ctx := errgroup.WithContext(ctx)

	// start scann to each scanners
	for _, scan := range d.Scanners {
		s := scan

		d.Reports[s.Network] = &Report{
			Network: s.Network,
		}

		// discovery network
		grp.Go(func() error {
			return d.runScanner(ctx, s)
		})
	}

	return grp.Wait()
}

func (d *Discovery) runScanner(ctx context.Context, s *scanner.Scanner) error {
	interval, err := time.ParseDuration(s.Interval)
	if err != nil {
		return fmt.Errorf("unable to parse discovery interval: %w", err)
	}

	runNow := time.Second
	ticker := time.NewTicker(runNow)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Infof("discovery network %s stopped", s.Network)

			return nil
		case <-ticker.C:
			log.Infof("scan network: %s", s.Network)

			d.Metrics.DiscoveryRunning.WithLabelValues(s.Network).Set(1)

			// run scan network
			hosts, duration, err := s.Scan(ctx)
			if err != nil {
				if strings.Contains(err.Error(), "nmap scan interrupted") {
					log.Infof("discovery network %s stopped", s.Network)

					return nil
				}

				return fmt.Errorf("unable to scan network %s: %w", s.Network, err)
			}

			// debug log
			for _, h := range hosts {
				fields := log.Fields{
					"network":          s.Network,
					"host":             h.Address,
					"icmp":             h.ICMP,
					"duration_seconds": duration.Seconds(),
				}
				log.WithFields(fields).Debug("discovered host")
				for _, p := range h.Ports {
					fieldsPorts := log.Fields{
						"protocol": p.Protocol,
						"port":     p.Port,
						"srv":      p.Service,
					}
					log.WithFields(fields).WithFields(fieldsPorts).Debug("discovered ports of host")
				}
			}

			// set metrics
			d.Reports[s.Network].DiscoveredHosts = hosts
			d.Metrics.CollectionDuration.WithLabelValues(s.Network).Observe(duration.Seconds())
			d.Metrics.CollectionDurationLast.WithLabelValues(s.Network).Set(duration.Seconds())
			d.Metrics.CollectionCount.WithLabelValues(s.Network).Inc()
			d.Metrics.DiscoveryRunning.WithLabelValues(s.Network).Set(0)

			// wait next scan (sleep)
			log.Infof("waiting next scan: %s", s.Network)
			ticker.Reset(interval)
		}
	}
}
