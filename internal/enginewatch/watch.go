// Package enginewatch 는 엔진 밖에서 엔진의 침묵을 판정하는 감시자임(openspec change a134, engine-safety 「엔진의 침묵은 엔진 밖에서 판정된다」).
//
// # 왜 엔진 밖인가
//
// 보호 주문이 UNWIRED 인 동안 손절 · 청산을 실행하는 것은 엔진 프로세스 하나뿐임. 죽은 프로세스는 스스로 통지할 수 없고,
// ntfy 는 「오지 않은 메시지」를 판정하지 않음(design D0). 그래서 판정은 엔진과 무관하게 사는 프로세스(호스트 systemd 타이머가
// 매 분 부르는 `tossctl engine heartbeat-check`)가 하고, 경보는 그 프로세스가 ntfy 로 보냄.
//
// # 생존 신호는 새로 만들지 않음
//
// 엔진은 이미 저널 디렉터리의 활성 마커(`engine-run.json`, internal/enginelock)를 1분마다 원자적으로 다시 쓰고 정상 정지 때 지움.
// 이 패키지는 그 마커의 수정 시각만 읽음 — 같은 사실을 두 파일이 말하면 생존의 정의가 둘이 됨(design D1).
//
// # 이 패키지가 하지 않는 것
//
// 엔진 · 저널 · 브로커 · 세션을 열지 않음. 엔진 잠금을 잡지 않음. 쓰는 파일은 자기 상태 파일 하나뿐.
package enginewatch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 감시자의 알림 이벤트 타입. obs/event.go 의 값과 겹치지 않음을 시험이 원문을 읽어 확인함(a134 시험).
// obs 등급표(criticalEvents)에 올리지 않음 — 감시자는 엔진 Notifier · outbox 를 거치지 않으므로 등급은 알림 자체의 Severity 로 정함.
const (
	// EventLivenessLost 는 엔진 생존 신호가 끊겼다는 critical 알림임.
	EventLivenessLost obs.EventType = "engine.liveness_lost"
	// EventLivenessRestored 는 끊김이 끝났다는 복구 알림임.
	EventLivenessRestored obs.EventType = "engine.liveness_restored"
)

// StateFileName 은 감시자 상태 파일 이름임. 저널 디렉터리(마커 옆)에 둠 — 엔진은 이 파일을 읽지 않음.
const StateFileName = "engine-heartbeat-watch.json"

// DefaultStaleAfter 는 마커가 이보다 오래되면 끊김으로 보는 기본 기준임.
//
// 3분 = 1분 갱신을 두 번 놓친 시점. 감지 지연 최악 ≈ 기준 + 타이머 1분. enginelock.StaleAfter(5분)와 따로인 이유:
// 그 창은 「유령 마커를 믿는 최대 시간」(콘솔 · autostart), 이 기준은 「사람을 깨우기 전 기다릴 최대 시간」이라 질문이 다름(design D2).
const DefaultStaleAfter = 3 * time.Minute

// Verdict 는 한 번의 관측 판정임.
type Verdict string

const (
	// VerdictUp 은 기준 안에 갱신된 마커가 있다는 판정임.
	VerdictUp Verdict = "up"
	// VerdictDown 은 엔진 생존 신호가 끊겼다는 판정임.
	VerdictDown Verdict = "down"
	// VerdictIdle 은 엔진이 돌 것으로 기대되지 않고(autostart OFF) 마커도 없다는 판정 — 감시 대상 아님.
	VerdictIdle Verdict = "idle"
)

// 끊김의 사유.
const (
	// ReasonStale 은 마커가 기준보다 오래됨 — 죽은 프로세스만 마커를 남긴 채 갱신을 멈춤.
	ReasonStale = "stale"
	// ReasonMissing 은 autostart ON 인데 마커가 없음 — 돌아야 할 엔진이 없음.
	ReasonMissing = "missing"
)

var (
	// ErrNoPublisher 는 알릴 수단이 없다는 오류임. 알릴 수 없는 감시자는 매 실행 실패해야 함(조용한 성공 금지, design D4).
	ErrNoPublisher = errors.New("enginewatch: 알림 전송 수단이 없어 엔진 정지를 알릴 수 없음")
	// ErrStaleAfterTooShort 는 기준이 마커 갱신 주기 이하라는 오류임 — 살아 있는 엔진을 죽었다고 읽게 됨.
	ErrStaleAfterTooShort = fmt.Errorf("enginewatch: 끊김 기준은 마커 갱신 주기(%s)보다 길어야 함", enginelock.RefreshEvery)
)

