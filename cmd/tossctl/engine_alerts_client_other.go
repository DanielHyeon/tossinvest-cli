//go:build !unix

package main

// Unix 소켓이 없는 곳에는 엔진의 알림 표면도 없다 (alert_control_transport_other.go).
//
// 여기서 원장을 직접 열어 승인하는 대체 경로를 만들지 않는다. 그 경로는 원장만
// 고치고 **이 프로세스 밖의** 진입 게이트는 못 푼다 — 운영자가 승인해도 진입은
// 재시작까지 막힌 채다. 그것이 design D7.1 이 없애려는 상태 그 자체다.

import (
	"context"
	"errors"
)

func dialAlertControl(context.Context, string) (*alertControlClient, error) {
	return nil, errors.New("engine alerts: Unix sockets are unsupported on this platform")
}
