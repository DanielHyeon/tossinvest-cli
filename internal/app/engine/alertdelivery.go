package engine

// alertdelivery.go is a098's answer to "nobody sends what the outbox keeps".
//
// A critical alert reaches the ledger through EnqueueAlert and then waits. The
// synchronous notify path only sends the alert it is holding, and the one thing
// that drains a backlog — Notifier.Flush — has no production caller at all. So a
// row written while the transport was down, or written by a sender that then
// died, stayed PENDING until the same condition happened to be observed again.
// That is the defect. This loop is the subject the sentence was missing.
//
// # Why it does not call Flush
//
// Flush takes the notifier's mutex and holds it across every publish in the
// backlog (notifier.go:734-735, and PendingAlerts(ctx, 0) at :737 asks for all of
// them). The synchronous stop-alert path takes the same mutex. A loop built on
// Flush would therefore make a stop wait for N network round trips with no bound
// on N — the very defect a092 exists to remove, rebuilt here. design D1.1
// rejected that shape; this loop has its own delivery path and takes no mutex at
// all over a publish.
//
// # What separates it from the synchronous sender, then
//
// The ledger claim a099 built. Every row is claimed before it is published and
// settled under that claim's token, so two senders cannot both be sending one
// row. That is why a098 cannot ship without a099 (deploy-pair.txt).
//
// # The settlement rules are a099's, not this file's
//
// They are asymmetric and the asymmetry is the point:
//
//	published + settled     MarkAlertDelivered releases the lease as part of
//	                        settling. Releasing again would be a second write
//	                        saying nothing.
//	published + not settled the lease stays held. Releasing here would let the
//	                        next cycle send a row that already went out — the
//	                        2026-08-08 storm, arriving through the success path.
//	not published           record the attempt, then hand the lease back: this
//	                        row's turn is over and holding it until expiry keeps
//	                        it out of the next cycle for no reason.
//
// # 지속 실패의 주인 (a124)
//
// a092 가 동기 시도를 엔진 루프에서 빼면 「전달 실패 지속 → 신규 진입 차단 · 운영 모드 승격」을 할 주체가 이 실행자뿐임.
// 판정은 세 가지이고 전부 judge 를 거침(design D1 · D7 · D8):
//
//	시도 한도      실패 기록이 커밋한 attempts ≥ alertAttemptLimit → 차단 + 승격 (반납 뒤 판정)
//	전달 기록 실패  발행은 됐는데 정산이 안 됨 → 임차를 쥔 채 즉시 차단 + 승격
//	기록 · 나열 실패 원장에 시도를 못 남긴 연속(행별 · 실행자 단위)이 한도 → 차단 + 승격
//
// 늦은 적용은 원칙 E(해제 세대)를 따르고, 실행자가 잡는 잠금은 진입 게이트 잠금뿐임. 진입을 실제로 막는 것은 게이트
// 래치이고 모드 행은 원장 기록임(design D10 — 투영기 미배선).

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// alertDeliveryInterval is how long the executor waits between cycles.
//
// It is measured rather than borrowed (a098 task 3.2). The measurement gives a
// band, not a number: an empty cycle costs 0.065ms of ledger read, so holding the
// steady-state duty cycle under 1/10,000 puts the floor at ~0.65s; keeping the
// wait from adding more than 5% to the 81s lease a dead sender's row sits behind
// puts the ceiling at ~4.05s. Two seconds sits near the middle of that band.
//
// The floor and the 5% are chosen; 0.065ms and 81s are measured. Anyone moving
// this should move the measurement first — `go test ./internal/app/engine/ -run
// '^$' -bench A098` reproduces it.
const alertDeliveryInterval = 2 * time.Second

// alertDeliveryBatch is how many rows one cycle takes.
//
// Reading is not what bounds it: 100 rows read in 0.449ms while settling a single
// row costs 5.584ms, two orders of magnitude apart. Ten rows is 56ms of ledger
// per cycle, 2.8% of the interval. A hundred would be 28%.
//
// The cost that is not on that scale is the network. With every publish timing
// out, a cycle of ten takes about 100s — which is why the cycle checks for
// cancellation between rows rather than only between cycles.
const alertDeliveryBatch = 10

// alertReleaseTimeout bounds handing a lease back when the cycle's own context is
// already done.
const alertReleaseTimeout = 5 * time.Second

