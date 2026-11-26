package host

import (
	"net"
	"strconv"
)

// localAddrs returns this computer's IPv4 addresses on its local
// networks, with port, for a pairing offer.
func localAddrs(port int) []string {
	var addrs []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		list, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, a := range list {
			ipnet, ok := a.(*net.IPNet)
			if !ok || ipnet.IP.To4() == nil || !ipnet.IP.IsPrivate() {
				continue
			}
			addrs = append(addrs, net.JoinHostPort(ipnet.IP.String(), strconv.Itoa(port)))
		}
	}
	return addrs
}
