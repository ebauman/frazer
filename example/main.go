package main

import (
	"context"
	"fmt"
	"github.com/ebauman/frazer/apiserver"
	"github.com/ebauman/frazer/example/api/v1/foo"
	"github.com/ebauman/frazer/example/types"
)

//func main() {
//	arguments := args.Default()
//	arguments.InputDirs = []string{"github.com/ebauman/frazer/example/handlers"}
//	if err := arguments.Execute(
//		generators.NameSystems(),
//		generators.DefaultNameSystem(),
//		generators.Packages,
//	); err != nil {
//		fmt.Printf("error: %v", err)
//		os.Exit(1)
//	}
//
//	fmt.Printf("completed successfully")
//}

func main() {
	host := "localhost"
	port := 8080

	a := apiserver.New(&apiserver.FrazerOptions{Package: "github.com/ebauman/frazer/example"})

	a.RegisterHandler(ListFoos, &apiserver.HandlerOptions{
		Path:   "/api/v1/foos",
		Method: "GET",
	})

	a.RegisterHandler(CreateFoo, &apiserver.HandlerOptions{
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

func CreateFoo(ctx context.Context, f *types.Foo) (types.Foo, error) {
	fmt.Println("calling HANDLER")
	return *f, nil
}