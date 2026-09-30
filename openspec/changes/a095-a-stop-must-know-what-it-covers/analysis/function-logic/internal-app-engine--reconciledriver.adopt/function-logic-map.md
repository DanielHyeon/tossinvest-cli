# Function Logic Map: `ReconcileDriver.adopt`

- Source: `internal/app/engine/adoption.go` (`202`–`249`)
- Qualified: `ReconcileDriver.adopt`
- AST evidence: `ast.json` (`source_sha256` 26a0601d9987c7dc…)
- Risk scan: `risk-pattern-report.md`
- 분기 8 · return 4 · 호출 11

**역할.** 후보를 한 번의 묶음 시세 읽기로 값 매기고 편입할 수 있는 것을 편입한다. **후보별 결과**(편입됨 · 시도 실패 · 연기)를 돌려준다(10판).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.observeCandidates`의 답 | 후보 시세 | 브로커 시세 경로 | 오류면 빈 map — 모든 후보가 연기 |
| `quotes[key]` | 종목별 관측 | 위 읽기 | 없으면 `cycle.Deferred`, 그 후보는 연기 |
| 시세 나이 | `PriceStaleness` | config · 기본값 | 넘으면 남은 후보 전부 연기로 return |
| `d.adoptOne`의 답 | 편입 시도 | 아래 함수 | 참이면 `adoptAdopted`, 거짓이면 `adoptFailed`(10판) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:205` `if len(candidates) == 0 {` | `d.observeCandidates`, `len` | :206 | 예 |
| B2 | if | `:210` `if err != nil {` | `len` | — | 예 |
| B3 | if | `:212` `if cycle.Err == nil {` | — | :215 | 예 |
| B4 | if | `:219` `if bound <= 0 {` | — | — | 예 |
| B5 | range | `:222` `for _, c := range candidates {` | `adoptionQuoteKey` | — | 예 |
| B6 | if | `:225` `if !ok {` | — | — | 예 |
| B7 | if | `:231` `if age := d.clk.Now().Sub(readAt); age > bound {` | `Sub`, `adoptedCount`, `d.clk.Now`, `d.logDeferred`, `fmt.Sprintf`, `len` | :238 | 예 |
| B8 | if | `:240` `if d.adoptOne(ctx, c, observed) {` | `d.adoptOne` | :248 | 예 |

## Calls and live bindings

`d.observeCandidates` · `adoptionQuoteKey` · `d.logDeferred`(묵음 창) · `adoptedCount`(10판) · `d.adoptOne`.

결과는 후보별 결과 map(`adoptResult` — 편입됨 · 시도 실패 · 연기, 10판)이다 — 오류를 돌려주지 않는다. 시세 읽기 오류는 `cycle.Err`에 담고 빈 map을 돌려준다(없는 후보 = 영값 연기).

## State mutations and fallbacks

편입(`d.adoptOne` 경유) · `cycle.Deferred` · `cycle.Adopted` 계수 · 결과 map.

## Safety conclusion

- **Safe edit boundary**: **10판 편집**: 결과 형태만 바뀌었다 — `map[string]bool`(편입됨) → `map[string]adoptResult`. 시도 실패 자리에 `adoptFailed`를 적고, 묵음 창의 연기 셈은 `adoptedCount`(편입된 수 — 옛 `len(adopted)`와 같은 값)로 셈을 보존한다. 분기 조건 · 순서 무변화. 영값이 연기인 것은 의도다 — 중간 반환이 남긴 후보는 시도되지 않았다.
- **High-risk impact**: yes — 편입의 유일한 입구다.
