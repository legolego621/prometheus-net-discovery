package scan

import (
	"fmt"
	"strings"

	"github.com/malfunkt/iprange"
)

func targetsToIPs(targets []string) ([]string, error) {
	targetsString := strings.Join(targets, ",")

	list, err := iprange.ParseList(targetsString)
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
