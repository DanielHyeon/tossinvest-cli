//go:build unix

package main

// dialModeControl 은 모드 제어 descriptor 를 읽고 소켓에 붙음 — 알림 제어 클라이언트와 같은 검사(positionpolicyrpc)를 씀.

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

// errEngineModeControlUnavailable 은 원인을 단정하지 않음(a109 D3a-2 와 같은 이유 — 엔진이 이 표면 없이 강등 부팅했을 수 있음).
var errEngineModeControlUnavailable = errors.New(
	"engine mode-release: 완화 표면이 없다 (descriptor 없음). " +
		"이 디렉터리로 도는 엔진이 없다면 `tossctl engine run` 을 살려라 — " +
		"엔진이 돌고 있다면 그 엔진이 이 표면 없이 강등 부팅한 것이고, 원인은 엔진 로그의 mode control 강등 보고에 있다. " +
		"원장을 직접 고치는 대체 경로는 없다 — 산 게이트는 엔진 안에 있다")

func dialModeControl(ctx context.Context, engineDir string) (*modeControlClient, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	dir := strings.TrimSpace(engineDir)
	if dir == "" {
		return nil, errors.New("engine mode-release: engine directory is empty")
	}
	descriptorPath := engine.ModeControlDescriptorPath(dir)
	f, err := positionpolicyrpc.OpenPrivateDescriptorFile(descriptorPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errEngineModeControlUnavailable
		}
		return nil, fmt.Errorf("engine mode-release: opening the descriptor: %w", err)
	}
	defer f.Close()
	var descriptor engine.ModeControlDescriptor
	body, err := io.ReadAll(io.LimitReader(f, 4<<10+1))
	if err != nil || len(body) > 4<<10 {
		return nil, errors.New("engine mode-release: descriptor is unreadable or oversized")
	}
	decoder := json.NewDecoder(strings.NewReader(string(body)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&descriptor); err != nil {
		return nil, errors.New("engine mode-release: descriptor JSON is invalid")
	}
	if descriptor.Socket != engine.ModeControlSocketFileName() || descriptor.PID <= 0 ||
		len(strings.TrimSpace(descriptor.Token)) < 32 {
		return nil, errors.New("engine mode-release: descriptor fields are invalid")
	}
	socketPath := filepath.Join(filepath.Dir(descriptorPath), descriptor.Socket)
	if err := positionpolicyrpc.ValidatePrivateSocket(socketPath); err != nil {
		return nil, fmt.Errorf("engine mode-release: validating the socket: %w", err)
	}
	transport := &http.Transport{Proxy: nil, DisableKeepAlives: true,
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, "unix", socketPath)
		}}
	return &modeControlClient{baseURL: "http://mode", token: descriptor.Token,
		http: &http.Client{Transport: transport, Timeout: 30 * time.Second}}, nil
}
