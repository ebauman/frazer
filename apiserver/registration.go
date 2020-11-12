package apiserver

import (
	"fmt"
	"github.com/ebauman/frazer/frazer"
	"github.com/ebauman/frazer/http"
	"github.com/ebauman/frazer/typecheckers"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)


func (a *APIServer) RegisterHandler(handler interface{}, options *frazer.HandlerOptions) {
	if !typecheckers.IsHandler(reflect.TypeOf(handler)) {
		panic(fmt.Sprintf("argument %s was not a valid handler", reflect.TypeOf(handler)))
	}

	a.registerServerHandler(runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name(),
		reflect.ValueOf(handler),
		options)
}

func (a *APIServer) registerServerHandler(handlerName string, handler reflect.Value, options *frazer.HandlerOptions) {
	var path string
	var prefix string
	var method http.Method

	if options != nil {
		path = options.Path
		prefix = options.Prefix
		method = options.Method
	}

	if len(path) == 0 || len(method) == 0 {
		p, m, err := decodePathAndMethod(handlerName)
		if len(path) == 0 {
			path = p
		}
		if len(method) == 0 {
			method = m
		}
		if err != nil {
			panic(fmt.Sprintf("could not decode method: %v", err))
		}
	}

	if len(prefix) > 0 && len(path) > 0 {
		path = fmt.Sprintf("%s/%s", prefix, path)
	} else if len(prefix) > 0 {
		path = prefix
	}

	_, exists := a.handlers[path]
	if !exists {
		a.handlers[path] = map[http.Method]reflect.Value{}
	}

	a.handlers[path][method] = handler
}

func decodePathAndMethod(handlerName string) (string, http.Method, error) {
	if strings.Contains(handlerName, ".") {
		// handlerName could come in as e.g. main.ListFoo, so fix that
		handlerName = strings.Split(handlerName, ".")[1]
	}
	// handlerNames that are automapped are
	// Query = GET, List = GET, Create = POST, Update = PUT,
	// Get, Put, Post, Patch, etc. are all automapped
	compoundMethod := regexp.MustCompile("(?i)(?P<method>get|head|post|put|patch|delete|options|query|list|create|update)(?P<action>\\w*)(?-i)")
	res := compoundMethod.FindStringSubmatch(handlerName)

	if len(res) < 3 {
		// there was a problem, should always be three even if the method name is simple (e.g. "Get" vs "GetFoo")
		return "", http.Get, fmt.Errorf("invalid handler name, incompatible with autodetection: %s", handlerName)
	}

	method, err := http.FromString(res[1])
	if err != nil {
		return "", http.Get, fmt.Errorf("could not autodetect http method for handler %s", handlerName)
	}

	path := fmt.Sprintf("%s", strings.ToLower(res[2]))

	return path, method, nil
}

func (a *APIServer) RegisterServer(server interface{}, options *frazer.ServerOptions) {
	var isPtr = false
	if reflect.ValueOf(server).Kind() != reflect.Struct {
		if reflect.ValueOf(server).Kind() == reflect.Ptr && reflect.ValueOf(server).Elem().Kind() != reflect.Struct {
			panic(fmt.Sprintf("server %v was not a struct or struct pointer", reflect.TypeOf(server)))
		}
		isPtr = true
	}

	// build the necessary prefix
	var prefix string
	if options != nil {
		prefix = options.Prefix
	} else {
		if isPtr {
			pkgpath := reflect.ValueOf(server).Elem().Type().PkgPath()
			prefix = strings.TrimPrefix(pkgpath, a.pkg)
		} else {
			pkgpath := reflect.TypeOf(server).PkgPath()
			prefix = strings.TrimPrefix(pkgpath, a.pkg)
		}
	}

	// register the handlers on the server
	// handlers can either provide their own HandlerOptions, or perform a lookup map
	// the difference is func(context.Context, interface{}) (interface{}, error) for a direct handler
	// or
	// func() (*HandlerOptions, func(context.Context, interface{}) (interface{}, error))
	var methodCount int
	var svrValue reflect.Value
	if reflect.TypeOf(server).Kind() == reflect.Ptr {
		methodCount = reflect.TypeOf(server).NumMethod()
		svrValue = reflect.ValueOf(server)
	} else {
		methodCount = reflect.TypeOf(server).NumMethod()
		svrValue = reflect.ValueOf(server)
	}

	for i := 0; i < methodCount; i++ {
		// determine if this function is a handler instantiator
		if typecheckers.IsHandlerInstantiator(reflect.ValueOf(server).Method(i).Type()) {
			// call and register
			res := svrValue.Method(i).Call([]reflect.Value{})
			hOptsIntf, ok := res[0].Interface().(*frazer.HandlerOptions)
			if !ok {
				fmt.Println("cannot assert interface{} to *HandlerOptions")
			}

			a.registerServerHandler(reflect.TypeOf(server).Method(i).Name, res[1], hOptsIntf) // register handler
		}

		// we are not dealing with a handler instantiator. are we dealing with a handler?
		if typecheckers.IsHandler(svrValue.Method(i).Type()) {
			// register this handler. no HandlerOptions are coming along for the ride
			// so we should construct our own based on the parent ServerOptions
			opts := &frazer.HandlerOptions{
				Prefix: prefix,
			}
			a.registerServerHandler(reflect.TypeOf(server).Method(i).Name, svrValue.Method(i), opts)
		}
	}
}


