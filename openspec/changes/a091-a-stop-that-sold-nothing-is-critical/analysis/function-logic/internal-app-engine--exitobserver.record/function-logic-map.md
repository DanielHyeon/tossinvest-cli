# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (`1225`–`1355`)
- Qualified: `ExitObserver.record`
- AST evidence: `ast.json` (`source_sha256` 0f943813a3efa423…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 16 · 반환 6

**편집.** 편집하지 않는다 — M1: `submit` 은 `orderable` 일 때만 불린다(B11)는 증거.

**역할.** 판정을 원장에 무장하고 제출로 넘긴다. `orderable = snapshot.Orderable && !proposal.Zero()`(`:1233`) 가 거짓이면 제출이 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.ProjectedQuantity` | 정수 정규형 | `EvaluateLadderSnapshot` · `EvaluateRatchetSnapshot` | `"0"` 이면 `Orderable=false` |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1228` `if !o.quoteUsable(quote) {` | 아니오 |
| B2 | if | `:1247` `if judgement.ObservationSource == "" {` | 아니오 |
| B3 | if | `:1248` `if quote.FetchedAt.IsZero() {` | 아니오 |
| B4 | else | `:1250` `} else {` | 아니오 |
| B5 | if | `:1271` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 |
| B6 | if | `:1272` `if m.reJudge && !isProtective(proposal) {` | 예 |
| B7 | else | `:1294` `} else {` | 예 |
| B8 | if | `:1296` `if err != nil {` | 예 |
| B9 | if | `:1299` `if !cleared.cleared {` | 예 |
| B10 | else | `:1306` `} else {` | 예 |
| B11 | if | `:1314` `if orderable {` | 예 |
| B12 | if | `:1316` `if intentID == "" {` | 아니오 |
| B13 | if | `:1328` `if err != nil {` | 예 |
| B14 | if | `:1329` `if errors.Is(err, journal.ErrProposalPending) {` | 예 |
| B15 | if | `:1335` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 |
| B16 | if | `:1348` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 |

Exact AST return positions: `1229:3`, `1297:5`, `1333:4`, `1346:3`, `1349:3`, `1353:2`

## Calls and live bindings

호출 좌표 전수는 `ast.json` `calls`(24) — 이 번들이 근거로 쓰는 것은 B11 → `o.submit`(`:1353`) 하나다.

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.submit` | `:1353` | 제출 | submit 번들 |

## State mutations and fallbacks

원장 무장 · 해제(판정 트랜잭션).

## Safety conclusion

- 0 투영 수량은 `orderable` 을 거짓으로 만들어 `submit` 에 닿지 않는다 — `applyFloor` 에 들어오는 `quantity` 는 항상 양수다.
