package foo

import (
	"context"
	"fmt"
	"github.com/ebauman/frazer/example/types"
	"github.com/ebauman/frazer/frazer"
)

type Server struct {
}

// Example of a handler instantiator.
func (s *Server) Update() (*frazer.HandlerOptions, func(context.Context, types.Foo) (types.Foo, error)) {
	h := &frazer.HandlerOptions{
		Path:   "/eamon/update/foo",
	}

	return h, func(ctx context.Context, f types.Foo) (types.Foo, error) {
		return f, nil
	}
}

// Example of a server handler. Prints "Calling SERVER" to stdout to demonstrate
// priority of registered handlers (in order of registration)
func (s *Server) Create(ctx context.Context, f *types.Foo) (*types.Foo, error) {
	fmt.Println("Calling SERVER")
	return f, nil
}

// Example of a server handler with path arguments. The path argument returns to the caller
// as the name field of the Foo struct
func (s *Server) List(ctx context.Context, _ interface{}, id string) ([]types.Foo, error) {
	return []types.Foo{
		{
			Name: id,
		},
		{
			Name: id,
		},
	}, nil
}