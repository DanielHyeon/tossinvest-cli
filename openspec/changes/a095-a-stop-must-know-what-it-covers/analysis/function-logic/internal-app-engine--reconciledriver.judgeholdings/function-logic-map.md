# Function Logic Map: `ReconcileDriver.judgeHoldings`

- Source: `internal/app/engine/adoption.go` (`75`–`156`)
- Qualified: `ReconcileDriver.judgeHoldings`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 15 · return 0 · 호출 23

**역할.** 안정 스냅샷의 보유를 게이트에 통과시키고, 편입하거나 무관리로 모은다. **`checkExternalIncrease`와 reconcile 쪽 `alertUnmanaged`의 유일한 호출자다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `p.ExitEligible()` | 진입 결정 또는 편입 기록이 있나 | 원장 `positions` | B7 창 — 참이면 `continue` |
| `p.Adopted()` | 편입 기록이 있나 | 원장 `positions.adoption_id` | B8 창 — 참일 때만 `checkExternalIncrease` |
| `d.blocked` · `fresh` | 전이 상태 | RECONCILE 추적기 · 스냅샷 나이 | B9 · B10 — 무알림 `continue` |
| `d.opts.Adoption` | 편입 설정 | 런타임 config | B11(exclude) · B12(off∧미지정) → `unmanaged` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:78` `if stale <= 0 {` | `d.clk.Now`, `snapshot.Age` | — | 예 |
| B2 | range | `:87` `for _, holding := range snapshot.Holdings {` | `strings.ToLower`, `strings.TrimSpace` | — | 예 |
| B3 | if | `:89` `if market == "" {` | `strings.ToLower`, `strings.ToUpper`, `strings.TrimSpace` | — | 아니오 |
| B4 | if | `:93` `if symbol == "" \|\| market == "" \|\| isZeroQuantity(holding.Quantity) {` | `d.opts.Journal.CurrentPosition`, `isZeroQuantity` | — | 아니오 |
| B5 | if | `:98` `if err != nil {` | — | — | 아니오 |
| B6 | if | `:104` `if p.State == journal.PositionClosed \|\| isZeroQuantity(p.Quantity) {` | `isZeroQuantity` | — | 예 |
| B7 | if | `:108` `if p.ExitEligible() {` | `p.ExitEligible` | — | 예 |
| B8 | if | `:109` `if p.Adopted() {` | `d.checkExternalIncrease`, `p.Adopted` | — | 예 |
| B9 | if | `:116` `if d.blocked(market, symbol) {` | `d.blocked` | — | 예 |
| B10 | if | `:119` `if !fresh {` | — | — | 예 |
| B11 | if | `:127` `if d.opts.Adoption.Excludes(symbol) {` | `append`, `d.opts.Adoption.Excludes` | — | 예 |
| B12 | if | `:135` `if !d.opts.Adoption.Enabled && !d.opts.Adoption.Included(symbol) {` | `append`, `d.adopt`, `d.opts.Adoption.Included` | — | 예 |
| B13 | range | `:143` `for _, c := range candidates {` | — | — | 예 |
| B14 | if | `:144` `if !adopted[c.position.ID] {` | `append` | — | 예 |
| B15 | range | `:152` `for _, p := range unmanaged {` | `d.alertUnmanaged` | — | 예 |

## Calls and live bindings

`d.opts.Journal.CurrentPosition` · `d.checkExternalIncrease`(B8 창) · `d.blocked`(B9) · `d.opts.Adoption.Excludes`(B11) · `d.opts.Adoption.Included`(B12) · `d.adopt` · `d.alertUnmanaged`(B15 창).

결과값이 없다 — 오류를 돌려주지 않는다. `CurrentPosition` 오류는 그 보유를 건너뛰고(B5), 편입 쪽 오류는 `adopt`가 `cycle.Err`에 담는다.

## State mutations and fallbacks

`cycle.Unmanaged` 계수 · 편입(`d.adopt` 경유) · 알림(`alertUnmanaged` 경유).

## Safety conclusion

- **Safe edit boundary**: **3판의 사실 셋이 이 함수에서 나온다.** ① B7 창은 `ExitEligible`이면 `continue`하고 B8만 `checkExternalIncrease`를 부른다 — **엔진이 직접 연 포지션(편입 기록 없음)의 수량 증가는 어디서도 검사되지 않는다**(결정 (3)(i)의 대상). ② 편입 기록이 없는 보유는 B8에 오지 않으므로 `checkExternalIncrease` B2가 받는 입력이 아니다 — 2판 R2-B2의 전제가 거짓이었다. ③ 무관리 보유는 B11(exclude) · B12(off∧미지정) · B14(편입 실패)로 모여 B15에서 알려진다 — 운영자가 고른 상태(B11 · B12)와 고르지 않은 상태(B14)가 **다른 분기**로 이미 갈린다.
- **High-risk impact**: yes — 편입과 무관리 보고의 입구다.
