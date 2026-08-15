//go:build unix

package engine

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// positionPolicyRuntimeStagingPrefix는 이 endpoint의 descriptor 발행이 os.CreateTemp에
// 넘기는 임시 이름의 앞머리다.
//
// 리터럴이 아니라 상수인 이유: 회수가 "우리 잔재"로 인정하는 이름 집합에 **같은 값**이
// 들어가야 한다. 두 곳에 따로 적으면 회수가 자기 잔재를 낯선 것으로 거부하고, 그 거부는
// 결정적이라 매 부팅 같은 자리에서 멈춘다.
const positionPolicyRuntimeStagingPrefix = ".position-policy-runtime-"

// positionPolicyRuntimeEndpointNames는 이 endpoint가 자기 control 디렉터리에 만들 수 있는
// 이름의 전부다. 완전성은 테스트가 고정한다 — 각 발행 경로의 **생성기를 실제로 돌려서**
// 그 결과가 이 집합에 속하는지 본다(design D1b).
func positionPolicyRuntimeEndpointNames() positionpolicyrpc.PrivateEndpointNames {
	return positionpolicyrpc.PrivateEndpointNames{
		Descriptor: positionpolicyrpc.RuntimeDescriptorFileName,
		Socket:     positionpolicyrpc.RuntimeSocketFileName,
		StagingPrefixes: []string{
			positionpolicyrpc.StagingPrefix,
			positionPolicyRuntimeStagingPrefix,
		},
	}
}

// PositionPolicyRuntimeServer is separate from the loopback command endpoint.
// Its Unix socket exports one authenticated GET and possession of its client
// cannot express Preview, Apply, or reconciliation mutation.
type PositionPolicyRuntimeServer struct {
	listener   net.Listener
	server     *http.Server
	descriptor string
	socket     string
	controlDir string
	once       sync.Once
}

func StartPositionPolicyRuntimeServer(engineDir string,
	reader positionpolicy.RuntimeReader) (*PositionPolicyRuntimeServer, error) {
	if reader == nil {
		return nil, errors.New("engine: position policy runtime server has no reader")
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine: position policy runtime server has no engine directory")
	}
	if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {
		return nil, fmt.Errorf("engine: validating position policy runtime engine directory: %w", err)
	}
	controlDir := positionpolicyrpc.RuntimeControlDirectory(dir)
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("engine: creating position policy runtime directory: %w", err)
		}
		// 디렉터리가 이미 있다는 것은 지난 기동의 잔재가 있다는 뜻이다. 회수는
		// 전체성이다 — 자기 수명주기가 만들 수 있는 부분 상태를 전부 치우고 디렉터리째
		// 비운 뒤 새로 만든다. **수락 중인 socket은 최종 이름이든 staging 이름이든
		// 지우지 않는다**(design D2 + a109 §1-fix F5). 생사는 권한 비트로 추정하지 않고
		// 연결해서 묻는다.
		if err := positionpolicyrpc.ReclaimStalePrivateEndpoint(controlDir,
			positionPolicyRuntimeEndpointNames()); err != nil {
			return nil, fmt.Errorf("engine: reclaiming position policy runtime directory: %w", err)
		}
		if err := os.Mkdir(controlDir, 0o700); err != nil {
			return nil, fmt.Errorf("engine: recreating position policy runtime directory: %w", err)
		}
	}
	// 회수가 끝났으므로 이 디렉터리는 언제나 **이번 기동이 만든 것**이다. 실패 정리가
	// 조건 없이 지워도 되는 근거가 그것이다.
	cleanupControlDir := func() { _ = os.Remove(controlDir) }
	if err := positionpolicyrpc.ValidateRuntimeControlDirectory(controlDir); err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: validating position policy runtime directory: %w", err)
	}
	socketPath := positionpolicyrpc.RuntimeSocketPath(dir)
	// 임시 이름에 bind → 0600 → rename. 최종 이름은 완성된 socket에만 붙으므로
	// pre-chmod 상태가 최종 이름을 가지는 순간이 사라진다(design D1).
	listener, err := positionpolicyrpc.ListenStagedPrivateSocket(controlDir, socketPath)
	if err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: publishing position policy runtime socket: %w", err)
	}
	cleanupListener := func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		cleanupControlDir()
	}
	// 발행 확인은 **정확-0600**이다. 회수의 좁은 완화가 발행으로 새지 않았다는 증거가
	// 이 호출이다(freeze P1-3).
	if err := positionpolicyrpc.ValidateRuntimeSocket(socketPath); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: validating position policy runtime socket: %w", err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: generating position policy runtime token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	server := &PositionPolicyRuntimeServer{
		listener: listener, descriptor: positionpolicyrpc.RuntimeDescriptorPath(dir),
		socket: socketPath, controlDir: controlDir,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/runtime", server.auth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "GET required")
			return
		}
		runtime, err := reader.Runtime(r.Context())
		if err != nil {
			writePositionPolicyRPCError(w, err)
			return
		}
		writeRPCJSON(w, http.StatusOK, runtime)
	}))
	server.server = &http.Server{
		Handler: mux, ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 8 << 10,
	}
	descriptor := positionpolicyrpc.RuntimeDescriptor{
		Socket: positionpolicyrpc.RuntimeSocketFileName, Token: token, PID: os.Getpid(),
	}
	if err := writePositionPolicyRuntimeDescriptor(server.descriptor, descriptor); err != nil {
		cleanupListener()
		return nil, err
	}
	go func() { _ = server.server.Serve(listener) }()
	return server, nil
}

