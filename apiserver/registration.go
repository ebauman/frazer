package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	frazerHttp "github.com/ebauman/frazer/http"
	"github.com/ebauman/frazer/methods"
	"github.com/ebauman/frazer/types"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"regexp"
	"runtime"
	"strings"
)

type APIServer struct {
	pkg            string
	handlers       map[string]map[methods.HTTPMethod]interface{} // '/api/v1/foos' => ['GET' => handler, 'POST' => handler]
	serverHandlers map[string]map[methods.HTTPMethod]reflect.Value
}

type HandlerOptions struct {
	Path   string
	Prefix string
	Method methods.HTTPMethod
}

type ServerOptions struct {
	Prefix         string
	HandlerOptions map[string]HandlerOptions
}

type FrazerOptions struct {
	Package string // package name to remove from prefix of import PkgPath()
}

func New(opts *FrazerOptions) *APIServer {
	a := &APIServer{}
	handlers := make(map[string]map[methods.HTTPMethod]interface{})
	a.handlers = handlers
	serverHandlers := make(map[string]map[methods.HTTPMethod]reflect.Value)
	a.serverHandlers = serverHandlers
	if opts != nil {
		if len(opts.Package) > 0 {
			a.pkg = opts.Package
		}
	}
	return a
}

func (a *APIServer) RegisterHandler(handler interface{}, options *HandlerOptions) {
	if reflect.ValueOf(handler).Kind() != reflect.Func {
		panic(fmt.Sprintf("handler %v was not a Func", reflect.TypeOf(handler)))
	}
	var path string
	var prefix string
	var method methods.HTTPMethod

	if options != nil {
		path = options.Path
		prefix = options.Prefix
		method = options.Method
	}

	if len(path) == 0 || len(method) == 0 {
		handlerName := runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name()
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

	if len(prefix) > 0 {
		path = fmt.Sprintf("%s/%s", prefix, path)
	}

	_, exists := a.handlers[path]
	if !exists {
		a.handlers[path] = map[methods.HTTPMethod]interface{}{}
	}

	a.handlers[path][method] = handler
}

func (a *APIServer) registerServerHandler(handlerName string, handler reflect.Value, options *HandlerOptions) {
	var path string
	var prefix string
	var method methods.HTTPMethod
	var err error

	if options != nil {
		path = options.Path
		prefix = options.Prefix
		method = options.Method
	}

	if len(path) == 0 || len(method) == 0 {
		path, method, err = decodePathAndMethod(handlerName)
		if err != nil {
			panic(fmt.Sprintf("could not decode method: %v", err))
		}
	}

	if len(prefix) > 0 && len(path) > 0 {
		path = fmt.Sprintf("%s/%s", prefix, path)
	} else if len(prefix) > 0 {
		path = prefix
	}

	_, exists := a.serverHandlers[path]
	if !exists {
		a.serverHandlers[path] = map[methods.HTTPMethod]reflect.Value{}
	}

	a.serverHandlers[path][method] = handler
}

func decodePathAndMethod(handlerName string) (string, methods.HTTPMethod, error) {
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
		return "", methods.Get, fmt.Errorf("invalid handler name, incompatible with autodetection: %s", handlerName)
	}

	method, err := methods.FromString(res[1])
	if err != nil {
		return "", methods.Get, fmt.Errorf("could not autodetect http method for handler %s", handlerName)
	}

	path := fmt.Sprintf("%s", strings.ToLower(res[2]))

	return path, method, nil
}

func (a *APIServer) RegisterServer(server interface{}, options *ServerOptions) {
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
		if isHandlerInstantiator(reflect.ValueOf(server).Method(i).Type()) {
			// call and register
			res := svrValue.Method(i).Call([]reflect.Value{})
			hOptsIntf, ok := res[0].Interface().(*HandlerOptions)
			if !ok {
				fmt.Println("cannot assert interface{} to *HandlerOptions")
			}

			hIntf := res[0].Interface()         // convert handler to interface{}
			a.RegisterHandler(hIntf, hOptsIntf) // register handler
		}

		// we are not dealing with a handler instantiator. are we dealing with a handler?
		if isHandler(svrValue.Method(i).Type()) {
			// register this handler. no HandlerOptions are coming along for the ride
			// so we should construct our own based on the parent ServerOptions
			opts := &HandlerOptions{
				Prefix: prefix,
			}
			a.registerServerHandler(reflect.TypeOf(server).Method(i).Name, svrValue.Method(i), opts)
		}
	}
}

