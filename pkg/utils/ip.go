package utils

import "net"

func GetLocalIp() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		return "127.0.0.1"
	}

	var ipv4, ipv6 string

	for _, ifc := range interfaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := ifc.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			default:
				continue
			}

			if ip.IsLoopback() {
				continue
			}

			// ipv4 firstly
			if ipv4 == "" {
				if ip4 := ip.To4(); ip4 != nil {
					ipv4 = ip4.String()
					continue
				}
			}

			if ipv6 == "" && ip.To4() == nil {
				if !ip.IsLinkLocalUnicast() {
					ipv6 = ip.String()
				}
			}
		}
	}

	if ipv4 != "" {
		return ipv4
	}

	if ipv6 != "" {
		return ipv6
	}

	return "127.0.0.1"
}