// alertAttemptLimit 는 전달 실패를 「지속」으로 보는 시도 수임 (a124 design D2 — 동기 경로의 기본 예산을 그대로 인용).
// 숫자를 복사하지 않음: 동기 경로와 같은 실패 양상에서 같은 시간에 잠그는 것이 이 값의 근거라서, 한쪽만
// 바뀌면 그 등가가 조용히 깨짐. 이 값을 **늘리는** 것은 진입 차단 완화이므로 사람 결정 대상임.
const alertAttemptLimit = obs.DefaultCriticalAttempts

// alertNoPublisherCause 는 전송 수단이 없을 때 원장에 남기는 고정 원인 문구임 (D3 — 없는 전송 수단도 실패 시도로 셈).
const alertNoPublisherCause = "no publisher is configured"

// alertJudgementLogMarker 는 이 change 가 새로 쓰는 판정 줄의 scope 값임.
// 새 줄은 허용 목록 필드만 씀(사건 종류 · 고정 detail · alert_id · 사유 · 수) — 정제 시험(2.11)이 이 표식으로 새 줄을 가려냄.
const alertJudgementLogMarker = "alert_delivery_judgement"

// 게이트 detail 은 고정 문구 + 수뿐임 (불변식 8, design D9). 행의 제목 · 본문 · payload · 원문 오류 · 계좌 · 토큰을
// 넣지 않음 — 게이트 상태는 그것을 읽는 모든 곳에 노출됨.
const (
	alertLatchAttemptLimit     = "a critical alert reached the delivery attempt limit (%d) without being delivered; an operator has to acknowledge the backlog"
	alertLatchUnrecorded       = "a critical alert was sent but its delivery could not be recorded; an operator has to acknowledge the backlog"
	alertLatchUnaccounted      = "a failed delivery attempt of a critical alert could not be matched to its row; an operator has to acknowledge the backlog"
	alertLatchRecordRun        = "the delivery attempts of a critical alert could not be recorded %d times in a row; an operator has to acknowledge the backlog"
	alertLatchListRun          = "the critical alert backlog could not be read %d cycles in a row; an operator has to acknowledge the backlog"
	alertLatchEscalationFailed = "an undelivered critical alert could not tighten the operating mode; an operator has to acknowledge the backlog"
)

// alertLedger 는 실행자가 원장에서 쓰는 것 전부임. 생산에서는 *journal.Journal 그대로이고, 시험이 원장 결함을
// 주입할 때만 감쌈(alertDeliverer.ledger).
type alertLedger interface {
	PendingAlertsForDelivery(ctx context.Context, limit, attemptLimit int) ([]journal.Alert, error)
	ClaimAlertByID(ctx context.Context, id int64, claimant string) (journal.ClaimResult, error)
	MarkAlertAttemptFailed(ctx context.Context, id int64, token, cause string) (journal.SettleResult, error)
	MarkAlertDelivered(ctx context.Context, id int64, token string) (journal.SettleResult, error)
	ReleaseAlertClaim(ctx context.Context, id int64, token string) (journal.SettleResult, error)
	EscalateOperatingMode(ctx context.Context, accountRef, trigger string,
		announcer journal.ModeAnnouncer) (journal.OperatingModeRecord, bool, error)
}

// alertJudgeStage 는 판정 한 건이 지나는 단계임. 시험 훅(judgeHook)이 해제를 정확한 자리에 끼워 원칙 E 의
// 인터리빙을 결정적으로 재기 위해 있음 — 생산에서는 훅이 nil 이라 아무 일도 없음.
type alertJudgeStage int

const (
	// alertStageSettled: 정산이 돌아온 뒤(실패 기록 경로는 반납까지 끝난 뒤), 해제 세대를 읽기 전.
	alertStageSettled alertJudgeStage = iota
	// alertStageEpochRead: 해제 세대를 읽은 뒤, 조건부 잠금 전.
	alertStageEpochRead
	// alertStageLatched: 조건부 잠금 뒤, 승격 전.
	alertStageLatched
)

// alertNoRow 는 행이 없는 판정(미전달 나열 실패)의 alert id 자리임. 새 줄은 이 값이면 alert_id 칸을 쓰지 않음 —
// 0 을 실제 행 id 로 읽지 않게.
const alertNoRow int64 = 0

