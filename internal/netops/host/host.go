package host

type Host struct {
	Address string
	ICMP    bool
	Ports   []*Port
}

type Port struct {
	Port     string
	Protocol string
	Service  string
}
