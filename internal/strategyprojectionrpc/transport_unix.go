//go:build unix

package strategyprojectionrpc

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

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
	"golang.org/x/sys/unix"
)

type Server struct {
	server     *http.Server
	listener   net.Listener
	descriptor string
	socket     string
	controlDir string
	once       sync.Once
}

func Start(engineDir string, reader strategyprojection.Reader) (*Server, error) {
	if reader == nil {
		return nil, errors.New("strategy projection runtime: reader is required")
	}
	root := strings.TrimSpace(engineDir)
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o022 != 0 {
		return nil, errors.New("strategy projection runtime: engine directory is unsafe")
	}
	controlDir := ControlDirectory(root)
	if err := os.Mkdir(controlDir, 0o700); err != nil {
		return nil, fmt.Errorf("strategy projection runtime: create control directory: %w", err)
	}
	cleanupDir := func() { _ = os.Remove(controlDir) }
	socketPath := SocketPath(root)
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		cleanupDir()
		return nil, err
	}
	cleanupListener := func() {
		_ = listener.Close()
		_ = os.Remove(socketPath)
		cleanupDir()
	}
	if err := os.Chmod(socketPath, 0o600); err != nil {
		cleanupListener()
		return nil, err
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		cleanupListener()
		return nil, err
	}
	token := base64.RawURLEncoding.EncodeToString(tokenBytes)
	server := &Server{listener: listener, descriptor: DescriptorPath(root), socket: socketPath, controlDir: controlDir}
	mux := http.NewServeMux()
	mux.HandleFunc(ProjectionPath, func(w http.ResponseWriter, request *http.Request) {
		provided := strings.TrimPrefix(request.Header.Get("Authorization"), "Bearer ")
		if len(provided) != len(token) || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"code": "unauthorized"})
			return
		}
		if request.Method != http.MethodGet && request.Method != http.MethodHead {
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"code": "method_not_allowed"})
			return
		}
		if hasRequestBody(request) || request.URL.RawQuery != "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"code": "input_not_supported"})
			return
		}
		snapshot, err := reader.Read(request.Context())
		if err != nil || strategyprojection.Validate(snapshot) != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"code": "runtime_unavailable"})
			return
		}
		if request.Method == http.MethodHead {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.Header().Set("Cache-Control", "no-store")
			w.WriteHeader(http.StatusOK)
			return
		}
		writeJSON(w, http.StatusOK, strategyprojection.Clone(snapshot))
	})
	server.server = &http.Server{Handler: mux, ReadHeaderTimeout: 2 * time.Second, ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second, IdleTimeout: 15 * time.Second, MaxHeaderBytes: 8 << 10}
	descriptor := Descriptor{SchemaVersion: descriptorSchema, Socket: SocketFileName, Token: token, PID: os.Getpid()}
	if err := writeDescriptor(server.descriptor, descriptor); err != nil {
		cleanupListener()
		return nil, err
	}
	go func() { _ = server.server.Serve(listener) }()
	return server, nil
}

func Dial(_ context.Context, descriptorPath string) (*Client, error) {
	descriptor, err := readDescriptor(strings.TrimSpace(descriptorPath))
	if err != nil {
		return nil, err
	}
	socketPath := filepath.Join(filepath.Dir(descriptorPath), descriptor.Socket)
	info, err := os.Lstat(socketPath)
	if err != nil || info.Mode()&os.ModeSocket == 0 || info.Mode().Perm() != 0o600 {
		return nil, errors.New("strategy projection runtime: socket is invalid")
	}
	transport := &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
	}}
	return &Client{baseURL: "http://unix", token: descriptor.Token, http: &http.Client{Transport: transport, Timeout: 5 * time.Second}}, nil
}

func (s *Server) Close() error {
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

func writeDescriptor(path string, descriptor Descriptor) error {
	body, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	if _, err := file.Write(body); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return err
	}
	return file.Close()
}

func openDescriptorNoFollow(path string) (*os.File, error) {
	fd, err := unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, err
	}
	return os.NewFile(uintptr(fd), path), nil
}

func unixHTTPClient(socket string) *http.Client {
	return &http.Client{Transport: &http.Transport{Proxy: nil, DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, "unix", socket)
	}}, Timeout: 5 * time.Second}
}
