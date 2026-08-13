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

// projectionProbeTimeout은 잔재 socket에 "누가 받아 주나"를 묻는 시간이다. 같은
// 호스트의 Unix socket 연결은 즉시 끝나므로, 이 시간을 다 쓰면 답이 안 온 것이고
// 그것은 죽었다는 증거가 아니다 — 그 경우는 살아 있다고 보고 기동을 거부한다.
const projectionProbeTimeout = 200 * time.Millisecond

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
		if !errors.Is(err, os.ErrExist) {
			return nil, fmt.Errorf("strategy projection runtime: create control directory: %w", err)
		}
		if err := reclaimStaleControlDirectory(root); err != nil {
			return nil, err
		}
		if err := os.Mkdir(controlDir, 0o700); err != nil {
			return nil, fmt.Errorf("strategy projection runtime: recreate control directory: %w", err)
		}
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

// reclaimStaleControlDirectory는 지난 기동이 남긴 잔재를 치운다.
//
// 이 함수가 다루는 상태는 Start·Close·이 함수 자신이 만들 수 있는 **전부**다.
// 어느 지점에서 프로세스가 죽어도 디스크에는 아래 넷 중 하나만 남는다.
//
//	빈 디렉터리    Mkdir 뒤 listen 전에 죽었다 — 검증할 것이 없다
//	descriptor만   종료가 socket을 unlink한 뒤 죽었다(2026-08-13 사고)
//	socket만       제거 순서(descriptor→socket)의 중간에서 죽었다
//	둘 다          SIGKILL·전원 단절
//
// 빠진 파일은 "덜 만들어진 것"이지 "낯선 것"이 아니다. 예전 코드는 둘 다 있어야만
// 회수해서, 넷 중 셋을 **어떤 재시도로도 벗어날 수 없는 영구 거부**로 만들었다.
//
// 주인이 살아 있는지는 PID가 아니라 socket에 실제로 연결해서 묻는다(design D2).
// 컨테이너를 다시 만들면 PID가 거의 같은 자리에 재배정되므로, descriptor에 적힌
// PID는 남의 프로세스를 가리킬 수 있다 — 그 오판은 살아 있다고 잘못 읽으면 영구
// 거부를, 죽었다고 잘못 읽으면 남의 socket 삭제를 만든다.
//
// 보안 검사(0700 디렉터리·소유 uid·symlink 금지·socket 0600·hard link 없음·
// descriptor 형식)는 회수 가능한 상태에서도 전부 한다. 넓히는 것은 회수의 범위이지
// 검증이 아니다.
func reclaimStaleControlDirectory(engineDir string) error {
	controlDir := ControlDirectory(engineDir)
	info, err := os.Lstat(controlDir)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o700 {
		return errors.New("strategy projection runtime: existing control directory is unsafe")
	}
	var directoryStat unix.Stat_t
	if err := unix.Lstat(controlDir, &directoryStat); err != nil || directoryStat.Uid != uint32(os.Geteuid()) {
		return errors.New("strategy projection runtime: existing control directory ownership is unsafe")
	}
	entries, err := os.ReadDir(controlDir)
	if err != nil {
		return err
	}
	// 우리가 만들 수 있는 이름은 둘뿐이다. 하나라도 낯선 것이 있으면 이 디렉터리는
	// 우리 잔재가 아니므로 아무것도 건드리지 않는다.
	seen := make(map[string]bool, len(entries))
	for _, entry := range entries {
		if entry.Name() != DescriptorFileName && entry.Name() != SocketFileName {
			return errors.New("strategy projection runtime: stale control directory has unexpected entries")
		}
		seen[entry.Name()] = true
	}
	if seen[DescriptorFileName] {
		if _, err := readDescriptor(DescriptorPath(engineDir)); err != nil {
			return fmt.Errorf("strategy projection runtime: stale descriptor is unsafe: %w", err)
		}
	}
	socketPath := SocketPath(engineDir)
	if seen[SocketFileName] {
		socketInfo, err := os.Lstat(socketPath)
		if err != nil || socketInfo.Mode()&os.ModeSocket == 0 || socketInfo.Mode()&os.ModeSymlink != 0 || socketInfo.Mode().Perm() != 0o600 {
			return errors.New("strategy projection runtime: stale socket is unsafe")
		}
		var socketStat unix.Stat_t
		if err := unix.Lstat(socketPath, &socketStat); err != nil || socketStat.Uid != uint32(os.Geteuid()) || socketStat.Nlink != 1 {
			return errors.New("strategy projection runtime: stale socket ownership is unsafe")
		}
		if projectionSocketAccepts(socketPath) {
			return errors.New("strategy projection runtime: projection owner is still alive")
		}
	}
	// 없는 파일을 지우라는 요청은 성공과 같다. 주인이 아직 자기 Close 도중이면 양쪽이
	// 같은 경로를 지우는데, 그 경합은 결과가 같으므로 양성이다(design D2).
	for _, path := range []string{DescriptorPath(engineDir), socketPath} {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("strategy projection runtime: remove stale endpoint: %w", err)
		}
	}
	if err := os.Remove(controlDir); err != nil {
		return fmt.Errorf("strategy projection runtime: remove stale control directory: %w", err)
	}
	return nil
}

// projectionSocketAccepts는 이 socket 경로에서 연결을 받아 주는 자가 있는지 본다.
//
// Start는 listen이 성공한 뒤에야 descriptor를 쓰므로, 수락하지 않는 socket 파일은
// 죽은 주인의 것이다. 반대로 수락한다면 그게 누구든 이 디렉터리는 남의 것이다.
//
// 죽었다고 읽는 것은 두 가지뿐이다: 연결 거부(listener 없음)와 파일 부재. 그 밖의
// 오류(권한·타임아웃 등)는 죽었다는 증거가 아니므로 살아 있다고 본다 — 잘못 살아
// 있다고 보면 이번 기동만 실패하지만, 잘못 죽었다고 보면 남의 socket을 지운다.
func projectionSocketAccepts(socketPath string) bool {
	conn, err := net.DialTimeout("unix", socketPath, projectionProbeTimeout)
	if err == nil {
		_ = conn.Close()
		return true
	}
	return !errors.Is(err, unix.ECONNREFUSED) && !errors.Is(err, os.ErrNotExist)
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
