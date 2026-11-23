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
	stageProofOverdue    attestationStage = "proof_overdue"    // a196: 감독 증거의 기한은 지났지만 파일은 아직 통과 — critical, 하루 한 번
)

// 기한의 종류(알림 필드 deadline_kind 값). a196: 재시작이 거절되기 시작할 시각이 무엇에서 왔는지 운영자가 읽는 칸.
const (
	deadlineExpiry = "expiry"           // attestation 자신의 만료
	deadlineProof  = "supervised_proof" // 실린 POST 감독 증거가 유효 기간을 벗어나는 시각
)

// attestationVerdict 는 판정 한 번의 결과임. 계좌는 담지 않음.
type attestationVerdict struct {
	Stage     attestationStage
	DaysLeft  int // notice · imminent 에서만 의미 있음(남은 시간을 하루 단위로 올림)
	IssuedAt  time.Time
	ExpiresAt time.Time
	Missing   []string // scope_lost 의 누락 endpoint(정렬됨)
	// ProofAt 은 POST endpoint 마다의 가장 새 감독 증거 중 가장 이른 것임 — 갱신이 POST 를 유지하려면 이 증거가 유효 기간
	// 안이어야 하므로 운영자가 「언제 감독 검증을 다시 해야 하나」를 읽는 칸. a196: 판정 입력이기도 함 — restartDeadline 이 이 값 + (만료 − 발급) 으로 기한을 정함.
	ProofAt time.Time
	// Deadline 은 재시작이 거절되기 시작할 시각(만료와 감독 증거의 기한 중 이른 쪽), DeadlineKind 는 그 출처임(a196).
	// Verify 를 통과한 판정에서만 채워짐 — 거절 단계는 기한 없이 날짜로 셈.
	Deadline     time.Time
	DeadlineKind string
	// EpisodeAt 은 거절 에피소드의 시작 시각임(G12) — 판정이 아니라 감시(prepare)가 채움. 거절 단계 key 에만 쓰임.
	EpisodeAt time.Time
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

	// a196: 날수는 만료가 아니라 기한(만료와 감독 증거의 기한 중 이른 쪽)까지 셈. 6시간 갱신이 만료를 계속 밀어도 감독 증거는 늙음.
	v.Deadline, v.DeadlineKind = restartDeadline(att, v.ProofAt)
	remaining := v.Deadline.Sub(now)
	if remaining <= 0 {
		// 기한 정각은 이미 지난 것(발급자는 나이 == 유효 기간인 증거를 뺌). 만료는 Verify 가 걸렀으므로 여기 오는 것은 감독 증거의 기한뿐.
		// 날수 단계로 접으면 같은 키가 하루 넘게 이어져 침묵하므로 별도 단계로 하루 한 번 알림.
		v.Stage = stageProofOverdue
		return v
	}
	// 하루 단위 올림: 6.5일 남음 = D-7, 정확히 7일 = D-7, 7일 1초 = D-8. remaining 은 여기서 양수.
	v.DaysLeft = daysLeftCeil(remaining)
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

// restartDeadline 은 재시작이 거절되기 시작할 시각과 그 출처를 돌려줌(a196 D1).
//
// 감독 증거의 기한 = 실린 증거 중 가장 이른 것 + (만료 − 발급). 발급자는 같은 유효 기간 값으로 만료와 증거 수용을 정하므로(soak 시험이
// 고정) 파일 값만으로 계산하고 발급자 규칙은 복제하지 않음. 증거가 없거나 유효 기간이 0 이하면 만료만 셈.
// 기한은 「실린」 증거 기준 — 다른 시장의 더 새 증거가 있으면 실제 상실은 더 늦음(이른 경보 쪽). 예외: 기본값보다 긴 `--validity` 로 손수 발급한 파일은 기본값 갱신 타이머가 이 기한보다 먼저 증거를 뺌(늦은 경보 쪽).
func restartDeadline(att attest.Attestation, proofAt time.Time) (time.Time, string) {
	deadline, kind := att.ExpiresAt, deadlineExpiry
	if validity := att.ExpiresAt.Sub(att.IssuedAt); !proofAt.IsZero() && validity > 0 {
		if due := proofAt.Add(validity); due.Before(deadline) {
			deadline, kind = due, deadlineProof
		}
	}
	return deadline, kind
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

	identity string          // 마지막으로 본 기한(a196 — 파일의 발급 시각이 아님, 거절 단계는 기한 없음)
	sent     map[string]bool // 이 신원에서 이미 알린 단계 키

	episode   string    // 직전 판정의 거절 신원(단계 + 누락 집합 + 기한, 거절이 아니면 "")
	episodeAt time.Time // 그 거절 에피소드가 시작된 판정 시각 — 프로세스 메모리뿐이라 재시작하면 새로 시작함
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

	// 기한이 바뀌면(새 감독 증거 · 실린 증거 교체 · 만료가 기한인 파일의 갱신) 이미 알린 단계를 잊음. a196: 같은 감독 증거를 실은
	// 재발급은 기한이 같으므로 새로 세지 않음 — 파일 신원으로 세면 6시간 갱신마다 같은 통지가 다시 나감.
	identity := v.Deadline.UTC().Format(time.RFC3339Nano)
	if identity != w.identity || w.sent == nil {
		w.identity = identity
		w.sent = map[string]bool{}
	}
	// 거절 에피소드: 거절 신원이 직전 판정과 다르면(직전이 거절이 아니었던 경우 포함) 이번 판정 시각에 새 에피소드를 시작함.
	// 에피소드 시작이 key 에 들어가므로 회복 뒤 같은 날 재발은 새 outbox 행임 — 옛 행의 재알림 창(1시간)에 삼켜져 「보냄」으로
	// 적히고 그날 침묵하지 않게(gstack review G12). 같은 에피소드 안의 재발급은 같은 key 라 하루 한 번이 유지됨.
	episode := ""
	switch v.Stage {
	case stageHealthy, stageNotice, stageImminent:
	default:
		episode = string(v.Stage) + "|" + strings.Join(v.Missing, ",") + "|" + v.Deadline.UTC().Format(time.RFC3339Nano)
	}
	if episode != w.episode {
		w.episode, w.episodeAt = episode, now
	}
	v.EpisodeAt = w.episodeAt
	if v.Stage == stageHealthy {
		return obs.Event{}, "", false
	}
	key = stageKey(v, now)
	if w.sent[key] {
		return obs.Event{}, "", false
	}
	return w.eventFor(v, key), key, true
}

// stageKey 는 중복 방지 키이자 outbox 키임(a196 — 파일의 발급 신원을 담지 않음, 재발급이 새 행을 만들지 않게).
// 날수 단계는 단계 + 날수 + 기한(D-7 과 D-6 은 다른 키라 하루 한 번씩). 기한이 지난 상태는 단계 + 기한 + KST 날짜, 재시작이 거절될 상태
// (만료 · 범위 상실 · 계좌 불일치 · 무효)는 단계 (+ 범위 상실이면 누락 endpoint 집합) + KST 날짜 — 고칠 때까지 하루 한 번.
func stageKey(v attestationVerdict, now time.Time) string {
	// 거절 단계 · 지남은 KST 날짜 + 에피소드 시작까지 — 날짜가 바뀌면 다음 날 재알림, 에피소드가 바뀌면(회복 뒤 재발 · 재시작) 새 알림.
	day := now.In(kst).Format("2006-01-02") + ":" + v.EpisodeAt.UTC().Format(time.RFC3339)
	deadline := v.Deadline.UTC().Format(time.RFC3339Nano)
	switch v.Stage {
	case stageNotice, stageImminent:
		return fmt.Sprintf("%s:%d:%s", v.Stage, v.DaysLeft, deadline)
	case stageProofOverdue:
		return string(v.Stage) + ":" + deadline + ":" + day
	case stageScopeLost:
		// 누락 집합은 MissingEndpoints 가 정렬해 둠 — 같은 집합이면 같은 키.
		return string(v.Stage) + ":" + strings.Join(v.Missing, ",") + ":" + day
	}
	return string(v.Stage) + ":" + day
}

// mutationEndpoint 는 endpoint 가 주문 실행(POST)인지임 — 메서드 접두로만 판정(internal/soak 를 import 하지 않음).
func mutationEndpoint(e string) bool {
	return strings.HasPrefix(strings.ToUpper(strings.TrimSpace(e)), "POST ")
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
	deadline := v.Deadline.In(kst).Format("2006-01-02 15:04 KST")
	// a196: 재시작이 거절될 상태는 결과를 제목과 본문 첫 문장 둘 다에서 단정함 — 기기의 제목 표시를 저장소로 검증할 수 없어서.
	const refused = "지금 엔진을 재시작하면 기동이 거부된다"
	const consequence = "재시작이 거부되면 청산(손절 · 익절) 관측 루프가 돌지 않아 자동 손절이 멈춘다. 지금 실행 중인 엔진은 계속 돈다."
	// 날수 · 지남 통지는 「지금 도는 엔진의 손절」이 아니라 「재시작이 거부된 뒤」를 말함(구현 리뷰 P2-5).
	const afterRefusal = "재시작이 거부되면 그때부터 자동 손절이 돌지 않는다. 지금 실행 중인 엔진은 계속 돈다."
	// 감독 검증은 엔진과 같은 실행 잠금을 잡으므로 엔진을 멈춰야 돎 — 그 사실과 그동안 손절이 없다는 것을 숨기지 않음(gstack review G10).
	const reverifySteps = "감독하 실주문 검증을 다시 해야 한다. 이 검증은 엔진과 같은 잠금을 써서 엔진을 멈춰야 실행되고, 멈춘 동안에는 " +
		"자동 손절이 돌지 않는다. 보유 종목의 시장이 닫힌 시간에 엔진을 멈추고, 터미널에서 `tossctl verify run`(주문을 받는 시장을 " +
		"`--market` 으로) → `tossctl soak attest` → 엔진 재시작 순서로 하라."
	const reverify = "할 일: " + reverifySteps
	const renew = "할 일: 갱신(`tossctl soak attest`)이 왜 실패하는지 확인한다 — 갱신에는 soak 통과와 유효 기간 안의 감독하 실주문 검증(POST 주문 · 취소 증거)이 함께 필요하다."

	var title, body string
	switch v.Stage {
	case stageNotice, stageImminent:
		if v.DeadlineKind == deadlineProof {
			title = fmt.Sprintf("감독하 실주문 검증 만료 D-%d — 그 뒤 엔진 재시작이 거부될 수 있다", v.DaysLeft)
			body = fmt.Sprintf("감독하 실주문 검증(POST 주문 · 취소 증거)이 D-%d(%s)에 만료된다. 다른 유효한 감독 증거가 없으면 그 뒤 첫 갱신부터 "+
				"엔진 재시작이 거부된다. %s attestation 자체의 만료는 %s 이다. 할 일: 기한 전에 %s", v.DaysLeft, deadline, afterRefusal, expires, reverifySteps)
		} else {
			title = fmt.Sprintf("자동매매 검증서(attestation) 만료 D-%d", v.DaysLeft)
			body = fmt.Sprintf("capability attestation 이 D-%d(%s)에 만료된다. 그 뒤 엔진 재시작이 거부된다. %s %s",
				v.DaysLeft, expires, afterRefusal, renew)
		}
	case stageProofOverdue:
		title = "감독하 실주문 검증의 기한이 지났다 — 지금 감독 검증을 다시 해야 한다"
		body = fmt.Sprintf("감독하 실주문 검증의 기한이 지났다(%s). 다른 유효한 감독 증거가 없으면 다음 갱신부터 엔진 재시작이 거부된다 — "+
			"갱신이 실패하고 있으면 이 파일의 만료(%s)에 거부된다. %s %s", deadline, expires, afterRefusal, reverify)
	case stageExpired:
		title = refused + " — 자동매매 검증서(attestation)가 만료되었다"
		body = fmt.Sprintf("%s. capability attestation 이 %s 에 만료되었다. %s %s", refused, expires, consequence, renew)
	case stageScopeLost:
		// 원인 · 할 일은 빠진 endpoint 의 종류로 가름(구현 리뷰 P1-2) — 감독하 실주문 검증이 덮는 것은 주문 실행(POST)뿐임.
		missing := strings.Join(v.Missing, ", ")
		mutations := 0
		for _, e := range v.Missing {
			if mutationEndpoint(e) {
				mutations++
			}
		}
		if mutations == len(v.Missing) {
			title = refused + " — 주문 실행 검증이 attestation 에 없다"
			body = fmt.Sprintf("%s. 현재 attestation 에 엔진이 쓰는 endpoint 가 빠져 있다: %s. 갱신이 유효 기간 안의 감독하 실주문 검증 증거 없이 쓰였다. %s %s",
				refused, missing, consequence, reverify)
		} else {
			title = refused + " — 엔진이 쓰는 endpoint 가 attestation 에 없다"
			remedy := "할 일: 빠진 endpoint 를 덮는 능력 검증을 다시 실행해 attestation 을 재발급하라."
			if mutations > 0 {
				remedy += " 주문 실행 endpoint 는 감독하 실주문 검증으로만 덮인다 — " + reverifySteps
			}
			body = fmt.Sprintf("%s. 현재 attestation 에 엔진이 쓰는 endpoint 가 빠져 있다: %s. %s %s", refused, missing, consequence, remedy)
		}
	case stageIncomplete:
		title = refused + " — 자동매매 검증서(attestation)에 필수 칸이 비어 있다"
		body = refused + ". 엔진이 읽는 attestation 에 계좌 · 만료 · 검증된 endpoint 중 비어 있는 칸이 있다. " + consequence + " " + renew
	case stageAccountMismatch:
		title = refused + " — 자동매매 검증서(attestation)가 엔진의 계좌와 다르다"
		body = refused + ". 현재 attestation 이 엔진 자격 증명의 계좌와 다른 계좌를 가리킨다. " + consequence +
			" 할 일: 엔진 자격 증명의 계좌로 attestation 을 다시 발급한다."
	default:
		title = refused + " — 자동매매 검증서(attestation)를 쓸 수 없다"
		body = refused + ". 엔진이 읽는 attestation 파일이 없거나 읽을 수 없다. " + consequence + " " + renew
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
	if !v.Deadline.IsZero() {
		fields["deadline_at"] = v.Deadline.UTC().Format(time.RFC3339)
		fields["deadline_kind"] = v.DeadlineKind
	}

	return obs.Event{
		Type: kind,
		// outbox 키 = 단계 키(a196 — 발급 신원 없음). 재발급 · 재시작이 같은 상태를 같은 행으로 읽게 하고, 단계 · 날짜 · 기한이 다르면 다른 행.
		Key:    "attestation-watch:" + key,
		Title:  title,
		Body:   body,
		Fields: fields,
	}
}
