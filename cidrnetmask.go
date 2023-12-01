package main

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

func CidrNetmask(prefix string) (string, error) {
	_, network, err := net.ParseCIDR(prefix)
	if err != nil {
		errTxt := fmt.Sprintf("invalid CIDR expression: %s", err)
		return "", errors.New(errTxt)
	}
	return net.IP(network.Mask).String(), nil
}
