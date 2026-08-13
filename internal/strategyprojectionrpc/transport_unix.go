//go:build unix

package strategyprojectionrpc

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
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

// stagingPrefix는 **아직 발행되지 않은** 산출물의 이름 앞머리다.
//
// descriptor도 socket도 임시 이름에 먼저 만들고 rename으로 최종 이름을 준다. 그래서
// 최종 이름이 보이면 그것은 항상 완성된 파일이다 — 반쯤 쓰인 상태가 최종 이름을
// 가지는 순간이 없다(design D1-2). 대신 임시 이름으로 죽는 새 잔재가 생기는데,
// 그것은 낯선 엔트리가 아니라 우리 잔재이므로 회수 목록에 들어간다.
//
// # 왜 이렇게 짧은가
//
// bind 하는 이름은 최종 이름이 아니라 **이 임시 이름**이고, unix socket 경로에는
// sun_path 상한(108바이트 — 실측상 107자까지 bind 된다)이 있다. 임시 이름이 최종
// 이름보다 길면, 최종 경로는 상한 안인데 발행만 `invalid argument` 로 죽는 배포가
// 생긴다. 운영자에게 보이는 것은 「매 부팅 projection 이 없다」뿐이고 최종 경로는
// 멀쩡하므로 원인이 보이지 않는다. 그래서 임시 basename 을 최종 basename
// (runtime.sock 12자 · endpoint.json 13자) 이하로 유지한다
// (`TestStagingNamesAreNeverLongerThanTheNamesTheyBecome` 가 고정한다).
const stagingPrefix = ".s-"

// 임시 이름의 종류 한 글자. 사람이 잔재를 볼 때 무엇이 되려다 만 것인지 알려 준다.
const (
	stagingSocketKind     = "s"
	stagingDescriptorKind = "e"
)

// discardPublication은 실패한 기동이 남길 수 있는 산출물을 **전부** 치운다.
//
// descriptor 가 이 목록에 있는 이유: `writeDescriptor` 의 rename 이 성공한 **뒤에도**
// 실패하는 줄이 있다(발행 확인의 SameFile, 디렉터리 sync). 그 경로에서 descriptor 는
// 이미 최종 이름을 갖고 있으므로, 치우지 않으면 당회 부팅 내내 이번 사고와 같은 S1
// 잔재가 남고 디렉터리 제거도 ENOTEMPTY 로 실패한다.
//
// 없는 파일을 지우라는 요청은 성공과 같으므로 오류를 전부 버린다 — 이 함수가 불리는
// 시점은 이미 기동이 실패한 뒤이고, 정리의 실패로 그 사유를 덮지 않는다.
func discardPublication(controlDir, descriptorPath, socketPath string) {
	_ = os.Remove(descriptorPath)
	_ = os.Remove(socketPath)
	_ = os.Remove(controlDir)
}

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
	descriptorPath := DescriptorPath(root)
	listener, err := listenPrivateSocket(controlDir, socketPath)
	if err != nil {
		cleanupDir()
		return nil, err
	}
	// 경로를 지우는 것은 언제나 우리 손이다 — listener는 자기 이름을 unlink하지
	// 않는다(listenPrivateSocket의 SetUnlinkOnClose(false), design D2-2).
	//
	// descriptor 까지 치우는 이유는 discardPublication 의 주석에 있다: rename 이
	// 성공한 뒤에 실패하는 줄이 있고, 그 경로의 descriptor 는 이미 최종 이름이다.
	cleanupListener := func() {
		_ = listener.Close()
		discardPublication(controlDir, descriptorPath, socketPath)
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

// stagingPath는 아직 발행되지 않은 산출물의 임시 경로를 만든다. 이름이 겹치면
// bind가 EADDRINUSE로 실패하므로 난수를 붙인다 — 회수가 방금 비운 디렉터리라
// 겹칠 일은 없지만, 겹치지 않는 것에 기대지 않는다.
//
// basename 은 `.s-` + 종류 한 글자 + 8자 hex = 12자로 고정이다. 그 길이가 왜
// 계약인지는 stagingPrefix 의 주석에 있다.
func stagingPath(controlDir, kind string) (string, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", err
	}
	return filepath.Join(controlDir, stagingPrefix+kind+hex.EncodeToString(suffix)), nil
}

