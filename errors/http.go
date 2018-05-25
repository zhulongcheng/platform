package errors

import (
	"context"
	"fmt"
	"net/http"
)

// HandleHTTP sets the X-Influx-Error and X-Influx-Reference headers on the response,
// and sets the response status to the corresponding status code.
func HandleHTTP(ctx context.Context, e TypedError, w http.ResponseWriter) {
	if e == nil {
		return
	}

	w.Header().Set("X-Influx-Error", e.Error())
	w.Header().Set("X-Influx-Reference", fmt.Sprintf("%d", e.Reference()))
	w.WriteHeader(typCode[e.Reference()])
}
