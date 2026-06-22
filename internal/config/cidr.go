package config

import (
	"fmt"
	"net"
	"strings"
)

// ParseCIDRs parses comma-separated CIDR strings into *net.IPNet slice.
func ParseCIDRs(input string) ([]*net.IPNet, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}
	parts := strings.Split(input, ",")
	result := make([]*net.IPNet, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		_, cidr, err := net.ParseCIDR(p)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", p, err)
		}
		result = append(result, cidr)
	}
	return result, nil
}

// IsTrustedIP checks if ip belongs to any of the trusted CIDR networks.
func IsTrustedIP(ip net.IP, trusted []*net.IPNet) bool {
	if ip == nil {
		return false
	}
	for _, cidr := range trusted {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
