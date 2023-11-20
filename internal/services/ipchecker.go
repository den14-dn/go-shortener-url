package services

import (
	"fmt"
	"net"

	"golang.org/x/exp/slog"
)

// IPChecker describes the operation of the service for checking an IP address for inclusion in a trusted subnet.
type IPChecker interface {
	IsTrustedSubnet(net.IP) bool
}

// IPCheckService contains a field describing the IP network.
type IPCheckService struct {
	ipNet *net.IPNet
}

// IsTrustedSubnet checks the IP address for inclusion in a trusted subnet.
func (s IPCheckService) IsTrustedSubnet(ip net.IP) bool {
	return s.ipNet.Contains(ip)
}

// InitIpCheckService constructor for IPCheckService.
func InitIpCheckService(trustedSubnet string) IPChecker {
	_, ipNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		slog.Error(fmt.Sprintf("err happened when parsing trusted subnet: %v", err))
		return nil
	}

	return IPCheckService{ipNet}
}
