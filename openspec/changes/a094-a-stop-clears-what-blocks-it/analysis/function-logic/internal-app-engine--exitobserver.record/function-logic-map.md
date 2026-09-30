# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (`1195`–`1325`)
- Qualified: `ExitObserver.record`
- AST evidence: `ast.json` (`source_sha256` 9b7d9a800069b100…) — 구현 로트(2026-09-30) 편집 뒤. 편집 전 AST 는 `analysis/implementation/pre-edit/`
- Risk scan: `risk-pattern-report.md`
- 분기 16

**역할.** 판정 결과를 원장에 기록하고, 제출 가능하면 발의를 무장해 submit 에 넘긴다. a094: 청소 결과(clearResult)를 받아 지연 사유 · 연속 실패 · 연속 종료를 처리.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cleared` | 청소 결과(cleared · countable · awaitingClose) | clearTheSymbol | 미완료면 무장 안 함(B9) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 `go test ./internal/<pkg>/ -count=1 -covermode=set` 프로파일(2026-09-30, `analysis/harness/flm_tables.py`)에서 그 줄로 시작하는 블록의 count.

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1198` `if !o.quoteUsable(quote) {` | 아니오 |
| B2 | if | `:1217` `if judgement.ObservationSource == "" {` | 아니오 |
| B3 | if | `:1218` `if quote.FetchedAt.IsZero() {` | 아니오 |
| B4 | else | `:1220` `} else {` | 아니오 |
| B5 | if | `:1241` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | 예 |
| B6 | if | `:1242` `if m.reJudge && !isProtective(proposal) {` | 예 |
| B7 | else | `:1264` `} else {` | 예 |
| B8 | if | `:1266` `if err != nil {` | 예 |
| B9 | if | `:1269` `if !cleared.cleared {` | 예 |
| B10 | else | `:1276` `} else {` | 예 |
| B11 | if | `:1284` `if orderable {` | 예 |
| B12 | if | `:1286` `if intentID == "" {` | 아니오 |
| B13 | if | `:1298` `if err != nil {` | 예 |
| B14 | if | `:1299` `if errors.Is(err, journal.ErrProposalPending) {` | 예 |
| B15 | if | `:1305` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | 예 |
| B16 | if | `:1318` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | 예 |

## Calls and live bindings

`clearTheSymbol` · `noteDelay`(사유 = `clearResult.why()` — 형태 B 면 복구 절차 문장) · **`noteClearFailure`** · `clearDelay` · **`endClearStreak`** · `RecordExitJudgementResult` · `submit`.

## State mutations and fallbacks

판정 기록 · 발의 무장(원장) · 관측자 연속 상태(메모리).

## Safety conclusion

- B9: 미완료면 무장 · 제출 없음(종전). 연속 실패 경보는 기존 지연 타이머(delayedSince · delayAlerted)를 건드리지 않는 별개 key(D−2.7). High-risk: yes.
