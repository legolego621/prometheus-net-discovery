package scan

import (
	"context"
	"errors"
	"fmt"
	"prometheus-net-discovery/internal/netops/host"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/Ullaakut/nmap/v3"
	"github.com/alitto/pond/v2"
)

type Scan struct {
	concurrency int
	deep        bool
	ping        bool
	ports       []string
	targets     []string
}

func NewScan(options ...Option) (*Scan, error) {
	s := &Scan{
		concurrency: 0,
		deep:        false,
		ping:        false,
		ports:       []string{},
		targets:     []string{},
	}

	for _, opt := range options {
		opt(s)
	}

	// simple validate
	if len(s.targets) == 0 {
		return nil, errors.New("failed to start scan: targets is empty")
	}
	if !s.ping && len(s.ports) == 0 {
		return nil, errors.New("failed to start scan: ports and ping scan are empty")
	}
	if s.concurrency > 0 && len(s.ports) == 0 {
		return nil, errors.New("failed to start scan: unsupported ports empty when concurrency is set")
	}

	return s, nil
}

func (s *Scan) Scan(ctx context.Context) ([]*host.Host, time.Duration, error) {
	var hostsPing, hostsPorts []*host.Host
	var durationPing, durationPorts time.Duration
	var err error

	// ports scan
	if len(s.ports) > 0 {
		if s.concurrency > 0 {
			hostsPorts, durationPorts, err = s.portsScanParralel(ctx)
		} else {
			hostsPorts, durationPorts, err = s.portsScan(ctx, s.targets)
		}
		if err != nil {
			return nil, 0, err
		}
	}

	// icmp ping scan
	if s.ping {
		hostsPing, durationPing, err = s.pingScan(ctx)
		if err != nil {
			return nil, 0, err
		}
	}

	seen := make(map[string]bool)
	var hostsFull []*host.Host

	// add icmp to ports hosts
	for _, h := range hostsPorts {
		if slices.IndexFunc(hostsPing, func(p *host.Host) bool { return p.Address == h.Address }) != -1 {
			h.ICMP = true
		}

		hostsFull = append(hostsFull, h)
		seen[h.Address] = true
	}

	// add unique icmp hosts
	for _, h := range hostsPing {
		if !seen[h.Address] {
			h.ICMP = true
			hostsFull = append(hostsFull, h)
		}
	}

	durationFull := durationPing + durationPorts

	return hostsFull, durationFull, nil
}

func (s *Scan) pingScan(ctx context.Context) ([]*host.Host, time.Duration, error) {
	if !s.ping {
		return nil, 0, errors.New("failed to start ping scan: ping scan is disabled")
	}

	nmapOptons := []nmap.Option{
		nmap.WithTargets(s.targets...),
		nmap.WithPingScan(),
	}

	scanner, err := nmap.NewScanner(ctx, nmapOptons...)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create nmap scanner: %w", err)
	}

	start := time.Now()

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("unable to run nmap scan: %w", err)
	}

	if len(*warnings) > 0 {
		return nil, 0, fmt.Errorf("nmap scan has problems: %s", *warnings)
	}

	var hosts []*host.Host
	for _, h := range result.Hosts {
		if h.Status.State != "up" {
			continue
		}

		hosts = append(hosts, &host.Host{
			Address: h.Addresses[0].String(),
			ICMP:    true,
		})
	}

	duration := time.Since(start)

	return hosts, duration, nil
}

func (s *Scan) portsScan(ctx context.Context, targets []string) ([]*host.Host, time.Duration, error) {
	if len(s.ports) == 0 {
		return nil, 0, errors.New("failed to start ports scan: target ports is empty")
	}

	nmapOptons := []nmap.Option{
		nmap.WithTargets(targets...),
		nmap.WithPorts(s.ports...),
	}

	if s.deep {
		nmapOptons = append(nmapOptons, nmap.WithSkipHostDiscovery())
	}

	scanner, err := nmap.NewScanner(ctx, nmapOptons...)
	if err != nil {
		return nil, 0, fmt.Errorf("unable to create nmap scanner: %w", err)
	}

	start := time.Now()

	result, warnings, err := scanner.Run()
	if err != nil {
		return nil, 0, fmt.Errorf("unable to run nmap scanner: %w", err)
	}

	if len(*warnings) > 0 {
		return nil, 0, fmt.Errorf("nmap scanner has problems: %s", *warnings)
	}

	var hosts []*host.Host
	for _, h := range result.Hosts {
		if len(h.Ports) == 0 || len(h.Addresses) == 0 {
			continue
		}

		var ports []*host.Port
		for _, p := range h.Ports {
			if p.State.String() != string(nmap.Open) {
				continue
			}

			ports = append(ports, &host.Port{
				Port:     strconv.Itoa(int(p.ID)),
				Protocol: p.Protocol,
				Service:  p.Service.Name,
			})
		}

		if len(ports) == 0 {
			continue
		}

		host := &host.Host{
			Address: h.Addresses[0].String(),
			Ports:   ports,
		}
		// if deep scan is disabled - detected hosts with online
		if !s.deep {
			host.ICMP = true
		}

		hosts = append(hosts, host)
	}

	duration := time.Since(start)

	return hosts, duration, nil
}

func (s *Scan) portsScanParralel(ctx context.Context) ([]*host.Host, time.Duration, error) {
	if s.concurrency == 0 {
		return nil, 0, errors.New("error run parralel scan: concurrency is 0. Set concurrency to greater than 0")
	}

	ips, err := targetsToIPs(s.targets)
	if err != nil {
		return nil, 0, err
	}

	start := time.Now()

	pool := pond.NewPool(s.concurrency)
	var hosts []*host.Host
	var mu sync.Mutex

	errCh := make(chan error, len(ips))

	for _, ip := range ips {
		pool.SubmitErr(func() error {
			result, _, err := s.portsScan(ctx, []string{ip})
			if err != nil {
				errCh <- err
				return err
			}

			mu.Lock()
			hosts = append(hosts, result...)
			mu.Unlock()

			return nil
		})
	}

	pool.StopAndWait()

	close(errCh)
	for err := range errCh {
		if err != nil {
			return nil, 0, err
		}
	}

	duration := time.Since(start)

	return hosts, duration, nil
}
