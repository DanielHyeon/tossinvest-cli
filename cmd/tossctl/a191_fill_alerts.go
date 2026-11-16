package main

// a191_fill_alerts.go 는 체결 감지의 미해결 경보 통로를 엔진 알림기에 묶음.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/filldetect"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// engineFillAlerts 는 nil *obs.Notifier 를 nil 인터페이스로 바꿈 — 그대로 넣으면 nil 이 아닌 인터페이스가 되어 Notify 가
// 패닉함(attestation_watch.go 의 같은 거름 선례).
func engineFillAlerts(n *obs.Notifier) filldetect.Alerter {
	if n == nil {
		return nil
	}
	return n
}
