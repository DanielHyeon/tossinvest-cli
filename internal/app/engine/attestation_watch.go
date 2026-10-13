package engine

// attestation_watch.go — a135: 엔진이 재시작할 때 읽을 capability attestation 의 만료 · 범위 상실을 재시작 **전에** 알림.
//
// # 왜 필요한가
//
// 게이트가 켜진 엔진은 기동할 때만 attestation 을 읽음(verifyGate 4~6). 만료되었거나 필요한 endpoint 를 잃은 파일로
// 재시작하면 기동이 거절되고, 그때는 청산 관측 루프도 돌지 않음 — 보호 주문이 UNWIRED 인 지금 손절이 없어짐.
// engine-safety 「등급화된 알림」은 「자격증명 만료 임박」을 critical 로 정했지만 그 알림을 내는 곳이 없었음.
//
// # 판정은 인터록의 것을 그대로 씀
//
// 「지금 재시작하면 무장하는가」는 attest.Attestation.Verify + RequiredEndpoints() + 엔진 자격 증명의 계좌로 묻는 인터록의
// 술어 그대로임. 여기서 규칙을 다시 쓰면 두 판정이 갈릴 수 있으므로 사본을 두지 않음. 이 파일이 더하는 것은 「통과했다면
// 며칠 남았나」 하나뿐임.
//
// # 엔진을 해치지 않는 구조
//
//   - 감독 루프가 아니라 보조 실행자(Auxiliary) — 정지해도 다른 루프를 내리지 않음(auxiliary.go).
//   - 판정 한 번의 패닉은 그 판정 안에서 회수되고 다음 판정이 돎. 실행자의 정지는 자기 이벤트 타입으로 로그만 남김(래치 없음).
//   - 읽는 것은 로컬 파일 하나, 1시간마다. 브로커 호출 0.
//   - Verified · EntryPermitted · 게이트 · 운영 모드를 쓰지 않음. 단, critical 알림이 전달 한도를 넘겨 실패하면 알림기의 기존
//     규칙대로 진입이 잠김(청산 · 손절은 무관) — 알림이 꺼진 엔진은 그래서 critical 을 내지 않음(a091 · a095 규칙).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	// attestationWatchName 은 보조 실행자 이름임(감독 루프가 아님 — Runtime.LoopNames 에 나오면 안 됨).
	attestationWatchName = "attestation-watch"
	// attestationWatchInterval 은 판정 주기임. 갱신 타이머(6시간)보다 짧아서 갱신이 범위를 잃으면 한 시간 안에 알림.
	attestationWatchInterval = time.Hour
	// attestationNoticeDays 이하 남으면 normal 사전 통지, attestationCriticalDays 이하면 critical.
	attestationNoticeDays   = 7
	attestationCriticalDays = 2
)

// attestationStage 는 판정 단계임. 값은 알림 필드 · 중복 방지 키에 그대로 쓰는 영문 코드.
type attestationStage string

const (
	stageHealthy         attestationStage = "healthy"          // D-8 이상 — 알릴 것 없음
	stageNotice          attestationStage = "notice"           // D-7 ~ D-3 — normal
	stageImminent        attestationStage = "imminent"         // D-2 ~ D-1 — critical
	stageExpired         attestationStage = "expired"          // 만료 — critical
	stageScopeLost       attestationStage = "scope_lost"       // 필요 endpoint 누락 — critical
	stageAccountMismatch attestationStage = "account_mismatch" // 다른 계좌의 attestation — critical
	stageUnreadable      attestationStage = "unreadable"       // 없음 · 깨짐 · 형식 초과 — critical
	stageIncomplete      attestationStage = "incomplete"       // 읽혔지만 계좌 · 만료 · endpoint 칸이 빔 — critical
)