// alertFailureRun 은 원장에 쓰지 못한 **연속** 횟수와, 직전 증가 직후 읽은 해제 세대임 (D8, AA2).
type alertFailureRun struct {
	count int
	epoch uint64
}

// alertDeliverer drains the critical-alert outbox.
type alertDeliverer struct {
	Journal   *journal.Journal
	Publisher obs.Publisher
	Log       *obs.Logger
	Clock     clock.Clock
	Interval  time.Duration
	Batch     int
	Claimant  string

	// heldReported is which rows this executor has already said are held, keyed
	// by the lease it saw. It is written only from Run's own goroutine (cycle →
	// deliverOne), so it carries no lock.
	//
	// It exists because "held" is worth saying once and worthless said forty
	// times: a stuck row is re-listed every cycle, and at a 2s interval under an
	// 81s lease one abandoned row would write ~40 identical lines before the
	// lease even expires. That is not observation, it is a log storm that buries
	// the line an operator needs (a098 R21).
	heldReported map[int64]time.Time

	// Gate 는 전달 실패 사유(ReasonAlertUndelivered)를 거는 진입 게이트임 (a124). 실행자가 잡는 잠금은 이
	// 게이트의 잠금뿐이고 그 안에서는 map 연산만 함 — Notifier 의 뮤텍스는 잡지 않음(design D7, Q1).
	// nil 이면 잠그지 않음(시험 조립 전용; 생산 조립 Context.AlertDeliverer 는 항상 넘김).
	Gate *execgw.EntryGate
	// AccountRef 는 운영 모드를 승격할 계정임. 비어 있으면 승격하지 않음 — Notifier.escalate 와 같은 규칙.
	AccountRef string

	// ledger 는 시험이 원장 결함을 주입할 때만 채움. nil 이면 Journal 을 씀.
	ledger alertLedger
	// judgeHook 은 판정 단계마다 불리는 시험 훅임. 생산에서는 nil.
	judgeHook func(stage alertJudgeStage, alertID int64)

	// recordRuns 는 행별 연속 기록 실패(D8)이고, listRun 은 나열 실패의 연속임. 둘 다 Run 의 goroutine
	// 에서만 쓰므로 잠금이 없음(heldReported 와 같은 이유). 재시작이 비우지만 기동 복원이 PENDING 으로 다시 잠금.
	recordRuns map[int64]alertFailureRun
	listRun    alertFailureRun
}

// Run cycles until ctx ends.
//
// It cycles first and waits afterwards, so an engine starting with a backlog
// begins draining it immediately rather than one interval later.
//
// A cancelled context is a clean stop and is reported *as a cancellation*:
// ctx.Err(), never nil.
//
// ⛔ The first draft returned nil here, reasoning that "a clean stop is not a
// failure". Under the runtime that reads this, the reasoning inverts: the
// runtime judges a stop with Runtime.gracefulStop, which requires the error to
// be the context's (runtime.go). A nil return has no context in it, so every
// ordinary shutdown would be classified as the sender dying — and that is the
// one thing that latches the entry gate. The judgement is deliberately the same
// predicate the supervised loops get (design D8.3), so the convention is the
// return's to match, not the predicate's.
//
// A cycle that fails does not stop the loop. The backlog is exactly what should
// outlive a transient ledger or transport fault, and a loop that exits on the
// first error would leave nobody sending again.
func (d *alertDeliverer) Run(ctx context.Context) error {
	for {
		if err := d.cycle(ctx); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			d.logf(obs.EventAlertUndelivered, err, "a delivery cycle failed")
		}
		// Both clocks return ctx.Err() when the wait is cancelled (clock.go,
		// fake.go), so this is the cancellation travelling out unchanged rather
		// than a second judgement about what happened.
		if err := d.Clock.Sleep(ctx, d.interval()); err != nil {
			return err
		}
	}
}

