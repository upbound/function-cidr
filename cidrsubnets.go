package main

import (
	"fmt"
	"net"

	"github.com/apparentlymart/go-cidr/cidr"
	"github.com/pkg/errors"
)

const Bits32 = 32
const Bits128 = 128

func CidrSubnets(prefix string, newbits ...int) ([][]byte, error) {
	_, network, err := net.ParseCIDR(prefix)
	if err != nil {
		return nil, err
	}

	startPrefixLen, _ := network.Mask.Size()

	prefixLengthArgs := newbits
	if len(prefixLengthArgs) == 0 {
		return nil, nil
	}

	var firstLength int
	firstLength = newbits[0]
	firstLength += startPrefixLen

	retVals := make([][]byte, len(prefixLengthArgs))

	current, _ := cidr.PreviousSubnet(network, firstLength)
	for i, lengthArg := range prefixLengthArgs {
		var length int
		length = lengthArg

		if length < 1 {
			return nil, errors.New("must extend prefix by at least one bit")
		}
		// For portability with 32-bit systems where the subnet number
		// will be a 32-bit int, we only allow extension of 32 bits in
		// one call even if we're running on a 64-bit machine.
		// (Of course, this is significant only for IPv6.)
		if length > Bits32 {
			return nil, errors.New("may not extend prefix by more than 32 bits")
		}
		length += startPrefixLen
		if length > (len(network.IP) * 8) {
			protocol := "IP"
			switch len(network.IP) * 8 {
			case Bits32:
				protocol = "IPv4"
			case Bits128:
				protocol = "IPv6"
			}
			errTxt := fmt.Sprintf("would extend prefix to %d bits, which is too long for an %s address", length, protocol)
			return nil, errors.New(errTxt)
		}

		next, rollover := cidr.NextSubnet(current, length)
		if rollover || !network.Contains(next.IP) {
			// If we run out of suffix bits in the base CIDR prefix then
			// NextSubnet will start incrementing the prefix bits, which
			// we don't allow because it would then allocate addresses
			// outside of the caller's given prefix.
			errTxt := fmt.Sprintf("not enough remaining address space for a subnet with a prefix of %d bits after %s", length, current.String())
			return nil, errors.New(errTxt)
		}

		current = next
		retVals[i] = []byte(current.String())
	}

	return retVals, nil
}
