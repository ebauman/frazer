package middleware

import (
	"context"
	"fmt"
)

func LogRequestMiddleware(ctx context.Context, body interface{}, params ...interface{}) (context.Context, interface{}, []interface{}) {
	fmt.Print("calling log request middleware")
	fmt.Println(body)

	return ctx, body, params
}
