package actionlog

import (
	"context"
)

// Key to use when setting the request ID.
type ctxKey int

const (
	RequestOrigin_APP_Init            = "app/init"
	RequestOrigin_APP_Serve           = "app/serve"
	RequestOrigin_APP_Upgrade         = "app/upgrade"
	RequestOrigin_APP_Activate        = "app/activate"
	RequestOrigin_APP_Provision       = "app/provision"
	RequestOrigin_APP_Run             = "app/run"
	RequestOrigin_HTTPServer_API_REST = "app/http-server/api/rest"
	RequestOrigin_HTTPServer_API_GRPC = "app/http-server/api/grpc"
	RequestOrigin_CLI                 = "app/cli"
)

// RequestOriginKey is the key that holds th unique request ID in a request context.
const requestOriginKey ctxKey = 0

// RequestOriginToContext stores request origin to context
func RequestOriginToContext(ctx context.Context, origin string) context.Context {
	return context.WithValue(ctx, requestOriginKey, origin)
}

// RequestOriginFromContext returns remote IP address from context
func RequestOriginFromContext(ctx context.Context) string {
	v := ctx.Value(requestOriginKey)
	if str, ok := v.(string); ok {
		return str
	}

	return ""
}