// attestationVerdict 는 판정 한 번의 결과임. 계좌는 담지 않음.
type attestationVerdict struct {
	Stage     attestationStage
	DaysLeft  int // notice · imminent 에서만 의미 있음(남은 시간을 하루 단위로 올림)
	IssuedAt  time.Time
	ExpiresAt time.Time
	Missing   []string // scope_lost 의 누락 endpoint(정렬됨)
	// ProofAt 은 POST endpoint 마다의 가장 새 감독 증거 중 가장 이른 것임 — 갱신이 POST 를 유지하려면 이 증거가 유효 기간
	// 안이어야 하므로 운영자가 「언제 감독 검증을 다시 해야 하나」를 읽는 칸. 판정에는 쓰지 않음.
	ProofAt time.Time
}

// critical 은 재시작이 거절될 상태이거나 이틀 안에 그렇게 될 상태인지임.
func (v attestationVerdict) critical() bool {
	switch v.Stage {
	case stageHealthy, stageNotice:
		return false
	default:
		return true
	}
}

// judgeAttestation 은 엔진이 지금 재시작하면 이 attestation 으로 무장하는지, 무장한다면 며칠 남았는지 판정함.
//
// loadErr 는 attest.Load 의 오류임 — 없는 파일 · 깨진 파일 모두 재시작 거절이므로 unreadable 로 묶음.
// Verify 의 순서(필수 칸 → 만료 → 계좌 → endpoint)를 그대로 따르므로 만료와 범위 상실이 겹치면 만료로 보고됨.
func judgeAttestation(att attest.Attestation, loadErr error, now time.Time, accountRef string) attestationVerdict {
	if loadErr != nil {
		return attestationVerdict{Stage: stageUnreadable}
	}
	v := attestationVerdict{IssuedAt: att.IssuedAt, ExpiresAt: att.ExpiresAt, ProofAt: binding(att.SupervisedBy)}
	required := RequiredEndpoints()
	if err := att.Verify(now, accountRef, required); err != nil {
		switch {
		case errors.Is(err, attest.ErrExpired):
			v.Stage = stageExpired
		case errors.Is(err, attest.ErrEndpointNotAttested):
			v.Stage = stageScopeLost
			v.Missing = att.MissingEndpoints(required)
		case errors.Is(err, attest.ErrAccountMismatch):
			v.Stage = stageAccountMismatch
		case errors.Is(err, attest.ErrIncomplete):
			v.Stage = stageIncomplete
		default:
			// 알 수 없는 거절 — 인터록이 거절할 파일이라는 점에서 읽을 수 없는 파일과 같음.
			v.Stage = stageUnreadable
		}
		return v
	}

	// Verify 를 통과했으므로 남은 시간은 양수임. 하루 단위 올림: 6.5일 남음 = D-7, 정확히 7일 = D-7, 7일 1초 = D-8.
	v.DaysLeft = daysLeftCeil(att.ExpiresAt.Sub(now))
	switch {
	case v.DaysLeft > attestationNoticeDays:
		v.Stage = stageHealthy
	case v.DaysLeft > attestationCriticalDays:
		v.Stage = stageNotice
	default:
		v.Stage = stageImminent
	}
	return v
}

// daysLeftCeil 은 남은 시간을 하루 단위로 올림함. remaining 은 양수로 들어옴.
func daysLeftCeil(remaining time.Duration) int {
	const day = 24 * time.Hour
	days := int(remaining / day)
	if remaining%day != 0 {
		days++
	}
	return days
}

// binding 은 endpoint 마다 가장 새 감독 증거를 고른 뒤 그중 가장 이른 시각을 돌려줌 — 가장 먼저 기간을 벗어날 POST 근거.
// 같은 endpoint 의 오래된 증거가 새 증거를 가리지 않게 endpoint 별 최신을 먼저 고름. 없으면 0.
func binding(proofs []attest.Proof) time.Time {
	newest := map[string]time.Time{}
	for _, p := range proofs {
		key := strings.Join(strings.Fields(strings.ToUpper(p.Endpoint)), " ")
		if p.At.IsZero() || key == "" {
			continue
		}
		if p.At.After(newest[key]) {
			newest[key] = p.At
		}
	}
	var earliest time.Time
	for _, at := range newest {
		if earliest.IsZero() || at.Before(earliest) {
			earliest = at
		}
	}
	return earliest
}

