package main

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/salemove/crossplane-function-javascript/input/v1beta1"
	"github.com/salemove/crossplane-function-javascript/internal/js"
	"google.golang.org/protobuf/encoding/protojson"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	fnv1beta1 "github.com/crossplane/function-sdk-go/proto/v1beta1"
	"github.com/crossplane/function-sdk-go/request"
	"github.com/crossplane/function-sdk-go/response"
)

// Function returns whatever response you ask it to.
type Function struct {
	fnv1beta1.UnimplementedFunctionRunnerServiceServer

	log logging.Logger
}

// RunFunction runs the Function.
func (f *Function) RunFunction(_ context.Context, req *fnv1beta1.RunFunctionRequest) (*fnv1beta1.RunFunctionResponse, error) {
	f.log.Info("Running function", "tag", req.GetMeta().GetTag())

	rsp := response.To(req, response.DefaultTTL)

	in := &v1beta1.Input{}
	if err := request.GetInput(req, in); err != nil {
		response.Fatal(rsp, errors.Wrapf(err, "cannot get Function input from %T", req))
		return rsp, nil
	}

	source := strings.TrimSpace(in.Spec.Source.Inline)
	if source == "" {
		response.Fatal(rsp, errors.New("invalid function input: empty source"))
		return rsp, nil
	}

	reqObj, err := convertToMap(req)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	respObj, err := NewResponse(req)
	if err != nil {
		response.Fatal(rsp, err)
		return rsp, nil
	}

	runtime := js.NewRuntime()

	_, err = runtime.RunScript("<inline>.js", source, reqObj, respObj)
	if err != nil {
		response.Fatal(rsp, errors.Wrap(err, "function error"))
		return rsp, nil
	}

	if err := respObj.setFunctionResponse(rsp); err != nil {
		response.Fatal(rsp, err)
	}
	return rsp, nil
}

func convertToMap(req *fnv1beta1.RunFunctionRequest) (map[string]any, error) {
	jReq, err := protojson.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal request from proto to json")
	}

	var mReq map[string]any
	if err := json.Unmarshal(jReq, &mReq); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal json to map[string]any")
	}

	return mReq, nil
}