func isHandler(t reflect.Type) bool {
	if t.Kind() != reflect.Func {
		return false
	}

	if t.NumIn() != 2 {
		return false
	}

	if t.In(0).Name() != "Context" {
		return false
	}

	// don't check the second argument as it can be interface{}
	return true
}

func isHandlerOptionsPtr(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return false
	}

	if t.Elem().Name() != "HandlerOptions" {
		return false
	}

	return true
}

func isHandlerInstantiator(t reflect.Type) bool {
	// handler instantiators take no arguments
	if t.NumIn() > 0 {
		return false
	}

	// handler instantiators return exactly two outs
	if t.NumOut() != 2 {
		return false
	}

	return isHandlerOptionsPtr(t.Out(0)) && isHandler(t.Out(1))
}

func (a *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.dispatch(w, r)
}

func (a *APIServer) dispatch(w http.ResponseWriter, r *http.Request) {
	// first attempt to match a strict handler for this
	if item, exists := a.handlers[r.URL.Path]; exists {
		m, _ := methods.FromString(r.Method)
		if h, exists := item[m]; exists {
			a.handle(w, r, h)
		}
	} else if item, exists := a.serverHandlers[r.URL.Path]; exists {
		m, _ := methods.FromString(r.Method)
		if h, exists := item[m]; exists {
			a.handle(w, r, h)
		}
	} else {
		handleError(w, r, frazerHttp.New(400, "not found"))
	}
}

func (a *APIServer) handle(w http.ResponseWriter, r *http.Request, handler interface{}) {
	var handlerValue reflect.Value
	var handlerType reflect.Type
	handlerValue, ok := handler.(reflect.Value)

	if !ok {
		handlerType = reflect.TypeOf(handler)
		handlerValue = reflect.ValueOf(handler)
		if reflect.ValueOf(handler).Kind() != reflect.Func {
			panic(fmt.Sprintf("handler %v was not a Func", reflect.TypeOf(handler)))
		}
	} else {
		handlerType = handlerValue.Type()
	}

	// a valid function should have the signature
	// func(ctx context.Context, body interface{})

	// construct a context and arguments, send into the handler.
	queryMap, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		handleError(w, r, err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "queryMap", queryMap)

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(ctx)

	// get the type associated with the second parameter
	bodyType := handlerType.In(1)
	obj := reflect.New(bodyType)

	// reflected pointer
	newP := obj.Interface()

	if r.Method != "GET" {
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			handleError(w, r, err)
		}

		err = json.Unmarshal(data, newP)
		if err != nil {
			handleError(w, r, err)
		}
	}

	if handlerType.In(1).Kind() == reflect.Interface {
		params[1] = obj
	} else {
		params[1] = obj.Elem()
	}

	defer recoverHandlerCall()

	results := handlerValue.Call(params)
	// results[0] should be interface{}, results[1] should be error
	if len(results) < 2 {
		panic(fmt.Sprintf("did not get expected results from handler call, received %v", results))
	}

	if reflect.TypeOf(results[1]).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !results[0].IsNil() {
			err, ok := results[1].Interface().(error)
			if !ok {
				panic("could not turn non-nil error into error type")
			}
			handleError(w, r, err)
		}
	}

	handleData(w, r, results[0].Interface())
}

func recoverHandlerCall() {
	if r := recover(); r != nil {
		fmt.Println("recovered from failed handler call ", r)
	}
}

func handleData(w http.ResponseWriter, r *http.Request, data interface{}) {
	// build response
	// is data a collection?
	// just shit out json for now
	j, err := json.Marshal(data)
	if err != nil {
		marshalError(w, r, err)
	}

	w.WriteHeader(200)
	_, _ = w.Write(j)
}

func marshalError(w http.ResponseWriter, r *http.Request, err error) {
	w.WriteHeader(500)
	_, writeErr := w.Write([]byte(fmt.Sprintf("error while encoding response: %s", err)))
	if writeErr != nil {
		log.Fatalf("unable to write error response: %s", writeErr)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var code int
	if ee, ok := err.(frazerHttp.Error); ok {
		code = ee.Code()
	} else {
		code = 500
	}

	// build error
	e := types.NewError(uint(code), "error", err.Error(), err.Error())

	eb, err := json.Marshal(e)
	if err != nil {
		marshalError(w, r, err)
	}

	w.WriteHeader(code)
	_, _ = w.Write(eb)
}

func (a *APIServer) ListenAndServe(host string, port int) error {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), a)
	return err
}
