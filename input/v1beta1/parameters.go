// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=cidr.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This isn't a custom resource, in the sense that we never install its CRD.
// It is a KRM-like object, so we generate a CRD to describe its schema.

// MultiPrefix defines an item in a list of CIDR blocks to NewBits mappings
type MultiPrefix struct {
	// Prefix is a CIDR block that is used as input for CIDR calculations
	//
	// +required
	// +kubebuilder:validation:Pattern="^([0-9]{1,3}.){3}[0-9]{1,3}/[0-9]{1,2}$"
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Required
	Prefix string `json:"prefix"`

	// NewBits is a list of bits to allocate to the subnet
	//
	// +required
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:Required
	// +listType=atomic
	NewBits []int `json:"newBits"`

	// Offset is the number of bits to offset the subnet mask by when generating
	// subnets.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=32
	// +kubebuilder:default=0
	Offset int `json:"offset,omitempty"`
}

// Parameters can be used to provide input to this Function.
//
// Almost all parameters can be provided as literals or as references to
// fields on the claim, allowing defaults to be set in the composition and then
// overridden by the claim.
//
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Parameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// cidrFunc is the name of the function to call
	//
	// +optional
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum={cidrhost,cidrnetmask,cidrsubnet,cidrsubnets,cidrsubnetloop,multiprefixloop}
	CidrFunc string `json:"cidrFunc"`

	// cidrFuncField is a reference to a location on the claim specifying the
	// cidrFunc to call
	//
	// +optional
	// +kubebuilder:validation:Type=string
	CidrFuncField string `json:"cidrFuncField,omitempty"`

	// multiPrefix is a list of CIDR blocks to NewBits mappings that are used as
	// input for the `multiprefixloop` function.
	//
	// +optional
	MultiPrefix []MultiPrefix `json:"multiPrefix,omitempty"`

	// multiPrefixField describes a location on the claim that contains the
	// multiPrefix to use as input for the `multiprefixloop` function.
	//
	// The location referenced should contain a list of MultiPrefix objects.
	//
	// +optional
	MultiPrefixField string `json:"multiPrefixField,omitempty"`

	// prefixField defines a location on the claim to take the prefix from
	//
	// +optional
	PrefixField string `json:"prefixField,omitempty"`

	// prefix is a CIDR block that is used as input for CIDR calculations
	//
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// hostNumField points to a field on the claim that contains the hostNum
	//
	// +optional
	HostNumField string `json:"hostNumField,omitempty"`

	// hostNum  is a whole number that can be represented as a binary integer
	// with no more than the number of digits remaining in the address after
	// the given prefix.
	//
	// +optional
	HostNum int `json:"hostNum,omitempty"`

	// newbitsField points to a field on the claim that contains the newBits
	//
	// +optional
	NewBitsField string `json:"newBitsField,omitempty"`

	// newbits is the number of additional bits with which to extend the prefix.
	// For example, if given a prefix ending in /16 and a newbits value of 4,
	// the resulting subnet address will have length /20.
	//
	// +optional
	NewBits []int `json:"newBits,omitempty"`

	// netNumField points to a field on the claim that contains the netNum
	//
	// +optional
	NetNumField string `json:"netNumField,omitempty"`

	// netNum is a whole number that can be represented as a binary integer with
	// no more than newbits binary digits, which will be used to populate the
	// additional bits added to the prefix.
	//
	// +optional
	NetNum int64 `json:"netNum,omitempty"`

	// netNumCountField points to a field on the claim that contains the
	// netNumCount
	//
	// +optional
	NetNumCountField string `json:"netNumCountField,omitempty"`

	// netNumCount defines how many networks to create from the given prefix
	//
	// +optional
	NetNumCount int64 `json:"netNumCount,omitempty"`

	// netNumItemsField points to a field on the claim that contains the
	// netNumItems
	//
	// +optional
	NetNumItemsField string `json:"netNumItemsField,omitempty"`

	// netNumItems is an array of items whose length may be used to determine
	// how many networks to create from the given prefix.
	//
	// When this field is defined, its length is compared against `netNumCount`
	// and the larger of the two values is used.
	//
	// +optional
	NetNumItems []string `json:"netNumItems,omitempty"`

	// offsetField defines a location on the claim to take the offset from
	//
	// This field is mutually exclusive with netNumCount and netNumItems
	//
	// +optional
	OffsetField string `json:"offsetField,omitempty"`

	// offset defines a starting point in the cidr block to start allocating
	// subnets from. If 0, will start from the beginning of the prefix.
	//
	// This field is mutually exclusive with netNumCount and netNumItems
	//
	// +optional
	Offset int `json:"offset,omitempty"`

	// outputField specifies a location on the XR to patch the results of the
	// function call to.
	//
	// If this field is not specified, the results will be patched to the status
	// field `status.atFunction.cidr`.
	//
	// +optional
	OutputField string `json:"outputField,omitempty"`
}
