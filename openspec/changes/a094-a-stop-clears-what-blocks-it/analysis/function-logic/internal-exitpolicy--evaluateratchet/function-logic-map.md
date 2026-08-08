# Function Logic Map: `EvaluateRatchet`

- Source: `internal/exitpolicy/ratchet.go` (`335`–`453`)
- Qualified: `EvaluateRatchet`
- AST evidence: `ast.json` (`source_sha256` 9aaa342d5ff8f142…)
- Risk scan: `risk-pattern-report.md`
- 분기 22 · return 14 · 호출 32

**역할.** RATCHET 정책의 같은 판정. **B17 `:423`이 LADDER B26과 같은 억제다** — 정책 종류가 달라도 동결 기전은 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.PendingAction` | 무장된 발의 | `exit_states.pending_action` | **B17 — `ActionBaselineBreach`면 억제** |
| `in.ObservedPrice` | 현재가 | 관측 루프 | B16 `:422` |
| `in.Baseline` | 유효 손절 | `exit_states.baseline_price` | B16의 비교 대상 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:337` `if in.Config != nil {` | — | — | 예 |
| B2 | if | `:340` `if err := cfg.Validate(); err != nil {` | `cfg.Validate`, `riskOf` | :341 | 예 |
| B3 | if | `:344` `if err != nil {` | `positive` | :345 | 예 |
| B4 | if | `:348` `if err != nil {` | `positive` | :349 | 예 |
| B5 | if | `:352` `if err != nil {` | `positive` | :353 | 예 |
| B6 | if | `:356` `if err != nil {` | `positive` | :357 | 예 |
| B7 | if | `:360` `if err != nil {` | `fraction` | :361 | 예 |
| B8 | if | `:364` `if err != nil {` | — | :365 | 예 |
| B9 | if | `:371` `if observed.Cmp(probe) > 0 {` | `Quo`, `Sub`, `new`, `observed.Cmp`, `ratchetCandidate` | — | 예 |
| B10 | if | `:378` `if err != nil {` | `formatPrice`, `formatRMultiple` | :379 | 아니오 |
| B11 | if | `:394` `if levelCandidate != "" {` | — | — | 예 |
| B12 | if | `:400` `if level.Rank() >= LevelBreakeven.Rank() {` | `BuildCandidates`, `ComputeProtectedStop`, `LevelBreakeven.Rank`, `level.Rank` | — | 예 |
| B13 | if | `:405` `if err != nil {` | `parseRat` | :406 | 아니오 |
| B14 | if | `:409` `if err != nil {` | `composed.Cmp`, `err.Error`, `parseRat`, `refusal` | :410 | 아니오 |
| B15 | if | `:419` `if err != nil {` | `err.Error`, `refusal` | :420 | 아니오 |
| B16 | if | `:422` `if observed.Cmp(baseline) < 0 {` | `observed.Cmp` | — | 예 |
| B17 | if | `:423` `if in.PendingAction == ActionBaselineBreach {` | `string` | :425, :433 | 예 |
| B18 | if | `:438` `if wantsPartial {` | — | — | 예 |
| B19 | switch | `:439` `switch {` | — | — | — |
| B20 | case | `:440` `case taken.Sign() > 0:` | `taken.Sign` | — | 예 |
| B21 | case | `:442` `case in.PendingAction != ActionNone:` | — | — | 예 |
| B22 | case | `:444` `default:` | `string` | :452 | 예 |

## Calls and live bindings

`parseRat` · `positive` · `percentOf` · 부분 청산 비율 계산(B18~B20).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 순수 계산이다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** LADDER와 같은 이유이며, **해동 경로가 두 정책 모두에 적용되어야 한다**는 근거가 이 함수다. 3판 이전 문서가 LADDER만 보고 있었다.
- **High-risk impact**: yes — 손절 발의의 유무를 정한다.
