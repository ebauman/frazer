package frazer

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

const (
	// plugin stages
	PluginBeforeDispatch PluginStage = "BeforeDispatch"
	PluginBeforeHandle   PluginStage = "BeforeHandle"
	PluginBeforeResponse PluginStage = "BeforeResponse"
	PluginResponse       PluginStage = "Response"
	PluginBeforeError    PluginStage = "BeforeError"
	PluginError          PluginStage = "Error"
)

type PluginStage string

type HandlePlugin func(ctx context.Context, data interface{}, params ...string)

type ResponseHandler func(w http.ResponseWriter, r *http.Request, data interface{})
type ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

type BeforeHandle func(next HandlePlugin) HandlePlugin


type PluginOptions struct {
	Paths      []string // follows the rules of gorilla/mux for path matching
	Methods    []string
}

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
