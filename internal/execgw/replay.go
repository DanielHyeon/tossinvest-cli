package execgw

// replay.go is the idempotent replay entry point (extend-execution-contract
// tasks 2.1/2.2, design D3).
//
// # A replay is not a retry
//
// A retry sends a mutation again in the hope that it works this time. A replay
// resends *the same bytes under the same idempotency key*, which the broker
// answers by handing back the original order's result instead of creating a
// second one: "동일 값으로 재요청 시 이전 주문 결과를 그대로 재반환합니다 …
// 멱등성 키는 10분간 유효하며" (openapi, OrderCreateRequest.clientOrderId). The
// purpose is identity recovery — learning *which* order exists — and inside the
// key's validity window a replay cannot create a second order at all. That is
// what makes it compatible with "주문 mutation은 어떤 오류에도 자동 재시도 금지".
//
// # Why the entry point guards itself
//
// Every precondition is checked here rather than assumed of the caller, because
// a caller that forgets one is indistinguishable from a caller that decided the
// rule did not apply. The entry point verifies, on its own:
//
//	the attempt is IN_DOUBT               — the only state a replay is defined for
//	the capability attestation is on      — off by default until 2b measures it
//	elapsed < TTL − margin, *every time*  — re-checked before each individual send
//	the replay cap and minimum interval   — counted durably, in the journal
//	the body is the stored one            — see ReplayBody; there is no other one
//
// # Why the input is an attempt id and nothing else
//
// ReplayInDoubt takes a context and an attempt id. It has no parameter that
// could carry a request body, and the only value it can hand the transport is a
// ReplayBody, whose single field is unexported — so no package outside this one
// can construct a non-empty one. "Replay something other than what was stored"
// is therefore not an API anyone can spell, which is the structural form of
// design D3's "저장된 wire body 외 전송 불가".
//
// # Why the dispatch classifier is not used here
//
// classifyMutation/journal.ClassifyHTTPMutation treat a 422 as a definitive
// rejection, which is correct for a first dispatch and wrong for a replay: a
// replay's `422 idempotency-key-conflict` says the key was once used for a
// different body, and says *nothing* about whether the original order exists.
// Running it through the dispatch classifier would settle the attempt
// FAILED_CONFIRMED on the strength of a fact about the replay. The rules below
// are therefore written out separately, as the spec requires.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// ReplayKeyTTL is how long the broker honours an idempotency key: "멱등성 키는
// 10분간 유효하며, 이후 동일 값으로 재요청 시 새 주문으로 처리됩니다" (openapi,
// OrderCreateRequest.clientOrderId).
//
// It is a constant rather than a configuration field because it is a documented
// broker fact, and because the only dangerous direction is *widening* it: a
// replay sent after the key expired is a second live order. The margin below is
// what a deployment tunes, and it can only make the usable window smaller.
const ReplayKeyTTL = 10 * time.Minute

// Replay defaults.
const (
	// DefaultReplayMargin keeps the last minute of the key's life unused. The
	// evidence that the window is still open is a *local* clock reading against a
	// timestamp this process wrote, so the boundary itself must never be used
	// (order-execution: "경과 근거가 로컬 시계이므로 마진 없는 경계 사용 금지").
	// Sixty seconds is provisional — verify-execution-capability measures the
	// real round-trip p99 [미측정 — 2b].
	DefaultReplayMargin = 60 * time.Second
	// DefaultMaxReplays is how many replays may be *counted* against one attempt.
	DefaultMaxReplays = 2
	// DefaultReplayMinInterval is the minimum gap between two replays of the same
	// attempt. It matches the resolver's poll interval: both are "how long to
	// wait before asking the broker the same question again".
	DefaultReplayMinInterval = 5 * time.Second
	// DefaultMaxReplayWaits bounds how many times one call will wait. It exists
	// because `409 request-in-progress` does not consume the cap, so the cap
	// alone does not bound a broker that answers 409 forever. How long the
	// original request can stay in progress is [미측정 — 2b].
	DefaultMaxReplayWaits = 3
)