// cycle claims, publishes and settles one batch.
//
// It returns an error only for a failure that made the whole cycle impossible —
// reading the backlog. One row failing is that row's problem, not the backlog's:
// returning early there is how a flush once abandoned every remaining critical
// alert because the first row had settled underneath it.
func (d *alertDeliverer) cycle(ctx context.Context) error {
	// An episode nobody can still be in is not worth remembering. Dropping the
	// lapsed ones here is what keeps the map bounded by *currently contended*
	// rows rather than by every row this engine has ever seen contended.
	d.forgetLapsedHeld(d.Clock.Now())
	// 한도 아래 행을 먼저 고름 — 한도에 이른 행 열이 배치를 채워 새 critical 행이 첫 시도조차 못 하는
	// 굶주림을 막음(a124 R2). 한도 행은 버리지 않고 잔여 자리에만 들어감.
	pending, err := d.led().PendingAlertsForDelivery(ctx, d.batch(), alertAttemptLimit)
	if err != nil {
		// 나열 실패도 지속 실패임(D8) — 행이 없으므로 실행자 단위 계수 하나로 셈.
		d.countListFailure(ctx)
		return fmt.Errorf("listing the critical alert backlog: %w", err)
	}
	d.listRun = alertFailureRun{}
	if len(pending) < d.batch() {
		// 잘리지 않은(완전한) 나열에 없는 행은 PENDING 을 떠났음 — 그 행의 기록 실패 계수를 지움.
		// 잘린 나열에 없다는 것은 그 증거가 아니므로 거르지 않음(Q5).
		d.pruneRecordRuns(pending)
	}
	for _, alert := range pending {
		// Between rows, not only between cycles: with a dead transport this loop
		// spends a publish timeout per row, and a batch of them is far longer
		// than the interval. Run's wait being cancellable is not enough on its own.
		if ctx.Err() != nil {
			return nil
		}
		d.deliverOne(ctx, alert)
	}
	return nil
}

// deliverOne is one row's whole life in this cycle.
//
// a124: 이 행의 시도가 원장에 남긴 결과로 지속 실패를 판정함 — 판정 입력은 정산이 커밋한 attempts 뿐이고
// (D1), 늦은 적용은 원칙 E(judge)를 따름. 원격 전송 · 원장 연산 동안 어떤 잠금도 쥐지 않음.
func (d *alertDeliverer) deliverOne(ctx context.Context, alert journal.Alert) {
	claim, err := d.led().ClaimAlertByID(ctx, alert.ID, d.claimant())
	if err != nil {
		d.logf(obs.EventAlertUndelivered, err, "claiming an alert failed", "alert_id", alert.ID)
		// 임차 요청 실패도 이 행의 시도를 원장에 남기지 못한 것임(D8).
		d.countRecordFailure(ctx, alert.ID)
		return
	}
	switch claim.Disposition {
	case journal.ClaimAcquired:
		// The row is ours now, so whatever episode of contention it was in is
		// over. Forgetting here is what makes the next one reportable.
		d.forgetHeld(alert.ID)
	case journal.ClaimHeldElsewhere:
		// Another sender holds a live lease. Skipping is right — but not
		// silently: a row that is held every cycle is not contention, it is a
		// sender that stopped without settling, and the ledger only reopens it
		// when the lease expires.
		//
		// Once per lease, though. Saying it every cycle would report the same
		// fact ~40 times before that lease expires, and forty copies of one line
		// is how the *next* line stops being read (R21).
		if d.reportHeld(alert.ID, claim.ExpiresAt) {
			d.logf(obs.EventAlertClaimHeld, nil, "an alert is held by another sender",
				"alert_id", alert.ID, "claimed_by", claim.ClaimedBy,
				"claim_expires_at", claim.ExpiresAt.UTC().Format(time.RFC3339))
		}
		return
	default:
		// Settled between the listing and here. Normal.
		d.forgetHeld(alert.ID)
		// PENDING 을 떠난 것이 관측됨 — 이 행의 기록 실패 계수도 끝(D8).
		d.forgetRecordRun(alert.ID)
		return
	}
	if claim.Stole {
		// A steal is a designed recovery and never silent: somebody died or
		// stalled while holding this row.
		// claim_stolen 으로 적음(a092 25라운드 보이스 A #2) — claim_held 는 알림기에서 정상 경합(INFO)이라, 죽은 발송자 신호가
		// 같은 이름을 쓰면 경보 규칙을 걸 수 없음. 두 발송 경로의 인수 신호가 이 이름 하나.
		d.logf(obs.EventAlertClaimStolen, nil, "an expired alert lease was taken over",
			"alert_id", alert.ID, "stole_from", claim.StoleBy)
	}

	var perr error
	cause := ""
	if d.Publisher == nil {
		// No transport at all. The row stays PENDING and the lease goes back;
		// saying so once per cycle is what keeps a silently unconfigured engine
		// from looking like a working one.
		//
		// a124 D3: 없는 전송 수단도 실패 시도로 셈 — 세지 않으면 무설정 엔진에 전달 실패 래치가 영구히
		// 서지 않음(newNotifier 문서의 「specified direction」). 원인은 고정 문구.
		d.logf(obs.EventAlertUndelivered, nil, "no publisher is configured, so nothing can be sent",
			"alert_id", alert.ID)
		perr, cause = errors.New(alertNoPublisherCause), alertNoPublisherCause
	} else {
		perr = d.Publisher.Publish(ctx, obs.Notification{
			Type:     obs.EventType(alert.Type),
			Severity: obs.SeverityCritical,
			Title:    alert.Title,
			Body:     alert.Body,
		})
		if perr != nil {
			cause = perr.Error()
		}
	}
	if perr != nil {
		d.recordFailedAttempt(ctx, alert.ID, claim.Token, cause)
		return
	}
	d.recordDelivery(ctx, alert.ID, claim.Token)
}

