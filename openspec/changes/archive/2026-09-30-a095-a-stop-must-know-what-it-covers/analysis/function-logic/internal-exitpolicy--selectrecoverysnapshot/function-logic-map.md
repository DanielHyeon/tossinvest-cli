# Function Logic Map: `SelectRecoverySnapshot`

- Source: `internal/exitpolicy/recovery.go` (`134`–`172`)
- Qualified: `SelectRecoverySnapshot`
- AST evidence: `ast.json` (`source_sha256` 292a78a44ead7910…)
- Risk scan: `risk-pattern-report.md`
- 분기 10 · return 11 · 호출 9

**역할.** 저장된 effective 스냅샷과 재계산 스냅샷 중 더 안전한 하나를 통째로 고른다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `saved` | 저장된 effective 스냅샷 또는 nil | 원장 effective JSON | B2 — nil이면 재계산값을 그대로 받음 |
| `recomputed` | 재계산 스냅샷 | 판정 | B9 · B10 — 보호가 · 워터마크 · 단계를 저장값과 비교 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:135` `if err := validateRecoverySnapshot(recomputed); err != nil {` | `validateRecoverySnapshot` | :136 | 예 |
| B2 | if | `:138` `if saved == nil {` | — | :139 | 예 |
| B3 | if | `:141` `if err := validateRecoverySnapshot(*saved); err != nil {` | `validateRecoverySnapshot` | :142 | 예 |
| B4 | if | `:144` `if saved.PositionID != recomputed.PositionID \|\| saved.PositionGeneration != recomputed.PositionGeneration \|\|` | `compareRecoveryDecimal`, `sameRecoveryPolicy`, `strings.TrimSpace` | :148 | 예 |
| B5 | if | `:151` `if err != nil {` | `compareRecoveryDecimal` | :152 | 아니오 |
| B6 | if | `:155` `if err != nil {` | `compareRecoveryStage` | :156 | 아니오 |
| B7 | if | `:159` `if err != nil {` | — | :160 | 아니오 |
| B8 | if | `:162` `if stage == 0 && (saved.NextTarget != recomputed.NextTarget \|\| saved.NextProtection != recomputed.NextProtection) {` | `fmt.Errorf` | :163 | 예 |
| B9 | if | `:165` `if protection >= 0 && high >= 0 && stage >= 0 {` | — | :166 | 예 |
| B10 | if | `:168` `if protection <= 0 && high <= 0 && stage <= 0 {` | — | :169, :171 | 예 |

## Calls and live bindings

`validateRecoverySnapshot` · `sameRecoveryPolicy` · `compareRecoveryDecimal` · `compareRecoveryStage`.

결과는 `(ExitLineSnapshot, RecoverySource, error)`다 — 검증 실패 · 신원 불일치 · 모호함을 오류로 돌려준다(B1 · B3 · B4 · B5~B8 · B10 창).

## State mutations and fallbacks

없다 — 순수 선택.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B2(`saved == nil`)의 창 return `:139`은 비교 없이 재계산값을 받는다. B9 · B10의 비교 대상은 `saved.CurrentProtection` — effective JSON의 보호가이지 스칼라 `baseline_price`가 아니다(r3 N5).
- **High-risk impact**: yes — 복구 시 손절선 선택이다.
