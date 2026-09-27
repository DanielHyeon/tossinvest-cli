# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go` (`493`–`612`)
- Qualified: `ExitObserver.workingSet`
- AST evidence: `ast.json` (`source_sha256` 522d5d81c4992c57…)
- Risk scan: `risk-pattern-report.md`
- 분기 22 · return 3 · 호출 28

**역할.** exit 관측이 판정할 포지션 집합을 만든다. 적격하지 않은 보유는 알리고 건너뛴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.ExitEligible()` | 진입 결정 또는 편입 기록 | 원장 | B6 — 거짓이면 `o.alertUnmanaged` 후 `continue` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:495` `if err != nil {` | `o.opts.Journal.OpenExitStateResults` | :496 | 아니오 |
| B2 | if | `:499` `if err != nil {` | `len`, `make` | :500 | 아니오 |
| B3 | range | `:503` `for _, result := range stateResults {` | — | — | 예 |
| B4 | range | `:508` `for _, p := range positions {` | — | — | 예 |
| B5 | if | `:509` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | `isZeroQuantity` | — | 예 |
| B6 | if | `:512` `if !p.ExitEligible() {` | `int64`, `o.alertUnmanaged`, `p.ExitEligible` | — | 예 |
| B7 | if | `:525` `if !ok {` | `o.openState` | — | 예 |
| B8 | if | `:527` `if err != nil {` | — | — | 아니오 |
| B9 | if | `:528` `if cycle.Err == nil {` | — | — | 아니오 |
| B10 | if | `:533` `if opened.PositionID == "" {` | — | — | 예 |
| B11 | if | `:542` `if result.Corruption != nil {` | `o.opts.Journal.QuarantineExitSnapshot`, `result.Corruption.Error` | — | 예 |
| B12 | if | `:545` `if qerr != nil {` | — | — | 아니오 |
| B13 | if | `:546` `if cycle.Err == nil {` | `append`, `fmt.Errorf`, `o.announceQuarantine` | — | 아니오 |
| B14 | if | `:556` `if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p.InstanceSeq); qerr != nil {` | `o.opts.Journal.ActiveExitSnapshotQuarantine` | — | 예 |
| B15 | else | `:561` `} else if active && !q.NeedsReJudgement() {` | `q.NeedsReJudgement` | — | 예 |
| B16 | if | `:557` `if cycle.Err == nil {` | — | — | 아니오 |
| B17 | if | `:561` `} else if active && !q.NeedsReJudgement() {` | `append`, `fmt.Errorf`, `q.NeedsReJudgement` | — | 예 |
| B18 | else | `:565` `} else if active {` | — | — | 예 |
| B19 | if | `:565` `} else if active {` | `managedPolicyIdentity`, `o.log`, `p.Adopted` | — | 예 |
| B20 | if | `:589` `if identityErr != nil {` | `identityErr.Error`, `o.opts.Journal.QuarantineExitSnapshot` | — | 예 |
| B21 | if | `:592` `if qerr != nil {` | — | — | 아니오 |
| B22 | if | `:593` `if cycle.Err == nil {` | `append`, `fmt.Errorf`, `o.announceQuarantine` | :611 | 아니오 |

## Calls and live bindings

`o.opts.Journal.Positions` · `o.opts.Journal.OpenExitStateResults` · `o.alertUnmanaged`(B6 창) · `o.openState`(B7) · `o.opts.Journal.QuarantineExitSnapshot`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`cycle.Unmanaged` 계수 · 알림(B6) · 격리 기록.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B6 창의 `o.alertUnmanaged`는 exit goroutine 안의 동기 호출이며, 결정 (1)이 그 자리를 normal로 둔다. B6에는 전이 상태 판정이 없다(reconcile 쪽 `judgeHoldings` B9 · B10과 다르다) — normal이므로 진입 차단에 닿지 않는다.
- **High-risk impact**: yes — 손절 판정 앞의 작업 집합이다.
