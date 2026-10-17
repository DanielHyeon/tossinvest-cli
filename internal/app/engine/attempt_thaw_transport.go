package engine

// attempt_thaw_transport.go 는 a094 park 해동 route 를 엔진의 인증 loopback endpoint 에 얹음 — a066 해제 route 와 같은
// listener · bearer 토큰 · private descriptor. capability 없는 빌드는 route 집합 불변. 콘솔은 이 route 를 부르지 않음.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
)

// AttemptThawPath 는 해동 route 임(클라이언트와 공유하는 값 — 시험이 두 값이 같음을 고정).
const AttemptThawPath = "/v1/attempt-thaw/resolve"

// attemptThawCommands 는 StartPositionPolicyCommandServer 가 찾는 선택 capability 임.
type attemptThawCommands interface {
	ResolveParkedAttempt(context.Context, attemptthaw.Request) (attemptthaw.Result, error)
}

func registerAttemptThawRoute(mux *http.ServeMux, server *PositionPolicyCommandServer, token string,
	commands attemptThawCommands) {
	mux.HandleFunc(AttemptThawPath, server.auth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "POST required")
			return
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			writeRPCError(w, http.StatusUnsupportedMediaType, "invalid", "application/json required")
			return
		}
		const maxAttemptThawRequestBytes = 8 << 10
		body, err := io.ReadAll(io.LimitReader(r.Body, maxAttemptThawRequestBytes+1))
		if err != nil || len(body) > maxAttemptThawRequestBytes {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
			return
		}
		var req attemptthaw.Request
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
		result, err := commands.ResolveParkedAttempt(r.Context(), req)
		if err != nil {
			writeAttemptThawRPCError(w, err)
			return
		}
		// 커밋된 해소는 발의 해제가 실패해도 200 임 — 결과의 ReleaseError 가 그것을 말함.
		writeRPCJSON(w, http.StatusOK, result)
	}))
}

func writeAttemptThawRPCError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal"
	switch {
	case errors.Is(err, attemptthaw.ErrInvalidRequest):
		status, code = http.StatusBadRequest, "invalid"
	case errors.Is(err, attemptthaw.ErrStale):
		status, code = http.StatusPreconditionFailed, "stale"
	case errors.Is(err, attemptthaw.ErrNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, attemptthaw.ErrAuditUnavailable):
		status, code = http.StatusServiceUnavailable, "audit_unavailable"
	case errors.Is(err, attemptthaw.ErrUnwired):
		status, code = http.StatusNotImplemented, "unwired"
	}
	writeRPCError(w, status, code, err.Error())
}
