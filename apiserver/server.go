package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ebauman/frazer/frazer"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

type APIServer struct {
	pkg      string
	handlers map[string]map[string]reflect.Value
	schemas  map[string]reflect.Type
	schemaNames map[string]string
	middleware map[string][]reflect.Value // key is path
	// handlers key (string) is http method as defined in https://github.com/golang/go/blob/master/src/net/http/method.go
	router *mux.Router
}

func New(opts *frazer.FrazerOptions, ) *APIServer {
	a := &APIServer{}
	a.router = mux.NewRouter()
	handlers := make(map[string]map[string]reflect.Value)
	a.handlers = handlers
	a.schemas = make(map[string]reflect.Type)
	a.middleware = make(map[string][]reflect.Value)
	a.schemaNames = make(map[string]string)
	if opts != nil {
		if len(opts.Package) > 0 {
			a.pkg = opts.Package
		}
	}
	return a
}

func (a *APIServer) dispatch(w http.ResponseWriter, r *http.Request) {
	rt := mux.CurrentRoute(r)
	p, err := rt.GetPathTemplate()
	if err != nil {
		handleError(w, r, frazer.NewError(http.StatusBadRequest, "invalid path"))
		return
	}
	if item, exists := a.handlers[p]; exists {
		m, _ := frazer.MethodFromString(r.Method)
		if h, exists := item[m]; exists {
			a.handle(w, r, p, h)
		}
	} else {
		handleError(w, r, frazer.NewError(400, "not found"))
	}
}

func (a *APIServer) handle(w http.ResponseWriter, r *http.Request, matchedPath string, handler reflect.Value) {
	var handlerType = handler.Type()

	// get any path parameters
	pathParams := mux.Vars(r)
	// get the number of inputs to the handler
	// and if inputs-2 (accounting for context, and body) does not equal path params
	// there is a problem and we should bail
	if handler.Type().NumIn()-2 != len(pathParams) {
		handleError(w, r, frazer.NewError(http.StatusInternalServerError, "invalid number of path parameter defined in handler"))
		return
	}

	// construct a context and arguments, send into the handler.
	queryMap, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		handleError(w, r, err)
		return
	}

	ctx := r.Context()                                 // base context is the request, handlers can then use it for cancellation
	ctx = context.WithValue(ctx, "queryMap", queryMap) // add the query map

	params := make([]reflect.Value, handler.Type().NumIn())
	params[0] = reflect.ValueOf(ctx)

	// get the type associated with the second parameter
	bodyType := handlerType.In(1)
	obj := reflect.New(bodyType)

	// reflected pointer
	newP := obj.Interface()

	// figure out how to handle situations where no arguments
	// are required. do we have a handler that only accepts
	// context? e.g. func(context)(interface{}, error)
	// or do we force the user to specify func(context, _ interface{})?

	// also what about situations where there is no return value?
	// is that ever an occurrence? e.g.
	// func(context, interface{}) (error)

	// for the marshalling of the body, probably perform ioutil.ReadAll(r.Body)
	// and then depending on the len(data), marshal or not.
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		handleError(w, r, err)
		return
	}

	if len(data) > 0 {
		err = json.Unmarshal(data, newP)
		if err != nil {
			handleError(w, r, err)
			return
		}
	}

	if handlerType.In(1).Kind() == reflect.Interface {
		params[1] = obj
	} else {
		params[1] = obj.Elem()
	}

	if len(params) > 2 {
		// there are path parameters
		i := 2
		for _, v := range pathParams {
			params[i] = reflect.ValueOf(v)
			i++
		}
	}

	// recovery func, but defined inline so as to use w, r
	defer func() {
		if rr := recover(); rr != nil {
			handleError(w, r, frazer.NewError(500, "internal error while calling http handler"))
		}
	}()

	params = Middleware(params, a.middleware[matchedPath])

	results := handler.Call(params)
	// results[0] should be interface{}, results[1] should be error
	if len(results) < 2 {
		err = frazer.NewError(500, "handler did not return as expected")
		handleError(w, r, err)
		return
	}

	if reflect.TypeOf(results[1]).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		if !results[0].IsNil() {
			err, ok := results[1].Interface().(error)
			if !ok {
				err = frazer.NewError(500, "could not convert non-nil error into error type")
			}
			handleError(w, r, err)
			return
		}
	}

	handleData(w, r, results[0].Interface())
}

func handleData(w http.ResponseWriter, r *http.Request, data interface{}) {
	j, err := json.Marshal(data)
	if err != nil {
		marshalError(w, r, err)
	}

	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(200)
	_, _ = w.Write(j)
}

func marshalError(w http.ResponseWriter, _ *http.Request, err error) {
	w.WriteHeader(500)
	_, writeErr := w.Write([]byte(fmt.Sprintf("error while encoding response: %s", err)))
	if writeErr != nil {
		log.Fatalf("unable to write error response: %s", writeErr)
	}
}

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	var code int
	if ee, ok := err.(frazer.Error); ok {
		code = ee.Code()
	} else {
		code = 500
	}

	// build error
	e := frazer.NewErrorResponse(uint(code), "error", err.Error(), err.Error())

	eb, err := json.Marshal(e)
	if err != nil {
		marshalError(w, r, err)
		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(eb)
}

func (a *APIServer) ListenAndServe(host string, port int) error {
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), a.router)
	return err
}

func Middleware(params []reflect.Value, middlewares []reflect.Value) []reflect.Value {
	for i := len(middlewares)-1; i >= 0; i-- {
		params = middlewares[i].Call(params)
	}

	return params
}
