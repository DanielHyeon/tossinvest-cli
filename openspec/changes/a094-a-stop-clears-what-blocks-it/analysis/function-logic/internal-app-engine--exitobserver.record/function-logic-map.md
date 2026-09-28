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
| B1 | if | `:1180` `if !o.quoteUsable(quote) {` | `o.quoteUsable` | :1181 | — (a111 이 더한 분기, base 실측 대상 아님) |
| B2 | if | `:1199` `if judgement.ObservationSource == "" {` | — | — | — (a111 이 더한 분기, base 실측 대상 아님) |
| B3 | if | `:1200` `if quote.FetchedAt.IsZero() {` | `quote.FetchedAt.IsZero` | — | 예 |
| B4 | else | `:1202` `} else {` | — | — | 예 |
| B5 | if | `:1223` `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal)) {` | `isFullExit` | — | 예 |
| B6 | if | `:1224` `if m.reJudge && !isProtective(proposal) {` | `isProtective` | — | 예 |
| B7 | else | `:1246` `} else {` | `o.clearTheSymbol` | — | 예 |
| B8 | if | `:1248` `if err != nil {` | — | :1143 | 아니오 |
| B9 | if | `:1251` `if !cleared {` | `o.noteDelay` | — | 예 |
| B10 | else | `:1255` `} else {` | `o.clearDelay` | — | 예 |
| B11 | if | `:1262` `if orderable {` | `exitIntentID` | — | 예 |
| B12 | if | `:1264` `if intentID == "" {` | `o.opts.Journal.RecordExitJudgementResult`, `o.opts.NewID`, `string` | — | 아니오 |
| B13 | if | `:1276` `if err != nil {` | — | — | 예 |
| B14 | if | `:1277` `if errors.Is(err, journal.ErrProposalPending) {` | `errors.Is` | :1175 | 예 |
| B15 | if | `:1283` `if errors.Is(err, journal.ErrExitSnapshotQuarantined) {` | `errors.Is`, `fmt.Errorf`, `o.announceQuarantineFromLedger` | :1188 | 예 |
| B16 | if | `:1296` `if recorded.ArmedProposal == nil \|\| recorded.ArmOutcome != journal.ExitArmArmed {` | `o.submit` | :1191, :1195 | 예 |

> **4판·5판 재번호(2026-09-27).** 이 표는 현재 AST(16 분기, `exitloop.go` 는 `3937e341` 이후 무변)의 번호와 줄이다. base 의 B1..B14 는 B3..B16 이 됐고 새 B1·B2 는 a111 `882a0b49` 이 더했다(difflib 대응, BTM 과 같다). 「창의 호출/return」 열의 줄 번호(`:1143` 등)는 base 좌표 그대로이며 「진입 실측」 은 base 에서 잰 값이다 — 재측정 안 함.

## Calls and live bindings

`o.clearTheSymbol`(**B7 안**, base `:1141`) · `o.opts.Journal.RecordExitObservation` 계열 · `errors.Is`(B14·B15) · `o.quoteUsable`(B1, a111).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

원장의 관측·제안 기록. 청소 경로는 `clearTheSymbol`이 소유한다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** 게이트 조건(현 **B5** `:1223`, base 번호 B3)은 그대로다. 4판 a094 는 그 안에서 불리는 `clearTheSymbol` 의 목록을 넓히지 않는다(엔진 귀속 주문만, design D−2.4). ~~475150은 `pending_action`이 `STOP_LOSS_LADDER`였으므로 그 게이트는 이미 참이었다~~ — **6판 정정(5라운드 R5-2): 거짓.** 무장 발의가 손절 자신이면 `EvaluateLadder` 가 `ladder.go:441-443` 에서 `SuppressedPending` 으로 발의 없이 돌아가고(`CancelPendingFirst` 를 세우는 `:446-447` 에 닿지 않는다), 그래서 `orderable`(`:1185`)이 거짓이라 B5 `:1223` 게이트에 **닿지 않는다.** 475150 의 청소는 이 함수에서 일어나지 않았다. 또 a111 의 **B1** `:1180-1182`(`!o.quoteUsable(quote)` → `return nil`)이 쓸 수 있는 시세가 없는 주기를 먼저 거른다.
- **High-risk impact**: yes — 청산 판단의 기록과 청소의 진입점.

## Refresh (2026-09-27, HEAD ddd39a83)

`ast.json` 을 현재 소스로 재생성했다(옛 파일은 base `ec29dc72` 소스). 위 본문의 줄 번호는 base 기준이며 현재 위치는 1077-1197 → 1177-1303 이다. 분기 번호의 정본은 `ast.json`·Branch Test Map 이다. 이웃 커밋 882a0b49 (a111) 이 이 함수를 바꿨다 — 3판 문서에 미치는 인용 영향은 `analysis/third-round-errata.md` 에 적는다.
