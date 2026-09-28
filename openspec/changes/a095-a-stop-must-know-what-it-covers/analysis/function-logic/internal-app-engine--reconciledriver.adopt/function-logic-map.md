# Function Logic Map: `ReconcileDriver.adopt`

- Source: `internal/app/engine/adoption.go` (`172`–`218`)
- Qualified: `ReconcileDriver.adopt`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 8 · return 4 · 호출 11

**역할.** 후보를 한 번의 묶음 시세 읽기로 값 매기고 편입할 수 있는 것을 편입한다. 편입된 id 집합을 돌려준다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.observeCandidates`의 답 | 후보 시세 | 브로커 시세 경로 | B2 — 오류면 빈 집합을 돌려준다 |
| `quotes[key]` | 종목별 관측 | 위 읽기 | B6 — 없으면 `cycle.Deferred`, 그 후보는 편입되지 않는다 |
| 시세 나이 | `PriceStaleness` | config · 기본값(B4) | B7 — 넘으면 남은 후보 전부를 편입하지 않고 return |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:175` `if len(candidates) == 0 {` | `d.observeCandidates`, `len` | :176 | 예 |
| B2 | if | `:180` `if err != nil {` | `len` | — | 아니오 |
| B3 | if | `:182` `if cycle.Err == nil {` | — | :185 | 아니오 |
| B4 | if | `:189` `if bound <= 0 {` | — | — | 예 |
| B5 | range | `:192` `for _, c := range candidates {` | `adoptionQuoteKey` | — | 예 |
| B6 | if | `:195` `if !ok {` | — | — | 예 |
| B7 | if | `:201` `if age := d.clk.Now().Sub(readAt); age > bound {` | `Sub`, `d.clk.Now`, `d.logDeferred`, `fmt.Sprintf`, `len` | :208 | 예 |
| B8 | if | `:210` `if d.adoptOne(ctx, c, observed) {` | `d.adoptOne` | :217 | 예 |

## Calls and live bindings

`d.observeCandidates`(B1 뒤) · `adoptionQuoteKey` · `d.logDeferred`(B7 창) · `d.adoptOne`(B8).

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

편입(`d.adoptOne` 경유) · `cycle.Deferred` · `cycle.Adopted` 계수.

## Safety conclusion

- **Safe edit boundary**: **5판: 결과 형태가 편집 경계 안이다(r4 R4-1, Manager 처분).** 오늘 이 함수는 편입된 id 집합만 돌려주고, 집합에 없는 후보는 호출자 `judgeHoldings` B14가 무관리로 모은다 — B2(시세 읽기 오류) · B6(관측 없음) · B7(관측 묵음)로 **연기된** 후보와 B8(`d.adoptOne` 거짓)의 **시도 실패**가 한 사유(`alertUnmanaged` B5)로 합쳐진다. critical 요구(시도 실패만)와 열린 Q2(c)(연기분의 등급)의 어느 답도 막지 않으려면 후보별 결과(편입 · 연기 · 시도 실패)를 호출자에 전해야 한다. 형태는 구현 로트가 정한다(design D1 「후보별 결과와 억제 키」).
- **High-risk impact**: yes — 편입의 유일한 입구다.
