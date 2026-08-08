# Function Logic Map: `ReconcileDriver.checkExternalIncrease`

- Source: `internal/app/engine/adoption.go` (`441`–`472`)
- Qualified: `ReconcileDriver.checkExternalIncrease`
- AST evidence: `ast.json` (`source_sha256` f121aba90cd05c31…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 3 · 호출 5

**역할.** 편입된 포지션의 수량이 편입 기록보다 늘었는지 보고 알린다. 주석이 t0 동결을 **의도적 설계(A8)**로 선언한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.grown[p.ID]` | 이미 알렸는가 | **프로세스 메모리 map** | **B1 — 재시작하면 초기화된다. 한 프로세스에서 한 번만 운다** |
| `AdoptionOf(p.ID)` | 편입 기록 | 원장 | **B2 — 없으면 조용히 반환한다.** 엔진이 직접 연 포지션과 미편입 보유가 여기로 온다 |
| `p.Quantity vs adoption.Quantity` | 현재 수량 대 편입 수량 | 계좌 스냅샷 대 원장 | B3 — 늘지 않았으면 반환 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/{obs,app/engine,journal,exitpolicy}/... -count=1 -covermode=set`의 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:442` `if d.grown[p.ID] {` | `d.opts.Journal.AdoptionOf` | :443 | 아니오 |
| B2 | if | `:446` `if err != nil {` | `riskcalc.CompareDecimal` | :447 | 아니오 |
| B3 | if | `:450` `if err != nil \|\| cmp <= 0 {` | `d.alert`, `d.label`, `string` | :451 | 예 |

## Calls and live bindings

`d.opts.Journal.AdoptionOf`(B2 앞) · `riskcalc.CompareDecimal`(B3 앞) · `d.alert`(끝) · `d.label`.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

`d.grown[p.ID] = true` (메모리) · `d.alert` 경유 알림 1건. **원장의 exit state는 건드리지 않는다.**

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수의 t0 동결 결정을 뒤집지 않는다.** 주석의 논거(*"moving any of them would rewrite the denominator every R on that position has already been expressed in"*)는 유효하다. a095가 고치는 것은 그 결정이 기대는 **보고**다 — 같은 주석이 *"Reporting it is what the operator can act on"*이라 쓰는데 그 보고가 outbox에 닿지 않는다. B2의 조용한 반환도 범위다.
- **High-risk impact**: yes — 보호 범위와 실제 보유의 차이를 알리는 유일한 자리다.
