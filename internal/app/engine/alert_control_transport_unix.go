//go:build unix

package engine

// alert_control_transport_unix.go is the socket the operator's CLI talks to.
//
// # Why a socket of its own
//
// The engine already owns two authenticated endpoints and neither can carry this:
//
//	PositionPolicyCommandServer   loopback TCP. design D7.1 chose filesystem
//	                              permissions as the access control, not a port.
//	PositionPolicyRuntimeServer   a Unix socket, and the model for this file — but
//	                              its own doc states the contract "possession of its
//	                              client cannot express Preview, Apply, or
//	                              reconciliation mutation".
//
// Acknowledging an alert clears an entry latch, which is a mutation. Adding it to
// the runtime socket would widen what that token can do for everyone already
// holding it. That is the same argument that ruled out `tossctl httpapi` — a
// surface's own statement of what it cannot do is a constraint, not a comment
// (a098 4라운드 A-P9).
//
// # What possession of *this* descriptor grants
//
// Reading the critical-alert backlog and acknowledging rows in it. It cannot
// place, cancel or modify an order, cannot flip a toggle, and cannot start or stop
// the engine: AlertOperations has two methods and neither reaches the gateway.

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// privateDescriptorStagingPrefix는 publishPrivateDescriptor가 os.CreateTemp에 넘기는
// 임시 이름의 앞머리다.
//
// 리터럴이 아니라 상수인 이유: 회수가 "우리 잔재"로 인정하는 이름 집합에 **같은 값**이
// 들어가야 한다. 두 곳에 따로 적으면 회수가 자기 잔재를 낯선 것으로 거부하고, 그 거부는
// 결정적이라 매 부팅 같은 자리에서 멈춘다. 이름-독립이라는 이 함수의 성질은 그대로다 —
// 접두는 endpoint 이름이 아니라 "발행 전"이라는 상태의 이름이다.
const privateDescriptorStagingPrefix = ".endpoint-"

// alertControlEndpointNames는 이 endpoint가 자기 control 디렉터리에 만들 수 있는 이름의
// 전부다. 완전성은 테스트가 고정한다(design D1b).
func alertControlEndpointNames() positionpolicyrpc.PrivateEndpointNames {
	return positionpolicyrpc.PrivateEndpointNames{
		Descriptor: alertControlDescriptorFileName,
		Socket:     alertControlSocketFileName,
		StagingPrefixes: []string{
			positionpolicyrpc.StagingPrefix,
			privateDescriptorStagingPrefix,
		},
	}
}

// AlertControlServer is the engine's alert operator endpoint.
type AlertControlServer struct {
	server *http.Server
	// listener는 Close가 **직접** 닫는다. 필드가 없던 판은 그 일을 Shutdown 에 맡겼고,
	// 늦게 도착한 정리가 후계자의 socket 을 지웠다(a108 A1 F5, design D2b).
	listener   net.Listener
	descriptor string
	socket     string
	controlDir string
	once       sync.Once
}

