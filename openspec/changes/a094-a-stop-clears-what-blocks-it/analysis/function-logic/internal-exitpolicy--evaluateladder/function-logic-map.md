# Function Logic Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go` (`307`–`481`)
- Qualified: `EvaluateLadder`
- AST evidence: `ast.json` (`source_sha256` ee9c91aa8a5d8f51…)
- Risk scan: `risk-pattern-report.md`
- 분기 32 · return 23 · 호출 54

**역할.** 관측 하나를 사다리 전이로 바꾼다. **B26 `:441`이 a094 3판의 출발점이다** — 발의가 무장된 채면 손절 조건이 성립해도 빈 전이를 돌려준다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.State.PendingAction` | 무장된 발의 | `exit_states.pending_action` | **B26 — `ActionLadderStop`이면 `SuppressedPending`으로 빈 전이** |
| `in.EntryPrice` | 고정 진입가 | `exit_states.entry_price` | `:329`. 모든 선의 원천 |
| `in.ObservedPrice` | 현재가 | 관측 루프 | B25 `:439`에서 baseline과 비교 |
| `in.Baseline` | 유효 손절 | `exit_states.baseline_price` | B25의 비교 대상 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy,filldetect}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:308` `if err := in.Policy.Validate(); err != nil {` | `in.Policy.Validate` | :309 | 예 |
| B2 | if | `:311` `if in.State.PolicyID != in.Policy.PolicyID {` | `fmt.Sprintf`, `in.Policy.Identity`, `refusal` | :314 | 예 |
| B3 | if | `:318` `if err != nil {` | — | :319 | 아니오 |
| B4 | if | `:321` `if in.State.PolicyVersion != "" && in.State.PolicyVersion != identity.Version {` | `fmt.Sprintf`, `refusal` | :322 | 예 |
| B5 | if | `:325` `if in.State.PolicyDigest != "" && in.State.PolicyDigest != identity.Digest {` | `fmt.Errorf`, `positive` | :326 | 예 |
| B6 | if | `:330` `if err != nil {` | `positive` | :331 | 아니오 |
| B7 | if | `:334` `if err != nil {` | `positive` | :335 | 아니오 |
| B8 | if | `:338` `if err != nil {` | `positive` | :339 | 아니오 |
| B9 | if | `:342` `if err != nil {` | — | :343 | 아니오 |
| B10 | if | `:345` `if _, err := fraction("taken ratio total", in.State.TakenRatioTotal); err != nil {` | `fraction` | :346 | 예 |
| B11 | if | `:348` `if in.State.ActivatedRung < NoRung \|\| in.State.ActivatedRung >= len(in.Policy.Rungs) {` | `fmt.Sprintf`, `len`, `refusal` | :349 | 예 |
| B12 | if | `:355` `if observed.Cmp(probe) > 0 {` | `formatPrice`, `formatRMultiple`, `observed.Cmp`, `percentOf` | — | 예 |
| B13 | range | `:374` `for i, rung := range in.Policy.Rungs {` | `parseRat` | — | 예 |
| B14 | if | `:376` `if err != nil {` | `err.Error`, `refusal` | :377 | 아니오 |
| B15 | if | `:379` `if i > newIndex && returnPct.Cmp(target) >= 0 {` | `returnPct.Cmp` | — | 예 |
| B16 | if | `:386` `if newIndex > NoRung {` | `lockPrice` | — | 예 |
| B17 | if | `:388` `if err != nil {` | `BuildCandidates` | :389 | 아니오 |
| B18 | if | `:392` `if newIndex == len(in.Policy.Rungs)-1 && in.Policy.RunnerTrailPct != "" {` | `len`, `parseRat` | — | 예 |
| B19 | if | `:394` `if err != nil {` | `ComputeProtectedStop`, `Mul`, `Quo`, `Sub`, `append`, `big.NewRat`, `err.Error`, `formatPrice`, `new`, `refusal` | :395 | 아니오 |
| B20 | if | `:404` `if err != nil {` | `parseRat` | :405 | 아니오 |
| B21 | if | `:408` `if err != nil {` | `err.Error`, `price.Cmp`, `refusal` | :409 | 아니오 |
| B22 | if | `:420` `if newIndex > in.State.ActivatedRung {` | — | — | 예 |
| B23 | if | `:427` `if in.State.Completed {` | `parseRat` | :430 | 예 |
| B24 | if | `:436` `if err != nil {` | `err.Error`, `refusal` | :437 | 아니오 |
| B25 | if | `:439` `if observed.Cmp(baseline) < 0 {` | `observed.Cmp` | — | 예 |
| B26 | if | `:441` `if in.State.PendingAction == ActionLadderStop {` | `rungLabel` | :443, :448 | 예 |
| B27 | if | `:453` `if out.RungPromotedTo == NoRung {` | — | :454 | 예 |
| B28 | switch | `:457` `switch {` | — | — | — |
| B29 | case | `:458` `case newIndex == len(in.Policy.Rungs)-1 && in.Policy.FinalTakeFull:` | `len`, `rungLabel` | — | 예 |
| B30 | case | `:461` `case isPositive(rung.PartialRatio):` | `isPositive`, `rungLabel` | — | 예 |
| B31 | case | `:466` `default:` | — | :469 | 예 |
| B32 | if | `:475` `if in.State.PendingAction != ActionNone {` | — | :480 | 예 |

## Calls and live bindings

`parseRat` · `positive` · `percentOf`(`:358`) · `lockPrice`(`:387`) · `rungLabel`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 순수 계산이다. 부작용은 호출자가 만든다.

## Safety conclusion

- **Safe edit boundary**: **a094는 이 함수를 바꾸지 않는다.** B26의 억제는 옳다 — 살아 있을지 모르는 매도 위에 두 번째를 얹지 않는 것이다. **바꾸는 것은 그 억제가 영원히 풀리지 않는 것**이고, 그것은 발의를 해제하는 쪽(`ResolveExitProposal`)에서 고친다.
- **High-risk impact**: yes — 손절 발의의 유무를 정한다.
