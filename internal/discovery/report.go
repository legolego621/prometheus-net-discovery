package discovery

import "prometheus-net-discovery/internal/netops/host"

// Report is a struct for save discovered hosts.
// Used for metrics collect function.
type Report struct {
	Network         string
	DiscoveredHosts []*host.Host
}
