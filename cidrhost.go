package main

import (
	"fmt"
	"math/big"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pkg/errors"
)

func CidrHost(prefix string, hostNumber int) (string, error) {
	hostNum := big.NewInt(int64(hostNumber))

	_, network, err := net.ParseCIDR(prefix)
	if err != nil {
		errTxt := fmt.Sprintf("invalid CIDR expression: %s", err)
		return "", errors.New(errTxt)
	}

	ip, err := cidr.HostBig(network, hostNum)
	if err != nil {
		return "", err
	}

	return ip.String(), nil
}
