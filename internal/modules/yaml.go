package modules

import (
	"github.com/dop251/goja"
	"gopkg.in/yaml.v3"
)

// YAML provides functions to parse YAML into object and to encode objects into YAML
//
// Example (js):
//
//	import * as YAML from "yaml";
//	let o = YAML.parse('a: 1');
//	let s = YAML.stringify(o);
func YAML(runtime *goja.Runtime, module *goja.Object) {
	o := module.Get("exports").(*goja.Object)

	_ = o.Set("stringify", func(call goja.FunctionCall) goja.Value {
		obj := call.Argument(0).Export()
		b, err := yaml.Marshal(obj)
		if err != nil {
			panic(runtime.ToValue(err.Error()))
		}

		return runtime.ToValue(string(b))
	})

	_ = o.Set("parse", func(call goja.FunctionCall) goja.Value {
		str := call.Argument(0).ToString().String()

		var result interface{}
		err := yaml.Unmarshal([]byte(str), &result)
		if err != nil {
			panic(runtime.ToValue(err.Error()))
		}

		return runtime.ToValue(result)
	})
}
