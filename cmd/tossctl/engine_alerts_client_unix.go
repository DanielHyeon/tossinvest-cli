//go:build unix

package main

// dialAlertControl 은 descriptor 를 읽고 소켓에 붙는다.
//
// 검사는 `positionpolicyrpc` 의 것을 그대로 쓴다 — 4.4b-1 이 그 검사들의 이름 없는
// 형태를 그 패키지에 만들어 둔 이유다. 여기에 다시 쓰면 서버가 거는 경계와 클라이언트가
// 확인하는 경계가 **두 곳**이 되고, 두 곳이 되면 갈라진다.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

// errEngineAlertsUnavailable 은 「엔진이 안 돌고 있다」를 dial 실패와 구분해서 말한다.
//
// descriptor 가 없다는 것은 오늘 사실상 하나를 뜻한다: 이 디렉터리로 도는 엔진이
// 없다. 그것을 `no such file` 로 찍으면 운영자는 자기 경로를 의심하고, 진짜로
// 의심해야 할 것 — 승인할 대상 프로세스가 없다 — 은 안 보인다.
var errEngineAlertsUnavailable = errors.New(
	"engine alerts: 이 디렉터리로 도는 엔진이 없다 (descriptor 없음). " +
		"`tossctl engine run` 이 살아 있어야 승인이 진입 게이트까지 닿는다")

const maxAlertDescriptorBytes = 4 << 10

func dialAlertControl(ctx context.Context, engineDir string) (*alertControlClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine alerts: engine directory is empty")
	}
	descriptorPath := engine.AlertControlDescriptorPath(dir)
	f, err := positionpolicyrpc.OpenPrivateDescriptorFile(descriptorPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errEngineAlertsUnavailable
		}
		return nil, fmt.Errorf("engine alerts: opening the descriptor: %w", err)
	}
	defer f.Close()
	descriptor, err := decodeAlertControlDescriptor(f)
	if err != nil {
		return nil, err
	}
	// 이름을 고정한다. 아무 이름이나 받으면 이 디렉터리에 쓸 수 있는 누군가가
	// **다른 엔드포인트를 가리키는** descriptor 를 놓아 둘 수 있다.
	if descriptor.Socket != engine.AlertControlSocketFileName() || descriptor.PID <= 0 ||
		len(strings.TrimSpace(descriptor.Token)) < 32 {
		return nil, errors.New("engine alerts: descriptor fields are invalid")
	}
	socketPath := filepath.Join(filepath.Dir(descriptorPath), descriptor.Socket)
	if err := positionpolicyrpc.ValidatePrivateSocket(socketPath); err != nil {
		return nil, fmt.Errorf("engine alerts: validating the socket: %w", err)
	}
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		}}
	return &alertControlClient{
		baseURL: "http://alerts",
		token:   descriptor.Token,
		// 목록에 상한이 없다 (alertops.go Pending) — 밀린 것이 수천이면 응답이 길다.
		// 그래서 runtime 읽기 소켓의 5초보다 넉넉하게 준다.
		http: &http.Client{Transport: transport, Timeout: 30 * time.Second},
	}, nil
}

func decodeAlertControlDescriptor(r io.Reader) (engine.AlertControlDescriptor, error) {
	var descriptor engine.AlertControlDescriptor
	body, err := io.ReadAll(io.LimitReader(r, maxAlertDescriptorBytes+1))
	if err != nil || len(body) > maxAlertDescriptorBytes {
		return descriptor, errors.New("engine alerts: descriptor is unreadable or oversized")
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&descriptor); err != nil {
		return descriptor, errors.New("engine alerts: descriptor JSON is invalid")
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return descriptor, errors.New("engine alerts: descriptor must contain one JSON value")
	}
	return descriptor, nil
}
