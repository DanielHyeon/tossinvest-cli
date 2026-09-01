# Function Logic Map: `strictMinuteFinalAttempt`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `strictMinuteFinalAttempt(attempts []AttemptTrace) (AttemptTrace, error)`
- Source range: `239:1`–`251:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 2, returns 3, calls 5, assignments 1, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Input is the ordered list of attempt traces captured by the chained observer for one `c.get` call; output is the single attempt whose bytes become the page.
- Ruling 28 fixes the rule as "the **last** attempt must be 2xx", not "the last 2xx attempt". The difference matters only if `send` ever retried after a success: taking the last *successful* attempt would then let an older body be presented as evidence for a call whose final attempt failed. Today's `send` retries only on 401 and stops at the first 2xx, so the two rules coincide — mutant N19 was declared equivalent by construction in the recheck.
- `AttemptTrace` supplies `Body`, `StatusCode` and `BodyReadComplete`; picking one attempt is what makes `ReadAt`, `StatusCode` and `BodyDigest` mutually consistent in `StrictMinutePage`.

## Branches and early returns

Exact AST return nodes: `242`, `247`, `250`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 240:2 | no attempt was observed at all → `NO_SUCCESSFUL_ATTEMPT` | not-applicable: unreachable defence declared in the source comment at 241 — if `c.get` returned nil, `doRequest` emitted at least one trace on every one of its three exit paths (D6, declared untested in the implementation report) |
| B2 | if | 245:2 | the last attempt carries an error or a non-2xx status → `NO_SUCCESSFUL_ATTEMPT` | not-applicable: unreachable defence declared in the source comment at 246 — `c.get` already returns an error in that case (`classifyStatus`), so the caller never reaches here; the untaken side (401 then 200 ⇒ the 200 attempt is used) is pinned by `TestStrictMinuteCandlesUsesTheLastSuccessfulAttemptAndChainsTheOuterObserver` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 240:5 |
| `strictMinuteRefuse` | 242:26 |
| `len` | 244:19 |
| `strictMinuteRefuse` | 247:26 |
| `strconv.Itoa` | 248:40 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `len(attempts)`, `attempts[len(attempts)-1]` | 240, 244 | last-attempt selection (ruling 28) |
| `strictMinuteRefuse(StrictReasonNoAttempt, …)` ×2 | 242, 247 | typed refusal, distinct from a transport error |
| `strconv.Itoa(last.StatusCode)` | 248 | the refusal names the status it saw |

## State mutations and fallbacks

- Locals only (1 AST assignment: `last`). No client state, no I/O, no goroutines, no defers, no clock read. The trace slice is read, never written — it is appended to by the observer closure in `Client.StrictMinuteCandles`, not here.
- No fallback: there is no "pick an earlier attempt" path. Either the last attempt is usable or the read is refused.

## Safety conclusion

- Fail-closed selector on the High-risk client path: it is the rule that prevents an older or unsuccessful attempt's body from becoming stored evidence. Both of its branches are declared unreachable through today's `send` and are kept deliberately as guards against a future change to the retry policy.
- Because both arms are unreachable, no test can drive them without editing a file this lot declared not-edited (`client.go`); the implementation report declares D6 as untested rather than claiming coverage. The untaken side — a 401 followed by a 200 — is exercised end to end, which is what pins the "last attempt" semantics in practice.
- Recorded residual (review.md 2026-08-17): `send`'s ≤2 refresh-on-401 may refresh a token shared with neighbouring products; the producer only runs when L5 wires it and a human approves.
