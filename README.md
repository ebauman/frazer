# Frazer - EXPERIMENTAL

Frazer is an experimental REST API server. 

It has two goals:

1. Allow for as much autosetup as possible while also exposing customization options. 

2. Not require API handlers to deal with http.ResponseWriter or http.Request (unless they want to)

To that end, one can define handlers in two ways. First, instantiate a Frazer server...

```go
package main

import (
	"github.com/ebauman/frazer/apiserver"
)

func main() {
    f := APIServer.New("github.com/your/package")
}
```

Specifying "github.com/your/package" serves to remove that string as a prefix for the package path of any
imported servers. e.g. If your API handlers are located at `github.com/your/package/api/v1/foo`, specifying
the above value will result in a URL path of `/api/v1/foo`

Once you have instantiated a Frazer server, register a handler or server...

```go
func GetFoo(ctx context.Context, id string) (Foo, error) {
    return Foo{}
}

func main() {
...
f.RegisterHandler(GetFoo, nil)
}
```

The second argument to `RegisterHandler` is an optional `HandlerOptions` type which you can use to specify
HTTP method, prefix, or path. 

In this case, the handler would be registered at `/foo` with the http method `GET`.

You can also register a server via...

```go
-- github.com/your/package/api/v1/foo/foo.go --
package foo

type Server struct {
}

func (s *Server) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
    return f, nil
}

-- github.com/your/package --
package main

func main() {
...
s := &Server{}

f.RegisterServer(s, nil)
}
```

In this case, the handler is registered at `/api/v1/foo` with the http method `POST`. This is because internally, 
'Create' maps to 'POST' for HTTP. 

The second argument to `RegisterServer` is an optional `ServerOptions` where you can define http prefix, 
or a map of string function names to their `HandlerOptions`. One may also register these `HandlerOptions` by
specifying a receiver method of the signature...

`func (s *Server) GetFoo() (*HandlerOptions, func(ctx context.Context, f Foo) (Foo, error))`

This allows you to either define the `HandlerOptions` as a map provided when registering the server with Frazer,
or having Frazer call your defined function to retrieve the handler options (should you want to define them closer
to the context in which they are used). 

## JSON Output

This is trying pretty weakly to implement https://github.com/rancher/api-spec. 

It is not feature-complete against that spec, but the types for implementing it mostly exist. 

It is a WIP to implement those in a way that conforms with goals #1 and #2. 

## Motivation

I was tired of writing out things like

`http.Handle("/api/v1/", handler)`

over and over again. I wanted something that does most of that work for me, so I can just _barely_ think about it
and focus instead on writing my handlers. 

## Caveat Emptor

Because of the heavy use of reflection in this library, you won't receive compile-time errors if you register a handler
or server that does not conform to a signature. Instead, Frazer will panic when it detects improper
received types at runtime. 

This is a trade-off I made in order to allow for the flexibility of not having to define paths statically. 
Because Go lacks covariance, I cannot allow for a generic "handler" function and still achieve the goal of 
a handler not having to know about the http context. 

You may say, why not just shove the request body into a Context and then assert it in the handler?

This is a valid argument, and certainly a way I could have gone.
I wanted to, however, offer as much type safety to the handler methods as possible while still offering flexibility.


If you're asking yourself if this is one big workaround for not having generics in Go, you're right. 

## Next Steps

I need to implement:

1. Better reflection safety
2. A more robust error system
3. Authn and Authz
4. Lots of other things