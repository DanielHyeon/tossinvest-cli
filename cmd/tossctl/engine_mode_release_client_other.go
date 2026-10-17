//go:build !unix

package main

// Unix 소켓이 없는 곳에는 엔진의 모드 완화 표면도 없음. 원장을 직접 고치는 대체 경로는 만들지 않음(산 게이트를 못 풂).

import (
	"context"
	"errors"
)

func dialModeControl(context.Context, string) (*modeControlClient, error) {
	return nil, errors.New("engine mode-release: Unix sockets are unsupported on this platform")
}
