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

// Example of a plain server handler. Added "Calling SERVER" to demonstrate priority of handler calls
// when two paths are registered, one with RegisterHandler, one with RegisterServer (handler takes priority)
func (s *Server) Create(ctx context.Context, f *types.Foo) (*types.Foo, error) {
	fmt.Println("Calling SERVER")
	return f, nil
}