// Broker error codes this file reacts to, quoted from openapi:
//
//	409 request-in-progress      "동일 주문 키에 대해 처리 중인 요청이 있습니다"
//	422 idempotency-key-conflict "동일한 clientOrderId 로 다른 내용의 주문을 요청할 수 없습니다"
const (
	replayCodeRequestInProgress = "request-in-progress"
	replayCodeKeyConflict       = "idempotency-key-conflict"
)

// The critical alert an UNRESOLVED attempt owes an operator.
//
// The event type is obs.EventOrderUnresolved. It is spelled out rather than
// imported because internal/obs imports *this* package (its notifier latches the
// entry gate), so the dependency can only run one way. The alert is written
// straight into the journal's outbox, which is the durable half of obs's
// critical path — the notifier's Flush picks the row up and delivers it — so an
// engine whose notifier is not wired still records the alert rather than losing
// it. TestReplayKeyConflictEnqueuesTheCriticalAlert pins the string against
// obs's own constant.
const (
	eventOrderUnresolved  = "order.unresolved_in_doubt"
	alertSeverityCritical = "critical"
)

// ReplayConfig tunes the replay procedure. Zero fields take the defaults above.
type ReplayConfig struct {
	// Margin is subtracted from ReplayKeyTTL before every send.
	Margin time.Duration
	// MaxReplays is the per-attempt cap on counted replays.
	MaxReplays int
	// MinInterval is the minimum gap between two replays of one attempt.
	MinInterval time.Duration
	// MaxWaits bounds the waits one call performs.
	MaxWaits int
}

func (c ReplayConfig) withDefaults() ReplayConfig {
	if c.Margin <= 0 {
		c.Margin = DefaultReplayMargin
	}
	if c.MaxReplays <= 0 {
		c.MaxReplays = DefaultMaxReplays
	}
	if c.MinInterval <= 0 {
		c.MinInterval = DefaultReplayMinInterval
	}
	if c.MaxWaits <= 0 {
		c.MaxWaits = DefaultMaxReplayWaits
	}
	return c
}

// ReplayBody is the stored request body, and the only thing a replay can send.
//
// The unexported field is the design, not an accident: no package outside this
// one can build a non-empty ReplayBody, so there is no way to hand a transport a
// body that did not come out of the journal — and the entry point that produces
// one takes an attempt id and nothing else. Together those two facts are what
// makes "replay a body I constructed" unrepresentable rather than merely
// discouraged (design D3).
type ReplayBody struct {
	stored string
}

// Bytes returns the stored body exactly as it was recorded at RECORDED time.
func (b ReplayBody) Bytes() []byte { return []byte(b.stored) }

// String returns the stored body.
func (b ReplayBody) String() string { return b.stored }

// Empty reports whether there is nothing to send. A zero ReplayBody — the only
// one an outside package can make — is empty, and transports refuse it.
func (b ReplayBody) Empty() bool { return b.stored == "" }

// ReplayResponse is a replay's answer, in the terms the rules below need.
//
// It is deliberately not domain.MutationResult: a replay is not a mutation, and
// the status code and broker error code are load-bearing here in a way they
// never are on the dispatch path.
type ReplayResponse struct {
	// Status is the HTTP status. Zero means no response arrived.
	Status int
	// OrderID is the identifier from a 2xx body, byte-exact as received.
	OrderID string
	// ClientOrderID is the echoed idempotency key, byte-exact, "" when absent.
	ClientOrderID string
	// ErrorCode is the broker's `error.code` for a non-2xx answer.
	ErrorCode string
	// ErrorMessage is its `error.message`, for the operator.
	ErrorMessage string
}

// ReplayTransport resends one stored body.
//
// It has no method that builds a request: the only input is a ReplayBody the
// caller cannot construct. Implementations must not retry — a replay is counted
// before it is sent, and a transport-level retry would send an uncounted one.
type ReplayTransport interface {
	ReplayPlace(ctx context.Context, body ReplayBody) (ReplayResponse, error)
}

