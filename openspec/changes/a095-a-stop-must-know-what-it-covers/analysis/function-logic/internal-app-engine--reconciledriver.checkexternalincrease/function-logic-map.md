# Function Logic Map: `ReconcileDriver.checkExternalIncrease`

- Source: `internal/app/engine/adoption.go` (`441`–`472`)
- Qualified: `ReconcileDriver.checkExternalIncrease`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 3 · 호출 5

**역할.** 편입된 포지션의 수량이 편입 기록보다 늘었는지 보고 알린다. 주석이 t0 동결을 의도적 설계(A8)로 선언한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.grown[p.ID]` | 이미 알렸나 | 프로세스 메모리 map | B1 창의 return |
| `AdoptionOf(p.ID)` | 편입 기록 | 원장 | B2 창의 return. 호출자 가드(`judgeHoldings` B8)로 편입된 포지션만 오고 `positions.adoption_id`는 `position_adoptions(id)`를 참조하므로, 여기 오는 입력은 조회 오류다 |
| `p.Quantity` 대 `adoption.Quantity` | 현재 수량 대 편입 수량 | 스냅샷 대 원장 | B3 — 늘지 않았으면 return |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:442` `if d.grown[p.ID] {` | `d.opts.Journal.AdoptionOf` | :443 | 아니오 |
| B2 | if | `:446` `if err != nil {` | `riskcalc.CompareDecimal` | :447 | 아니오 |
| B3 | if | `:450` `if err != nil \|\| cmp <= 0 {` | `d.alert`, `d.label`, `string` | :451 | 예 |

## Calls and live bindings

`d.opts.Journal.AdoptionOf`(B1 뒤) · `riskcalc.CompareDecimal`(B2 뒤) · `d.alert` · `d.label`.

결과값이 없다 — 오류를 돌려주지 않는다. `AdoptionOf` · `CompareDecimal`의 오류에서 조용히 반환하고(B2 · B3 창 return), `d.alert`는 `Notify`의 오류를 로그로만 남긴다.

## State mutations and fallbacks

`d.grown[p.ID] = true`(메모리) · 알림 1건(키 `…|grown|<posID>` — 수량이 없다). 원장의 exit state는 건드리지 않는다.

## Safety conclusion

- **Safe edit boundary**: **2판 FLM의 「B2 — 엔진이 직접 연 포지션과 미편입 보유가 여기로 온다」는 거짓이었다** — `judgeHoldings` B8 창이 편입된 포지션만 부른다. 결정 (3)이 R2-B2를 삭제했으므로 3판은 B2를 바꾸지 않는다. 알림 본문 스스로 *"늘어난 수량은 원래 수량 기준으로 산정된 손절의 보호를 받는다"*라고 쓴다 — 이 사실은 「무보호」가 아니며 그 종류·등급은 Q4다.
- **High-risk impact**: yes — 편입 후 수량 증가를 알리는 유일한 자리다.
