package modules

import "github.com/dop251/goja"

// RuntimeModule is a container for modules
type RuntimeModule struct {
	// Name of the module as it can be seen from the JS runtime
	Name string

	// The content of the module. Can be any type.
	Module interface{}
}

// Register exports the module into JS runtime
func (m *RuntimeModule) Register(vm *goja.Runtime) error {
	return vm.Set(m.Name, m.Module)
}