// recordFailedAttempt 는 실패한 시도를 기록하고, 임차를 돌려준 **뒤에** 판정함.
//
// 반납이 먼저인 이유(Y5): 판정은 게이트 잠금을 기다릴 수 있는데, 그동안 임차를 쥐고 있으면 이 행이 다음
// 사이클에서 빠질 이유가 없는데도 빠짐. 판정 표는 design D1 「실패 기록」.
func (d *alertDeliverer) recordFailedAttempt(ctx context.Context, id int64, token, cause string) {
	res, err := d.led().MarkAlertAttemptFailed(ctx, id, token, cause)
	if err != nil {
		d.logf(obs.EventAlertUndelivered, err, "recording a failed attempt failed", "alert_id", id)
	}
	d.release(ctx, id, token)
	d.hook(alertStageSettled, id)
	if err != nil {
		// 오류의 SettleResult{} 는 Outcome 영값이 Applied 라 결과를 읽으면 안 됨(F10) — err 가 먼저.
		// attempts 가 안 올랐으니 한도 판정이 설 수 없음 → 행별 연속 기록 실패로 셈(D8).
		d.countRecordFailure(ctx, id)
		return
	}
	switch res.Outcome {
	case journal.SettleApplied:
		// 원장이 이 행에 썼음 — 기록 실패의 연속은 끝.
		d.forgetRecordRun(id)
		if res.Attempts < alertAttemptLimit {
			return
		}
		d.judge(ctx, id, d.readEpoch(id), true, fmt.Sprintf(alertLatchAttemptLimit, alertAttemptLimit))
	case journal.SettleAlreadySettled, journal.SettleLeaseLost:
		// 승인(또는 남의 발송)이 먼저였음 — 잠그지도 승격하지도 않음. 0 행을 썼으므로 계수도 그대로(R3).
	default:
		// 행을 못 찾음 · 모르는 결과 → 잠금만(동기 경로 parity: lost → 승격 없음, N6). 모르는 값은
		// 동기 경로보다 한 칸 보수적(fail-safe).
		d.judge(ctx, id, d.readEpoch(id), false, alertLatchUnaccounted)
	}
}

// recordDelivery 는 발행이 성공한 행을 정산함. 정산이 안 되면 임차를 쥔 채 **즉시** 잠그고 승격함(D1 전달
// 정산 표, 동기 경로 parity) — 반납하면 다음 사이클이 이미 나간 알림을 곧장 다시 보냄(2026-08-08 폭풍).
func (d *alertDeliverer) recordDelivery(ctx context.Context, id int64, token string) {
	settled, err := d.led().MarkAlertDelivered(ctx, id, token)
	d.hook(alertStageSettled, id)
	if err != nil {
		// Published but not settled, and the write failed rather than being
		// refused. The lease stays: see the file comment.
		d.logf(obs.EventAlertUndelivered, err, "settling a delivered alert failed", "alert_id", id)
		// 취소 중이어도 판정함 — 이 오류가 취소 때문인지 원장 결함 때문인지 여기서 가를 수 없고, 그 사이 승인 ·
		// 남의 발송이 행을 PENDING 에서 뺄 수 있어 「기동 복원이 덮는다」도 보장이 아님(codex 구현 1회차 I1).
		// 취소된 ctx 의 승격 쓰기는 실패하고, 그러면 judge 가 무조건 차단으로 대신함(보수 방향).
		d.judge(ctx, id, d.readEpoch(id), true, alertLatchUnrecorded)
		return
	}
	if settled.Outcome == journal.SettleApplied {
		d.forgetRecordRun(id)
		return
	}
	switch settled.Outcome {
	case journal.SettleAlreadySettled, journal.SettleLeaseLost:
		// Published, but somebody else's row to settle. It is out; the holder
		// records it. Still no release. (승인 · 남의 임차가 먼저 — 게이트 사건 아님, 동기 경로 B8)
		d.logf(obs.EventAlertUndelivered, nil, "a delivered alert was settled by somebody else",
			"alert_id", id, "outcome", settled.Outcome.String())
	default:
		// 발행은 됐는데 행이 없거나 모르는 결과 — 정산되지 않은 것(동기 경로 B9·B10 → 즉시 잠금 + 승격).
		d.logf(obs.EventAlertUndelivered, nil, "a delivered alert could not be matched to its row",
			"alert_id", id, "outcome", settled.Outcome.String())
		d.judge(ctx, id, d.readEpoch(id), true, alertLatchUnrecorded)
	}
}

