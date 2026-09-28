package engine

// risk_relaxation_transport.go 는 a066 5.5 운영자 해제 두 route 를 엔진의 기존 인증 loopback endpoint 에 얹음.
//
// a079 격리 해제(exit_quarantine_transport.go)와 같은 구성임 — 같은 listener · 같은 bearer 토큰 · 같은 private
// descriptor. 신뢰 경계가 같음(0700 descriptor 를 읽을 수 있는 호출자는 이미 엔진 디렉터리 안에 있음). 오류 어휘는
// 따로 둠 — 해제의 거절(stale · audit 없음)을 정책 코드로 바꾸면 어느 제어면이 거절했는지 화면이 거짓말을 함.
// capability 가 없는 빌드는 이전과 정확히 같은 route 집합을 냄.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskrelaxation"
)

// RiskRelaxationEntryLockPath · RiskRelaxationLatchPath 는 두 해제 route 임(클라이언트와 공유하는 값).
const (
	RiskRelaxationEntryLockPath = "/v1/risk-relaxation/entry-lock-release"
	RiskRelaxationLatchPath     = "/v1/risk-relaxation/risk-latch-release"
)

// riskRelaxationCommands 는 StartPositionPolicyCommandServer 가 찾는 선택 capability 임.
type riskRelaxationCommands interface {
	ReleaseEntryLossLock(context.Context, riskrelaxation.EntryLockReleaseRequest) (riskrelaxation.Result, error)
	ReleaseRiskOverageLatch(context.Context, riskrelaxation.LatchReleaseRequest) (riskrelaxation.Result, error)
}

func registerRiskRelaxationRoutes(mux *http.ServeMux, server *PositionPolicyCommandServer,
	token string, commands riskRelaxationCommands) {
	mux.HandleFunc(RiskRelaxationEntryLockPath, server.auth(token,
		riskRelaxationRequestHandler(commands.ReleaseEntryLossLock)))
	mux.HandleFunc(RiskRelaxationLatchPath, server.auth(token,
		riskRelaxationRequestHandler(commands.ReleaseRiskOverageLatch)))
}

func riskRelaxationRequestHandler[Request any](call func(context.Context,
	Request) (riskrelaxation.Result, error)) http.HandlerFunc {
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
		const maxRiskRelaxationRequestBytes = 8 << 10
		body, err := io.ReadAll(io.LimitReader(r.Body, maxRiskRelaxationRequestBytes+1))
		if err != nil {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
			return
		}
		if len(body) > maxRiskRelaxationRequestBytes {
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
			writeRiskRelaxationRPCError(w, err)
			return
		}
		// 커밋된 해제는 통지 실패여도 200 임 — 결과의 Notified=false 가 「완화됨·통지 실패」를 말함.
		writeRPCJSON(w, http.StatusOK, result)
	}
}

func writeRiskRelaxationRPCError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal"
	switch {
	case errors.Is(err, riskrelaxation.ErrInvalidRequest):
		status, code = http.StatusBadRequest, "invalid"
	case errors.Is(err, riskrelaxation.ErrStale):
		status, code = http.StatusPreconditionFailed, "stale"
	case errors.Is(err, riskrelaxation.ErrStateMismatch):
		status, code = http.StatusConflict, "state_mismatch"
	case errors.Is(err, riskrelaxation.ErrAuditUnavailable):
		status, code = http.StatusServiceUnavailable, "audit_unavailable"
	case errors.Is(err, riskrelaxation.ErrUnwired):
		status, code = http.StatusNotImplemented, "unwired"
	}
	writeRPCError(w, status, code, err.Error())
}
