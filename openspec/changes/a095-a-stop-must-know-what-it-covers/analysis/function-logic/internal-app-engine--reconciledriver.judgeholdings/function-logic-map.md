# Function Logic Map: `ReconcileDriver.judgeHoldings`

- Source: `internal/app/engine/adoption.go` (`99`–`185`)
- Qualified: `ReconcileDriver.judgeHoldings`
- AST evidence: `ast.json` (`source_sha256` 26a0601d9987c7dc…)
- Risk scan: `risk-pattern-report.md`
- 분기 16 · return 0 · 호출 24

**역할.** 안정 스냅샷의 보유를 게이트에 통과시키고, 편입하거나 무관리로 모은다. **`checkExternalIncrease` · `checkEngineOpenedIncrease`와 reconcile 쪽 `alertUnmanaged`의 유일한 호출자다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.ExitEligible()` | 진입 결정 또는 편입 기록이 있나 | 원장 `positions` | B7 창 — 참이면 수량 증가만 검사하고 `continue` |
| `p.Adopted()` | 편입 기록이 있나 | 원장 `positions.adoption_id` | B8 — 참이면 `checkExternalIncrease`, 거짓(엔진 개설)이면 `checkEngineOpenedIncrease`(10판) |
| `d.blocked` · `fresh` | 전이 상태 | RECONCILE 추적기 · 스냅샷 나이 | 무알림 `continue` |
| `d.opts.Adoption` | 편입 설정 | 런타임 config | exclude · off∧미지정 → `unmanaged` |
| `d.adopt`의 결과 | 후보별 편입됨 · 시도 실패 · 연기 | 위 함수 | 편입됨이 아니면 `unmanaged`, 결과는 `alertUnmanaged`에 전달 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:102` `if stale <= 0 {` | `d.clk.Now`, `snapshot.Age` | — | 예 |
| B2 | range | `:111` `for _, holding := range snapshot.Holdings {` | `strings.ToLower`, `strings.TrimSpace` | — | 예 |
| B3 | if | `:113` `if market == "" {` | `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace` | — | 아니오 |
| B4 | if | `:117` `if symbol == "" \|\| market == "" \|\| isZeroQuantity(holding.Quantity) {` | `d.opts.Journal.CurrentPosition`, `isZeroQuantity` | — | 아니오 |
| B5 | if | `:122` `if err != nil {` | — | — | 아니오 |
| B6 | if | `:128` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | `isZeroQuantity` | — | 예 |
| B7 | if | `:132` `if p.ExitEligible() {` | `p.ExitEligible` | — | 예 |
| B8 | if | `:135` `if p.Adopted() {` | `d.checkExternalIncrease`, `p.Adopted` | — | 예 |
| B9 | else | `:137` `} else {` | `d.checkEngineOpenedIncrease` | — | 예 |
| B10 | if | `:144` `if d.blocked(market, symbol) {` | `d.blocked` | — | 예 |
| B11 | if | `:147` `if !fresh {` | — | — | 예 |
| B12 | if | `:155` `if d.opts.Adoption.Excludes(symbol) {` | `append`, `d.opts.Adoption.Excludes` | — | 예 |
| B13 | if | `:163` `if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {` | `append`, `d.adopt`, `d.opts.Adoption.Included` | — | 예 |
| B14 | range | `:171` `for _, c := range candidates {` | — | — | 예 |
| B15 | if | `:172` `if results[c.position.ID] != adoptAdopted {` | `append` | — | 예 |
| B16 | range | `:180` `for _, p := range unmanaged {` | `d.alertUnmanaged` | — | 예 |

## Calls and live bindings

`d.opts.Journal.CurrentPosition` · `d.checkExternalIncrease` · `d.checkEngineOpenedIncrease`(10판) · `d.blocked` · `d.opts.Adoption.Excludes` · `d.opts.Adoption.Included` · `d.adopt` · `d.alertUnmanaged(ctx, p, results[p.ID])`.

결과값이 없다 — 오류를 돌려주지 않는다. `CurrentPosition` 오류는 그 보유를 건너뛰고(B5), 편입 쪽 오류는 `adopt`가 `cycle.Err`에 담는다.

## State mutations and fallbacks

`cycle.Unmanaged` 계수 · 편입(`d.adopt` 경유) · 알림(`alertUnmanaged` · 수량 증가 검사 경유).

## Safety conclusion

- **Safe edit boundary**: **10판 편집(구현 로트)**: B8 에 else 갈래(엔진 개설 포지션의 수량 증가 검사 — 결정 (3)(i), Q3)를 더했고, `adopt`의 결과를 id 집합이 아니라 후보별 결과로 받아 `alertUnmanaged`에 넘긴다(Q1 · Q2(c)). 전이 상태 · 게이트 순서는 그대로다. 후보가 아닌 무관리 보유(제외 · 꺼짐 · 설정 거부)는 결과 map 에 없어 영값(연기)을 받지만 그 셋의 조건 칸은 결과를 읽지 않는다.
- **High-risk impact**: yes — 편입과 무관리 보고의 입구다.
