package engine

import (
	"bytes"
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

	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

type positionPolicyCommands interface {
	List(context.Context) ([]positionpolicy.State, error)
	Preview(context.Context, positionpolicy.Request) (positionpolicy.Preview, error)
	Apply(context.Context, positionpolicy.ApplyRequest) (positionpolicy.State, error)
}

// PositionPolicyCommandServer is an authenticated loopback endpoint whose
// lifetime is nested inside `engine run`. Its descriptor is a private
// capability file in the locked engine directory. A crashed engine loses the
// listener; a stale descriptor therefore cannot grant journal access.
type PositionPolicyCommandServer struct {
	listener   net.Listener
	server     *http.Server
	descriptor string
	once       sync.Once
}

func StartPositionPolicyCommandServer(engineDir string,
	commands positionPolicyCommands) (*PositionPolicyCommandServer, error) {
	if commands == nil {
		return nil, errors.New("engine: position policy command server has no service")
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine: position policy command server has no engine directory")
	}
	if err := positionpolicyrpc.ValidateEngineDirectory(dir); err != nil {
		return nil, fmt.Errorf("engine: validating position policy engine directory: %w", err)
	}
	controlDir := positionpolicyrpc.ControlDirectory(dir)
	createdControlDir := false
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("engine: creating position policy control directory: %w", err)
		}
	} else {
		createdControlDir = true
	}
	if err := positionpolicyrpc.ValidateControlDirectory(controlDir); err != nil {
		if createdControlDir {
			_ = os.Remove(controlDir)
		}
		return nil, fmt.Errorf("engine: validating position policy control directory: %w", err)
	}
	cleanupControlDir := func() {
		if createdControlDir {
			_ = os.Remove(controlDir)
		}
	}
	listener, err := net.Listen("tcp4", "127.0.0.1:0")
	if err != nil {
		cleanupControlDir()
		return nil, fmt.Errorf("engine: binding position policy control: %w", err)
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		listener.Close()
		cleanupControlDir()
		return nil, fmt.Errorf("engine: generating position policy control token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	mux := http.NewServeMux()
	server := &PositionPolicyCommandServer{
		listener: listener, descriptor: positionpolicyrpc.DescriptorPath(dir),
	}
	mux.HandleFunc("/v1/health", server.auth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "GET required")
			return
		}
		writeRPCJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}))
	mux.HandleFunc("/v1/positions", server.auth(token, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeRPCError(w, http.StatusMethodNotAllowed, "invalid", "GET required")
			return
		}
		states, err := commands.List(r.Context())
		if err != nil {
			writePositionPolicyRPCError(w, err)
			return
		}
		writeRPCJSON(w, http.StatusOK, states)
	}))
	mux.HandleFunc("/v1/preview", server.auth(token,
		positionPolicyRequestHandler(commands.Preview)))
	mux.HandleFunc("/v1/apply", server.auth(token,
		positionPolicyRequestHandler(commands.Apply)))
	// Change a063's quarantine release, discovered rather than injected: an
	// engine build without the capability serves exactly the route set above,
	// and no wiring point in `engine run` has to learn about a second service.
	if quarantines, ok := commands.(exitQuarantineCommands); ok {
		registerExitQuarantineRoutes(mux, server, token, quarantines)
	}
	server.server = &http.Server{
		Handler: mux, ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 8 << 10,
	}
	descriptor := positionpolicyrpc.Descriptor{
		Address: listener.Addr().String(), Token: token, PID: os.Getpid(),
	}
	if err := writePositionPolicyDescriptor(server.descriptor, descriptor); err != nil {
		listener.Close()
		cleanupControlDir()
		return nil, err
	}
	go func() {
		_ = server.server.Serve(listener)
	}()
	return server, nil
}

func (s *PositionPolicyCommandServer) Close() error {
	if s == nil {
		return nil
	}
	var result error
	s.once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		result = s.server.Shutdown(ctx)
		if err := os.Remove(s.descriptor); err != nil && !errors.Is(err, os.ErrNotExist) && result == nil {
			result = err
		}
		if err := os.Remove(filepath.Dir(s.descriptor)); err != nil && !errors.Is(err, os.ErrNotExist) && result == nil {
			result = err
		}
	})
	return result
}

func (s *PositionPolicyCommandServer) auth(token string,
	next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		provided := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeRPCError(w, http.StatusUnauthorized, "unauthorized", "local bearer token rejected")
			return
		}
		next(w, r)
	}
}

