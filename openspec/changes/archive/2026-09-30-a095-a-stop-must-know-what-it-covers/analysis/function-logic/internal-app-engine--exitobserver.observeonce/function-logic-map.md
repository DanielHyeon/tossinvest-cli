# Function Logic Map: `ExitObserver.ObserveOnce`

- Source: `internal/app/engine/exitloop.go` (`413`–`470`)
- Qualified: `ExitObserver.ObserveOnce`
- AST evidence: `ast.json` (`source_sha256` 2d34b5c57f25a2c8…)
- Risk scan: `risk-pattern-report.md`
- 분기 8 · return 5 · 호출 16

**역할.** exit 관측 한 사이클. 작업 집합을 만들고 시세를 읽은 뒤 포지션마다 판정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.workingSet`의 답 | 판정할 포지션 | 원장 | B2 창의 return(오류) |
| `o.observe`의 답 | 시세 | 브로커 | B4 창 — `o.checkOutage` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:417` `if o.opts.SLO != nil && o.opts.SLO.FillDetectionBehind() {` | `o.checkOutage`, `o.opts.SLO.FillDetectionBehind`, `o.workingSet` | :423 | 예 |
| B2 | if | `:427` `if err != nil {` | — | :429 | 아니오 |
| B3 | if | `:431` `if len(states) == 0 {` | `len`, `o.clk.Now`, `o.observe`, `symbolsOf` | :439 | 예 |
| B4 | if | `:442` `if err != nil {` | `len`, `o.checkOutage`, `o.clk.Now` | :445 | 예 |
| B5 | range | `:451` `for _, state := range states {` | `strings.ToUpper`, `strings.TrimSpace` | — | 예 |
| B6 | if | `:453` `if !ok {` | — | — | 예 |
| B7 | if | `:459` `if !o.quoteUsable(quote) {` | `o.quoteUsable` | — | 예 |
| B8 | if | `:465` `if err := o.judge(ctx, state, quote, observation, &cycle); err != nil && cycle.Err == nil {` | `o.judge` | :469 | 예 |

## Calls and live bindings

호출 순서(`ast.json` 좌표): `o.workingSet` `:426` → `o.observe` `:441` → `o.judge` `:465`.

결과는 `ExitCycle`이다 — 오류를 돌려주지 않고 `cycle.Err`에 담는다(B2 · B8 창).

## State mutations and fallbacks

사이클 계수 · 판정(`o.judge` 경유).

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수를 바꾸지 않는다.** `o.workingSet`이 `o.observe` · `o.judge`보다 **먼저** 불린다 — 작업 집합 안의 동기 알림은 이 사이클의 모든 손절 판정 앞에 선다. 결정 (1)이 그 자리에 critical을 두지 않게 한다.
- **High-risk impact**: yes — 손절 판정 루프다.
