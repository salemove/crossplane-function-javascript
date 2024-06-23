package main

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/crossplane/crossplane-runtime/pkg/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/resource"
	"github.com/crossplane/function-sdk-go/response"
)

const (
	xr = `{"apiVersion":"example.org/v1","kind":"XR","spec":{"region":"us-east-1"}}`
)

func TestRunFunction(t *testing.T) {

	type args struct {
		ctx context.Context
		req *fnv1beta1.RunFunctionRequest
	}
	type want struct {
		rsp *fnv1beta1.RunFunctionResponse
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"EmptySource": {
			reason: "The Function should return a fatal result if no input was specified",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta:  &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: scriptToInput(""),
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Results: []*fnv1beta1.Result{
						{
							Severity: fnv1beta1.Severity_SEVERITY_FATAL,
							Message:  "invalid function input: empty source",
						},
					},
				},
			},
		},
		"EmptyFunctionResponse": {
			reason: "The Function should return a normal result",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta:  &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: scriptToInput("export default (req, rsp) => {};"),
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
				},
			},
		},
		"BasicResource": {
			reason: "The Function should return a normal result",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta: &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: scriptToInput(`export default (req, rsp) => {
						rsp.setDesiredComposedResource("test", {
							apiVersion: 'example.org/v1',
							kind:       'Bucket',
							metadata: {
								annotations: { 'javascript.fn.crossplane.io/ready': 'True' }
							},
							spec: {
								forProvider: { region: req.observed.composite.resource.spec.region }
							}
						})
					};`),
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
						Resources: map[string]*fnv1beta1.Resource{
							"test": {
								Ready: fnv1beta1.Ready_READY_TRUE,
								Resource: resource.MustStructJSON(`{
									"apiVersion":"example.org/v1",
									"kind":"Bucket",
									"metadata":{"annotations":{}},
									"spec":{
										"forProvider":{"region":"us-east-1"}
									}
								}`),
							},
						},
					},
				},
			},
		},
		"ConnectionDetails": {
			reason: "The Function should return a normal result",
			args: args{
				req: &fnv1beta1.RunFunctionRequest{
					Meta: &fnv1beta1.RequestMeta{Tag: "hello"},
					Input: scriptToInput(`export default (req, rsp) => {
						rsp.setDesiredComposedResource("test", {
							apiVersion: 'example.org/v1',
							kind:       'Bucket',
							spec: {
								forProvider: { region: req.observed.composite.resource.spec.region }
							}
						});
						rsp.setConnectionDetails({ key: btoa('value') });
					};`),
					Observed: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource: resource.MustStructJSON(xr),
						},
					},
				},
			},
			want: want{
				rsp: &fnv1beta1.RunFunctionResponse{
					Meta: &fnv1beta1.ResponseMeta{Tag: "hello", Ttl: durationpb.New(response.DefaultTTL)},
					Desired: &fnv1beta1.State{
						Composite: &fnv1beta1.Resource{
							Resource:          resource.MustStructJSON(xr),
							ConnectionDetails: map[string][]byte{"key": []byte("value")},
						},
						Resources: map[string]*fnv1beta1.Resource{
							"test": {
								Resource: resource.MustStructJSON(`{
									"apiVersion":"example.org/v1",
									"kind":"Bucket",
									"spec":{
										"forProvider":{"region":"us-east-1"}
									}
								}`),
							},
						},
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			f := &Function{log: logging.NewNopLogger()}
			rsp, err := f.RunFunction(tc.args.ctx, tc.args.req)

			if diff := cmp.Diff(tc.want.rsp, rsp, protocmp.Transform()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want rsp, +got rsp:\n%s", tc.reason, diff)
			}

			if diff := cmp.Diff(tc.want.err, err, cmpopts.EquateErrors()); diff != "" {
				t.Errorf("%s\nf.RunFunction(...): -want err, +got err:\n%s", tc.reason, diff)
			}
		})
	}
}

func scriptToInput(script string) *structpb.Struct {
	return resource.MustStructObject(&unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "javascript.fn.glia-dev.com/v1beta1",
			"kind":       "Input",
			"spec": map[string]interface{}{
				"source": map[string]interface{}{
					"inline": script,
				},
			},
		},
	})
}
