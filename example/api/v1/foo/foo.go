package foo

import (
	"context"
	"fmt"
	"github.com/ebauman/frazer/apiserver"
	"github.com/ebauman/frazer/example/types"
)

type Server struct {
}

func (s *Server) Update() (*apiserver.HandlerOptions, func(context.Context, types.Foo) (types.Foo, error)) {
	h := &apiserver.HandlerOptions{
		Path:   "/eamon/update/foo",
	}

	return h, func(ctx context.Context, f types.Foo) (types.Foo, error) {
		return f, nil
	}
}

func (s *Server) Create(ctx context.Context, f types.Foo) (types.Foo, error) {
	fmt.Println("Calling SERVER")
	return f, nil
}