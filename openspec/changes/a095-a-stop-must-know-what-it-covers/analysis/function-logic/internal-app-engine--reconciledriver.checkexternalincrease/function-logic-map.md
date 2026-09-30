# Function Logic Map: `ReconcileDriver.checkExternalIncrease`

- Source: `internal/app/engine/adoption.go` (`556`–`586`)
- Qualified: `ReconcileDriver.checkExternalIncrease`
- AST evidence: `ast.json` (`source_sha256` 3d66976f07a50aa7…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 3 · 호출 6

**역할.** 편입된 포지션의 수량이 편입 기록보다 늘었는지 보고 알린다. 주석이 t0 동결을 의도적 설계(A8)로 선언한다. 10판: 새 최대 수량마다 다시 알린다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `AdoptionOf(p.ID)` | 편입 기록 | 원장 | 조회 오류면 조용히 return(무변화 — R2-B2 삭제) |
| `p.Quantity` 대 `adoption.Quantity` | 현재 수량 대 편입 수량 | 스냅샷 대 원장 | 늘지 않았으면 return |
| `d.newGrowthMaximum(p)` | 보고한 최대 수량보다 큰가 | 프로세스 메모리 `d.grown[p.ID]` | 아니면 return(10판 — 옛 bool 래치 대체) |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:558` `if err != nil {` | `riskcalc.CompareDecimal` | :559 | 예 |
| B2 | if | `:562` `if err != nil \|\| cmp <= 0 {` | — | :563 | 예 |
| B3 | if | `:565` `if !d.newGrowthMaximum(p) {` | `d.alert`, `d.label`, `d.newGrowthMaximum`, `string` | :566 | 예 |

## Calls and live bindings

`d.opts.Journal.AdoptionOf` · `riskcalc.CompareDecimal` · `d.newGrowthMaximum`(10판) · `d.alert` · `d.label`.

결과값이 없다 — 오류를 돌려주지 않는다. `AdoptionOf` · `CompareDecimal`의 오류에서 조용히 반환하고(표의 return 열), `newGrowthMaximum`의 비교 실패도 보고하지 않는 쪽이며, `d.alert`는 `Notify`의 오류를 로그로만 남긴다.

## State mutations and fallbacks

`d.grown[p.ID] = <보고한 수량>`(메모리, `newGrowthMaximum` 안) · normal 알림 1건(key `…|grown|<posID>`). exit state 무접촉.

## Safety conclusion

- **Safe edit boundary**: **10판 편집**: 함수 머리의 bool 래치(옛 B1)를 지우고 비교 뒤의 `newGrowthMaximum`(새 leaf)으로 옮겼다 — 래치 기준이 (포지션, 보고한 최대 수량)이다(Q4). 조회 오류의 조용한 반환은 무변화(3.1). 등급 normal 의 전제는 review §4.4.
- **High-risk impact**: yes — 편입 후 수량 증가를 알리는 유일한 자리다.