// readEpoch 는 판정 근거가 확정된 **직후** 해제 세대를 한 번 읽음(원칙 E). 그 뒤의 해제는 「근거 뒤」로
// 판정을 버리게 하고, 근거와 이 읽기 사이의 해제는 「앞」으로 보여 다시 잠금 — 틀리는 방향은 잠그는 쪽뿐(AB2).
func (d *alertDeliverer) readEpoch(id int64) uint64 {
	var epoch uint64
	if d.Gate != nil {
		epoch = d.Gate.ClearEpoch(execgw.ReasonAlertUndelivered)
	}
	d.hook(alertStageEpochRead, id)
	return epoch
}

// judge 는 판정 하나를 원칙 E 로 적용함 — 「늦은 적용은 제때 적용과 같아야 한다」 (design D7 판정 → 행동 표).
//
// 제때 적용은 차단과 승격을 함께 하고 그 뒤의 운영자 해제는 차단만 지움(모드는 사람 완화로만 풀림). 그래서
// epoch 뒤로 해제가 있었으면 차단은 버리고(제때 선 차단도 그 해제가 지웠음) 승격은 적용함. 승격 쓰기가
// 실패하면 조건부 차단 결과와 무관하게 무조건 차단함(Z2 · AA1 — 둘 다 잃지 않게).
//
// 잡는 잠금은 게이트 잠금 하나이고 그 안에서는 map 연산뿐임. 승격 · 로그는 잠금 밖.
func (d *alertDeliverer) judge(ctx context.Context, id int64, epoch uint64, escalate bool, detail string) {
	if d.Gate != nil {
		if _, inserted := d.Gate.BlockUnlessClearedSince(execgw.ReasonAlertUndelivered, epoch, detail); inserted {
			d.reportLatch(id, detail)
		}
	}
	d.hook(alertStageLatched, id)
	if !escalate {
		return
	}
	if err := d.escalate(ctx, id); err != nil && d.Gate != nil {
		d.Gate.Block(execgw.ReasonAlertUndelivered, alertLatchEscalationFailed)
	}
}

// escalate 는 운영 모드를 ENTRY_BLOCKED 로 승격함 — Notifier.escalate 와 같은 호출, announcer 는 nil
// (전달 수단이 막 실패한 참이라 알림을 다시 낼 이유가 없음).
//
// 새 줄에는 계좌 · 원문 오류를 넣지 않음(R2): 모드 전이 오류는 계좌를 담을 수 있으므로 err 대신 고정 분류만.
// 이 모드 행은 원장 기록이며, 생산에서 진입을 막는 것은 위의 게이트 래치임(design D10 — 투영기 미배선).
func (d *alertDeliverer) escalate(ctx context.Context, id int64) error {
	if strings.TrimSpace(d.AccountRef) == "" {
		return nil
	}
	_, changed, err := d.led().EscalateOperatingMode(ctx, d.AccountRef,
		journal.ModeTriggerCriticalAlertUndelivered, nil)
	if err != nil {
		if d.Log != nil {
			class := "ledger"
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				class = "context"
			}
			d.Log.Error(obs.EventOperatingMode, nil, withAlertID([]any{
				obs.FieldScope, alertJudgementLogMarker,
				obs.FieldReason, "escalation_failed:" + class,
				obs.FieldDetail, "an undelivered critical alert could not tighten the operating mode; the entry latch stands instead",
			}, id)...)
		}
		return err
	}
	if changed && d.Log != nil {
		d.Log.Warn(obs.EventOperatingMode, withAlertID([]any{
			obs.FieldScope, alertJudgementLogMarker,
			obs.FieldToState, journal.ModeEntryBlocked,
			obs.FieldReason, journal.ModeTriggerCriticalAlertUndelivered,
			obs.FieldDetail, "an undelivered critical alert tightened the operating mode",
		}, id)...)
	}
	return nil
}