func positionPolicyRequestHandler[Request, Response any](call func(context.Context,
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
		const maxPositionPolicyRequestBytes = 16 << 10
		body, err := io.ReadAll(io.LimitReader(r.Body, maxPositionPolicyRequestBytes+1))
		if err != nil {
			writeRPCError(w, http.StatusBadRequest, "invalid", "request body rejected")
			return
		}
		if len(body) > maxPositionPolicyRequestBytes {
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
			writePositionPolicyRPCError(w, err)
			return
		}
		writeRPCJSON(w, http.StatusOK, result)
	}
}

func writePositionPolicyRPCError(w http.ResponseWriter, err error) {
	status, code := http.StatusInternalServerError, "internal"
	switch {
	case errors.Is(err, positionpolicy.ErrInvalidRequest):
		status, code = http.StatusBadRequest, "invalid"
	case errors.Is(err, positionpolicy.ErrPositionNotFound):
		status, code = http.StatusNotFound, "not_found"
	case errors.Is(err, positionpolicy.ErrVersionMismatch):
		status, code = http.StatusPreconditionFailed, "stale"
	case errors.Is(err, positionpolicy.ErrExitConflict):
		status, code = http.StatusConflict, "exit_conflict"
	case errors.Is(err, positionpolicy.ErrUnknownState):
		status, code = http.StatusConflict, "unknown_state"
	case errors.Is(err, positionpolicy.ErrIneligible):
		status, code = http.StatusConflict, "ineligible"
	case errors.Is(err, positionpolicy.ErrCapabilityTooEarly):
		status, code = http.StatusTooEarly, "capability_too_early"
	case errors.Is(err, positionpolicy.ErrCapabilityExpired):
		status, code = http.StatusGone, "capability_expired"
	case errors.Is(err, positionpolicy.ErrCapabilityStale):
		status, code = http.StatusGone, "capability_stale"
	case errors.Is(err, positionpolicy.ErrCapabilityInvalid):
		status, code = http.StatusForbidden, "capability_invalid"
	case errors.Is(err, positionpolicy.ErrConfirmationRequired):
		status, code = http.StatusBadRequest, "confirmation_required"
	}
	writeRPCError(w, status, code, err.Error())
}

func writeRPCError(w http.ResponseWriter, status int, code, message string) {
	writeRPCJSON(w, status, map[string]string{"code": code, "message": message})
}

func writeRPCJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writePositionPolicyDescriptor(path string, descriptor positionpolicyrpc.Descriptor) (result error) {
	body, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := positionpolicyrpc.ValidateControlDirectory(dir); err != nil {
		return fmt.Errorf("engine: validating control directory before staging: %w", err)
	}
	temporary, err := os.CreateTemp(dir, ".position-policy-control-*")
	if err != nil {
		return fmt.Errorf("engine: staging control descriptor: %w", err)
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := positionpolicyrpc.ValidatePrivateOpenFile(temporary); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: validating staged control descriptor: %w", err)
	}
	stagedInfo, err := temporary.Stat()
	if err != nil {
		_ = temporary.Close()
		return fmt.Errorf("engine: inspecting staged control descriptor: %w", err)
	}
	err = writePositionPolicyDescriptorBody(temporary, body)
	if err != nil {
		return fmt.Errorf("engine: writing control descriptor: %w", err)
	}
	if err := positionpolicyrpc.ValidatePrivateFile(temporaryPath); err != nil {
		return fmt.Errorf("engine: validating closed control descriptor: %w", err)
	}
	closedInfo, err := os.Lstat(temporaryPath)
	if err != nil || !os.SameFile(stagedInfo, closedInfo) {
		return errors.New("engine: staged control descriptor changed before publication")
	}
	if err := positionpolicyrpc.ValidateControlDirectory(dir); err != nil {
		return fmt.Errorf("engine: validating control directory before publication: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("engine: publishing control descriptor: %w", err)
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
	if err := positionpolicyrpc.ValidatePublishedDescriptor(path); err != nil {
		return fmt.Errorf("engine: validating published control descriptor: %w", err)
	}
	publishedInfo, err := os.Lstat(path)
	if err != nil || !os.SameFile(stagedInfo, publishedInfo) {
		return errors.New("engine: published control descriptor does not match staged file")
	}
	directory, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("engine: opening control directory for sync: %w", err)
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return fmt.Errorf("engine: syncing control directory: %w", err)
	}
	if err := directory.Close(); err != nil {
		return fmt.Errorf("engine: closing control directory: %w", err)
	}
	return nil
}

type positionPolicyDescriptorTemp interface {
	Chmod(os.FileMode) error
	Write([]byte) (int, error)
	Sync() error
	Close() error
}

func writePositionPolicyDescriptorBody(temporary positionPolicyDescriptorTemp, body []byte) error {
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	written, err := temporary.Write(body)
	if err != nil {
		_ = temporary.Close()
		return err
	}
	if written != len(body) {
		_ = temporary.Close()
		return io.ErrShortWrite
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	return temporary.Close()
}