// attestationNotifier 는 감시가 쓰는 알림기의 한 메서드임 — 시험이 기록기를 끼울 수 있게.
type attestationNotifier interface {
	Notify(ctx context.Context, e obs.Event) error
}

// attestationWatch 는 판정을 주기적으로 돌리고 단계마다 한 번 알리는 보조 실행자의 몸통임.
// 상태(identity · sent)는 Run 의 goroutine 하나만 만지므로 잠금이 없음.
type attestationWatch struct {
	enabled    bool   // 게이트가 켜졌는지 — 꺼진 엔진은 attestation 에 의존하지 않음
	path       string // 인터록이 읽은 바로 그 파일(AutomationStatus.AttestationFile)
	accountRef string // 엔진 자격 증명의 계좌 — 판정 입력일 뿐 알림에 싣지 않음
	critical   bool   // notifications.enabled — 꺼져 있으면 critical 단계도 normal 타입으로
	notifier   attestationNotifier
	log        *obs.Logger
	clock      clock.Clock
	interval   time.Duration
	load       func(string) (attest.Attestation, error)

	identity string          // 마지막으로 본 attestation 의 신원(발급 · 만료 시각)
	sent     map[string]bool // 이 신원에서 이미 알린 단계 키
}

// newAttestationWatch 는 엔진 Context 에서 감시를 만듦.
func newAttestationWatch(c *Context, clk clock.Clock) *attestationWatch {
	w := &attestationWatch{interval: attestationWatchInterval, load: attest.Load, clock: clk}
	if c == nil {
		return w
	}
	if w.clock == nil {
		w.clock = c.clock()
	}
	w.enabled = c.Automation.Enabled
	w.path = c.Automation.AttestationFile
	w.accountRef = c.AccountRef
	w.critical = c.Config.Engine.Notifications.Enabled
	w.log = c.Log
	// nil *obs.Notifier 를 그대로 넣으면 nil 이 아닌 인터페이스가 되어 Notify 가 패닉함 — 명시적으로 걸러 냄.
	if c.Notifier != nil {
		w.notifier = c.Notifier
	}
	return w
}

// AttestationWatch 는 a135 attestation 감시 보조 실행자를 돌려줌. 정지 처리(OnStop)는 없음 — 감시의 정지가 진입이나
// 청산을 잠그면 안 되므로 자기 이벤트 타입의 로그 한 줄로 끝남.
func (c *Context) AttestationWatch(clk clock.Clock) AuxiliaryExecutor {
	w := newAttestationWatch(c, clk)
	return AuxiliaryExecutor{
		Name:      attestationWatchName,
		Run:       w.Run,
		StopEvent: obs.EventAttestationWatchFailed,
	}
}

// Run 은 기동 직후 한 번, 그 뒤 interval 마다 판정함. 취소되면 ctx.Err() 를 돌려줌(런타임은 그것을 정상 정지로 읽음).
func (w *attestationWatch) Run(ctx context.Context) error {
	if !w.enabled {
		// 게이트 꺼짐 — 읽을 것도 알릴 것도 없음. 일찍 돌아가면 런타임이 「정지」로 기록하므로 취소를 기다림.
		<-ctx.Done()
		return ctx.Err()
	}
	for {
		w.checkOnce(ctx)
		if err := w.clock.Sleep(ctx, w.interval); err != nil {
			return err
		}
	}
}