// reportLatch 는 이 실행자가 래치를 **새로** 세웠을 때만 한 줄을 씀 — 이미 서 있던 래치(동기 경로 · 기동 복원)의
// 재확인은 조용함(design D9, Eng 리뷰 F4 · codex 2회차).
func (d *alertDeliverer) reportLatch(id int64, detail string) {
	if d.Log == nil {
		return
	}
	d.Log.Warn(obs.EventAlertUndelivered, withAlertID([]any{
		obs.FieldScope, alertJudgementLogMarker,
		obs.FieldReason, string(execgw.ReasonAlertUndelivered),
		obs.FieldDetail, detail,
	}, id)...)
}

// withAlertID 는 행이 있는 판정에만 alert_id 칸을 붙임.
func withAlertID(fields []any, id int64) []any {
	if id == alertNoRow {
		return fields
	}
	return append(fields, "alert_id", id)
}

// countRecordFailure 는 한 행의 원장 쓰기 실패 하나를 셈(D8 「오류 하나의 전이」).
//
// 기록되지 않는 시도는 attempts 를 못 올리므로, 이 규칙이 없으면 원장 결함이 차단의 면제가 됨. 엔진 종료로
// 인한 취소는 세지 않음(취소는 원장 결함이 아님 — transport timeout 은 ntfy 의 자식 ctx 라 발행 실패로 옴).
func (d *alertDeliverer) countRecordFailure(ctx context.Context, id int64) {
	if ctx.Err() != nil {
		return
	}
	epoch := d.readEpoch(id)
	if d.recordRuns == nil {
		d.recordRuns = make(map[int64]alertFailureRun)
	}
	next, judged, blockEpoch := advanceFailureRun(d.recordRuns[id], epoch)
	if !judged {
		d.recordRuns[id] = next
		return
	}
	delete(d.recordRuns, id)
	d.judge(ctx, id, blockEpoch, true, fmt.Sprintf(alertLatchRecordRun, alertAttemptLimit))
}

// countListFailure 는 미전달 나열 실패 하나를 셈 — 행이 없으므로 실행자 단위 계수 하나(D8).
func (d *alertDeliverer) countListFailure(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	epoch := d.readEpoch(alertNoRow)
	next, judged, blockEpoch := advanceFailureRun(d.listRun, epoch)
	if !judged {
		d.listRun = next
		return
	}
	d.listRun = alertFailureRun{}
	d.judge(ctx, alertNoRow, blockEpoch, true, fmt.Sprintf(alertLatchListRun, alertAttemptLimit))
}

// advanceFailureRun 은 오류 하나가 연속을 어떻게 바꾸는지 정함 (D8 AA2 — 위에서 아래 우선).
//
//  1. 이번 오류가 한도째 → 판정(리셋 판단보다 **먼저**). 차단은 직전 증가 직후의 세대로 조건부 — 해제가
//     직전 증가와 한도째 사이였든(연속이 끊겼어야 함) 한도째 뒤였든(제때 선 차단도 지워졌음) 차단 쪽 답은
//     「없음」으로 같음. 승격 쪽은 모호하므로 보수(승격)로 기움(Z1).
//  2. 아니고 직전 증가 뒤 해제가 있었음 → 연속이 끊김, 1 부터 다시.
//  3. 그 밖 → +1.
func advanceFailureRun(run alertFailureRun, epoch uint64) (next alertFailureRun, judged bool, blockEpoch uint64) {
	// count 0 = 연속 없음(저장되는 연속은 늘 count ≥ 1 이고 판정 뒤에는 영값으로 돌아감).
	if run.count == 0 {
		if alertAttemptLimit <= 1 {
			return alertFailureRun{}, true, epoch
		}
		return alertFailureRun{count: 1, epoch: epoch}, false, 0
	}
	if run.count+1 >= alertAttemptLimit {
		return alertFailureRun{}, true, run.epoch
	}
	if epoch != run.epoch {
		return alertFailureRun{count: 1, epoch: epoch}, false, 0
	}
	return alertFailureRun{count: run.count + 1, epoch: epoch}, false, 0
}

