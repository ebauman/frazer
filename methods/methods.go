package methods

import (
	"fmt"
	"strings"
)

const (
	Get     HTTPMethod = "GET"
	Head    HTTPMethod = "HEAD"
	Post    HTTPMethod = "POST"
	Put     HTTPMethod = "PUT"
	Delete  HTTPMethod = "DELETE"
	Options HTTPMethod = "OPTIONS"
	Patch   HTTPMethod = "PATCH"
	// Connect HTTPMethod = "connect" // figure out how to handle websockets at some point
)
type HTTPMethod string

func FromString(method string) (HTTPMethod, error) {
	method = strings.ToUpper(method)
	switch method {
	case "GET":
		fallthrough
	case "QUERY":
		fallthrough
	case "LIST":
		return Get, nil
	case "POST":
		fallthrough
	case "CREATE":
		return Post, nil
	case "PUT":
		fallthrough
	case "UPDATE":
		return Put, nil
	case "DELETE":
		return Delete, nil
	case "PATCH":
		return Patch, nil
	case "HEAD":
		return Head, nil
	case "OPTIONS":
		return Options, nil
	default:
		return Get, fmt.Errorf("unable to convert method string %s into methods.HTTPMethod", method)
	}
}