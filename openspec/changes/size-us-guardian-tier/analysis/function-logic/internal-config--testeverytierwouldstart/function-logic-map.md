# Function Logic Map: `TestEveryTierWouldStart`

- Source: `internal/config/limits_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 전 티어 | 다섯 값 양수·유한, 비율 (0,1], 통화 비지 않음 | `GuardianLimits.Validate` (인터록 동치) | 거부는 `t.Errorf` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 전 티어 순회 | 없음 | 없음 | 자기 자신 |
| B2 | `Validate` 오류 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `GuardianTiers` | 레지스트리 | 오류 없음 | CodeGraph + AST |
| `GuardianLimits.Validate` | 기동 인터록과 동일 규칙 | error를 돌려준다 | CodeGraph + AST + `internal/app/engine/limits_equivalence_test.go` |

## State mutations and fallbacks

- 상태 변경 없음. 신규 테스트다.
- 프리셋은 클릭 한 번으로 파일에 기록된다. 인터록이 거부할 값을 실은 프리셋은 "누르면 엔진이 안 뜨는 버튼"이고, 화면은 그것을 미리 말해 줄 방법이 없다.
- `Validate`와 `execgw.Limits.Validate`의 동치는 이 파일이 아니라 `internal/app/engine`에서 생성 코퍼스로 고정된다(import cycle 때문).
- fallback 없음.

## Safety conclusion

- Safe edit boundary: 없음 — 티어를 추가하는 모든 사람이 지나야 하는 관문이다.
- High-risk impact: yes(Guardian) — 기동 가능성의 최소 조건이다.
