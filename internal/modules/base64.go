package modules

import (
	"encoding/base64"

	"github.com/dop251/goja"
)

// Base64 module provides functions to encode/decode strings to Base64
//
// Example (js):
//
// const encoded = btoa('hello');
// const decoded = atob(encoded);
var Base64 = &Base64module{}

type Base64module struct{}

func (b *Base64module) Enable(runtime *goja.Runtime) {
	_ = runtime.Set("btoa", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).ToString().String()
		result := base64.StdEncoding.EncodeToString([]byte(str))
		return runtime.ToValue(result)
	})

	_ = runtime.Set("atob", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).ToString().String()
		result, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			panic(runtime.ToValue(err.Error()))
		}

		return runtime.ToValue(string(result))
	})
}
