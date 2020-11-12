package http

import (
	"fmt"
	"strings"
)

const (
	Get     Method = "GET"
	Head    Method = "HEAD"
	Post    Method = "POST"
	Put     Method = "PUT"
	Delete  Method = "DELETE"
	Options Method = "OPTIONS"
	Patch   Method = "PATCH"
	// Connect Method = "connect" // figure out how to handle websockets at some point
)
type Method string

func FromString(method string) (Method, error) {
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
		return Get, fmt.Errorf("unable to convert method string %s into methods.Method", method)
	}
}