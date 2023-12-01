package main

import (
	"fmt"
	"math/big"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pkg/errors"
)

// CidrSubnet
func CidrSubnet(prefix string, newbits int, netnum int64) ([]byte, error) {
	_, network, err := net.ParseCIDR(prefix)
	if err != nil {
		errStr := fmt.Sprintf("prefix: %s, newbits: %d, netnum: %d", prefix, newbits, netnum)
		return nil, errors.Wrap(err, errStr)
	}

	newNetwork, err := cidr.SubnetBig(network, newbits, big.NewInt(netnum))
	if err != nil {
		errStr := fmt.Sprintf("cidr.SubnetBig: %v", err)
		return nil, errors.New(errStr)
	}
	return []byte(newNetwork.String()), nil
}
