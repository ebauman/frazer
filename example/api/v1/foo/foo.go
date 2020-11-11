package foo

import (
	"context"
	"fmt"
	"github.com/ebauman/frazer/example/types"
)

type Server struct {
}


func (s *Server) Create(ctx context.Context, f types.Foo) (types.Foo, error) {
	fmt.Println("Calling SERVER")
	return f, nil
}