func (s *PositionPolicyRuntimeServer) auth(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeRPCError(w, http.StatusUnauthorized, "unauthorized", "local bearer token rejected")
			return
		}
		next(w, r)
	}
}

func (s *PositionPolicyRuntimeServer) Close() error {
	if s == nil {
		return nil
	}
	var result error
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result = s.server.Shutdown(ctx)
		// listener는 **여기서 우리가** 닫는다. Shutdown이 Serve의 listener 등록을
		// 앞지르면 Go는 정리를 Serve goroutine의 defer로 미루는데, 그 늦은 정리가
		// 도착할 때 경로에는 이미 후계자의 socket이 앉아 있을 수 있다(a108 A1 F5는
		// 300라운드 중 3회 재현). 닫는 시점을 예정된 goroutine에 맡기지 않는다.
		// 이미 Shutdown이 닫았으면 net.ErrClosed가 오는데, 그것은 성공과 같다.
		if s.listener != nil {
			if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) &&
				result == nil {
				result = err
			}
		}
		// 경로를 지울 권한은 이 루프 하나뿐이다 — listener는 자기 이름을 unlink하지
		// 않고(SetUnlinkOnClose(false)), 기억하는 이름도 이미 사라진 임시 이름이다.
		for _, path := range []string{s.descriptor, s.socket, s.controlDir} {
			if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) && result == nil {
				result = err
			}
		}
	})
	return result
}

func writePositionPolicyRuntimeDescriptor(path string,
	descriptor positionpolicyrpc.RuntimeDescriptor) (result error) {
	body, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := positionpolicyrpc.ValidateRuntimeControlDirectory(dir); err != nil {
		return fmt.Errorf("engine: validating runtime directory before staging: %w", err)
	}
	temporary, err := os.CreateTemp(dir, positionPolicyRuntimeStagingPrefix+"*")
	if err != nil {
		return fmt.Errorf("engine: staging runtime descriptor: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := positionpolicyrpc.ValidatePrivateOpenFile(temporary); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: validating staged runtime descriptor: %w", err)
	}
	stagedInfo, err := temporary.Stat()
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: inspecting staged runtime descriptor: %w", err)
	}
	if err := writePositionPolicyDescriptorBody(temporary, body); err != nil {
		return fmt.Errorf("engine: writing runtime descriptor: %w", err)
	}
	if err := positionpolicyrpc.ValidatePrivateFile(temporaryPath); err != nil {
		return fmt.Errorf("engine: validating closed runtime descriptor: %w", err)
	}
	closedInfo, err := os.Lstat(temporaryPath)
	if err != nil || !os.SameFile(stagedInfo, closedInfo) {
		return errors.New("engine: staged runtime descriptor changed before publication")
	}
	if err := positionpolicyrpc.ValidateRuntimeControlDirectory(dir); err != nil {
		return fmt.Errorf("engine: validating runtime directory before publication: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("engine: publishing runtime descriptor: %w", err)
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
	if err := positionpolicyrpc.ValidateRuntimePublishedDescriptor(path); err != nil {
		return fmt.Errorf("engine: validating published runtime descriptor: %w", err)
	}
	publishedInfo, err := os.Lstat(path)
	if err != nil || !os.SameFile(stagedInfo, publishedInfo) {
		return errors.New("engine: published runtime descriptor does not match staged file")
	}
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("engine: opening runtime directory for sync: %w", err)
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return fmt.Errorf("engine: syncing runtime directory: %w", err)
	}
	if err := directory.Close(); err != nil {
		return fmt.Errorf("engine: closing runtime directory: %w", err)
	}
	return nil
}
