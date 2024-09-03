package main

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"github.com/pkg/errors"

	"github.com/tidwall/gjson"

	"k8s.io/apimachinery/pkg/util/validation/field"

	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"

	"github.com/upbound/function-cidr/input/v1beta1"
)

// ExtractKeys extracts keys from a dotted list of keys while considering quoted strings a single value.
func ExtractKeys(input string) []string {
	var keys []string
	var keyBuilder strings.Builder
	inQuotes := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if char == '\'' {
			inQuotes = !inQuotes
		} else if char == '.' && !inQuotes {
			keys = append(keys, keyBuilder.String())
			keyBuilder.Reset()
		} else {
			keyBuilder.WriteByte(char)
		}
	}

	if keyBuilder.Len() > 0 {
		keys = append(keys, keyBuilder.String())
	}

	return keys
}

// GetPrefixField returns the prefix value from the defined field
func GetPrefixField(prefixField string, oxr *resource.Composite, req *fnv1beta1.RunFunctionRequest) (string, error) {
	prefix := ""
	if strings.HasPrefix(prefixField, "desired.") {
		if strings.HasPrefix(prefixField, "desired.composite.") {
			dxr, err := request.GetDesiredCompositeResource(req)
			if err != nil {
				return "", errors.Wrapf(err, "cannot get desired composite resource from %s for %s", prefixField, dxr.Resource.GetKind())
			}
			dxrPrefix, err := dxr.Resource.GetString(strings.Replace(prefixField, "desired.composite.resource.", "", 1))
			prefix = dxrPrefix
			if err != nil {
				return "", errors.Wrapf(err, "cannot get prefix from field %s for %s", prefixField, dxr.Resource.GetKind())
			}
		} else if strings.HasPrefix(prefixField, "desired.resources.") {
			properties := ExtractKeys(strings.Replace(prefixField, "desired.resources.", "", 1))
			resourceName := resource.Name(properties[0])
			dxr, err := request.GetDesiredComposedResources(req)
			if err != nil {
				return "", errors.Wrapf(err, "cannot get desired composed resource from %s", prefixField)
			}
			if val, ok := dxr[resourceName]; ok {
				dxrPrefix, err := val.Resource.GetString(strings.Replace(prefixField, "desired.resources."+properties[0]+".resource.", "", 1))
				prefix = dxrPrefix
				if err != nil {
					return "", errors.Wrapf(err, "cannot get prefix for resource with name %s from field %s", resourceName, prefixField)
				}
			} else {
				return "", errors.New(fmt.Sprintf("No composed resource with name %s found for field %s", resourceName, prefixField))
			}
		}
	} else if strings.HasPrefix(prefixField, "context.") {
		ctxField := strings.Replace(prefixField, "context.", "", 1)
		ctx := req.Context
		if ctx == nil {
			return "", errors.New("No context available")
		}
		json, err := json.Marshal(ctx)
		if err != nil {
			return "", errors.Wrapf(err, "failed to marshall context to json for extraction of field %s", prefixField)
		}
		prefixValue := gjson.GetBytes(json, ctxField)
		if !prefixValue.Exists() {
			return "", errors.New(fmt.Sprintf("Failed to extract value for %s from json context %s", ctxField, json))
		}
		prefix = prefixValue.Str
	} else {
		prefixValue, err := oxr.Resource.GetString(prefixField)
		prefix = prefixValue
		if err != nil {
			return "", errors.Wrapf(err, "cannot get prefix from field %s for %s", prefixField, oxr.Resource.GetKind())
		}
	}
	return prefix, nil
}

// ValidatePrefixParameter validates prefix parameter
func ValidatePrefixParameter(prefix, prefixField string, oxr *resource.Composite, req *fnv1beta1.RunFunctionRequest) *field.Error {
	if len(prefix) > 0 && len(prefixField) > 0 {
		return field.Required(field.NewPath("parameters"), "specify only one of prefix or prefixField to avoid ambiguous function input")
	}
	if prefix == "" {
		if prefixField == "" {
			return field.Required(field.NewPath("parameters"), "either prefix or prefixField function input is required")
		}
		oxrPrefix, err := GetPrefixField(prefixField, oxr, req)
		prefix = oxrPrefix
		if err != nil {
			return field.Required(field.NewPath("parameters"), errors.Wrapf(err, "cannot get prefix at prefixField "+prefixField).Error())
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

func ValidateMultiCidrPrefixParameter(p *v1beta1.Parameters, oxr *resource.Composite) *field.Error {
	if len(p.MultiPrefix) > 0 && len(p.MultiPrefixField) > 0 {
		return field.Required(field.NewPath("parameters"), "specify only one of multiPrefix or multiPrefixField to avoid ambiguous function input")
	}

	if len(p.MultiPrefix) == 0 && p.MultiPrefixField == "" {
		return field.Required(field.NewPath("parameters"), "either multiPrefix or multiPrefixField function input is required")
	}

	var multiPrefixes []v1beta1.MultiPrefix = p.MultiPrefix
	if len(p.MultiPrefix) == 0 {
		err := oxr.Resource.GetValueInto(p.MultiPrefixField, &multiPrefixes)
		if err != nil {
			return field.Required(field.NewPath("parameters"), "cannot get multiPrefixes at multiPrefixField "+p.MultiPrefixField)
		}
	}

	for _, mp := range multiPrefixes {
		_, _, err := net.ParseCIDR(mp.Prefix)
		if err != nil {
			return field.Required(field.NewPath("parameters"), "invalid CIDR prefix address "+mp.Prefix)
		}

		if len(mp.NewBits) == 0 {
			return field.Required(field.NewPath("parameters"), "newBits is required for each prefix in multiPrefixField")
		}
	}

	return nil
}

// ValidateParameters validates the Parameters object.
func ValidateParameters(p *v1beta1.Parameters, oxr *resource.Composite, req *fnv1beta1.RunFunctionRequest) *field.Error {
	var cidrFunc string = p.CidrFunc
	var err error

	if p.CidrFuncField != "" {
		cidrFunc, err = oxr.Resource.GetString(p.CidrFuncField)
		if err != nil {
			return field.Required(field.NewPath("parameters"), "cannot get cidrFunc at cidrFuncField "+p.CidrFuncField)
		}
	}

	if cidrFunc != "multiprefixloop" {
		fieldError := ValidatePrefixParameter(p.Prefix, p.PrefixField, oxr, req)
		if fieldError != nil {
			return fieldError
		}
	}

	switch cidrFunc {
	case "":
		return field.Required(field.NewPath("parameters"), "cidrFunc is required")
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
	case "multiprefixloop":
		return ValidateMultiCidrPrefixParameter(p, oxr)
	default:
		return field.Required(field.NewPath("parameters"), "unexpected cidrFunc "+cidrFunc)
	}
}
