package main

import (
	"encoding/base64"

	"dario.cat/mergo"

	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/function-sdk-go/errors"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

type Response struct {
	desiredComposite *resource.Composite
	desiredComposed  map[resource.Name]*resource.DesiredComposed
}

const (
	AnnotationReadyKey = "javascript.fn.crossplane.io/ready"
)

// NewResponse creates a Response object passed as a second argument to the
// JavaScript handler function.
func NewResponse(req *fnv1beta1.RunFunctionRequest) (*Response, error) {
	desiredCompositeResource, err := request.GetDesiredCompositeResource(req)
	if err != nil {
		return nil, err
	}

	desiredComposedResources, err := request.GetDesiredComposedResources(req)
	if err != nil {
		return nil, err
	}

	return &Response{
		desiredComposite: desiredCompositeResource,
		desiredComposed:  desiredComposedResources,
	}, nil
}

// SetDesiredComposedResource sets the desired composed resource in the
// function response. The caller must be sure to avoid overwriting the desired
// state that may have been accumulated by previous Functions in the pipeline,
// unless they intend to.
func (r *Response) SetDesiredComposedResource(name string, obj map[string]any) error {
	if obj == nil {
		return errors.Errorf(`invalid resource "%s": expected a non-nil Object`, name)
	}

	res := resource.NewDesiredComposed()
	res.Resource.Object = obj

	if res.Resource.GetAPIVersion() == "" {
		return errors.Errorf(`invalid resource "%s": APIVersion must be set`, name)
	}

	if res.Resource.GetKind() == "" {
		return errors.Errorf(`invalid resource "%s": Kind must be set`, name)
	}

	annotations := res.Resource.GetAnnotations()

	if annotations != nil {
		if val, ok := annotations[AnnotationReadyKey]; ok {
			val := resource.Ready(val)

			switch val {
			case resource.ReadyTrue, resource.ReadyFalse, resource.ReadyUnspecified:
				res.Ready = val
			default:
				res.Ready = resource.ReadyUnspecified
			}

			delete(annotations, AnnotationReadyKey)
			res.Resource.SetAnnotations(annotations)
		}
	}

	r.desiredComposed[resource.Name(name)] = res

	return nil
}

// UpdateCompositeStatus merges the desired composite resource status in the
// function response. In case of conflict, new values have priority over existing ones.
func (r *Response) UpdateCompositeStatus(status map[string]any) error {
	dst := make(map[string]interface{})
	if err := r.desiredComposite.Resource.GetValueInto("status", &dst); err != nil && !fieldpath.IsNotFound(err) {
		return errors.Wrap(err, "cannot get desired composite status")
	}

	if err := mergo.Merge(&dst, status, mergo.WithOverride); err != nil {
		return errors.Wrap(err, "cannot merge desired composite status")
	}

	if err := r.desiredComposite.Resource.SetValue("status", dst); err != nil {
		return errors.Wrap(err, "cannot set desired composite status")
	}

	return nil
}

// SetConnectionDetails sets the desired composite resource connection details
// in the function response.
func (r *Response) SetConnectionDetails(details map[string]string) {
	for key, val := range details {
		decoded, _ := base64.StdEncoding.DecodeString(val)
		r.desiredComposite.ConnectionDetails[key] = decoded
	}
}

func (r *Response) setFunctionResponse(rsp *fnv1beta1.RunFunctionResponse) error {
	err := response.SetDesiredComposedResources(rsp, r.desiredComposed)
	if err != nil {
		return errors.Wrap(err, "cannot set desired composed resources")
	}

	err = response.SetDesiredCompositeResource(rsp, r.desiredComposite)
	if err != nil {
		return errors.Wrap(err, "cannot set desired composite resource")
	}

	return nil
}
