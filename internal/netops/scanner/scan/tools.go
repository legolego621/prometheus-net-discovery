package scan

import (
	"fmt"
	"strings"

	"github.com/malfunkt/iprange"
)

func targetsToIPs(targets []string) ([]string, error) {
	largetsString := strings.Join(targets, ",")

	list, err := iprange.ParseList(largetsString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse targets: %w", err)
	}

	listIPs := list.Expand()

	ips := make([]string, len(listIPs))
	for i := range listIPs {
		ips[i] = listIPs[i].String()
	}

	return ips, nil
}
