package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ebauman/frazer/frazer"
	frazerHttp "github.com/ebauman/frazer/http"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
)

type APIServer struct {
	pkg      string
	handlers map[string]map[frazerHttp.Method]reflect.Value
}

func New(opts *frazer.FrazerOptions) *APIServer {
	a := &APIServer{}
	handlers := make(map[string]map[frazerHttp.Method]reflect.Value)
	a.handlers = handlers
	if opts != nil {
		if len(opts.Package) > 0 {
			a.pkg = opts.Package
		}
	}
	return a
}

func (a *APIServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.dispatch(w, r)
}

func (a *APIServer) dispatch(w http.ResponseWriter, r *http.Request) {
	if item, exists := a.handlers[r.URL.Path]; exists {
		m, _ := frazerHttp.FromString(r.Method)
		if h, exists := item[m]; exists {
			a.handle(w, r, h)
		}
	} else {
		handleError(w, r, frazerHttp.New(400, "not found"))
	}
}

func (a *APIServer) handle(w http.ResponseWriter, r *http.Request, handler reflect.Value) {
	var handlerType = handler.Type()

	// a valid function should have the signature
	// func(ctx context.Context, body interface{})

	// construct a context and arguments, send into the handler.
	queryMap, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		handleError(w, r, err)
	}

	ctx := r.Context() // base context is the request, handlers can then use it for cancellation
	ctx = context.WithValue(ctx, "queryMap", queryMap) // add the query map

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

	results := handler.Call(params)
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
	e := frazer.NewError(uint(code), "error", err.Error(), err.Error())

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