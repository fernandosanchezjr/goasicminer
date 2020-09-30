package utils

import (
	"net"
)

func GetLocalIPs() ([]net.IP, error) {
	if addrs, err := net.InterfaceAddrs(); err != nil {
		return nil, err
	} else {
		var ips []net.IP
		for _, ifAddr := range addrs {
			switch v := ifAddr.(type) {
			case *net.IPNet:
				ips = append(ips, v.IP)
			case *net.IPAddr:
				ips = append(ips, v.IP)
			}
		}
		return ips, nil
	}

}