// ReplayOutcome is what one call to ReplayInDoubt did.
type ReplayOutcome struct {
	AttemptID string
	Symbol    string
	// State is the attempt's state after the procedure: CONFIRMED when the
	// identity was recovered, UNRESOLVED_IN_DOUBT when the key conflicted, and
	// otherwise whatever it already was.
	State journal.AttemptState
	// BrokerOrderID is the identifier recovered from a replay, verbatim.
	BrokerOrderID string
	Reason        ReasonCode
	Detail        string
	// Sent is how many replay requests this call actually put on the wire.
	Sent int
	// ReplayCount is the attempt's durable counter afterwards. A 409 does not
	// raise it.
	ReplayCount int
	// QueryFallback reports that the attempt is still unsettled and the caller
	// must run the query fallback. It is the normal outcome, not an error.
	QueryFallback bool
}

// ReplayInDoubt is the resolution procedure's first step: recover the identity
// of an IN_DOUBT mutation by resending exactly what was sent.
//
// The attempt id is the whole input. Everything else — the bytes, the key, the
// serializer version, the dispatch timestamp, the replay count — is read from
// the journal, and every precondition is checked here.
//
// It returns an error only when the journal itself could not be read or written.
// "Not eligible", "out of replays" and "the broker is still processing" are all
// ordinary outcomes with QueryFallback set: replay is the fast path, and the
// query fallback is what the procedure falls back *to*.
func (g *Gateway) ReplayInDoubt(ctx context.Context, attemptID string) (ReplayOutcome, error) {
	cfg := g.replayCfg.withDefaults()
	out := ReplayOutcome{AttemptID: attemptID}

	rec, err := g.journal.LookupAttempt(ctx, attemptID)
	if err != nil {
		return out, err
	}
	out.State = rec.State
	intent, err := g.journal.LookupIntent(ctx, rec.IntentID)
	if err != nil {
		return out, err
	}
	out.Symbol = intent.Symbol

	// A settled attempt has nothing to recover and needs no fallback. Reported,
	// not refused, so a recovery sweep can run this over every attempt it finds.
	if rec.State.IsTerminal() {
		out.Reason = ReasonCode(rec.ReasonCode)
		out.Detail = rec.Detail
		out.BrokerOrderID = rec.BrokerOrderID
		return out, nil
	}
	if g.replay == nil {
		return declineReplay(out, ReasonReplayIneligible,
			"no replay transport is wired, so identity can only be recovered by observation"), nil
	}
	// The attestation gate. Off by default and off in every build until
	// verify-execution-capability has measured the replay path against the real
	// broker: until then the fallback is the P1 procedure, which needs no
	// capability we have not tested. [미측정 — 2b 전 비활성]
	if g.attested == nil || !g.attested(ctx) {
		return declineReplay(out, ReasonReplayNotAttested,
			"the replay capability is not attested, so no request is resent [미측정 — 2b 전 비활성]"), nil
	}

	// Serialise against this process's own mutations on the symbol. The journal
	// already refuses a new mutation while this attempt is unsettled; this closes
	// the in-process window and stops two replays of one attempt overlapping.
	symbolKey := strings.ToLower(intent.Market) + "|" + strings.ToUpper(intent.Symbol)
	if !g.claimSymbol(symbolKey) {
		return declineReplay(out, ReasonSymbolInFlight,
			"another mutation on "+intent.Symbol+" is in flight in this process"), nil
	}
	defer g.releaseSymbol(symbolKey)

	attempt, err := g.journal.Resume(ctx, attemptID)
	if err != nil {
		return out, err
	}

	waits := 0
	for {
		// Re-read every time round. The guards below are checked against the row,
		// not against a copy this goroutine has been holding since the top: a
		// replay's authority is the record, and the record can change.
		rec, err = g.journal.LookupAttempt(ctx, attemptID)
		if err != nil {
			return out, err
		}
		out.State = rec.State
		out.ReplayCount = rec.ReplayCount
		if rejected := g.replayGuards(rec, cfg); rejected != nil {
			return declineReplay(out, rejected.Reason, rejected.Detail), nil
		}

		// The minimum interval, enforced before the count is taken so a refused
		// wait does not spend one.
		if wait := replayWait(rec, cfg, g.clk.Now()); wait > 0 {
			if waits >= cfg.MaxWaits {
				return declineReplay(out, ReasonReplayExhausted, fmt.Sprintf(
					"the broker was still processing after %d wait(s); the query fallback takes over", waits)), nil
			}
			waits++
			if err := g.clk.Sleep(ctx, wait); err != nil {
				return out, err
			}
			continue
		}

		state, err := g.journal.MarkReplayStarted(ctx, attemptID, cfg.MaxReplays)
		switch {
		case errors.Is(err, journal.ErrReplayCapReached):
			out.ReplayCount = state.Count
			return declineReplay(out, ReasonReplayExhausted, fmt.Sprintf(
				"the attempt has used its %d replay(s) without recovering an identity", cfg.MaxReplays)), nil
		case errors.Is(err, journal.ErrReplayNotInDoubt):
			return declineReplay(out, ReasonReplayIneligible, err.Error()), nil
		case err != nil:
			return out, err
		}
		out.ReplayCount = state.Count
		out.Sent++

		// The stored bytes, and only the stored bytes.
		resp, callErr := g.replay.ReplayPlace(ctx, ReplayBody{stored: rec.WireBody})
		verdict := classifyReplay(rec, resp, callErr)

		switch verdict.kind {
		case replayRecovered:
			detail := fmt.Sprintf(
				"replaying the stored body under key %s returned order %s (serializer %s, replay %d)",
				rec.ClientOrderID, verdict.orderID, rec.SerializerVersion, state.Count)
			if err := attempt.ResolveConfirmed(ctx, verdict.orderID,
				journal.ReasonReplayRecovered, detail); err != nil {
				return out, err
			}
			out.State = journal.StateConfirmed
			out.BrokerOrderID = verdict.orderID
			out.Reason = ReasonReplayRecovered
			out.Detail = detail
			return out, nil

		case replayInProgress:
			// The cap is not consumed: the broker told us it is still working on
			// the original request, which is the most common answer there is and
			// establishes nothing either way.
			refunded, rerr := g.journal.RefundReplay(ctx, attemptID)
			if rerr != nil {
				return out, rerr
			}
			out.ReplayCount = refunded.Count
			out.Reason = ReasonReplayInProgress
			out.Detail = verdict.detail
			continue

		case replayKeyConflict:
			// Never FAILED_CONFIRMED. The conflict is a fact about the *key*, and
			// the original order is exactly as unknown as it was a moment ago.
			if err := attempt.ResolveUnresolved(ctx, journal.ReasonReplayKeyConflict, verdict.detail); err != nil {
				return out, err
			}
			g.parkAlert(ctx, rec, intent, verdict.detail)
			out.State = journal.StateUnresolvedInDoubt
			out.Reason = ReasonReplayKeyConflict
			out.Detail = verdict.detail
			return out, nil

		default:
			// Inconclusive: recorded by the counter, and the loop either spends
			// the cap or the key's window runs out. Either way the fallback runs.
			out.Reason = ReasonBrokerOutcomeUnknown
			out.Detail = verdict.detail
			continue
		}
	}
}