// StartAlertControlServer opens the socket and serves both routes.
//
// It refuses rather than degrades on every failure and unwinds what it made: a
// half-built endpoint is a socket an operator's command can connect to and get
// nothing from, which reads as "there is no backlog".
func StartAlertControlServer(engineDir string, ops *AlertOperations) (*AlertControlServer, error) {
	if ops == nil {
		return nil, errors.New("engine: the alert control server has no operations")
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine: the alert control server has no engine directory")
	}
	if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {
		return nil, fmt.Errorf("engine: validating the alert control engine directory: %w", err)
	}
	controlDir := AlertControlDirectory(dir)
	// Mkdir → 회수 → 재Mkdir 의례는 형제와 **같은 기계**를 쓴다(a109 §2b.3 G9 —
	// 회수를 한 곳에 두는 이유가 그 부름에는 적용되지 않을 근거가 없다).
	if err := openReclaimedControlDirectory(controlDir, "the alert control directory",
		alertControlEndpointNames()); err != nil {
		return nil, err
	}
	// 회수가 끝났으므로 이 디렉터리는 언제나 **이번 기동이 만든 것**이다. 실패 정리가
	// 조건 없이 지워도 되는 근거가 그것이다.
	cleanupControlDir := func() { _ = os.Remove(controlDir) }
	if err := positionpolicyrpc.ValidatePrivateControlDirectory(controlDir); err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: validating the alert control directory: %w", err)
	}
	socketPath := AlertControlSocketPath(dir)
	// 임시 이름에 bind → 0600 → rename. 0600 이 붙기 전 상태는 최종 이름을 가진 적이
	// 없으므로, "권한이 덜 붙은 socket 위에서 승인 라우트가 열리는" 순간이 없다
	// (design D1). 예전 판은 최종 경로에 bind 한 뒤 chmod 했고, 그 두 줄 사이의 죽음이
	// 매 부팅 거부되는 잔재를 만들었다.
	listener, err := positionpolicyrpc.ListenStagedPrivateSocket(controlDir, socketPath)
	if err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: publishing the alert control socket: %w", err)
	}
	cleanupListener := func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		cleanupControlDir()
	}
	// 발행 확인은 **정확-0600**이다. 회수의 좁은 완화가 발행으로 새지 않았다는 증거가
	// 이 호출이다(freeze P1-3).
	if err := positionpolicyrpc.ValidatePrivateSocket(socketPath); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: validating the alert control socket: %w", err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: generating the alert control token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)

	server := &AlertControlServer{
		listener:   listener,
		descriptor: AlertControlDescriptorPath(dir),
		socket:     socketPath,
		controlDir: controlDir,
	}
	server.server = &http.Server{
		Handler:           alertControlRoutes(token, ops),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       15 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}
	body, err := json.Marshal(AlertControlDescriptor{
		Socket: alertControlSocketFileName, Token: token, PID: os.Getpid(),
	})
	if err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: encoding the alert control descriptor: %w", err)
	}
	if err := publishPrivateDescriptor(server.descriptor, body); err != nil {
		cleanupListener()
		return nil, err
	}
	go func() { _ = server.server.Serve(listener) }()
	return server, nil
}

// Close shuts the endpoint and takes its files with it. A descriptor outliving
// its engine is a token pointing at a socket nobody is listening on, and the
// command that reads it would report "the engine is not running" as a dial error
// rather than as the fact it is.
func (s *AlertControlServer) Close() error {
	if s == nil {
		return nil
	}
	var result error
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result = s.server.Shutdown(ctx)
		// listener 해체와 경로 제거는 형제와 **같은 기계**를 쓴다(a109 §2b.3 G9).
		// 왜 listener를 여기서 우리가 닫는지, 왜 제거가 이 한 곳뿐인지는 그 기계의
		// 자기 문서에 있다(`closePrivateEndpointFiles`).
		result = closePrivateEndpointFiles(result, s.listener, s.descriptor, s.socket, s.controlDir)
	})
	return result
}

func alertControlRoutes(token string, ops *AlertOperations) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(AlertControlListPath, alertControlAuth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "GET required")
			return
		}
		rows, err := ops.Pending(r.Context())
		if err != nil {
			writeRPCError(w, http.StatusInternalServerError, "internal", err.Error())
			return
		}
		if rows == nil {
			// "none" in one shape only: a null and an empty list are the same fact
			// to a decoder and different words on an operator's screen.
			rows = []PendingAlert{}
		}
		writeRPCJSON(w, http.StatusOK, rows)
	}))
	mux.HandleFunc(AlertControlAcknowledgePath, alertControlAuth(token,
		func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "POST required")
				return
			}
			req, ok := decodeAlertControlRequest(w, r)
			if !ok {
				return
			}
			result, err := ops.Acknowledge(r.Context(), req.Operator, req.IDs...)
			if err != nil {
				// A blank operator is the caller's mistake, not the engine's — and
				// it is the one refusal the audit trail depends on, so it must not
				// arrive as a 500 the CLI prints as "engine error".
				if strings.Contains(err.Error(), "operator's identity") {
					writeRPCError(w, http.StatusBadRequest, "invalid", err.Error())
					return
				}
				writeRPCError(w, http.StatusInternalServerError, "internal", err.Error())
				return
			}
			writeRPCJSON(w, http.StatusOK, result)
		}))
	return mux
}

