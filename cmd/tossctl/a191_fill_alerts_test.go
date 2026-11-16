package main

import (
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// TestA191FillAlertsFilterTheNilNotifier — nil *obs.Notifier 는 nil 인터페이스가 되어야 함(그대로 넣으면 Notify 가 패닉).
func TestA191FillAlertsFilterTheNilNotifier(t *testing.T) {
	if got := engineFillAlerts(nil); got != nil {
		t.Fatalf("engineFillAlerts(nil) = %#v, want a nil interface", got)
	}
	n := &obs.Notifier{}
	if got := engineFillAlerts(n); got != n {
		t.Fatalf("engineFillAlerts(n) = %#v, want the notifier itself", got)
	}
}

// TestA191TheDetectorIsWiredToTheEngineNotifier — 운영 조립이 감지기의 경보 통로를 엔진 알림기에 묶음.
func TestA191TheDetectorIsWiredToTheEngineNotifier(t *testing.T) {
	ectx := verifiedEngine()
	d, err := engineFillDetector(ectx, clock.System(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.Alerts != nil {
		t.Fatalf("Alerts = %#v with no engine notifier, want nil", d.Alerts)
	}
	ectx.Notifier = &obs.Notifier{}
	d, err = engineFillDetector(ectx, clock.System(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if d.Alerts != ectx.Notifier {
		t.Fatalf("Alerts = %#v, want the engine notifier", d.Alerts)
	}
}
