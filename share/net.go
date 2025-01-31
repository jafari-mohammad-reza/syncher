package share

import (
	"fmt"
	"log/slog"
	"net"
)

func getOutboundInterface() (string, string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := localAddr.IP.String()

	interfaces, err := net.Interfaces()
	if err != nil {
		return "", "", err
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.Equal(localAddr.IP) {
				return iface.Name, ip, nil
			}
		}
	}

	return "", "", fmt.Errorf("could not determine main network interface")
}

func GetIPv4() (string, error) {
	_, ip, err := getOutboundInterface()
	if err != nil {
		slog.Error("GetIPv4", "err", err.Error())
		return "", err
	}
	return ip, err
}