// replayGuards applies the per-send preconditions to the record as it is now.
//
// The TTL check is here, and therefore re-run before every individual send,
// because the spec says "재생 1회마다": a first replay that started well inside
// the window says nothing about whether a second one still is.
func (g *Gateway) replayGuards(rec journal.AttemptRecord, cfg ReplayConfig) *RejectedError {
	if rec.State != journal.StateInDoubt {
		return reject(ReasonReplayIneligible,
			"attempt %s is %s; only an IN_DOUBT attempt is replayable", rec.ID, rec.State)
	}
	if rec.Kind != journal.KindPlace {
		// Cancel and modify take no clientOrderId (openapi puts it on
		// OrderCreateRequest and nowhere else), so there is no key to replay
		// under and resending would be a second cancel, not a replay.
		return reject(ReasonReplayIneligible,
			"a %s carries no idempotency key, so it is resolved by observation", rec.Kind)
	}
	if strings.TrimSpace(rec.ClientOrderID) == "" {
		return reject(ReasonReplayIneligible,
			"attempt %s was dispatched without an idempotency key", rec.ID)
	}
	if rec.WireBody == "" {
		return reject(ReasonReplayIneligible,
			"attempt %s has no stored request body, and a rebuilt one is not the same request", rec.ID)
	}
	// The serializer version is *recorded*, not enforced. Refusing a body written
	// by an older serializer is precisely the failure the stored bytes exist to
	// avoid: "저장된 wire body가 그대로 사용되어 본문 불일치가 발생하지 않는다".
	stamp := strings.TrimSpace(rec.DispatchStartedAt)
	if stamp == "" {
		return reject(ReasonReplayIneligible,
			"attempt %s records no dispatch time, so the key's remaining life cannot be bounded", rec.ID)
	}
	dispatched, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return reject(ReasonReplayIneligible,
			"attempt %s has an unreadable dispatch time %q", rec.ID, stamp)
	}
	usable := ReplayKeyTTL - cfg.Margin
	if elapsed := g.clk.Now().Sub(dispatched); elapsed >= usable {
		return reject(ReasonReplayExpired,
			"%s have passed since dispatch and the key is only usable for %s of its %s life",
			elapsed.Round(time.Second), usable, ReplayKeyTTL)
	}
	return nil
}