// checkOnce 는 판정 한 번임.
//
// 회수 경계는 이 파일의 코드(읽기 · 판정 · 문구)에만 둠. 알림기 호출은 회수 경계 밖에서 함 — 알림기는 잠금(n.mu)을 defer 없이
// 놓으므로 그 안의 패닉을 회수하면 잠금이 영구히 잡혀 청산 관측 루프의 알림이 조용히 멈춤(a135 리뷰 P1). 그래서 Notify 는
// 별도 goroutine 에서 불러 그 패닉이 런타임의 회수(runAuxiliaryBody)에도 닿지 않게 함 — 감독 루프의 호출자와 같은 결과(프로세스
// 종료)이지, 살아 있는 채 멈춘 엔진이 아님.
func (w *attestationWatch) checkOnce(ctx context.Context) {
	event, key, ok := w.prepare()
	if !ok {
		return
	}
	if w.notifier == nil {
		// 알림기 없음 — 로그만. 알림기가 있으면 Notify 가 같은 줄을 스스로 남김.
		if w.log != nil {
			w.log.Warn(event.Type, obs.FieldReason, event.Fields[obs.FieldReason], obs.FieldDetail, event.Title)
		}
		w.sent[key] = true
		return
	}
	if ctx.Err() != nil {
		// 종료 중 — 기록을 시작하지 않음(중간에 끊긴 기록이 래치 · 오류 줄을 남기지 않게). 다음 기동이 다시 판정함.
		return
	}
	if err := w.notifyOutsideRecovery(ctx, event); err != nil {
		// critical 기록 실패 — 보낸 것으로 세지 않음. 다음 판정이 다시 시도함.
		if w.log != nil {
			w.log.Error(obs.EventAttestationWatchFailed, err,
				obs.FieldReason, "notify_failed",
				obs.FieldDetail, "the attestation expiry alert could not be recorded; it is retried on the next check")
		}
		return
	}
	w.sent[key] = true
}

// notifyOutsideRecovery 는 알림기를 회수 경계 밖 goroutine 에서 부르고 끝날 때까지 기다림(checkOnce 주석).
func (w *attestationWatch) notifyOutsideRecovery(ctx context.Context, event obs.Event) error {
	done := make(chan error, 1)
	go func() { done <- w.notifier.Notify(ctx, event) }()
	return <-done
}

// prepare 는 읽기 · 판정 · 중복 확인 · 문구를 함. 알릴 것이 있으면 ok. 패닉은 여기서 회수됨 — 판정 결함이 실행자를 끝내지 않게.
// 이 함수 안에는 다른 패키지의 잠금을 잡는 호출이 없음(attest.Load · 순수 판정 · 문구 생성뿐).
func (w *attestationWatch) prepare() (event obs.Event, key string, ok bool) {
	defer func() {
		if rec := recover(); rec != nil {
			ok = false
			if w.log != nil {
				w.log.Warn(obs.EventAttestationWatchFailed,
					obs.FieldReason, "check_panicked",
					obs.FieldDetail, "the attestation expiry check panicked; the next check runs on schedule")
			}
		}
	}()

	att, err := w.load(w.path)
	now := w.clock.Now()
	v := judgeAttestation(att, err, now, w.accountRef)

	// 신원이 바뀌면(갱신 · 파일 교체 · 사라짐) 이미 알린 단계를 잊음 — 새 파일의 단계는 새로 셈.
	identity := v.IssuedAt.UTC().Format(time.RFC3339Nano) + "|" + v.ExpiresAt.UTC().Format(time.RFC3339Nano)
	if identity != w.identity || w.sent == nil {
		w.identity = identity
		w.sent = map[string]bool{}
	}
	if v.Stage == stageHealthy {
		return obs.Event{}, "", false
	}
	key = stageKey(v, now)
	if w.sent[key] {
		return obs.Event{}, "", false
	}
	return w.eventFor(v, key), key, true
}

// stageKey 는 중복 방지 키임. 날수 단계는 단계 + 날수(D-7 과 D-6 은 다른 키라 하루 한 번씩). 재시작이 거절될 상태
// (만료 · 범위 상실 · 계좌 불일치 · 무효)는 단계 + KST 날짜 — 고칠 때까지 하루 한 번 다시 알림(한 번 알리고 침묵하지 않게).
func stageKey(v attestationVerdict, now time.Time) string {
	if v.Stage == stageNotice || v.Stage == stageImminent {
		return fmt.Sprintf("%s:%d", v.Stage, v.DaysLeft)
	}
	return string(v.Stage) + ":" + now.In(kst).Format("2006-01-02")
}