// forgetRecordRun 은 원장이 이 행에 정산을 **적용**했거나 행이 PENDING 을 떠난 것이 관측됐을 때 부름.
// 반납의 Applied · LeaseLost · AlreadySettled 는 정산 쓰기가 아니므로 부르지 않음(R3 · X3).
func (d *alertDeliverer) forgetRecordRun(id int64) {
	delete(d.recordRuns, id)
}

// pruneRecordRuns 는 완전한 나열에 없는 행의 계수를 지움 — PENDING 전체를 봤고 그 행은 없음.
func (d *alertDeliverer) pruneRecordRuns(pending []journal.Alert) {
	if len(d.recordRuns) == 0 {
		return
	}
	present := make(map[int64]struct{}, len(pending))
	for _, a := range pending {
		present[a.ID] = struct{}{}
	}
	for id := range d.recordRuns {
		if _, ok := present[id]; !ok {
			delete(d.recordRuns, id)
		}
	}
}

func (d *alertDeliverer) hook(stage alertJudgeStage, id int64) {
	if d.judgeHook != nil {
		d.judgeHook(stage, id)
	}
}

// led 는 실행자가 쓰는 원장임 — 생산에서는 Journal.
func (d *alertDeliverer) led() alertLedger {
	if d.ledger != nil {
		return d.ledger
	}
	return d.Journal
}

// release hands a lease back for a row this cycle is done with.
//
// The context is detached because a cycle cancelled partway must not leave the
// rows it claimed locked behind it: the ledger only reopens them when the lease
// expires, and until then nobody sends them.
func (d *alertDeliverer) release(ctx context.Context, id int64, token string) {
	relCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), alertReleaseTimeout)
	defer cancel()
	if _, err := d.led().ReleaseAlertClaim(relCtx, id, token); err != nil {
		d.logf(obs.EventAlertUndelivered, err, "releasing an alert lease failed", "alert_id", id)
	}
}

// reportHeld says whether this held row is worth a line, and remembers that it
// answered yes.
//
// The episode is keyed by the *lease*, not by the row. A second sender taking
// the row after the first one's lease lapsed is a new fact about a different
// sender, and collapsing the two would hide the handover.
func (d *alertDeliverer) reportHeld(id int64, expires time.Time) bool {
	if d.heldReported == nil {
		d.heldReported = make(map[int64]time.Time)
	}
	if seen, ok := d.heldReported[id]; ok && seen.Equal(expires) {
		return false
	}
	d.heldReported[id] = expires
	return true
}

func (d *alertDeliverer) forgetHeld(id int64) {
	delete(d.heldReported, id)
}

// forgetLapsedHeld drops episodes whose lease has already run out.
//
// It bounds the map by what is contended *now*. Without it a long-running engine
// keeps one entry per row it ever found held, including rows that were settled
// by somebody else and will never be listed again.
func (d *alertDeliverer) forgetLapsedHeld(now time.Time) {
	for id, expires := range d.heldReported {
		if !expires.After(now) {
			delete(d.heldReported, id)
		}
	}
}

func (d *alertDeliverer) interval() time.Duration {
	if d.Interval <= 0 {
		return alertDeliveryInterval
	}
	return d.Interval
}

func (d *alertDeliverer) batch() int {
	if d.Batch <= 0 {
		return alertDeliveryBatch
	}
	return d.Batch
}

func (d *alertDeliverer) claimant() string {
	if d.Claimant == "" {
		return "engine.alert_delivery"
	}
	return d.Claimant
}

// logf never carries the claim token: whoever reads it can settle somebody
// else's send (a099, ClaimResult.Token).
func (d *alertDeliverer) logf(event obs.EventType, err error, detail string, args ...any) {
	if d.Log == nil {
		return
	}
	args = append([]any{obs.FieldDetail, detail}, args...)
	if err != nil {
		d.Log.Error(event, err, args...)
		return
	}
	d.Log.Warn(event, args...)
}
