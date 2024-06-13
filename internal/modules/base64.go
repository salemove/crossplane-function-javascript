package modules

import (
	"encoding/base64"

	"github.com/dop251/goja"
)

// Base64 module provides functions to encode/decode strings to Base64
//
// Example (js):
//
//	import * as Base64 from "base64";
//	export default () => {
//		return Base64.decode(Base64.encode("foo"))
//	};
func Base64(runtime *goja.Runtime, module *goja.Object) {
	o := module.Get("exports").(*goja.Object)

	_ = o.Set("encode", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).ToString().String()
		result := base64.StdEncoding.EncodeToString([]byte(str))
		return runtime.ToValue(result)
	})

	_ = o.Set("decode", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).ToString().String()
		result, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			panic(runtime.ToValue(err.Error()))
		}

		return runtime.ToValue(string(result))
	})
}
