package net

import (
	"github.com/pkg/errors"
	"net"
)

func RetrieveLocalIP() (string, error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addresses {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}

	return "", errors.New("unable to determine local IP")
}
