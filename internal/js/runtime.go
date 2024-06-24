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

type Script struct {
	Name   string
	Source string
	Args   []interface{}

	runtime *Runtime
}

type ScriptOption func(s *Script) error

// NewRuntime creates a new JavaScript runtime. The runtime is set up to uncapitalize
// struct fields and methods passed into runtime as objects.
func NewRuntime() *Runtime {
	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	registry := new(require.Registry)
	registry.Enable(vm)

	modules.Base64.Enable(vm)
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

func (runtime *Runtime) Script(name string, source string, args ...interface{}) *Script {
	return &Script{
		Name:   name,
		Source: source,
		Args:   args,

		runtime: runtime,
	}
}

// Run runs a script, and then invokes the function exported by the script ("export default function")
// with the script's arguments.
func (s *Script) Run(opts ...ScriptOption) (interface{}, error) {
	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}
	exports, err := s.runtime.compile(s.Name, s.Source)
	if err != nil {
		return nil, err
	}

	def := exports.Get("default")

	if fn, ok := goja.AssertFunction(def); ok {
		values := s.runtime.asValues(s.Args)

		if val, err := fn(exports, values...); err == nil {
			return val.Export(), nil
		} else {
			return nil, err
		}
	} else if def == nil {
		return nil, fmt.Errorf("%s must export default function", s.Name)
	} else {
		return nil, fmt.Errorf("%s must export default function, %s exported", s.Name, def.ExportType())
	}
}

// TranspileToES5 transforms the script source code to ES5.1 using Babel
func TranspileToES5(val bool) ScriptOption {
	return func(s *Script) error {
		if !val {
			return nil
		}

		code, err := babel.TransformString(s.Source, map[string]interface{}{
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

		if err != nil {
			return err
		}

		s.Source = code
		return nil
	}
}

// compile runs a script from file and returns the value of "export default function" expression from the script.
// The exported function then should be run separately.
func (runtime *Runtime) compile(name string, source string) (*goja.Object, error) {
	// CommonJS exports work, but we need to manually define an object with this property
	_ = runtime.vm.Set("exports", map[string]interface{}{})

	if _, err := runtime.vm.RunScript(name, source); err != nil {
		return nil, err
	}

	exports := runtime.vm.Get("exports").ToObject(runtime.vm)

	return exports, nil
}

func (runtime *Runtime) asValues(args []interface{}) (ret []goja.Value) {
	for _, arg := range args {
		ret = append(ret, runtime.vm.ToValue(arg))
	}

	return ret
}

func init() {
	_ = babel.Init(4)
}
