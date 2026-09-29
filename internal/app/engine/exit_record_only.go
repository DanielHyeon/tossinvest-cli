package engine

// exit_record_only.go 는 a092 C1 의 주입 지점별 기록 전용 배선 도우미임. Context.ExitObserver 가 부름.

import (
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// exitSideRetrier 는 공유 Retrier 의 값 복사본에 기록 전용 통지자만 바꿔 끼움.
//
// 복사가 안전한 이유: execgw.Retrier 는 설정값만 가진 구조체이고(뮤텍스 · 계수기 없음) 상태(신선도 · 래치)는 공유
// Gate 포인터에 있음. 그래서 복사본의 신선도 기록과 401 래치는 같은 Gate 로 가고, 공유 Retrier 를 쓰는 범위 밖 소비자
// (대사 루프 · 운영 명령 · 편입 등)의 동기 통지는 그대로임(a092 Q1 문자 해석).
func exitSideRetrier(shared *execgw.Retrier, announcer journal.ModeAnnouncer) *execgw.Retrier {
	if shared == nil {
		return nil
	}
	exit := *shared
	exit.Announcer = announcer
	return &exit
}

// exitSideFloor 는 청산 수량 상한 공급자의 복사본이 exit 쪽 Retrier 로 조회하게 함 — 그 조회의 401 강화 통지도
// exit goroutine 에서 일어나기 때문임. 엔진 공급자가 없으면 오늘과 같이 그 nil 을 그대로 넘김(ConfirmedFloor 가 nil 수신자를 처리함).
func exitSideFloor(shared *reconcileFloor, retrier *execgw.Retrier) *reconcileFloor {
	if shared == nil {
		return shared
	}
	exit := *shared
	exit.retrier = retrier
	return &exit
}