// replayWait reports how long is left of the minimum interval.
func replayWait(rec journal.AttemptRecord, cfg ReplayConfig, now time.Time) time.Duration {
	stamp := strings.TrimSpace(rec.LastReplayAt)
	if stamp == "" {
		return 0
	}
	last, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		// An unreadable stamp is treated as "just now": waiting one interval too
		// many costs a few seconds, and not waiting costs a duplicate request.
		return cfg.MinInterval
	}
	if remaining := cfg.MinInterval - now.Sub(last); remaining > 0 {
		return remaining
	}
	return 0
}

// declineReplay finishes the outcome as "nothing was settled; run the fallback".
func declineReplay(out ReplayOutcome, reason ReasonCode, detail string) ReplayOutcome {
	out.Reason = reason
	out.Detail = detail
	out.QueryFallback = true
	return out
}

// --- response rules ---------------------------------------------------------

type replayVerdictKind int

const (
	// replayInconclusive: the answer proves nothing. Counted, then retried or
	// handed to the fallback.
	replayInconclusive replayVerdictKind = iota
	// replayRecovered: 2xx carrying the original order's identifier.
	replayRecovered
	// replayInProgress: 409 request-in-progress. Wait; the cap is refunded.
	replayInProgress
	// replayKeyConflict: the key does not name our order. Park, never fail.
	replayKeyConflict
)

type replayVerdict struct {
	kind    replayVerdictKind
	orderID string
	detail  string
}

