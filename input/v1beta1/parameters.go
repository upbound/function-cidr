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

// Parameters can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Parameters struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// cidrfunc is one of cidrhost, cidrnetmast, cidesubnet, cidrsubnets, cidrsubnetloop
	CidrFunc string `json:"cidrFunc"`

	// prefix field
	PrefixField string `json:"prefixField,omitempty"`

	// prefix is a CIDR block that is used as input for CIDR calculations
	Prefix string `json:"prefix"`

	// hostnum field
	HostNumField string `json:"hostNumField,omitempty"`

	// hostnum
	HostNum int `json:"hostNum,omitempty"`

	// newbits field
	NewBitsField string `json:"newBitsField,omitempty"`

	// newbits
	NewBits []int `json:"newBits,omitempty"`

	// netnum field
	NetNumField string `json:"netNumField,omitempty"`

	// netnum
	NetNum int64 `json:"netNum,omitempty"`

	// netnumcount field
	NetNumCountField string `json:"netNumCountField,omitempty"`

	// netnumcount
	NetNumCount int64 `json:"netNumCount,omitempty"`

	// netnumitems field
	NetNumItemsField string `json:"netNumItemsField,omitempty"`

	// netnumitems
	NetNumItems []string `json:"netNumItems,omitempty"`

	// offset field
	OffsetField string `json:"offsetField,omitempty"`

	// offset is only used by cidrsubnetloop
	Offset int `json:"offset,omitempty"`

	// output field
	OutputField string `json:"outputField,omitempty"`
}