// kst 는 운영자가 읽는 시각대임. tz 데이터 없이도 같은 값이 나오게 고정 오프셋으로 둠.
var kst = time.FixedZone("KST", 9*60*60)

// eventFor 는 판정을 알림 하나로 바꿈. 제목 · 본문은 한국어 고정 문구, 계좌 · 파일 경로 · 원문 오류는 싣지 않음(불변식 8).
func (w *attestationWatch) eventFor(v attestationVerdict, key string) obs.Event {
	kind := obs.EventCredentialExpiryNotice
	if v.critical() && w.critical {
		kind = obs.EventCredentialExpiring
	}

	expires := "알 수 없음"
	if !v.ExpiresAt.IsZero() {
		expires = v.ExpiresAt.In(kst).Format("2006-01-02 15:04 KST")
	}
	const consequence = "만료되거나 무효인 attestation 으로 엔진이 재시작하면 자동화 게이트가 무장하지 않고 청산(손절) 관측 루프도 돌지 않는다. " +
		"지금 실행 중인 엔진은 영향받지 않는다."
	const remedy = "갱신에는 soak 통과와, 유효 기간(30일) 안의 감독하 실주문 검증(POST 주문 · 취소 증거)이 함께 필요하다."

	var title, body string
	switch v.Stage {
	case stageNotice, stageImminent:
		title = fmt.Sprintf("자동매매 검증서(attestation) 만료 D-%d", v.DaysLeft)
		body = fmt.Sprintf("capability attestation 이 %s 에 만료된다(%d일 남음). %s %s", expires, v.DaysLeft, consequence, remedy)
	case stageExpired:
		title = "자동매매 검증서(attestation) 만료됨"
		body = fmt.Sprintf("capability attestation 이 %s 에 만료되었다. %s %s", expires, consequence, remedy)
	case stageScopeLost:
		title = "자동매매 검증서(attestation)에 필요한 주문 권한이 없다"
		body = fmt.Sprintf("현재 attestation 에 엔진이 쓰는 endpoint 가 빠져 있다: %s. 갱신이 감독 증거 없이 쓰였을 수 있다. %s %s",
			strings.Join(v.Missing, ", "), consequence, remedy)
	case stageIncomplete:
		title = "자동매매 검증서(attestation)에 필수 칸이 비어 있다"
		body = "엔진이 읽는 attestation 에 계좌 · 만료 · 검증된 endpoint 중 비어 있는 칸이 있다. " + consequence + " " + remedy
	case stageAccountMismatch:
		title = "자동매매 검증서(attestation)가 엔진의 계좌와 다르다"
		body = "현재 attestation 이 엔진 자격 증명의 계좌와 다른 계좌를 가리킨다. " + consequence
	default:
		title = "자동매매 검증서(attestation)를 쓸 수 없다"
		body = "엔진이 읽는 attestation 파일이 없거나 읽을 수 없다. " + consequence + " " + remedy
	}

	fields := map[string]any{
		obs.FieldReason: string(v.Stage),
	}
	if !v.ExpiresAt.IsZero() {
		fields["expires_at"] = v.ExpiresAt.UTC().Format(time.RFC3339)
	}
	if v.Stage == stageNotice || v.Stage == stageImminent {
		fields["days_left"] = v.DaysLeft
	}
	if len(v.Missing) > 0 {
		fields["missing_endpoints"] = strings.Join(v.Missing, ", ")
	}
	if !v.ProofAt.IsZero() {
		fields["supervised_proof_at"] = v.ProofAt.UTC().Format(time.RFC3339)
	}

	return obs.Event{
		Type: kind,
		// outbox 키는 신원 · 단계마다 하나 — 다른 단계를 같은 조건으로 읽어 「이미 보냄」으로 삼키지 않게.
		Key:    "attestation-watch:" + key + ":" + w.identity,
		Title:  title,
		Body:   body,
		Fields: fields,
	}
}