// listenPrivateSocket은 socket을 **임시 이름에 bind → 0600 → rename** 순으로 발행한다.
//
// 왜 이 의례인가(design D1-2). `net.Listen`이 만드는 socket 파일의 권한은 umask가
// 정한다 — 컨테이너 실측 umask 077이면 0700이다. 예전 코드는 최종 이름에 바로 bind한
// 뒤 chmod했으므로, 그 두 줄 사이에서 죽으면 **최종 이름 자리에 0600이 아닌 socket**이
// 남았다. 그리고 회수의 정확-0600 검사가 그것을 "unsafe"로 영구 거부했다. 버그가 보안
// 핀으로 봉인돼 있었던 것이다(A1 F1 실측 3/3).
//
// rename은 같은 디렉터리 안이라 원자적이고, Linux에서 unix socket의 rename은 유효하다:
// bind는 경로가 아니라 inode에 붙으므로 이름이 바뀌어도 그 listener는 계속 수락한다.
// 그래서 최종 이름은 "완성된 socket이 나타나는 순간"에만 존재한다.
//
// `SetUnlinkOnClose(false)`가 design D2-2다. Go의 unix listener는 닫힐 때 **자기가
// 기억하는 경로**를 unlink한다. `Shutdown`이 `Serve`의 listener 등록을 앞지르면 그
// 정리가 Serve goroutine의 defer로 밀리고, 늦게 도착한 unlink가 그 경로에 이미 앉아
// 있는 **후계자의 socket**을 지운다(A1 F5는 300라운드 중 3회 재현). 여기서 두 겹으로
// 막는다 — listener가 기억하는 이름은 임시 이름이고, unlink 자체를 끈다.
func listenPrivateSocket(controlDir, socketPath string) (net.Listener, error) {
	staged, err := stagingPath(controlDir, stagingSocketKind)
	if err != nil {
		return nil, err
	}
	listener, err := net.Listen("unix", staged)
	if err != nil {
		return nil, err
	}
	if unixListener, ok := listener.(*net.UnixListener); ok {
		unixListener.SetUnlinkOnClose(false)
	}
	discard := func() {
		_ = listener.Close()
		_ = os.Remove(staged)
	}
	if err := os.Chmod(staged, 0o600); err != nil {
		discard()
		return nil, err
	}
	if err := os.Rename(staged, socketPath); err != nil {
		discard()
		return nil, err
	}
	return listener, nil
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
	if err != nil || !controlDirectoryModeIsSafe(info.Mode()) {
		return errors.New("strategy projection runtime: existing control directory is unsafe")
	}
	var directoryStat unix.Stat_t
	if err := unix.Lstat(controlDir, &directoryStat); err != nil || !ownedByEffectiveUser(directoryStat.Uid) {
		return errors.New("strategy projection runtime: existing control directory ownership is unsafe")
	}
	entries, err := os.ReadDir(controlDir)
	if err != nil {
		return err
	}
	// 우리가 만들 수 있는 이름은 최종 두 개와 **발행 전 임시 이름**뿐이다. 하나라도
	// 낯선 것이 있으면 이 디렉터리는 우리 잔재가 아니므로 아무것도 건드리지 않는다.
	seen := make(map[string]bool, len(entries))
	staged := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		switch {
		case name == DescriptorFileName || name == SocketFileName:
			seen[name] = true
		case strings.HasPrefix(name, stagingPrefix):
			// 최종 이름을 가진 적이 없는 산출물이다. 아무도 이것을 읽지도 연결하지도
			// 않았으므로 **내용**은 검증할 것이 없다 — 우리 잔재로 치운다(design D1-2).
			//
			// 다만 **모양**은 본다. 이름만 보고 지우면, 우리가 임시 이름으로 만들 수
			// 없는 모양(디렉터리 등)까지 지우려 든다: 비어 있으면 남의 디렉터리를
			// 지우고, 비어 있지 않으면 ENOTEMPTY 로 회수 전체가 매 부팅 멈춘다.
			// 우리가 만드는 것은 정규 파일(descriptor)과 socket 둘뿐이므로, 그 밖의
			// 모양이면 아무것도 건드리지 않고 거부한다 — 낯선 엔트리와 같은 규칙이다.
			path := filepath.Join(controlDir, name)
			info, statErr := os.Lstat(path)
			if statErr != nil || !(info.Mode().IsRegular() || info.Mode()&os.ModeSocket != 0) {
				return errors.New("strategy projection runtime: stale staging entry has an unexpected shape")
			}
			staged = append(staged, path)
		default:
			return errors.New("strategy projection runtime: stale control directory has unexpected entries")
		}
	}
	// 주인의 생사를 먼저 판정한다. 그 답이 descriptor를 얼마나 엄격하게 볼지 정하기
	// 때문이다 — 사망이 증명되지 않은 채로 내용 실패를 회수로 읽으면 살아 있는 주인의
	// 파일을 지우게 된다.
	socketPath := SocketPath(engineDir)
	if seen[SocketFileName] {
		if err := verifyStaleSocketShape(socketPath); err != nil {
			return err
		}
		if projectionSocketAccepts(socketPath) {
			return errors.New("strategy projection runtime: projection owner is still alive")
		}
	}
	// 여기 왔다면 주인은 죽었다: socket이 없으면 listener도 없고(Start는 listen이
	// 성공한 뒤에만 descriptor를 쓴다), 있어도 아무도 수락하지 않았다.
	//
	// 그래서 descriptor는 **형식**만 본다. 0바이트·잘린 JSON은 O_EXCL→write→sync가
	// 원자적이지 않아서 생기는 우리 자신의 반쯤 쓴 파일이고, D2 이후 그 내용은 어떤
	// 판정에도 쓰이지 않는다 — 내용 파싱 실패를 거부로 읽으면 그 거부는 순수한
	// 손실이다(A1 F2 실측 3/3, design D1-2).
	if seen[DescriptorFileName] {
		file, err := openVerifiedDescriptor(DescriptorPath(engineDir))
		if err != nil {
			return fmt.Errorf("strategy projection runtime: stale descriptor is unsafe: %w", err)
		}
		_ = file.Close()
	}
	// 없는 파일을 지우라는 요청은 성공과 같다. 주인이 아직 자기 Close 도중이면 양쪽이
	// 같은 경로를 지우는데, 그 경합은 결과가 같으므로 양성이다(design D2). 디렉터리도
	// 같은 규칙을 쓴다 — 파일에만 ErrNotExist를 용인하고 rmdir에는 용인하지 않으면
	// 같은 경합이 파일에서는 양성이고 디렉터리에서는 기동 실패가 되는데, 그 비대칭에
	// 근거가 없다(A1 F6).
	//
	// probe와 제거 **사이**에 새 주인이 들어오는 경합의 방어는 이 함수가 아니다.
	// 부팅 1단계의 journal flock이 엔진을 하나로 강제하므로(cmd/tossctl 의 `runEngineRun`,
	// "flock on the journal directory FIRST"), 같은 디렉터리에 회수와 발행을 동시에
	// 하는 두 번째 엔진은 존재할 수 없다. 이 함수의 원자성이 아니라 그 잠금이 방어다.
	for _, path := range append(staged, DescriptorPath(engineDir), socketPath) {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("strategy projection runtime: remove stale endpoint: %w", err)
		}
	}
	if err := os.Remove(controlDir); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("strategy projection runtime: remove stale control directory: %w", err)
	}
	return nil
}

