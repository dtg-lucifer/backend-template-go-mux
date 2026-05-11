// Package utils provides shared core utilities used across all modules.
package utils

import (
	"fmt"
	"reflect"
	"time"

	"github.com/your-username/go-mux-backend-template/pkg"
)

// Dispatcher is a reflective method dispatcher that logs every call.
// Go has no runtime Proxy like TypeScript, so the pattern is:
//  1. Define a service interface (ServiceIface).
//  2. Implement it with a concrete struct (*Service).
//  3. Write a thin proxy struct (serviceDebugProxy) that also implements
//     the interface and forwards every call through Dispatcher.Call().
//  4. Expose a WithDebug() constructor on the service package.
type Dispatcher struct {
	target  any
	prefix  string
	logger  *pkg.Logger
	methods map[string]reflect.Value
}

// NewDispatcher builds a Dispatcher for target, pre-building the method map
// so Call() has no per-invocation reflection overhead beyond m.Call(in).
func NewDispatcher(target any, prefix string, logger *pkg.Logger) *Dispatcher {
	targetVal := reflect.ValueOf(target)
	targetType := targetVal.Type()

	methods := make(map[string]reflect.Value, targetType.NumMethod())
	for i := range targetType.NumMethod() {
		name := targetType.Method(i).Name
		methods[name] = targetVal.Method(i)
	}

	return &Dispatcher{
		target:  target,
		prefix:  prefix,
		logger:  logger,
		methods: methods,
	}
}

// Call invokes the named method, logging start/end/error around it.
// args must match the method's parameter types exactly.
// Returns []reflect.Value — extract typed results with .Interface().(YourType).
// Panics if the method name does not exist (fail-fast at startup).
func (d *Dispatcher) Call(name string, args ...any) []reflect.Value {
	start := time.Now()
	d.logger.Debug(fmt.Sprintf("[%s.%s] --> START", d.prefix, name),
		"args_count", len(args),
	)

	m, ok := d.methods[name]
	if !ok {
		panic(fmt.Sprintf("[DEBUG_PROXY] method %q not found on %s", name, d.prefix))
	}

	in := make([]reflect.Value, len(args))
	for i, a := range args {
		in[i] = reflect.ValueOf(a)
	}

	results := m.Call(in)
	duration := time.Since(start)

	errType := reflect.TypeOf((*error)(nil)).Elem()
	if len(results) > 0 {
		last := results[len(results)-1]
		if last.Type().Implements(errType) && !last.IsNil() {
			d.logger.Debug(fmt.Sprintf("[%s.%s] <-- ERROR", d.prefix, name),
				"duration", duration.String(),
				"error", last.Interface().(error).Error(),
			)
			return results
		}
	}

	d.logger.Debug(fmt.Sprintf("[%s.%s] <-- END", d.prefix, name),
		"duration", duration.String(),
	)
	return results
}
