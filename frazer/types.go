package frazer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Middleware func(ctx context.Context, body interface{}, params ...interface{}) (context.Context, interface{}, []interface{})

type HandlerOptions struct {
	Path   string
	Prefix string
	Method string
}

type ServerOptions struct {
	Prefix         string
	HandlerOptions map[string]HandlerOptions
}

type FrazerOptions struct {
	Package string
}

type ErrorResponse struct {
	Type    string `json:"type,omitempty"`
	Status  uint   `json:"status,omitempty"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func NewErrorResponse(status uint, code string, message string, detail string) ErrorResponse {
	return ErrorResponse{
		Type:    "error",
		Status:  status,
		Code:    code,
		Message: message,
		Detail:  detail,
	}
}

func MethodFromString(method string) (string, error) {
	method = strings.ToUpper(method)
	switch method {
	case "QUERY":
		fallthrough
	case "LIST":
		fallthrough
	case "GET":
		return http.MethodGet, nil
	case "CREATE":
		fallthrough
	case "POST":
		return http.MethodPost, nil
	case "UPDATE":
		fallthrough
	case "PUT":
		return http.MethodPut, nil
	case "DELETE":
		return http.MethodDelete, nil
	case "HEAD":
		return http.MethodHead, nil
	case "PATH":
		return http.MethodPatch, nil
	case "CONNECT":
		return http.MethodConnect, nil
	case "OPTIONS":
		return http.MethodOptions, nil
	case "TRACE":
		return http.MethodTrace, nil
	default:
		return http.MethodGet, fmt.Errorf("unable to convert method string %s into methods.Method", method)
	}
}
