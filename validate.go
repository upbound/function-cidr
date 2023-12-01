package main

import (
	"net"

	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/crossplane/function-sdk-go/resource"
	"github.com/humoflife/function-cidr/input/v1beta1"
)

// ValidatePrefixParameter validates prefix parameter
func ValidatePrefixParameter(prefix, prefixField string, oxr *resource.Composite) *field.Error {
	if len(prefix) > 0 && len(prefixField) > 0 {
		return field.Required(field.NewPath("parameters"), "specify only one of prefix or prefixfield to avoid ambiguous function input")
	}
	if prefix == "" {
		if prefixField == "" {
			return field.Required(field.NewPath("parameters"), "either prefix or prefixfield function input is required")
		}
		oxrPrefix, err := oxr.Resource.GetString(prefixField)
		prefix = oxrPrefix
		if err != nil {
			return field.Required(field.NewPath("parameters"), "cannot get prefix at prefixfield "+prefixField)
		}
	}

	_, _, err := net.ParseCIDR(prefix)
	if err != nil {
		return field.Required(field.NewPath("parameters"), "invalid CIDR prefix address "+prefix)
	}
	return nil
}

// ValidateCidrHostParameters validates the Parameters object
// in the context of cidrhost
func ValidateCidrHostParameters(p *v1beta1.Parameters, oxr resource.Composite) *field.Error {
	if p.HostNum > 0 && len(p.HostNumField) > 0 {
		return field.Required(field.NewPath("parameters"), "specify only one of hostnum or hostnumfield to avoid ambiguous function input")
	}
	hostNum := p.HostNum
	if hostNum == 0 {
		if p.HostNumField == "" {
			return field.Required(field.NewPath("parameters"), "either hostnum or hostnumfield function input is required")
		}
		_, err := oxr.Resource.GetInteger(p.HostNumField)
		if err != nil {
			return field.Required(field.NewPath("parameters"), "cannot get hostnum at hostnumfield "+p.HostNumField)
		}
	}

	return nil
}

// ValidateCidrSubnetParameters validates the Parameters object
// in the context of cidrsubnet
func ValidateCidrSubnetParameters(p *v1beta1.Parameters) *field.Error {
	if len(p.NewBits) > 0 && len(p.NewBitsField) > 0 {
		return field.Required(field.NewPath("parameters"), "specify only one of newbits or newbitsfield to avoid ambiguous function input")
	}
	if len(p.NewBits) == 0 && p.NewBitsField == "" {
		return field.Required(field.NewPath("parameters"), "either newbits or newbitsfield function input is required")
	}

	if p.NewBitsField == "" {
		if len(p.NewBits) != 1 {
			return field.Required(field.NewPath("parameters"), "cidrFunc cidrsubnet requires exactly 1 parameter in the array")
		}
	}

	if p.NetNum > 0 && len(p.NetNumField) > 0 {
		return field.Required(field.NewPath("parameters"), "cidrFunc cidrsubnet requires either one of netnum or netnumfield")
	}

	return nil
}

// ValidateCidrSubnetsParameters validates the Parameters object
// in the context of cidrsubnet
func ValidateCidrSubnetsParameters(p *v1beta1.Parameters, oxr resource.Composite) *field.Error {
	var newBits []int
	if len(p.NewBits) > 0 && len(p.NewBitsField) > 0 {
		return field.Required(field.NewPath("parameters"), "cidrFunc cidrsubnets requires either one of newbits or newbitsfield")
	}

	if len(p.NewBitsField) > 0 {
		err := oxr.Resource.GetValueInto(p.NewBitsField, &newBits)
		if err != nil {
			return field.Required(field.NewPath("parameters"), "cannot get newbits at newbitsfield "+p.NewBitsField)
		}
	}

	return nil
}

// ValidateCidrSubnetloopParameters validates the Parameters object
// in the context of cidrsubnetloop
func ValidateCidrSubnetloopParameters(p *v1beta1.Parameters) *field.Error {
	if p.NetNumCount > 0 && len(p.NetNumCountField) > 0 {
		// only one of netnumcount or NetNumCountField
		errStr := "cidrFunc cidrsubnetloop requires either one of netnumcount or netnumcountfield, "
		errStr += "but only if nonetnumitems or netnumitemsfield have been specified"
		return field.Required(field.NewPath("parameters"), errStr)
	}
	if len(p.NetNumItems) > 0 && len(p.NetNumItemsField) > 0 {
		// only one of netnumitems or netnumitemsfield
		errStr := "cidrFunc cidrsubnetloop requires either one of netnumitems or netnumitemsfield, "
		errStr += "but only if nonetnumcount or netnumcountfield have been specified"
		return field.Required(field.NewPath("parameters"), errStr)
	}

	netNumCountSpecified := bool(p.NetNumCount > 0 || len(p.NetNumCountField) > 0)
	netNumItemsSpecified := bool(len(p.NetNumItems) > 0 || len(p.NetNumItemsField) > 0)
	if netNumCountSpecified && netNumItemsSpecified {
		// only either netnumcount or items
		errStr := "cidrFunc cidrsubnetloop requires either one of netnumitems or netnumitemsfield, "
		errStr += "or mutually exclusive one of netnumcount or netnumcountfield, but not both counts and items"
		return field.Required(field.NewPath("parameters"), errStr)
	}
	if len(p.NewBits) > 0 && len(p.NewBitsField) > 0 {
		return field.Required(field.NewPath("parameters"), "cidrFunc cidrsubnetloop requires either one of newbits or newbitsfield")
	}
	if p.Offset > 0 && len(p.OffsetField) > 0 {
		return field.Required(field.NewPath("parameters"), "cidrFunc cidrsubnetloop requires either one of offset or offsetfield")
	}

	return nil
}

// ValidateParameters validates the Parameters object.
func ValidateParameters(p *v1beta1.Parameters, oxr *resource.Composite) *field.Error {
	if p.CidrFunc == "" {
		return field.Required(field.NewPath("parameters"), "cidrFunc is required")
	}

	fieldError := ValidatePrefixParameter(p.Prefix, p.PrefixField, oxr)
	if fieldError != nil {
		return fieldError
	}

	switch p.CidrFunc {
	case "cidrhost":
		return ValidateCidrHostParameters(p, *oxr)
	case "cidrnetmask":
		return nil // cidrnetmask only relies on prefix which was checked above
	case "cidrsubnet":
		return ValidateCidrSubnetParameters(p)
	case "cidrsubnets":
		return ValidateCidrSubnetsParameters(p, *oxr)
	case "cidrsubnetloop":
		return ValidateCidrSubnetloopParameters(p)
	default:
		return field.Required(field.NewPath("parameters"), "unexpected cidrFunc "+p.CidrFunc)
	}
}
