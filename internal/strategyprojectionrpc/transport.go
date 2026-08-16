// Package strategyprojectionrpc exposes only the bounded authenticated read
// projection over a private runtime Unix socket.
package strategyprojectionrpc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyprojection"
)

const (
	ControlDirectoryName = ".strategy-runtime-read"
	DescriptorFileName   = "endpoint.json"
	SocketFileName       = "runtime.sock"
	ProjectionPath       = "/v1/strategy-runtime"
	MaxProjectionBytes   = 1 << 20
	maxDescriptorBytes   = 4096
	descriptorSchema     = "tossos.strategy-runtime-unix/v1"
)

type Descriptor struct {
	SchemaVersion string `json:"schemaVersion"`
	Socket        string `json:"socket"`
	Token         string `json:"token"`
	PID           int    `json:"pid"`
}

func ControlDirectory(engineDir string) string {
	return filepath.Join(strings.TrimSpace(engineDir), ControlDirectoryName)
}

func DescriptorPath(engineDir string) string {
	return filepath.Join(ControlDirectory(engineDir), DescriptorFileName)
}

func SocketPath(engineDir string) string {
	return filepath.Join(ControlDirectory(engineDir), SocketFileName)
}

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

func (c *Client) Read(ctx context.Context) (strategyprojection.Snapshot, error) {
	var result strategyprojection.Snapshot
	if c == nil || c.http == nil || len(c.token) < 32 {
		return result, errors.New("strategy projection runtime: invalid client")
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+ProjectionPath, nil)
	if err != nil {
		return result, err
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	response, err := c.http.Do(request)
	if err != nil {
		return result, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, MaxProjectionBytes+1))
	if err != nil || len(body) > MaxProjectionBytes {
		return result, errors.New("strategy projection runtime: response unreadable or oversized")
	}
	if response.StatusCode != http.StatusOK {
		return result, fmt.Errorf("strategy projection runtime: HTTP %d", response.StatusCode)
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return result, fmt.Errorf("strategy projection runtime: invalid response: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return result, errors.New("strategy projection runtime: response must contain one value")
	}
	if err := strategyprojection.Validate(result); err != nil {
		return strategyprojection.Snapshot{}, err
	}
	return strategyprojection.Clone(result), nil
}

// Close는 이 client가 쥔 **유휴 연결**을 놓아 준다 — a109 §2b.3 G5.
//
// # 왜 필요해졌는가
//
// a109의 조회 데몬은 엔진이 돌아오면 전략 reader 자리를 갈아끼운다(재부착). 그때
// 밀려나는 값이 이 client인데, 놓아 줄 방법이 없으면 그것이 쥔 연결과 그 연결을 지키는
// goroutine이 데몬 수명 내내 남는다 — 엔진이 오르내릴 때마다 한 벌씩.
//
// # 무엇을 닫고 무엇을 닫지 않는가
//
// `http.Client.CloseIdleConnections`는 이름 그대로 **유휴** 연결만 닫는다. 진행 중인
// 요청은 끊지 않고, 닫은 뒤에도 이 client로 다시 읽을 수 있다. 그래서 재부착 쪽이
// 잠금 밖에서, 옛 자리를 아직 읽고 있는 요청과 나란히 불러도 안전하다
// (`cmd/tossctl`의 `closeEvictedStrategyReader`).
//
// nil 수신자와 nil transport에 안전한 이유는 자리에 **부재**도 앉기 때문이다.
// 오류를 돌려주는 시그니처인 것은 io.Closer 계약을 만족시키기 위해서다 — 자리는
// 값의 구체 타입을 모른 채 io.Closer로만 본다.
func (c *Client) Close() error {
	if c == nil || c.http == nil {
		return nil
	}
	c.http.CloseIdleConnections()
	return nil
}

// controlDirectoryModeIsSafe는 control 디렉터리의 **모양** 세 절을 한자리에 모은다.
// 세 절은 각각 다른 것을 막는다.
//
//	디렉터리가 아니다   우리가 만든 것이 아니다
//	symlink다           검사한 곳과 지우는 곳이 달라진다
//	0700이 아니다       남이 토큰을 들여다볼 수 있다
//
// 한 줄짜리 조건을 이름 있는 함수로 뽑은 이유는 절 하나하나를 **따로 죽여 볼 수**
// 있게 하기 위해서다(A1 F6 — 절 단위 미핀). 조건이 호출부에 붙어 있으면 그 절만
// 지운 뮤테이션이 어떤 테스트도 죽이지 않고 살아남는다.
func controlDirectoryModeIsSafe(mode os.FileMode) bool {
	return mode.IsDir() && mode&os.ModeSymlink == 0 && mode.Perm() == 0o700
}

// sameDescriptorFile은 "검사한 파일 · 연 파일 · 연 뒤에도 그 자리에 있는 파일"이
// 전부 같은 inode인지 본다. 하나라도 다르면 검사와 사용 사이에 누가 바꿔치기한
// 것이고, 그러면 우리가 검증한 것과 읽는 것이 다른 파일이다.
func sameDescriptorFile(inspected, opened, current os.FileInfo) bool {
	return os.SameFile(inspected, opened) && os.SameFile(opened, current)
}

// errDescriptorChanged는 "검사한 파일과 읽는 파일이 다르다"는 신호다.
//
// 이것은 손상이 아니다. 발행이 rename이므로(design D1-2) 다시 발행하는 그 순간에
// 읽으면 정상적으로 이 답이 나온다 — 반쯤 쓰인 파일을 본 것이 아니라 **다른 완성
// 파일로 바뀐 것**이다. 형식·내용이 깨진 답과 구분되는 이름을 준 이유는, 그
// 구분이 없으면 "발행 중에 반쯤 쓰인 것이 보이는가"를 재는 테스트가 rename의
// 정상 경합을 손상으로 읽기 때문이다.
var errDescriptorChanged = errors.New("strategy projection runtime: descriptor changed while it was opened")

// openVerifiedDescriptor는 descriptor의 **형식**만 확인하고 연 파일을 돌려준다.
//
// 형식과 내용을 가르는 것이 design D1-2다. 형식(정확히 0600인 정규 파일 · symlink를
// 따라가지 않음 · 열기 전후로 같은 inode)은 "이것이 우리가 만든 파일인가"를 묻고,
// 그 답이 아니오면 지금도 거부다. 내용(JSON·필드)은 D2 이후 **어떤 판정에도 쓰이지
// 않으므로**, 반쯤 쓰인 JSON을 거부하는 것은 거부만 생산한다. 회수는 형식만 보고
// (transport_unix.go의 reclaimStaleControlDirectory), 실제로 읽어 쓰는 Dial만
// 내용까지 본다.
func openVerifiedDescriptor(path string) (*os.File, error) {
	if filepath.Base(path) != DescriptorFileName || filepath.Base(filepath.Dir(path)) != ControlDirectoryName {
		return nil, errors.New("strategy projection runtime: unexpected descriptor path")
	}
	directory, err := os.Lstat(filepath.Dir(path))
	if err != nil || !controlDirectoryModeIsSafe(directory.Mode()) {
		return nil, errors.New("strategy projection runtime: control directory is not exact 0700")
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Mode().Perm() != 0o600 {
		return nil, errors.New("strategy projection runtime: descriptor is not exact 0600 regular file")
	}
	file, err := openDescriptorNoFollow(path)
	if err != nil {
		return nil, err
	}
	opened, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	current, currentErr := os.Lstat(path)
	if currentErr != nil || !sameDescriptorFile(info, opened, current) ||
		!opened.Mode().IsRegular() || opened.Mode().Perm() != 0o600 {
		_ = file.Close()
		return nil, errDescriptorChanged
	}
	return file, nil
}

func readDescriptor(path string) (Descriptor, error) {
	file, err := openVerifiedDescriptor(path)
	if err != nil {
		return Descriptor{}, err
	}
	defer file.Close()
	body, err := io.ReadAll(io.LimitReader(file, maxDescriptorBytes+1))
	if err != nil || len(body) > maxDescriptorBytes {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor unreadable or oversized")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	var descriptor Descriptor
	if err := decoder.Decode(&descriptor); err != nil {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor JSON invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor must contain one value")
	}
	if descriptor.SchemaVersion != descriptorSchema || descriptor.Socket != SocketFileName || len(descriptor.Token) < 32 || descriptor.PID <= 0 {
		return Descriptor{}, errors.New("strategy projection runtime: descriptor fields invalid")
	}
	return descriptor, nil
}

func hasRequestBody(request *http.Request) bool {
	return request.ContentLength > 0 || len(request.TransferEncoding) > 0
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

var _ strategyprojection.Reader = (*Client)(nil)
