# Function Logic Map: `usBar`

- Source: `internal/officialbars/producer_test.go`
- Source SHA-256: `6c6f064d3ecceb46ed4df8546cb5d1f16c40e2831dbfa99c223c0c99f2b5eda4` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `usBar(t *testing.T, open time.Time, closePrice string) official.RawMinuteCandle`
- Source range: `115:1`–`118:2` (ast.json `start`/`end`), revision `current`
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 0, returns 1, calls 3, assignments 0, defers 0, go statements 0.
- Disposition: **Test helper / test function caught by the gate because it existed at the frozen base `d9d9f71f` and the decision-30 correction edited it. Enumerated for gate completeness: these carry no production branch.**
- Risk scan: `risk-pattern-report.md` (no configured risk pattern matched).

## Inputs and invariants

- A three-line US fixture builder: given the instant a bar **opens**, return one `official.RawMinuteCandle` as the broker would put it on the wire.
- **What decision 30 changed.** At base it built the struct inline with `Timestamp: brokerTimestamp(t, open)` — i.e. it wrote the *open* instant onto the wire. The 2026-08-18 03:29 KST US probe proved the broker labels a bar with its **close** (the bar labelled `03:30:00` was the one still growing while the wall clock sat inside `[03:29, 03:30)`), so the helper now delegates to `usBarLabelled(t, open.Add(time.Minute), closePrice)`. The wire value is `open + 1m`; the argument is still the open instant.
- The invariant that makes the re-basing safe for ~30 call sites: **every caller still reads "a bar opening at X"**, so no expectation in those tests had to move. Only the bytes underneath changed, and they changed to what the broker really sends. `adoptPage` subtracts the minute back (`internal-officialbars--adoptpage`).
- The five non-timestamp fields are fixed literals (`Open` 231.4350, `High` 231.5000, `Low` 231.4000, `Volume` 1234, `Currency` USD); only `closePrice` varies, so a test can tell two bars apart by their close alone.
- `t.Helper()` is called first so a failure inside the fixture is reported at the caller's line, not here.

## Branches and early returns

The AST reports **0 branches**: the body is `t.Helper()` followed by a single unconditional `return` (AST return node `117`). There is nothing to enumerate, so the table below carries the one `—` row rather than a fabricated branch.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| — | (none) | 115–118 | branchless: `t.Helper()` then one unconditional return | happy path only; see `branch-test-map.md` B1 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 116:2 |
| `usBarLabelled` | 117:9 |
| `open.Add` | 117:26 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `t.Helper()` | 116 | failures attribute to the calling test, not to the fixture |
| `open.Add(time.Minute)` | 117 | **the decision-30 conversion, in the test direction**: open instant → broker close label |
| `usBarLabelled(t, …, closePrice)` | 117 | the label-space builder introduced by the correction (`producer_test.go:122`); it owns the literal field values and the `brokerTimestamp` formatting |

## State mutations and fallbacks

- No assignments (AST `assignments` 0), no package state, no I/O, no goroutines, no defers, no clock read. It is a pure function of its three arguments.
- No fallback and no failure path of its own: any failure surfaces from `usBarLabelled` → `brokerTimestamp` → `mustLocation`, which fails the test outright rather than returning a partial candle.

## Safety conclusion

- Safe edit boundary: a test fixture builder. It never runs in production and touches no order, stop-loss, sizing, guardian, journal or toggle surface.
- High-risk impact: no. There is no production branch behind it. Its correctness is load-bearing for the *evidence* the L1b bundles cite, though: if the `+1m` were removed, ~30 tests would silently start asserting against open-labelled wire data and would stop proving the conversion at all. That is the mirror of the "do not fix this subtraction back" comment in `producer.go:402-414`.
- Honest scope note: the helper cannot prove the broker's convention — the 03:29 KST measurement did. This helper only encodes the measured convention into the fixtures.
