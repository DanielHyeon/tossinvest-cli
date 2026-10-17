# Function Logic Map: `Journal.RecordFill`

- Source: `internal/journal/fills.go` (편집 전 `313`–`558`, base `233d5d9c`)
- AST evidence: `ast.json` (`source_sha256` 6214babe8bfa3047…) — branches 38 · returns 26
- Risk scan: `risk-pattern-report.md`
- 작성 시점: **편집 전**(2판, 리뷰 F-5 — 설계가 이 함수 내부를 근거로 씀).

**역할.** 체결 누적 스냅숏 하나를 적용 — 한 트랜잭션에서 스냅숏 · 이벤트 · 위험 · 훅 · 커밋.
호출자: 생산 체결 감지(filldetect) 외 다수.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 같은 트랜잭션의 원장 | — | 원장 | 각 분기 표 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:314` `if strings.TrimSpace(obs.OrderID) == "" {` | — | — | 무편집 갈래 |
| B2 | `:318` `if !scope.complete() && !scope.legacyUnscoped() {` | — | — | 무편집 갈래 |
| B3 | `:324` `if err != nil {` | — | — | 무편집 갈래 |
| B4 | `:328` `if math.IsNaN(filled) \|\| math.IsInf(filled, 0) {` | — | — | 무편집 갈래 |
| B5 | `:340` `if err != nil {` | — | — | 무편집 갈래 |
| B6 | `:346` `switch {` | — | — | 무편집 갈래 |
| B7 | `:347` `case err == nil:` | — | — | 무편집 갈래 |
| B8 | `:348` `case errors.Is(err, ErrFillNotFound):` | — | — | 무편집 갈래 |
| B9 | `:350` `default:` | — | — | 무편집 갈래 |
| B10 | `:356` `if ownerErr != nil {` | — | — | 무편집 갈래 |
| B11 | `:364` `if hadPrev && scope.complete() {` | — | — | 무편집 갈래 |
| B12 | `:368` `if owners == 1 && prev.CommittedAt <= ownershipAt {` | — | — | 무편집 갈래 |
| B13 | `:374` `if err != nil {` | — | — | 무편집 갈래 |
| B14 | `:380` `if hadPrev {` | — | — | 무편집 갈래 |
| B15 | `:385` `if refusal := classifyFillRefusal(obs, prev, hadPrev, filled, prevFilled); refusal != nil ` | — | — | 무편집 갈래 |
| B16 | `:392` `if err := markFillRefused(ctx, tx, scope, obs, refusal, now, hadPrev); err != nil {` | — | — | 무편집 갈래 |
| B17 | `:401` `if err != nil {` | — | — | 무편집 갈래 |
| B18 | `:405` `if err := tx.Commit(); err != nil {` | — | — | 무편집 갈래 |
| B19 | `:413` `if nearlyZero(delta, filled) {` | — | — | 무편집 갈래 |
| B20 | `:420` `if hadPrev && delta == 0 && sameSnapshot(obs, prev) {` | — | — | 무편집 갈래 |
| B21 | `:421` `if err := tx.Commit(); err != nil {` | — | — | 무편집 갈래 |
| B22 | `:437` `if err := upsertFillSnapshot(ctx, tx, scope, obs, now); err != nil {` | — | — | 무편집 갈래 |
| B23 | `:441` `if correction {` | — | — | 무편집 갈래 |
| B24 | `:442` `if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{` | — | — | 무편집 갈래 |
| B25 | `:464` `if delta > 0 {` | — | — | 무편집 갈래 |
| B26 | `:465` `if _, err := tx.ExecContext(ctx,` | — | — | 무편집 갈래 |
| B27 | `:504` `if err != nil {` | — | — | 무편집 갈래 |
| B28 | `:512` `if obs.Terminal && locallyOwned {` | — | — | 무편집 갈래 |
| B29 | `:516` `if err != nil {` | — | — | 무편집 갈래 |
| B30 | `:526` `if locallyOwned {` | — | — | 무편집 갈래 |
| B31 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | — | — | 무편집 갈래 |
| B32 | `:527` `if err := j.applyRiskBucketFillInTx(ctx, tx, applied); err != nil {` | — | — | 무편집 갈래 |
| B33 | `:530` `if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, applied); err != nil {` | — | — | 무편집 갈래 |
| B34 | `:533` `if err := applyWeeklyReservationLifecycleInTx(ctx, tx, applied); err != nil {` | — | — | 무편집 갈래 |
| B35 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | — | — | 무편집 갈래 |
| B36 | `:547` `if locallyOwned {` | — | — | 무편집 갈래 |
| B37 | `:548` `if err := j.runApplyHooks(ctx, tx, applied); err != nil {` | — | — | 무편집 갈래 |
| B38 | `:553` `if err := tx.Commit(); err != nil {` | — | — | 무편집 갈래 |

## Calls and live bindings

AST `calls` 목록 참조. Live config binding: 없음.

## State mutations and fallbacks

- 아래 「a130 편집 계획」 의 근거 자리 참조.

## a130 편집 계획 (2판)

B36 앞에 `var refusals []TradeOutcomeRefusal`, B37(:548) 호출의 ctx 인자를 `withOutcomeRefusals(ctx, &refusals)` 로, B38(:553) 커밋 성공 뒤 `j.reportOutcomeRefusals(refusals)` 한 줄(새 메서드, 새 파일 — 싱크 호출마다 recover). 커밋 실패 갈래(:554)와 그 밖 37 분기 무변경.

## Safety conclusion

- Safe edit boundary: 위 계획만.
- High-risk impact: yes(체결 반영 트랜잭션) — 커밋 뒤 보고 한 줄 / 호출 모양만.

## 편집 뒤 (a130 4판)

- 편집 뒤 `313`–`563` · `source_sha256` 92bc44f947eead2a… — branches 38 · returns 26
  (편집 전 38 · 26, 편집 전 AST 는 `ast-base.json`). 위 분기 표의 좌표는 편집 전, 아래 BTM 은 편집 뒤 좌표.
