# function-cidr

A [Crossplane](https://www.crossplane.io/)
[Composition Function](https://docs.crossplane.io/latest/concepts/composition-functions/)
for calculating Classless Inter-Domain Routing
([CIDR](https://en.wikipedia.org/wiki/Classless_Inter-Domain_Routing))
numbers.
A CIDR is an IP address allocation method that is used to improve
data routing efficiency on the internet.

## Overview

This composition function offers 4 HashiCorp compatible
IP Network Functions plus one custom wrapper. Follow the
function links for detailed explanations of the function
semantics.
- [cidrhost](https://developer.hashicorp.com/terraform/language/functions/cidrhost)
- [cidrnetmask](https://developer.hashicorp.com/terraform/language/functions/cidrnetmask)
- [cidrsubnet](https://developer.hashicorp.com/terraform/language/functions/cidrsubnet)
- [cidrsubnets](https://developer.hashicorp.com/terraform/language/functions/cidrsubnets)
- cidrsubnetloop wraps [cidrsubnet](https://developer.hashicorp.com/terraform/language/functions/cidrsubnet)

To use this function, apply the following
[functions.yaml](examples/functions.yaml)
to your Crossplane management cluster.
```
cat <<EOF|kubectl apply -f -
apiVersion: pkg.crossplane.io/v1beta1
kind: Function
metadata:
  name: upbound-function-cidr
spec:
  package: xpkg.upbound.io/upbound/function-cidr:v0.1.0
EOF
```
Call the function from a Crossplane composition as described below.

## Terminology
The `cidrfunc` IP Network Functions have various input parameters.
Below are brief descriptions for context.

- `prefix` must be given in CIDR notation, as defined in [RFC 4632 section 3.1.](https://datatracker.ietf.org/doc/html/rfc4632#section-3.1)
- `hostnum` is a whole number that can be represented as a binary integer with no more than the number of digits remaining in the address after the given prefix.
- `newbits` is the number of additional bits with which to extend the prefix. For example, if given a prefix ending in /16 and a newbits value of 4, the resulting subnet address will have length /20.
- `netnum` is a whole number that can be represented as a binary integer with no more than newbits binary digits, which will be used to populate the additional bits added to the prefix.

## Usage
Specify the `cidrfunc` calculation type in the composition function input.
Valid values are as follows:
```
- cidrhost
- cidrnetmask
- cidrsubnet
- cidrsubnets
- cidrsubnetloop
```
Specify a custom `outputfield` in the function input parameters
when the output should appear at a different path
than the respective `status.atFunction.cidr` sub field default path.

All `cidrfunc` IP Network Functions require a CIDR `prefix` as input.
Provide the `prefix` directly in the
function input or specify a `prefixfield` in the XR where
the function shall pick up the `prefix` value.

Function input field names ending in `field` indicate that
the function shall read the field path value from the specified
field path in the XR.

### cidrhost
The `cidrhost cidrfunc` requires a `hostnum` or `hostnumfield` as
function input. `hostnum` is an integer.

### cidrnetmask
The `cidrnetmask cidrfunc` does not require additional parameters
beyond the `prefix`. The `prefix` can be read from an XR field
when the `prefixField` path is specified in the function input
instead of a `prefix` value.

### cidrsubnet
The `cidrhost cidrsubnet` requires a `netnum` or `netnumfield`,
and a `newbits` or `newbitsfield` as function input.

`netNum` is an integer.
`newBits` is one integer in an array of integers.

### cidrsubnets
The `cidrhost cidrsubnets` requires a `newBits`
 or `newBitsField` as function input.

`newBits` is an array of integers.

### cidrsubnetloop
The `cidrhost cidrsubnetloop` requires the following input fields.
- `newBits` (integer array) or `newBitsField`
- `netNumCount` (integer) or `netNumCountField`
- `netNumItems` (string array) or `netNumItemsField`
- `offset` or `offsetfield`
** netNumCount and netNumItems are mutually exclusive **

The `cidrsubnetloop` wrapper calculates `cidrsubnet` CIDRs using
the `prefix` and `newBits` parameters as input. It performs the
calculations in a loop. The `netnum` is calculated during each
iteration from `iteration`+`offset`. The iterations are either from
0 to `netNumCount` -1 or from 0 to number of items in `netNumItemsCount`
or their respective values from their XR field references.

## Testing The Function
Clone the repo. Run `make debug` and in a second terminal run `make render`
and examine the output. Corresponding compositions and XR yaml can be
found in the `examples` folder.
