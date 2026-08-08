# Function Logic Map: `EvaluateLadder`

- Source: `internal/exitpolicy/ladder.go` (`307`–`481`)
- Qualified: `EvaluateLadder`
- AST evidence: `ast.json` (`source_sha256` ee9c91aa8a5d8f51…)
- Risk scan: `risk-pattern-report.md`
- 분기 32 · return 23 · 호출 54

**역할.** 관측 하나를 사다리 전이로 바꾼다. **모든 선이 `EntryPrice`에서 나온다** — 수익률도 잠금가도.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `in.EntryPrice` | 고정된 진입가 | `exit_states.entry_price` | **`:329` `positive("entry price")`. 평단이 아니다** |
| `in.ObservedPrice` | 현재가 | 관측 루프 | `percentOf(probe, entry)`의 분자 |
| `in.State` | 현 사다리 상태 | 원장 | 정책 version/digest 대조 |
| `in.Policy` | 불변 사다리 표 | config | `StopPct`는 **진입가 대비** 백분율(`:99-100`) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

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

`positive`(`:329` 등) · `fraction` · `percentOf`(`:358`) · `lockPrice`(`:387`) · `PolicyIdentityOf`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다 — 순수 계산이고 전이 값을 돌려준다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** `lockPrice(entry, pct)`(`:503-509`)가 `entry × (1 + pct/100)`이고 그것이 설계다. 평단을 여기 넣으면 이미 보고된 R의 분모가 바뀐다 — `checkExternalIncrease` 주석이 금지한 바로 그것이다. a095가 손절을 올린다면 **이 함수의 산출을 바꾸는 것이 아니라 `entry_price` 자체를 다시 세우는 경로**로 해야 하고, 그것은 `resetExitStateForReadoptTx`의 형제다.
- **High-risk impact**: yes — 손절가와 익절선을 계산한다.
