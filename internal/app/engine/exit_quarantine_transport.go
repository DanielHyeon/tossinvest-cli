package engine

// exit_quarantine_transport.go carries change a079's three routes on the engine's
// existing authenticated loopback endpoint.
//
// They ride the same listener, the same bearer token and the same private
// descriptor as the policy routes because the trust boundary is identical: a
// caller that can read the 0700 descriptor is already inside the engine's own
// directory. What is deliberately not shared is the error vocabulary — a
// quarantine release has refusals the policy pipeline has no word for, and
// mapping them onto the policy codes would make an operator's screen lie about
// which control plane refused.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/exitquarantine"
)

// exitQuarantineCommands is the optional capability StartPositionPolicyCommandServer
// looks for. An engine build without it serves exactly the route set it served
// before a079 — that is how §0.2 is satisfied at the transport layer.
type exitQuarantineCommands interface {
	Quarantines(context.Context) ([]exitquarantine.Row, error)
	PreviewQuarantineRelease(context.Context, exitquarantine.Request) (exitquarantine.Preview, error)
	ReleaseQuarantine(context.Context, exitquarantine.ApplyRequest) (exitquarantine.Result, error)
}

func registerExitQuarantineRoutes(mux *http.ServeMux, server *PositionPolicyCommandServer,
	token string, commands exitQuarantineCommands) {
	mux.HandleFunc("/v1/quarantines", server.auth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "GET required")
			return
		}
		rows, err := commands.Quarantines(r.Context())
		if err != nil {
			writeExitQuarantineRPCError(w, err)
			return
		}
		if rows == nil {
			// An empty list and a null are the same fact to a Go decoder but not
			// to a reader tailing this endpoint; say "none" in one shape only.
			rows = []exitquarantine.Row{}
		}
		writeRPCJSON(w, http.StatusOK, rows)
	}))
	mux.HandleFunc("/v1/quarantine/preview", server.auth(token,
		exitQuarantineRequestHandler(commands.PreviewQuarantineRelease)))
	mux.HandleFunc("/v1/quarantine/release", server.auth(token,
		exitQuarantineRequestHandler(commands.ReleaseQuarantine)))
}

func exitQuarantineRequestHandler[Request, Response any](call func(context.Context,
	Request) (Response, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "POST required")
			return
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeRPCError(w, http.StatusUnsupportedMediaType, "invalid", "application/json required")
			return
		}
		const maxExitQuarantineRequestBytes = 8 << 10
		body, err := io.ReadAll(io.LimitReader(r.Body, maxExitQuarantineRequestBytes+1))
		if err != nil {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
			return
		}
		if len(body) > maxExitQuarantineRequestBytes {
			writeRPCError(w, http.StatusRequestEntityTooLarge, "invalid", "request body is too large")
			return
		}
		var req Request
		decoder := json.NewDecoder(bytes.NewReader(body))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&req); err != nil {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request JSON rejected")
			return
		}
		var trailing any
		if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request must contain one JSON value")
			return
		}
		result, err := call(r.Context(), req)
		if err != nil {
			writeExitQuarantineRPCError(w, err)
			return
		}
		writeRPCJSON(w, http.StatusOK, result)
	}
}

func writeExitQuarantineRPCError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal"
	switch {
	case errors.Is(err, exitquarantine.ErrInvalidRequest):
		status, code = http.StatusBadRequest, "invalid"
	case errors.Is(err, exitquarantine.ErrNotQuarantined):
		status, code = http.StatusNotFound, "not_quarantined"
	case errors.Is(err, exitquarantine.ErrVersionMismatch):
		status, code = http.StatusPreconditionFailed, "stale"
	case errors.Is(err, exitquarantine.ErrCapabilityTooEarly):
		status, code = http.StatusTooEarly, "capability_too_early"
	case errors.Is(err, exitquarantine.ErrCapabilityExpired):
		status, code = http.StatusGone, "capability_expired"
	case errors.Is(err, exitquarantine.ErrCapabilityInvalid):
		status, code = http.StatusForbidden, "capability_invalid"
	case errors.Is(err, exitquarantine.ErrConfirmationRequired):
		status, code = http.StatusBadRequest, "confirmation_required"
	case errors.Is(err, exitquarantine.ErrUnwired):
		status, code = http.StatusNotImplemented, "unwired"
	}
	writeRPCError(w, status, code, err.Error())
}
