package verifylive

// korean.go is the display language (openspec change verify-execution-capability,
// task 1.8 ③).
//
// # What is translated, and what is frozen
//
// The sole operator of this tool is Korean and drives it from a Korean web console,
// so everything a person reads is Korean: the runner's progress, the approval
// prompt, the step catalogue, the screens. What stays in the original is everything
// a machine or a later reader joins on:
//
//	step ids            read-fixtures, conditional-persist. They are the record's
//	                    join key; a translated key orphans every line already written.
//	verdicts            pass, fail, skipped, awaiting-restart.
//	observation keys    conditional.survives_process_exit, and their values.
//	broker vocabulary   order-hours-closed, HTTP 422, POST /api/v1/orders,
//	                    clientOrderId, sellableQuantity, WATCHING. Quoting the
//	                    broker in translation would make the evidence uncheckable.
//	Step.Title          it is copied verbatim into the record's `title` field, and
//	                    records already written are in English. Comparability wins
//	                    over display language — which is why the labels below exist
//	                    as a separate mapping instead of as a change to Title.
//	the plan's own text StepMutation.Ends, .Note and the pricing rule are hashed
//	                    into approval.plan_digest. See the KO fields in plan.go:
//	                    the Korean is carried alongside, never instead.
//
// # The label map is exhaustive, and a test says so
//
// A step added to the catalogue without a label here would render as a bare
// identifier on the one screen where a person decides whether to send live orders.
// TestEveryCatalogueStepHasAKoreanLabel fails instead.

import "fmt"

// stepLabels is the operator-facing name of every step, in Korean.
//
// They are short on purpose: they sit in a table column next to the step id, which
// is the thing that is looked up, cited and grepped. The full explanation is
// Step.Proves and the procedure list.
var stepLabels = map[StepID]string{
	StepReadFixtures:        "기존 주문 이력에서 상태 enum 수집",
	StepSellableBaseline:    "매도가능수량 기준선",
	StepIdempotency:         "clientOrderId 재생·충돌·스코프",
	StepIdempotencyTTLEdge:  "멱등키 유효 창 경계",
	StepOrderCancel:         "KR 주문 접수·취소",
	StepOrderAmend:          "KR 주문 정정",
	StepSellBoundary:        "매도 경계: 부분·전량·보유초과",
	StepConditionalRegister: "조건주문 등록·조회",
	StepSellableReserved:    "조건주문 예약 중 매도가능수량",
	StepConditionalPersist:  "조건주문이 프로세스 종료보다 오래 사는가",
	StepConditionalTrigger:  "발동 관측과 triggeredOrderId 지연",
	StepConditionalModify:   "조건주문 정정은 새 ID를 발급하는가",
	StepConditionalCancel:   "조건주문 취소",
	StepCosts:               "실현 비용 수집",
}

// StepLabel is the Korean name of a step, for a screen or a progress line.
//
// An unknown id comes back as itself. That is the honest answer for a record
// written by a newer build than the one reading it, and it is never silently blank.
func StepLabel(id StepID) string {
	// The cleanup prologue is labelled here rather than in stepLabels: that map is
	// asserted to hold the catalogue and nothing but the catalogue, and the
	// prologue is deliberately not a catalogue step (cleanup.go).
	if id == StepCleanup {
		return CleanupLabel
	}
	if label, ok := stepLabels[id]; ok {
		return label
	}
	return string(id)
}

// StepLabels exposes the mapping for a drift test. It is a copy: a caller that
// mutated the map would change what an operator reads before approving live orders.
func StepLabels() map[StepID]string {
	out := make(map[StepID]string, len(stepLabels))
	for id, label := range stepLabels {
		out[id] = label
	}
	return out
}

// verdictLabels gloss the record's own verdict values.
//
// The value itself is always shown next to the gloss and never replaced by it: it
// is what the record holds, what `verify report` prints and what an operator pastes
// into a thread.
var verdictLabels = map[Verdict]string{
	VerdictPass:            "통과",
	VerdictFail:            "실패",
	VerdictSkipped:         "생략",
	VerdictRefused:         "거부됨",
	VerdictDeferred:        "보류",
	VerdictAwaitingRestart: "새 프로세스 대기",
}

// VerdictLabel renders a verdict as "pass (통과)": the recorded value, glossed.
func VerdictLabel(v Verdict) string {
	if v == "" {
		return "-"
	}
	if label, ok := verdictLabels[v]; ok {
		return fmt.Sprintf("%s (%s)", v, label)
	}
	return string(v)
}

// VerdictLabels exposes the gloss map for a drift test.
func VerdictLabels() map[Verdict]string {
	out := make(map[Verdict]string, len(verdictLabels))
	for v, label := range verdictLabels {
		out[v] = label
	}
	return out
}

// Verdicts is every verdict the runner can record. It is data so the drift test can
// walk it — a seventh verdict added without a gloss fails rather than rendering
// bare on the screen a person approves live orders from.
func Verdicts() []Verdict {
	return []Verdict{
		VerdictPass, VerdictFail, VerdictSkipped,
		VerdictRefused, VerdictDeferred, VerdictAwaitingRestart,
	}
}

// VerbKO is MutationKind.Verb in Korean: the action line of the approval summary.
//
// The English Verb stays where it is. It is what the per-mutation prompt was built
// around and what the plan's own tests read; this is the display form, chosen at
// render time.
func (k MutationKind) VerbKO() string {
	switch k {
	case MutatePlaceOrder:
		return "실제 LIMIT 주문 접수"
	case MutateReplayOrder:
		return "동일 clientOrderId로 같은 주문 재전송"
	case MutateConflictProbe:
		return "본문이 다른 요청으로 멱등키 충돌 관측"
	case MutateCancelOrder:
		return "살아 있는 주문 취소"
	case MutateAmendOrder:
		return "살아 있는 주문 정정"
	case MutateRegisterConditional:
		return "조건주문 등록"
	case MutateReplayConditional:
		return "동일 clientOrderId로 같은 조건주문 재전송"
	case MutateModifyConditional:
		return "조건주문 정정"
	case MutateCancelConditional:
		return "조건주문 취소"
	default:
		return string(k)
	}
}

// MutationKinds is every class of live request, as data, for the drift test.
func MutationKinds() []MutationKind {
	return []MutationKind{
		MutatePlaceOrder, MutateReplayOrder, MutateConflictProbe, MutateCancelOrder,
		MutateAmendOrder, MutateRegisterConditional, MutateReplayConditional,
		MutateModifyConditional, MutateCancelConditional,
	}
}
