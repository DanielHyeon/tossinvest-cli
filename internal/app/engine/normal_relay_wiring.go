package engine

// normal_relay_wiring.go 는 a092 C8 의 조립 한 조각 — exit 관측기와 런타임이 같은 일반 등급 이관을 씀.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// normalAlertRelayName 은 일반 등급 이관 실행자의 이름임(감독 루프가 아님 — Runtime.LoopNames 에 나오면 안 됨).
const normalAlertRelayName = "normal-alert-relay"

// NormalAlertRelay 는 이 엔진의 일반 등급 이관을 돌려줌(처음 부를 때 만듦). nil Context 는 nil.
func (c *Context) NormalAlertRelay() *obs.NormalRelay {
	if c == nil {
		return nil
	}
	c.normalRelayOnce.Do(func() {
		c.normalRelay = obs.NewNormalRelay(c.Notifier, obs.DefaultNormalRelayCapacity)
	})
	return c.normalRelay
}

// NormalAlertRelayExecutor 는 그 이관을 비우는 보조 실행자임. 정지는 로그만(자기 이벤트 타입) — best-effort 등급이라 래치 없음.
func (c *Context) NormalAlertRelayExecutor() AuxiliaryExecutor {
	return AuxiliaryExecutor{
		Name:      normalAlertRelayName,
		Run:       c.NormalAlertRelay().Run,
		StopEvent: obs.EventNormalAlertRelayStopped,
	}
}
