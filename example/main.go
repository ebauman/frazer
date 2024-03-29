package main

import (
	"context"
	"fmt"
	"github.com/ebauman/frazer/apiserver"
	"github.com/ebauman/frazer/example/api/v1/foo"
	"github.com/ebauman/frazer/example/middleware"
	"github.com/ebauman/frazer/example/types"
	"github.com/ebauman/frazer/frazer"
)

type Bar struct {

}

func main() {
	host := "localhost"
	port := 8080

	a := apiserver.New(&frazer.FrazerOptions{Package: "github.com/ebauman/frazer/example"})

	a.RegisterMiddleware("/api/v1/foos", middleware.LogRequestMiddleware)

	a.RegisterHandler(ListFoos, &frazer.HandlerOptions{
		Path:   "/api/v1/foos",
		Method: "GET",
	})

	a.RegisterHandler(CreateFoo, &frazer.HandlerOptions{
		Prefix: "/api/v1",
	})

	s := foo.Server{}

	a.RegisterServer(&s, nil)

	err := a.ListenAndServe(host, port)
	fmt.Print(err)
}

func ListFoos(ctx context.Context, _ interface{}) ([]types.Foo, error) {
	return []types.Foo{
		{
			Name: "Eamon",
		},
		{
			Name: "Courtney",
		},
	}, nil
}

func CreateFoo(ctx context.Context, f *types.Foo) (Bar, error) {
	fmt.Println("calling HANDLER")
	return Bar{}, nil
}