// Observation 은 마커 한 번 관측의 결과임.
type Observation struct {
	Verdict Verdict
	// Reason 은 Down 일 때만 채움(ReasonStale · ReasonMissing).
	Reason string
	// LastBeat 는 마커의 수정 시각 — 마커가 없으면 0.
	LastBeat time.Time
	// Age 는 now - LastBeat. 미래 시각이면 음수일 수 있음.
	Age time.Duration
	// Ready 는 마커 본문에 ready_at 이 있다는 뜻 — 재시작 복구(Recover)를 마친 엔진만 씀(a102).
	// 복구 판정(끊김 → 정상)은 이것까지 요구함: 마커는 부팅 6단계(Recover 전)에 쓰이고 오류 종료 때 지워지므로,
	// 재기동 루프의 몇 초짜리 마커 하나를 「복구」로 읽으면 거짓 안심 알림이 됨(리뷰 P2-1).
	Ready bool
}

// Observe 는 마커의 수정 시각으로 판정함(design D2 표).
//
// expected 는 「엔진이 돌아야 한다」는 사람의 승인(engine.autostart)임. 마커가 있으면 expected 와 무관하게 시각만 봄 —
// 남아 있는데 멈춘 마커는 수동 운용에서도 죽은 프로세스의 흔적이기 때문.
// 부재 외의 stat 실패는 오류로 돌려줌: 볼 수 없음은 없음이 아님(runlock.Fresh 처럼 접지 않음).
func Observe(markerPath string, now time.Time, staleAfter time.Duration, expected bool) (Observation, error) {
	info, err := os.Stat(markerPath)
	if errors.Is(err, os.ErrNotExist) {
		if expected {
			return Observation{Verdict: VerdictDown, Reason: ReasonMissing}, nil
		}
		return Observation{Verdict: VerdictIdle}, nil
	}
	if err != nil {
		return Observation{}, fmt.Errorf("enginewatch: 엔진 활성 마커를 볼 수 없음(부재로 단정하지 않음): %w", err)
	}
	beat := info.ModTime()
	age := now.Sub(beat)
	ready := markerReady(markerPath)
	// 미래 시각(age < 0)은 시계 불일치이지 죽음이 아님 — 기준 이하로 떨어져 Up 이 됨(runlock.Fresh 선례).
	if age > staleAfter {
		return Observation{Verdict: VerdictDown, Reason: ReasonStale, LastBeat: beat, Age: age, Ready: ready}, nil
	}
	return Observation{Verdict: VerdictUp, LastBeat: beat, Age: age, Ready: ready}, nil
}

// markerReady 는 마커 본문의 ready_at 유무를 읽음. 읽기 · 해석 실패는 「준비 안 됨」 — 복구를 늦추는 방향(거짓 안심 금지).
func markerReady(markerPath string) bool {
	body, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}
	var m enginelock.Marker
	if json.Unmarshal(body, &m) != nil {
		return false
	}
	return m.ReadyAt != nil
}

// State 는 감시자가 기억하는 것 — 「이번 끊김을 이미 알렸는가」뿐. 계좌 칸 없음(시험이 고정).
type State struct {
	Down   bool      `json:"down"`
	Reason string    `json:"reason,omitempty"`
	Since  time.Time `json:"since,omitempty"`
}

// loadState 는 상태 파일을 읽음. 없거나 읽을 수 없거나 손상됐으면 0 값(down 아님)을 돌려줌 —
// 중복 알림은 허용하고 억제는 허용하지 않는 방향(design D3).
func loadState(path string) State {
	body, err := os.ReadFile(path)
	if err != nil {
		return State{}
	}
	var s State
	if json.Unmarshal(body, &s) != nil {
		return State{}
	}
	return s
}