// ownedByEffectiveUser는 이 엔트리를 만든 것이 우리인지 본다. 남이 만든 socket·
// 디렉터리는 회수 대상이 아니다 — 치우면 남의 것을 치우는 것이다.
//
// 한 줄짜리 비교를 이름 있는 함수로 뽑은 이유는 절 하나를 따로 죽여 볼 수 있게
// 하기 위해서다(A1 F6). 비root로 도는 테스트는 파일 소유자를 바꿀 수 없어서 이
// 절만 실패시키는 디스크 상태를 만들 수 없다.
func ownedByEffectiveUser(uid uint32) bool {
	return uid == uint32(os.Geteuid())
}

// verifyStaleSocketShape는 잔재 socket이 **우리가 만든 모양**인지 본다.
//
// 권한 검사가 정확-0600이 아니라 "group/other 비트 없음"인 것이 design D1-2의 좁은
// 완화다. 구버전 바이너리가 listen과 chmod 사이에서 죽으면 umask가 정한 권한(실측
// 077 → 0700)의 socket이 남고, 정확-0600 검사는 **자기 자신이 만든 그 잔재**를 영구
// 거부했다. 완화의 폭은 딱 거기까지다 — group이나 other 비트가 하나라도 있으면
// 여전히 거부다(0644 socket은 지금도 unsafe).
//
// 이 완화가 안전한 것은 나머지 조건이 전부 성립할 때뿐이고, 그 전부를 여기와
// 호출부에서 확인한다: 우리 uid 소유 · 0700 디렉터리 안 · symlink 아님 ·
// hard link 없음 · 그리고 아무도 수락하지 않음(주인의 사망 입증).
func verifyStaleSocketShape(socketPath string) error {
	info, err := os.Lstat(socketPath)
	if err != nil || info.Mode()&os.ModeSocket == 0 || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm()&0o077 != 0 {
		return errors.New("strategy projection runtime: stale socket is unsafe")
	}
	var socketStat unix.Stat_t
	if err := unix.Lstat(socketPath, &socketStat); err != nil || !ownedByEffectiveUser(socketStat.Uid) || socketStat.Nlink != 1 {
		return errors.New("strategy projection runtime: stale socket ownership is unsafe")
	}
	return nil
}

