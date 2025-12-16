package sfid

import (
	"errors"
	"github.com/sony/sonyflake/v2/types"
	"net"
)

func lower16BitPrivateIP(interfaceAddrs types.InterfaceAddrs) (int, error) {
	ip, err := privateIPv4(interfaceAddrs)
	if err != nil {
		return 0, err
	}

	return int(ip[2])<<8 + int(ip[3]), nil
}

func privateIPv4(interfaceAddrs types.InterfaceAddrs) (net.IP, error) {
	as, err := interfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range as {
		ipnet, ok := a.(*net.IPNet)
		if !ok || ipnet.IP.IsLoopback() {
			continue
		}

		ip := ipnet.IP.To4()
		if isPrivateIPv4(ip) {
			return ip, nil
		}
	}

	return nil, errors.New("no private ip address")
}

func isPrivateIPv4(ip net.IP) bool {
	// Allow private IP addresses (RFC1918) and link-local addresses (RFC3927)
	return ip != nil &&
		(ip[0] == 10 || ip[0] == 172 && (ip[1] >= 16 && ip[1] < 32) || ip[0] == 192 && ip[1] == 168 || ip[0] == 169 && ip[1] == 254)
}