func decodeAlertControlRequest(w http.ResponseWriter, r *http.Request) (AcknowledgeRequest, bool) {
	var req AcknowledgeRequest
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeRPCError(w, http.StatusUnsupportedMediaType, "invalid", "application/json required")
		return req, false
	}
	const maxAlertControlRequestBytes = 8 << 10
	body, err := io.ReadAll(io.LimitReader(r.Body, maxAlertControlRequestBytes+1))
	if err != nil {
		writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
		return req, false
	}
	if len(body) > maxAlertControlRequestBytes {
		writeRPCError(w, http.StatusRequestEntityTooLarge, "invalid", "request body is too large")
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

func alertControlAuth(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(token) ||
			subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeRPCError(w, http.StatusUnauthorized, "unauthorized", "local bearer token rejected")
			return
		}
		next(w, r)
	}
}

// publishPrivateDescriptor stages a descriptor privately and publishes it by
// rename, verifying at every step that the file it is about to publish is still
// the file it made.
//
// ⛔ This is the **third** copy of this ceremony in this package
// (writePositionPolicyDescriptor, writePositionPolicyRuntimeDescriptor). Folding
// all three into one is the right cleanup and it is not a098's: the other two
// belong to the position-policy control plane, and rewriting them would edit
// existing functions on a surface this change has no mandate over. This one is
// written name-independently so the fold has something to fold to.
func publishPrivateDescriptor(path string, body []byte) (result error) {
	dir := filepath.Dir(path)
	if err := positionpolicyrpc.ValidatePrivateControlDirectory(dir); err != nil {
		return fmt.Errorf("engine: validating the control directory before staging: %w", err)
	}
	temporary, err := os.CreateTemp(dir, privateDescriptorStagingPrefix+"*")
	if err != nil {
		return fmt.Errorf("engine: staging a descriptor: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := positionpolicyrpc.ValidatePrivateOpenFile(temporary); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: validating a staged descriptor: %w", err)
	}
	stagedInfo, err := temporary.Stat()
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: inspecting a staged descriptor: %w", err)
	}
	if err := writePositionPolicyDescriptorBody(temporary, body); err != nil {
		return fmt.Errorf("engine: writing a descriptor: %w", err)
	}
	if err := positionpolicyrpc.ValidatePrivateFile(temporaryPath); err != nil {
		return fmt.Errorf("engine: validating a closed descriptor: %w", err)
	}
	closedInfo, err := os.Lstat(temporaryPath)
	if err != nil || !os.SameFile(stagedInfo, closedInfo) {
		return errors.New("engine: a staged descriptor changed before publication")
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("engine: publishing a descriptor: %w", err)
	}
	defer func() {
		if result == nil {
			return
		}
		publishedInfo, err := os.Lstat(path)
		if err == nil && os.SameFile(stagedInfo, publishedInfo) {
			_ = os.Remove(path)
		}
	}()
	publishedInfo, err := os.Lstat(path)
	if err != nil || !os.SameFile(stagedInfo, publishedInfo) {
		return errors.New("engine: a published descriptor does not match the staged file")
	}
	if err := positionpolicyrpc.ValidatePrivateFile(path); err != nil {
		return fmt.Errorf("engine: validating a published descriptor: %w", err)
	}
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("engine: opening the control directory for sync: %w", err)
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return fmt.Errorf("engine: syncing the control directory: %w", err)
	}
	return directory.Close()
}
