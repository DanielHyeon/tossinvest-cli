# Function Logic Map: `ExitObserver.record`

- Source: `internal/app/engine/exitloop.go` (`1077`–`1197`)
- Qualified: `ExitObserver.record`
- AST evidence: `ast.json` (`source_sha256` 6625c92061d5b05f…)
- Risk scan: `risk-pattern-report.md`
- 분기 14 · return 5 · 호출 19

**역할.** 관측 한 번의 청산 판단을 기록하고, 필요하면 길을 치운 뒤 제출로 넘긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.CancelPendingFirst` | 이미 걸린 동작이 있는가 (`ladder.go:447` — `PendingAction != ActionNone`) | `exitpolicy` | **B3이 청소 경로를 여는 조건의 절반** |
| `isFullExit(proposal)` | 전량 청산인가 | `exitpolicy` | B3의 나머지 절반 |
| `orderable` | 주문 가능한 상태인가 | 관측 | B3·B9 |
| `m.reJudge` | 재판정 중인가 | 엔진 상태 | B4가 보호 청산을 예외로 둔다 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:1095` `if quote.FetchedAt.IsZero() {` | `quote.FetchedAt.IsZero` | — | 예 |
| B2 | else | `:1097` `} else {` | — | — | 예 |
| B3 | if | `:1117` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | `isFullExit` | — | 예 |
| B4 | if | `:1118` `if m.reJudge && !isProtective(proposal) {` | `isProtective` | — | 예 |
| B5 | else | `:1140` `} else {` | `o.clearTheSymbol` | — | 예 |
| B6 | if | `:1142` `if err != nil {` | — | :1143 | 아니오 |
| B7 | if | `:1145` `if !cleared {` | `o.noteDelay` | — | 예 |
| B8 | else | `:1149` `} else {` | `o.clearDelay` | — | 예 |
| B9 | if | `:1156` `if orderable {` | `exitIntentID` | — | 예 |
| B10 | if | `:1158` `if intentID == "" {` | `o.opts.Journal.RecordExitJudgementResult`, `o.opts.NewID`, `string` | — | 아니오 |
| B11 | if | `:1170` `if err != nil {` | — | — | 예 |
| B12 | if | `:1171` `if errors.Is(err, journal.ErrProposalPending) {` | `errors.Is` | :1175 | 예 |
| B13 | if | `:1177` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | `errors.Is`, `fmt.Errorf`, `o.announceQuarantineFromLedger` | :1188 | 예 |
| B14 | if | `:1190` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | `o.submit` | :1191, :1195 | 예 |

## Calls and live bindings

`o.clearTheSymbol`(**B5 안, `:1141`**) · `o.opts.Journal.RecordExitObservation` 계열 · `errors.Is`(B12·B13).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

원장의 관측·제안 기록. 청소 경로는 `clearTheSymbol`이 소유한다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** B3의 게이트 조건은 그대로다 — a094는 그 안에서 불리는 `clearTheSymbol`이 **무엇을 볼 수 있는지**만 바꾼다. 475150은 `pending_action`이 `STOP_LOSS_LADDER`였으므로 B3은 **이미 참이었다.**
- **High-risk impact**: yes — 청산 판단의 기록과 청소의 진입점.
