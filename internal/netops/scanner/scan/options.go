package scan

type Option func(*Scan)

func WithCuncurrency(n int) Option {
	return func(s *Scan) {
		s.concurrency = n
	}
}

func WithDeep() Option {
	return func(s *Scan) {
		s.deep = true
	}
}

func WithPing() Option {
	return func(s *Scan) {
		s.ping = true
	}
}

func WithTargets(targets []string) Option {
	return func(s *Scan) {
		s.targets = targets
	}
}

func WithPorts(ports []string) Option {
	return func(s *Scan) {
		s.ports = ports
	}
}
