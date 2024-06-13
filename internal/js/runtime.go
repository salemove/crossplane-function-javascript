// Package js is a thin wrapper around Goja runtime.
//
// Because Goja itself doesn't support most of modern JS features,
// the source code is transpiled by Babel before executing it,
// so we can actually write state migrations in ES5.1+.
package js

import (
	"fmt"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	babel "github.com/jvatic/goja-babel"
	"github.com/salemove/crossplane-function-javascript/internal/modules"
)

type Runtime struct {
	vm *goja.Runtime
}

// NewRuntime creates a new JavaScript runtime. The runtime is set up to uncapitalize
// struct fields and methods passed into runtime as objects.
func NewRuntime() *Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	registry := new(require.Registry)
	registry.Enable(vm)
	registry.RegisterNativeModule("base64", modules.Base64)
	registry.RegisterNativeModule("yaml", modules.YAML)
	console.Enable(vm)

	return &Runtime{vm: vm}
}

// Set the specified variable in the global context.
func (runtime *Runtime) Set(name string, val interface{}) error {
	return runtime.vm.Set(name, val)
}

// Get the specified variable in the global context.
func (runtime *Runtime) Get(name string) goja.Value {
	return runtime.vm.Get(name)
}

// RunScript runs a script, and then invokes the function exported by the script ("export default function")
// with the given arguments.
func (runtime *Runtime) RunScript(name string, source string, args ...any) (interface{}, error) {
	exports, err := runtime.compile(name, source)

	if err != nil {
		return nil, err
	}

	def := exports.Get("default")

	if fn, ok := goja.AssertFunction(def); ok {
		values := runtime.asValues(args)

		if val, err := fn(exports, values...); err == nil {
			return val.Export(), nil
		} else {
			return nil, err
		}
	} else if def == nil {
		return nil, fmt.Errorf("%s must export default function", name)
	} else {
		return nil, fmt.Errorf("%s must export default function, %s exported", name, def.ExportType())
	}
}

// compile runs a script from file and returns the value of "export default function" expression from the script.
// The exported function then should be run separately.
func (runtime *Runtime) compile(name string, source string) (*goja.Object, error) {
	code, err := transpileFile(source)

	if err != nil {
		return nil, err
	}

	// CommonJS exports work, but we need to manually define an object with this property
	_ = runtime.vm.Set("exports", map[string]interface{}{})

	if _, err := runtime.vm.RunScript(name, code); err != nil {
		return nil, err
	}

	exports := runtime.vm.Get("exports").ToObject(runtime.vm)

	return exports, nil
}

// read the file, transpile it to ES5.1 (using Babel) and return the transpiled source code
func transpileFile(source string) (string, error) {
	return babel.TransformString(source, map[string]interface{}{
		"plugins": []interface{}{
			[]interface{}{"transform-modules-commonjs", map[string]interface{}{"loose": false}},
		},
		"ast":            false,
		"sourceMaps":     "inline", // include source maps in the output for better stack traces
		"babelrc":        false,
		"inputSourceMap": true, // if the function source already includes a source map, use it instead
		"compact":        false,
		"retainLines":    true,
		"highlightCode":  false,
	})
}

func (runtime *Runtime) asValues(args []interface{}) (ret []goja.Value) {
	for _, arg := range args {
		ret = append(ret, runtime.vm.ToValue(arg))
	}

	return ret
}