// projectionSocketAccepts는 이 socket 경로에서 연결을 받아 주는 자가 있는지 본다.
//
// Start는 listen이 성공한 뒤에야 descriptor를 쓰므로, 수락하지 않는 socket 파일은
// 죽은 주인의 것이다. 반대로 수락한다면 그게 누구든 이 디렉터리는 남의 것이다.
//
// 죽었다고 읽는 것은 세 가지뿐이다: 연결 거부(listener 없음), 파일 부재, 그리고
// **owner 쓰기 비트가 없는 socket**. 그 밖의 오류(타임아웃 등)는 죽었다는 증거가
// 아니므로 살아 있다고 본다 — 잘못 살아 있다고 보면 이번 기동만 실패하지만, 잘못
// 죽었다고 보면 남의 socket을 지운다.
//
// # 왜 owner 쓰기 비트가 판정인가
//
// unix socket 에 connect 하려면 그 파일에 쓰기 권한이 있어야 하고, 없으면 커널은
// EACCES 를 준다(실측). 그 오류를 "답이 안 왔다 = 살아 있다"로 읽으면 아무도 없는
// socket 하나가 **영구 거부**를 만든다 — 이 change 가 지우려는 모양 그대로다.
//
// 그렇게 읽어도 되는 근거는 우리 발행 수명주기다. 최종 이름을 가진 socket 은 반드시
// chmod 0600 을 지난 뒤 rename 된 것이고(listenPrivateSocket), 구버전이 최종 이름에
// 바로 bind 하고 남긴 pre-chmod 잔재도 `net.Listen` 이 umask 로 만든 권한이라
// `UMask=0277` 같은 배포에서만 owner 쓰기가 깎인다. 어느 쪽이든 **owner 가 쓸 수 없는
// socket 은 우리가 발행한 산 endpoint 일 수 없다.** 그리고 회수 경로에서 이 함수에
// 닿으려면 이미 우리 uid 소유 · 0700 디렉터리 안 · 비symlink · nlink 1 을 통과했다.
func projectionSocketAccepts(socketPath string) bool {
	conn, err := net.DialTimeout("unix", socketPath, projectionProbeTimeout)
	if err == nil {
		_ = conn.Close()
		return true
	}
	if errors.Is(err, unix.ECONNREFUSED) || errors.Is(err, os.ErrNotExist) {
		return false
	}
	if info, statErr := os.Lstat(socketPath); statErr == nil && info.Mode().Perm()&0o200 == 0 {
		return false
	}
	return true
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
	// 파일이 있다고 주인이 사는 것은 아니다. 전원이 끊기면 unlink할 기회가 없으므로
	// socket 파일이 그대로 남는데(design D1의 S3 — 전원 단절의 기본 모양), 그 위에서
	// lazy client를 돌려주면 실패는 **첫 Read**에서야 나타난다. 그때는 소비자가 이미
	// "붙었다"고 판단한 뒤라 강등할 기회를 놓치고, httpapi에서는 집계 스냅샷 전체가
	// 함께 죽었다(A2 F3 실측). 회수와 같은 원시로 지금 묻는다(design D4-2).
	if !projectionSocketAccepts(socketPath) {
		return nil, errors.New("strategy projection runtime: socket has no listener")
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
		// listener는 **여기서 우리가** 닫는다. `Shutdown`이 `Serve`의 listener 등록을
		// 앞지르면 Go는 정리를 Serve goroutine의 defer로 미루는데, 그 늦은 정리가
		// 도착할 때 경로에는 이미 후계자의 socket이 앉아 있을 수 있다(A1 F5는 300
		// 라운드 중 3회 재현). 닫는 시점을 예정된 goroutine에 맡기지 않는다.
		// 이미 Shutdown이 닫았으면 net.ErrClosed가 오는데, 그것은 성공과 같다.
		if s.listener != nil {
			if err := s.listener.Close(); err != nil && !errors.Is(err, net.ErrClosed) && result == nil {
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

// writeDescriptor는 descriptor를 임시 이름에 다 쓴 뒤 rename으로 발행한다.
//
// 예전 코드는 최종 이름에 O_EXCL로 만들고 거기에 write·sync했다. 그 셋은 원자적이지
// 않으므로 사이에서 죽으면 **최종 이름을 가진 0바이트·잘린 JSON**이 남았고, 회수의
// 내용 검증이 그것을 영구 거부했다(A1 F2). rename은 원자적이라 최종 이름은 완성된
// 파일에만 붙는다 — 같은 저장소의 다른 세 endpoint가 이미 쓰는 의례다
// (internal/app/engine의 publishPrivateDescriptor).
//
// O_EXCL이 사라진 자리는 회수가 채운다: 이 함수가 불리는 시점의 control 디렉터리는
// 방금 Mkdir한 빈 디렉터리여서 덮어쓸 최종 이름이 애초에 없다.
func writeDescriptor(path string, descriptor Descriptor) error {
	body, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}
	// 이름을 `os.CreateTemp` 에 맡기지 않는다. 그쪽은 접두에 임의 길이의 숫자를
	// 붙이므로 임시 이름의 길이를 계약으로 잡을 수 없는데, 그 길이가 socket 쪽에서는
	// bind 가능성 자체를 정한다(stagingPrefix 주석). 두 산출물이 같은 규칙을 쓰는 것이
	// 회수 쪽에서도 읽기 쉽다 — 접두 하나로 둘 다 걸린다.
	stagedPath, err := stagingPath(filepath.Dir(path), stagingDescriptorKind)
	if err != nil {
		return err
	}
	staged, err := os.OpenFile(stagedPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	// rename이 성공하면 이 이름은 이미 비어 있어 ErrNotExist로 끝난다. 실패한 모든
	// 경로에서는 반쯤 쓴 파일이 여기서 사라진다 — 임시 잔재조차 남기지 않는다.
	defer func() { _ = os.Remove(stagedPath) }()
	if err := staged.Chmod(0o600); err != nil {
		_ = staged.Close()
		return err
	}
	stagedInfo, err := staged.Stat()
	if err != nil {
		_ = staged.Close()
		return err
	}
	if _, err := staged.Write(body); err != nil {
		_ = staged.Close()
		return err
	}
	if err := staged.Sync(); err != nil {
		_ = staged.Close()
		return err
	}
	if err := staged.Close(); err != nil {
		return err
	}
	// 발행하기 직전에 "지금 발행하려는 것이 내가 만든 그 파일인가"를 다시 확인한다.
	closed, err := os.Lstat(stagedPath)
	if err != nil || !os.SameFile(stagedInfo, closed) {
		return errors.New("strategy projection runtime: staged descriptor changed before publication")
	}
	if err := os.Rename(stagedPath, path); err != nil {
		return err
	}
	published, err := os.Lstat(path)
	if err != nil || !os.SameFile(stagedInfo, published) {
		return errors.New("strategy projection runtime: published descriptor is not the staged file")
	}
	// 디렉터리 엔트리까지 내려가야 전원이 끊겨도 rename이 살아남는다.
	directory, err := os.Open(filepath.Dir(path))
	if err != nil {
		return err
	}
	if err := directory.Sync(); err != nil {
		_ = directory.Close()
		return err
	}
	return directory.Close()
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
