package typecheckers

import "reflect"

// A handler has the signature
// func(context.Context, interface{}, ...string) (interface{}, error)
// first argument is context
// second argument is the body of the request
// third...n arguments are path parameters
// omitting the third...n arguments results in no query parameters being added to the path
func IsHandler(t reflect.Type) bool {
	// A handler must be a function
	if t.Kind() != reflect.Func {
		return false
	}

	// A handler must take at least two arguments
	if t.NumIn() < 2 {
		return false
	}

	// A handler must return two results
	if t.NumOut() != 2 {
		return false
	}

	// A handler's first argument must be a context
	if t.In(0).Name() != "Context" {
		return false
	}

	// A handler's second argument can be anything except for a func, we cannot
	// unmarshal JSON into a function.
	if t.In(1).Kind() == reflect.Func {
		return false
	}

	// A handler's third..n arguments, if present, must be strings
	for i := 2; i < t.NumIn(); i++ {
		if t.In(i).Kind() != reflect.String {
			return false
		}
	}

	// A handler's first output can be anything except for a func, we cannot
	// marshal a func into JSON
	if t.Out(0).Kind() == reflect.Func {
		return false
	} else if t.Out(0).Kind() == reflect.Ptr {
		if t.Out(0).Elem().Kind() == reflect.Ptr || t.Out(0).Elem().Kind() == reflect.Func {
			// this allows for returning pointer objects so long as they aren't
			// pointing at another pointer or a func
			return false
		}
	}

	// A handler's second output must be an error
	if t.Out(1).Name() != "error" {
		return false
	}

	return true
}

// Determine if t is *HandlerOptions
func IsHandlerOptionsPtr(t reflect.Type) bool {
	// Must be a pointer
	if t.Kind() != reflect.Ptr {
		return false
	}

	// That pointer must point to a HandlerOptions
	if t.Elem().Name() != "HandlerOptions" {
		return false
	}

	return true
}

// A handler instantiator has the signature
// func() (*HandlerOptions, func(context.Context, interface{}) (interface{}, error))
func IsHandlerInstantiator(t reflect.Type) bool {
	// A handler instantiator must not take any arguments
	if t.NumIn() > 0 {
		return false
	}

	// A handler instantiator must return exactly two outputs
	if t.NumOut() != 2 {
		return false
	}

	// type check the two outputs
	return IsHandlerOptionsPtr(t.Out(0)) && IsHandler(t.Out(1))
}