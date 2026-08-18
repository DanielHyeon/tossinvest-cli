# Design: the minute timestamp is the bar's close

## 1. The measured fact this change encodes

A bar labelled `t` covers `[t − interval, t)`. Measured 2026-08-18 in both markets and recorded as a112 decisions 30 (US) and 31 (KR); the KR reading is the decisive one because the wall clock sat inside `[09:06:00, 09:07:00)` while only the bar labelled `09:07` accumulated volume and the bar labelled `09:06` stayed frozen. Under open labels the forming bar would have had to be `09:06` and grow, and a bar labelled `09:07` could not have existed at all.

## 2. What that means for a five-minute KRX bucket

The KRX regular session is `09:00`–`15:30` KST (`openspec/specs/strategy-engine/spec.md`, *KRX open은 `09:00 KST`로 봉인*). Read through close labels:

| Bucket | Minute labels | Opens | Closes |
|---|---|---|---|
| first | `09:01 09:02 09:03 09:04 09:05` | `09:00` | `09:05` |
| second | `09:06 … 09:10` | `09:05` | `09:10` |
| last | `15:26 … 15:30` | `15:25` | `15:30` |

Three rules follow, and they are the whole change:

1. **Open** = first label − 1 minute.
2. **Grid alignment** is checked on that derived open, not on the label. Anchored at `09:00`, `(open − 09:00) mod 5 == 0`.
3. **Session membership** for a label is `09:01 ≤ label ≤ 15:30` — inclusive at both ends. The label `09:00` covers `08:59`–`09:00` and is outside; the label `15:30` covers `15:29`–`15:30` and is inside. The wall-clock window is unchanged; only its expression in label space moves.

`closedAt` is then `open + 5m`, which is identically the last label — the aggregator can compute it either way and the tests pin the value, not the expression.

## 3. Why the fixtures shift and nothing else does

Every existing fixture named a bucket by its **opening** instant and then emitted labels starting at that instant, which was correct under the old reading and wrong under the measured one. Shifting each fixture's labels by one minute keeps the bucket it names identical, so `OpenAt()` and `ClosedAt()` come out exactly as before and no downstream lane assertion moves. That is why `internal/strategyengine`'s session-boundary and cutoff tests are untouched despite depending on both instants.

## 4. Direction of the behaviour change

The correction is conservative where it can be: a minute that is currently admitted as regular (`08:59`–`09:00`, pre-open) becomes a refusal. It also admits the session's genuine last minute, which is a widening — but only back to the session boundary the spec already froze, never past it. No refusal kind, threshold, sizing, stop or exposure value is touched.

## 5. What is deliberately not done

- `docs/migration/openapi.latest.json` keeps the broker's wording. It is a mirrored external document; correcting it here would fork it silently from its source. `candle_reads.go` now says *why* it disagrees with that document, which is the useful place to say it.
- The aggregator is not deleted even though it has no production caller and `internal/officialbars` supersedes it. Deleting a047's surface is an architecture decision with its own spec consequences; this change corrects a false value and stops there (YAGNI).
- No production wiring, no toggle, no container.