// saveState 는 상태 파일을 원자적으로 바꿈(같은 디렉터리 임시 파일 → 0600 → rename). 실패 경로마다 임시 파일을 지움.
func saveState(path string, s State) error {
	body, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("enginewatch: 상태 직렬화 실패: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".engine-heartbeat-watch-*.json")
	if err != nil {
		return fmt.Errorf("enginewatch: 상태 임시 파일 생성 실패: %w", err)
	}
	staged := tmp.Name()
	if _, err := tmp.Write(append(body, '\n')); err != nil {
		tmp.Close()
		os.Remove(staged)
		return fmt.Errorf("enginewatch: 상태 쓰기 실패: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginewatch: 상태 닫기 실패: %w", err)
	}
	if err := os.Chmod(staged, 0o600); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginewatch: 상태 권한 설정 실패: %w", err)
	}
	if err := os.Rename(staged, path); err != nil {
		os.Remove(staged)
		return fmt.Errorf("enginewatch: 상태 게시 실패: %w", err)
	}
	return nil
}

// Check 는 감시 한 번의 입력임.
type Check struct {
	// Dir 는 엔진 저널 디렉터리(마커와 상태 파일이 사는 곳). 필수.
	Dir string
	// StaleAfter 는 끊김 기준. 0 이면 DefaultStaleAfter. 마커 갱신 주기 이하는 거절.
	StaleAfter time.Duration
	// Expected 는 engine.autostart — 마커 부재를 끊김으로 볼지 정함.
	Expected bool
	// Publisher 는 알림 전송 수단. nil 이면 ErrNoPublisher(판정보다 먼저 검사).
	Publisher obs.Publisher
	// Clock 은 현재 시각. nil 이면 clock.System().
	Clock clock.Clock
}

// Outcome 은 감시 한 번의 결과임.
type Outcome struct {
	Observation Observation
	// Previous 는 이번 실행 전의 상태.
	Previous State
	// Sent 는 이번에 보낸 알림의 타입 — 보내지 않았으면 빈 값.
	Sent obs.EventType
}

// Run 은 관측 → 상태 비교 → (바뀐 순간에만) 알림 한 건 → 전송 성공 뒤 상태 기록 순서로 한 번 감시함(design D3 표).
//
// 핵심 순서: 상태는 전송이 **성공한 뒤에만** 옮김. 그래서 「알렸다고 기록했는데 안 나감」이 생길 수 없고, 실패한 전송은
// 다음 실행(1분 뒤)이 다시 보냄. 상태 기록만 실패하면 다음 실행이 같은 알림을 한 번 더 보낼 수 있음 — 중복은 허용, 누락은 금지.
func (c Check) Run(ctx context.Context) (Outcome, error) {
	// 알릴 수 없는 감시자는 판정과 무관하게 실패 — 엔진이 살아 있을 때도 매번 실패해야 운영자가 빈 감시를 알아챔.
	if c.Publisher == nil {
		return Outcome{}, ErrNoPublisher
	}
	staleAfter := c.StaleAfter
	if staleAfter == 0 {
		staleAfter = DefaultStaleAfter
	}
	if staleAfter <= enginelock.RefreshEvery {
		return Outcome{}, ErrStaleAfterTooShort
	}
	clk := c.Clock
	if clk == nil {
		clk = clock.System()
	}
	now := clk.Now()

	obsv, err := Observe(enginelock.MarkerPath(c.Dir), now, staleAfter, c.Expected)
	if err != nil {
		return Outcome{}, err
	}
	statePath := filepath.Join(c.Dir, StateFileName)
	prev := loadState(statePath)
	out := Outcome{Observation: obsv, Previous: prev}

	down := obsv.Verdict == VerdictDown
	// 상태가 바뀌지 않으면(끊김 지속 · 정상 지속) 아무것도 하지 않음 — 중복 억제의 전부.
	if down == prev.Down {
		return out, nil
	}
	// 끊김에서 벗어나는 것은 복구를 마친(ready_at) 엔진의 신선한 마커 또는 감시 대상 아님(IDLE)일 때만 —
	// 신선하지만 준비 안 된 마커는 끊김을 유지함(재기동 루프 · Recover 에서 멈춘 엔진, 리뷰 P2-1).
	if prev.Down && obsv.Verdict == VerdictUp && !obsv.Ready {
		return out, nil
	}

	var msg obs.Notification
	var next State
	if down {
		msg = lostNotification(obsv, staleAfter)
		next = State{Down: true, Reason: obsv.Reason, Since: now.UTC()}
	} else {
		msg = restoredNotification(obsv, prev)
	}
	if err := c.Publisher.Publish(ctx, msg); err != nil {
		return out, fmt.Errorf("enginewatch: 알림 전송 실패 — 상태를 옮기지 않으므로 다음 실행이 다시 보냄: %w", err)
	}
	out.Sent = msg.Type
	if err := saveState(statePath, next); err != nil {
		return out, err
	}
	return out, nil
}