// classifyReplay applies the replay response rules.
//
// It shares no code with classifyMutation, and that is the requirement rather
// than an oversight (order-execution: "재생 응답 분류에 dispatch 분류기를 사용해서는
// 안 된다"). The two disagree about 422 in a way that decides whether an attempt
// is settled FAILED_CONFIRMED or parked for a human.
func classifyReplay(rec journal.AttemptRecord, resp ReplayResponse, err error) replayVerdict {
	if err != nil {
		return replayVerdict{kind: replayInconclusive,
			detail: "the replay produced no usable answer: " + err.Error()}
	}

	switch {
	case resp.Status >= 200 && resp.Status < 300:
		if strings.TrimSpace(resp.OrderID) == "" {
			return replayVerdict{kind: replayInconclusive,
				detail: fmt.Sprintf("the broker answered HTTP %d without an order id, so no identity was recovered",
					resp.Status)}
		}
		// The echo is documented to come back "요청 시 전달한 값 그대로"
		// (openapi, OrderResponse.clientOrderId), so a *different* non-empty key
		// describes a different order. Byte-exact, because `orderId` and the key
		// are opaque tokens and we have no rule that says whitespace or case do
		// not matter.
		if resp.ClientOrderID != "" && resp.ClientOrderID != rec.ClientOrderID {
			return replayVerdict{kind: replayKeyConflict, detail: fmt.Sprintf(
				"the replay came back with clientOrderId %q but this attempt sent %q — "+
					"the answer is about somebody else's order, so it cannot confirm this one",
				resp.ClientOrderID, rec.ClientOrderID)}
		}
		return replayVerdict{kind: replayRecovered, orderID: resp.OrderID}

	case resp.Status == http.StatusConflict && resp.ErrorCode == replayCodeRequestInProgress:
		// openapi: "동일 주문 키에 대해 처리 중인 요청이 있습니다. 잠시 후 다시
		// 시도해 주세요." The original request is still being processed.
		return replayVerdict{kind: replayInProgress, detail: fmt.Sprintf(
			"the broker is still processing the original request under this key (409 %s): %s",
			resp.ErrorCode, resp.ErrorMessage)}

	case resp.Status == http.StatusUnprocessableEntity && resp.ErrorCode == replayCodeKeyConflict:
		// openapi: "동일한 clientOrderId 로 다른 내용의 주문을 요청할 수 없습니다."
		// Somewhere, this key was used for different bytes. That is a defect in
		// this program, and it says nothing whatsoever about the original order.
		return replayVerdict{kind: replayKeyConflict, detail: fmt.Sprintf(
			"the broker refused the replay with 422 %s: the key %s has been used for a different body, "+
				"which proves nothing about the original order — an operator has to establish what exists",
			resp.ErrorCode, rec.ClientOrderID)}

	default:
		// Everything else, including other 4xx codes. A first dispatch would call
		// several of these definitive refusals; a replay may not, because they
		// describe the replay request and not the order that may already exist.
		return replayVerdict{kind: replayInconclusive, detail: fmt.Sprintf(
			"HTTP %d (%s) describes the replay, not the original order, so it settles nothing",
			resp.Status, firstNonBlank(resp.ErrorCode, resp.ErrorMessage, "no error code"))}
	}
}

// parkAlert records the critical alert a parked attempt owes an operator and
// latches the entry gate, exactly as the resolver's own park does.
//
// The outbox write is best-effort in the sense that a failure to enqueue does
// not un-park the attempt — the park is already durable and is the safety
// property; losing the notification is bad but not unsafe, and it is visible in
// the returned detail.
func (g *Gateway) parkAlert(ctx context.Context, rec journal.AttemptRecord, intent journal.Intent, detail string) {
	if g.entry != nil {
		g.entry.Block(ReasonUnresolvedInDoubt, fmt.Sprintf(
			"attempt %s on %s is unresolved: %s", rec.ID, intent.Symbol, detail))
	}
	payload, err := json.Marshal(map[string]any{
		"attempt_id": rec.ID,
		"intent_id":  rec.IntentID,
		"symbol":     intent.Symbol,
		"market":     intent.Market,
		"account":    intent.AccountRef,
		"reason":     journal.ReasonReplayKeyConflict,
		"detail":     detail,
	})
	if err != nil {
		payload = nil
	}
	_, _ = g.journal.EnqueueAlert(ctx, journal.Alert{
		EventKey: eventOrderUnresolved + "|" + rec.ID,
		Type:     eventOrderUnresolved,
		Severity: alertSeverityCritical,
		Title:    "UNRESOLVED_IN_DOUBT: " + intent.Symbol,
		Body:     detail,
		Payload:  string(payload),
	})
}

func firstNonBlank(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}
