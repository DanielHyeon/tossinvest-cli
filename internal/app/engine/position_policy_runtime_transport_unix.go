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
	createdControlDir := false
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("engine: creating position policy runtime directory: %w", err)
		}
	} else {
		createdControlDir = true
	}
	cleanupControlDir := func() {
		if createdControlDir {
			_ = os.Remove(controlDir)
		}
	}
	if err := positionpolicyrpc.ValidateRuntimeControlDirectory(controlDir); err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: validating position policy runtime directory: %w", err)
	}
	socketPath := positionpolicyrpc.RuntimeSocketPath(dir)
	if err := positionpolicyrpc.PrepareRuntimeSocket(socketPath); err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: preparing position policy runtime socket: %w", err)
	}
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: binding position policy runtime socket: %w", err)
	}
	cleanupListener := func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		cleanupControlDir()
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		cleanupListener()
		return nil, fmt.Errorf("engine: securing position policy runtime socket: %w", err)
	}
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
	temporary, err := os.CreateTemp(dir, ".position-policy-runtime-*")
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