// EventWatchTest 는 설치 때 사람이 보내는 전달 확인 알림의 타입임(--send-test, 리뷰 P1-1).
const EventWatchTest obs.EventType = "engine.liveness_watch_test"

// SendTest 는 감시자의 전송 수단으로 normal 알림 한 건을 보냄 — 판정 · 상태와 무관.
//
// 판정은 상태가 바뀔 때만 보내므로, 토큰 누락 · 잘못된 토픽 · 401 같은 「수단은 있는데 전달이 안 됨」은 첫 끊김 때에야 드러남.
// 설치 단계에서 단위와 같은 환경으로 이것을 한 번 돌려 전달을 실측함. 실패는 오류(성공으로 끝나지 않음).
func SendTest(ctx context.Context, publisher obs.Publisher) error {
	if publisher == nil {
		return ErrNoPublisher
	}
	if err := publisher.Publish(ctx, obs.Notification{
		Type:     EventWatchTest,
		Severity: obs.SeverityNormal,
		Title:    "TossOS 엔진 생존 감시 시험",
		Body:     "이 메시지가 보이면 엔진 생존 감시(tossctl engine heartbeat-check)의 끊김 · 복구 알림이 이 채널로 도착함.",
	}); err != nil {
		return fmt.Errorf("enginewatch: 시험 알림 전송 실패 — 감시자가 끊김을 알릴 수 없음: %w", err)
	}
	return nil
}

// kst 는 알림 본문의 시각 표기용 고정 시간대(사람이 읽는 운영 시각).
var kst = time.FixedZone("KST", 9*60*60)

// lostNotification 은 끊김 알림을 만듦. 본문은 판정 사실(마커 · 시각 · 기준)과 걸린 것(손절)과 할 일만 — 계좌 정보 없음.
func lostNotification(o Observation, staleAfter time.Duration) obs.Notification {
	var cause string
	if o.Reason == ReasonStale {
		cause = fmt.Sprintf("엔진 활성 마커(%s)가 %s 동안 갱신되지 않음 — 마지막 갱신 %s, 기준 %s. 엔진 프로세스가 마커를 지우지 못하고 멈췄거나(OOM · 강제 종료 · 호스트 정지) 마커 쓰기가 실패하는 중(디스크).",
			enginelock.MarkerFileName, o.Age.Round(time.Second), o.LastBeat.In(kst).Format("2006-01-02 15:04:05 MST"), staleAfter)
	} else {
		cause = fmt.Sprintf("엔진 활성 마커(%s)가 없음 — 자동 기동(engine.autostart)이 켜져 있는데 엔진이 돌고 있지 않거나 마커를 쓰지 못함.",
			enginelock.MarkerFileName)
	}
	return obs.Notification{
		Type:     EventLivenessLost,
		Severity: obs.SeverityCritical,
		Title:    "TossOS 엔진 생존 신호 끊김",
		Body: cause + "\n엔진이 멈추면 손절 · 청산이 돌지 않음(보호 주문 UNWIRED). " +
			"콘솔의 엔진 상태와 컨테이너 프로세스(docker top)를 확인하고 엔진을 다시 시작할 것.",
	}
}

// restoredNotification 은 복구 알림을 만듦. 지금 Up 인지 Idle(정상 정지 · 자동 기동 OFF)인지 구분해 말함.
func restoredNotification(o Observation, prev State) obs.Notification {
	var now string
	if o.Verdict == VerdictUp {
		now = fmt.Sprintf("엔진 활성 마커가 다시 갱신되고 재시작 복구를 마침 — 마지막 갱신 %s.", o.LastBeat.In(kst).Format("2006-01-02 15:04:05 MST"))
	} else {
		now = "엔진이 정상 정지 상태이고 자동 기동이 꺼져 있어 감시 대상에서 벗어남."
	}
	if !prev.Since.IsZero() {
		now += fmt.Sprintf(" 끊김 감지 시각 %s.", prev.Since.In(kst).Format("2006-01-02 15:04:05 MST"))
	}
	return obs.Notification{
		Type:     EventLivenessRestored,
		Severity: obs.SeverityNormal,
		Title:    "TossOS 엔진 생존 신호 복구",
		Body:     now,
	}
}
