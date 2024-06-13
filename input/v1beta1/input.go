// Package v1beta1 contains the input type for this Function
// +kubebuilder:object:generate=true
// +groupName=javascript.fn.crossplane.io
// +versionName=v1beta1
package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This isn't a custom resource, in the sense that we never install its CRD.
// It is a KRM-like object, so we generate a CRD to describe its schema.

// Input can be used to provide input to this Function.
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=crossplane
type Input struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InputSpec `json:"spec"`
}

// InputSpec defines input parameters for the function
type InputSpec struct {
	// Source is the function source spec
	Source InputSource `json:"source"`

	// Values is the map of string variables to be passed into the request context
	Values map[string]string `json:"values,omitempty"`
}

// InputSource defines function source parameters
type InputSource struct {
	// Type defines the input source type (currently, only `Inline` is supported).
	// +kubebuilder:validation:Enum=Inline
	// +kubebuilder:default:=Inline
	Type string `json:"type,omitempty"`

	// Inline is the inline form input of the function source
	Inline string `json:"inline,omitempty"`
}
