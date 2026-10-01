# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (`1232`–`1362`)
- Qualified: `ExitObserver.record`
- AST evidence: `ast.json` (`source_sha256` 0733bd8641ed8c41…) — 편집 뒤 `540aebe6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 16 · 반환 6

**편집.** 편집하지 않는다 — M1: `submit` 은 `orderable` 일 때만 불린다(B11)는 증거.

**역할.** 판정을 원장에 무장하고 제출로 넘긴다. `orderable = snapshot.Orderable && !proposal.Zero()`(`:1233`) 가 거짓이면 제출이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.ProjectedQuantity` | 정수 정규형 | `EvaluateLadderSnapshot` · `EvaluateRatchetSnapshot` | `"0"` 이면 `Orderable=false` |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `540aebe6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1235` `if !o.quoteUsable(quote) {` | 아니오 |
| B2 | if | `:1254` `if judgement.ObservationSource == "" {` | 아니오 |
| B3 | if | `:1255` `if quote.FetchedAt.IsZero() {` | 아니오 |
| B4 | else | `:1257` `} else {` | 아니오 |
| B5 | if | `:1278` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 |
| B6 | if | `:1279` `if m.reJudge && !isProtective(proposal) {` | 예 |
| B7 | else | `:1301` `} else {` | 예 |
| B8 | if | `:1303` `if err != nil {` | 예 |
| B9 | if | `:1306` `if !cleared.cleared {` | 예 |
| B10 | else | `:1313` `} else {` | 예 |
| B11 | if | `:1321` `if orderable {` | 예 |
| B12 | if | `:1323` `if intentID == "" {` | 아니오 |
| B13 | if | `:1335` `if err != nil {` | 예 |
| B14 | if | `:1336` `if errors.Is(err, journal.ErrProposalPending) {` | 예 |
| B15 | if | `:1342` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 |
| B16 | if | `:1355` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 |

Exact AST return positions: `1236:3`, `1304:5`, `1340:4`, `1353:3`, `1356:3`, `1360:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(24) — 이 번들이 근거로 쓰는 것은 B11 → `o.submit`(`:1353`) 하나다.

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.submit` | `:1360` | 제출 | submit 번들 |

## State mutations and fallbacks

원장 무장 · 해제(판정 트랜잭션).

## Safety conclusion

- 0 투영 수량은 `orderable` 을 거짓으로 만들어 `submit` 에 닿지 않는다 — `applyFloor` 에 들어오는 `quantity` 는 항상 양수다.
