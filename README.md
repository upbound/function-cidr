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
  name: function-cidr
  annotations:
    render.crossplane.io/runtime: Docker
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

#### Composition With Default Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: vpc
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrhost
      prefix: "10.0.0.0/20"
      hostNum: 111
```

#### Composition With Custom Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: vpc
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrhost
      prefixField: spec.forFunction.cidr
      hostNumField: spec.forFunction.hostNum
      outputField: "status.atFunction.cidr.hostAddress"
```
#### XR For Composition With Custom Input / Output Fields
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
  forFunction:
    cidr: 10.0.0.0/20
    hostNum: 111
```
The function writes its output into the specified `outputfield`.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
status:
  atFunction:
    cidr:
      hostAddress: 10.0.0.111
```

### cidrnetmask
The `cidrnetmask cidrfunc` does not require additional parameters
beyond the `prefix`. The `prefix` can be read from an XR field
when the `prefixField` path is specified in the function input
instead of a `prefix` value.

#### Composition With Default Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: cidr-example
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrnetmask
      prefix: 172.16.0.0/12
```
#### XR
The `cidrnetmask` function does not rely on any fields in the XR.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
```
The function writes its output to the default `status.cidr`
field unless the function input `outputfield` has been specified.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
status:
  atFunction:
    cidr:
      netmask: 255.240.0.0
```

### cidrsubnet
The `cidrhost cidrsubnet` reauires a `netnum` or `netnumfield`,
and a `newbits` or `newbitsfield` as function input.

`netNum` is an integer.
`newBits` is one integer in an array of integers.

#### Composition With Custom Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: cidr-example
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrsubnet
      prefixField: spec.forFunction.cidr
      newBitsField: spec.forFunction.newBits
      netNumField: spec.forFunction.netNum
      outputField: status.atFunction.cidr.subnet-1
```
#### XR For Composition With Custom Input / Output Fields
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
  forFunction:
    cidr: 10.0.0.0/20
    newBits:
      - 8
    netNum: 3
```
The function writes its output into the specified `outputfield`.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
status:
  atFunction:
    cidr:
      subnet-1: 10.0.0.48/28
```

### cidrsubnets
The `cidrhost cidrsubnets` reauires a `newBits`
 or `newBitsField` as function input.

`newBits` is an array of integers.

#### Composition With Hybrid Default and Custom Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: cidr-example
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrsubnets
      prefix: 10.0.0.0/20
      newBitsField: spec.forFunction.newBits
```
#### XR For Hybrid Composition With Default and Custom Input / Output Fields
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
  forFunction:
    newBits:
      - 8
      - 4
      - 2
```
The function writes its output into the default `outputfield`.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
status:
  atFunction:
    cidr:
      subnets:
      - 10.0.0.0/28
      - 10.0.1.0/24
      - 10.0.4.0/22
```

### cidrsubnetloop
The `cidrhost cidrsubnetloop` reauires the following input fields.
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

#### Composition With Custom Input / Output Fields
```
apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata:
  name: cidr-example
spec:
  compositeTypeRef:
    apiVersion: platform.upbound.io/v1alpha1
    kind: XVPC
  mode: Pipeline
  pipeline:
  - step: cidr
    functionRef:
      name: function-cidr
    input:
      apiVersion: cidr.fn.crossplane.io/v1beta1
      kind: Parameters
      cidrFunc: cidrsubnetloop
      prefixField: spec.forFunction.cidrBlock
      newBitsField: spec.forFunction.newBits
      netNumItemsField: spec.forFunction.azs
      offsetField: spec.forFunction.offset
      outputfield: spec.atFunction.cidr.subnets
```
#### XR For Composition With Custom Input / Output Fields
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
  forFunction:
    cidr: 10.0.0.0/20
    newBits:
      - 8
    netNum: 3
```
The function writes its output into the specified `outputField`.
```
apiVersion: platform.upbound.io/v1alpha1
kind: XVPC
metadata:
  name: cidr-example
spec:
  atFunction:
    cidr:
      subnets:
      - 10.0.0.48/32
      - 10.0.0.49/32
      - 10.0.0.50/32
      - 10.0.0.51/32
      - 10.0.0.52/32
```

## Testing The Function
Clone the repo. Run `make debug` and in a second terminal run `make render`
and examine the output. Corresponding compositions and XR yaml can be
found in the `examples` folder.
