//go:build unix

package engine

// mode_control_transport_unix.go 는 운영자의 모드 완화 CLI 가 붙는 소켓임(a092 착지 단위 ④).
//
// 알림 제어 소켓과 **같은 기계**(회수 · staged bind · 0600 검증 · descriptor 발행 · 닫기)를 쓰고, 이름 · 토큰만 따로 가짐.
// descriptor 발행 의례는 `publishPrivateDescriptor`(alert_control_transport_unix.go)를 그대로 부름 — 네 번째 사본을 만들지
// 않음. 그래서 staging 접두도 그 공유 상수(privateDescriptorStagingPrefix)임.
//
// # 이 descriptor 를 가진 사람이 할 수 있는 것
//
// 계좌의 운영 모드를 사람이 완화(또는 조임)하는 것 하나. 주문 · 취소 · 토글 · 엔진 기동/정지는 못 함 — ModeOperations 는
// 메서드가 하나이고 게이트웨이에 닿지 않음. 완화 판정(OPERATOR · 승인 · audit 가 commit 앞)은 원장이 함.

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// modeControlEndpointNames 는 이 엔드포인트가 자기 control 디렉터리에 만들 수 있는 이름의 전부임.
func modeControlEndpointNames() positionpolicyrpc.PrivateEndpointNames {
	return positionpolicyrpc.PrivateEndpointNames{
		Descriptor: modeControlDescriptorFileName,
		Socket:     modeControlSocketFileName,
		StagingPrefixes: []string{
			positionpolicyrpc.StagingPrefix,
			privateDescriptorStagingPrefix,
		},
	}
}

// ModeControlServer 는 엔진의 모드 완화 엔드포인트임.
type ModeControlServer struct {
	server     *http.Server
	listener   net.Listener
	descriptor string
	socket     string
	controlDir string
	once       sync.Once
}

// StartModeControlServer 는 소켓을 열고 라우트를 섬김. 실패하면 만든 것을 되감고 거절함(반쯤 만든 엔드포인트는 붙을 수 있는데
// 아무 일도 안 하는 소켓이라 「완화 경로가 있다」로 읽힘).
func StartModeControlServer(engineDir string, ops *ModeOperations) (*ModeControlServer, error) {
	if ops == nil {
		return nil, errors.New("engine: the mode control server has no operations")
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine: the mode control server has no engine directory")
	}
	if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {
		return nil, fmt.Errorf("engine: validating the mode control engine directory: %w", err)
	}
	controlDir := ModeControlDirectory(dir)
	if err := openReclaimedControlDirectory(controlDir, "the mode control directory",
		modeControlEndpointNames()); err != nil {
		return nil, err
	}
	cleanupControlDir := func() { _ = os.Remove(controlDir) }
	if err := positionpolicyrpc.ValidatePrivateControlDirectory(controlDir); err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: validating the mode control directory: %w", err)
	}
	socketPath := ModeControlSocketPath(dir)
	listener, err := positionpolicyrpc.ListenStagedPrivateSocket(controlDir, socketPath)
	if err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: publishing the mode control socket: %w", err)
	}
	cleanupListener := func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		cleanupControlDir()
	}
	if err := positionpolicyrpc.ValidatePrivateSocket(socketPath); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: validating the mode control socket: %w", err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: generating the mode control token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	server := &ModeControlServer{
		listener:   listener,
		descriptor: ModeControlDescriptorPath(dir),
		socket:     socketPath,
		controlDir: controlDir,
	}
	server.server = &http.Server{
		Handler:           modeControlRoutes(token, ops),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		// 원격 호출은 없음. 한 요청이 audit · 커밋 · 통지 기록 · 재읽기 넷을 단일 연결 원장에서 차례로 하므로 알림 제어(5s)보다
		// 여유를 둠(주석과 값이 갈렸던 것을 게이트 준비 gstack 리뷰가 찾음 — 값은 그대로, 주석을 값에 맞춤).
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    15 * time.Second,
		MaxHeaderBytes: 8 << 10,
	}
	body, err := json.Marshal(ModeControlDescriptor{Socket: modeControlSocketFileName, Token: token, PID: os.Getpid()})
	if err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: encoding the mode control descriptor: %w", err)
	}
	if err := publishPrivateDescriptor(server.descriptor, body); err != nil {
		cleanupListener()
		return nil, err
	}
	go func() { _ = server.server.Serve(listener) }()
	return server, nil
}

// Close 는 엔드포인트를 닫고 파일을 치움 — 알림 제어와 같은 기계.
func (s *ModeControlServer) Close() error {
	if s == nil {
		return nil
	}
	var result error
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result = s.server.Shutdown(ctx)
		result = closePrivateEndpointFiles(result, s.listener, s.descriptor, s.socket, s.controlDir)
	})
	return result
}

func modeControlRoutes(token string, ops *ModeOperations) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(ModeControlReleasePath, alertControlAuth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "POST required")
			return
		}
		req, ok := decodeModeReleaseRequest(w, r)
		if !ok {
			return
		}
		result, err := ops.Release(r.Context(), req)
		if err != nil {
			switch {
			// 운영자 입력 실수(빈 칸 · 승인 없음 · 허용 밖 모드)는 400 — 엔진 고장(500)과 가름.
			case errors.Is(err, ErrModeReleaseInvalid), errors.Is(err, journal.ErrModeApprovalRequired),
				errors.Is(err, journal.ErrInvalidRequest):
				writeRPCError(w, http.StatusBadRequest, "invalid", err.Error())
			case errors.Is(err, ErrModeReleaseUnavailable):
				writeRPCError(w, http.StatusServiceUnavailable, "unavailable", err.Error())
			default:
				// 본문은 고정 문구 — 원장 오류 문구는 계좌를 담을 수 있음(26라운드 보이스 B #2). 원문은 가려서 로그에.
				ops.logFailure("the mode release failed", err)
				writeRPCError(w, http.StatusInternalServerError, "internal", modeReleaseInternalFailure)
			}
			return
		}
		writeRPCJSON(w, http.StatusOK, result)
	}))
	return mux
}

func decodeModeReleaseRequest(w http.ResponseWriter, r *http.Request) (ModeReleaseRequest, bool) {
	var req ModeReleaseRequest
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeRPCError(w, http.StatusUnsupportedMediaType, "invalid", "application/json required")
		return req, false
	}
	const maxModeControlRequestBytes = 8 << 10
	body, err := io.ReadAll(io.LimitReader(r.Body, maxModeControlRequestBytes+1))
	if err != nil || len(body) > maxModeControlRequestBytes {
		writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
		return req, false
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&req); err != nil {
		writeRPCError(w, http.StatusBadRequest, "invalid", "request JSON rejected")
		return req, false
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		writeRPCError(w, http.StatusBadRequest, "invalid", "request must contain one JSON value")
		return req, false
	}
	return req, true
}
