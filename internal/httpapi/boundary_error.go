package httpapi

import (
	"net/http"
	"time"
)

// WriteBoundaryRejection keeps failures from the outer network boundary inside
// the same stable JSON/no-store contract as router and mutation refusals.
func WriteBoundaryRejection(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.WriteHeader(http.StatusForbidden)
	if r != nil && r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(errorResponseBody("BOUNDARY_REFUSED",
		"The request was refused by the private canonical transport boundary.", nil, time.Now()))
}
