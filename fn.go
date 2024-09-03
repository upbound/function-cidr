package main

import (
	"context"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"

	"github.com/upbound/function-cidr/input/v1beta1"
)

// Function runs CIDR calculations and composes CIDR resources.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	rsp := response.To(req, response.DefaultTTL)

	input := &v1beta1.Parameters{}
	if err := request.GetInput(req, input); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get Function input"))
		return rsp, nil
	}

	oxr, err := request.GetObservedCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get observed composite resource from %T", req))
		return rsp, nil
	}

	if err := ValidateParameters(input, oxr, req); err != nil {
		response.Fatal(rsp, errors.Wrap(err, "invalid Function input"))
		return rsp, nil
	}

	log := f.log.WithValues(
		"oxr-version", oxr.Resource.GetAPIVersion(),
		"oxr-kind", oxr.Resource.GetKind(),
		"oxr-name", oxr.Resource.GetName(),
	)

	dxr, err := request.GetDesiredCompositeResource(req)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "cannot get desired composite resource"))
		return rsp, nil
	}

	dxr.Resource.SetAPIVersion(oxr.Resource.GetAPIVersion())
	dxr.Resource.SetKind(oxr.Resource.GetKind())

	cidrFunc := input.CidrFunc
	if len(input.CidrFuncField) > 0 {
		cidrFunc, err = oxr.Resource.GetString(input.CidrFuncField)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot get cidrFunc from field %s for %s", input.CidrFunc, oxr.Resource.GetKind()))
			return rsp, nil
		}
	}
	log.Info("Running function", "cidrFunc", cidrFunc)

	var prefix string = input.Prefix
	if cidrFunc != "multiprefixloop" && len(input.PrefixField) > 0 {
		prefix, err = GetPrefixField(input.PrefixField, oxr, req)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot get prefix from field %s for %s", input.PrefixField, oxr.Resource.GetKind()))
			return rsp, nil
		}
	}

	var field string = input.OutputField
	if field == "" {
		field = "status.atFunction.cidr"
	}

	switch cidrFunc {
	// cidrhost calculates the host CIDR from a prefix and a host number.
	// https://developer.hashicorp.com/terraform/language/functions/cidrhost
	case "cidrhost":
		hostNum := int64(input.HostNum)
		if len(input.HostNumField) > 0 {
			hostNum, err = oxr.Resource.GetInteger(input.HostNumField)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get hostnum from field %s for %s", input.HostNumField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}
		host, cidrHostErr := CidrHost(prefix, int(hostNum))
		if cidrHostErr != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot calculate CIDR host number for %s", oxr.Resource.GetKind()))
			return rsp, nil
		}

		err = dxr.Resource.SetString(field, host)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, host, oxr.Resource.GetKind()))
			return rsp, nil
		}

	// cidrnetmask calculates the netmask from a prefix.
	// https://developer.hashicorp.com/terraform/language/functions/cidrnetmask
	case "cidrnetmask":
		netmask, cidrNetmaskErr := CidrNetmask(prefix)
		if cidrNetmaskErr != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot calculate CIDR netmask for %s", oxr.Resource.GetKind()))
			return rsp, nil
		}

		err = dxr.Resource.SetString(field, netmask)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, netmask, oxr.Resource.GetKind()))
			return rsp, nil
		}

	// cidrsubnet calculates a subnet CIDR from a prefix, a net number
	// and a new bits.
	// https://developer.hashicorp.com/terraform/language/functions/cidrsubnet
	case "cidrsubnet":
		var newBits []int
		newBits = input.NewBits
		if len(input.NewBitsField) > 0 {
			err = oxr.Resource.GetValueInto(input.NewBitsField, &newBits)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get newbits from field %s of %s", input.NewBitsField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}
		netNum := input.NetNum
		if len(input.NetNumField) > 0 {
			netNum, err = oxr.Resource.GetInteger(input.NetNumField)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get netnum from field %s for %s", input.NetNumField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}
		cidr, cidrSubnetErr := CidrSubnet(prefix, newBits[0], netNum)
		if cidrSubnetErr != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot calculate subnet CIDR for %s", oxr.Resource.GetKind()))
			return rsp, nil
		}

		err = dxr.Resource.SetString(field, string(cidr))
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, string(cidr), oxr.Resource.GetKind()))
			return rsp, nil
		}

	// cidrsubnets calculates a sequence of consecutive
	// IP address ranges within a particular CIDR prefix.
	// https://developer.hashicorp.com/terraform/language/functions/cidrsubnets
	case "cidrsubnets":
		var newBits []int
		newBits = input.NewBits
		if len(input.NewBitsField) > 0 {
			err = oxr.Resource.GetValueInto(input.NewBitsField, &newBits)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get newbits from field %s of %s", input.NewBitsField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}
		cidrs, err := CidrSubnets(prefix, newBits...)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot calculate Subnet CIDRs for %s", oxr.Resource.GetKind()))
			return rsp, nil
		}

		var cidrSubnetsStringArray []string
		for _, cidr := range cidrs {
			cidrSubnetsStringArray = append(cidrSubnetsStringArray, string(cidr))
		}

		err = dxr.Resource.SetValue(field, cidrSubnetsStringArray)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, cidrSubnetsStringArray, oxr.Resource.GetKind()))
			return rsp, nil
		}

	// cidrsubnetloop is a convenience wrapper around cidrsubnet
	// that loops over a range of items, e.g. AZs or subnets
	// or takes a count for its iterations.
	case "cidrsubnetloop":
		var cidrSubnetLoopStringArray []string
		var netNum int64
		var netNumItems []string
		var newBits []int

		newBits = input.NewBits
		if len(input.NewBitsField) > 0 {
			err = oxr.Resource.GetValueInto(input.NewBitsField, &newBits)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get newbits from field %s of %s", input.NewBitsField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}
		offset := int64(input.Offset)
		if len(input.OffsetField) > 0 {
			offset, err = oxr.Resource.GetInteger(input.OffsetField)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get offset from field %s for %s", input.OffsetField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}

		netNumItems = input.NetNumItems
		if len(input.NetNumItemsField) > 0 {
			err = oxr.Resource.GetValueInto(input.NetNumItemsField, &netNumItems)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get netnumitems from field %s for %s", input.NetNumItemsField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}

		netNumCount := input.NetNumCount
		if int64(len(netNumItems)) > netNumCount {
			netNumCount = int64(len(netNumItems))
		}

		if len(input.NetNumCountField) > 0 {
			netNumCount, err = oxr.Resource.GetInteger(input.NetNumCountField)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get netnumcount from field %s for %s", input.NetNumCountField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}

		for netNum = 0; netNum < netNumCount; netNum++ {
			cidr, cidrSubnetErr := CidrSubnet(prefix, newBits[0], netNum+offset)
			if cidrSubnetErr != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot calculate subnet CIDR for %s", oxr.Resource.GetKind()))
				return rsp, nil
			}
			cidrSubnetLoopStringArray = append(cidrSubnetLoopStringArray, string(cidr))
		}

		err = dxr.Resource.SetValue(field, cidrSubnetLoopStringArray)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, cidrSubnetLoopStringArray, oxr.Resource.GetKind()))
			return rsp, nil
		}

	// multiprefix is a convenience wrapper around cidrsubnets that	loops over a
	// range of prefixes to create a list of subnets for each prefix.
	case "multiprefixloop":
		var subnetsByCidr map[string][]string = make(map[string][]string)
		var multiPrefixes []v1beta1.MultiPrefix

		multiPrefixes = input.MultiPrefix
		if len(input.MultiPrefixField) > 0 {
			err = oxr.Resource.GetValueInto(input.MultiPrefixField, &multiPrefixes)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot get multiprefix from field %s for %s", input.MultiPrefixField, oxr.Resource.GetKind()))
				return rsp, nil
			}
		}

		for _, multiPrefix := range multiPrefixes {
			prefix := multiPrefix.Prefix
			if len(prefix) == 0 {
				continue
			}

			newBits := multiPrefix.NewBits
			if len(newBits) == 0 {
				continue
			}

			if multiPrefix.Offset > 0 {
				newBits = append([]int{multiPrefix.Offset}, newBits...)
			}

			cidrs, err := CidrSubnets(prefix, newBits...)
			if err != nil {
				response.Fatal(rsp, errors.Wrapf(err, "cannot calculate Subnet CIDRs for %s", oxr.Resource.GetKind()))
				return rsp, nil
			}

			var cidrSubnetsStringArray []string
			for _, cidr := range cidrs {
				cidrSubnetsStringArray = append(cidrSubnetsStringArray, string(cidr))
			}

			subnetsByCidr[prefix] = cidrSubnetsStringArray
			if multiPrefix.Offset > 0 {
				subnetsByCidr[prefix] = cidrSubnetsStringArray[1:]
			}
		}

		err = dxr.Resource.SetValue(field, subnetsByCidr)
		if err != nil {
			response.Fatal(rsp, errors.Wrapf(err, "cannot set field %s to %s for %s", field, subnetsByCidr, oxr.Resource.GetKind()))
			return rsp, nil
		}

	default:
		log.Info("internal error: sub function not supported: ", "cidrFunc", input.CidrFunc)
	}

	if err := response.SetDesiredCompositeResource(rsp, dxr); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot set desired composite resources from %T", req))
		return rsp, nil
	}

	return rsp, nil
}
