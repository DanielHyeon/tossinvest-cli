# Function Logic Map: `TestTheGateEditingExemptionIsNotIdle`

- Source: `internal/console/engineproc_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Narrative context: `../../function-logic-map.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 면제 파일명 2개 | `settings.go`, `settings_limits.go` | 짝 테스트와 손으로 맞춘 목록 | 목록이 갈라지면 이 테스트가 먼저 실패한다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 면제 파일 순회 | 없음 | 없음 | 자기 자신 |
| B2 | 면제 파일이 패키지에 없음 | 없음 | `t.Errorf` | 자기 자신 (파일 삭제 시) |
| B3 | 면제 파일이 더 이상 블록을 이름하지 않음 | 없음 | `t.Errorf` | 자기 자신 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `packageFiles` | 패키지 전 파일 | 실패는 t.Fatal | CodeGraph + AST |
| `nonCommentLines` | 주석 제외 | 없음 | CodeGraph + AST |

## State mutations and fallbacks

- 신규 테스트. 가드의 가드다: 면제가 필요 없어진 뒤에도 남아 있으면, 나중에 그 파일에 들어온 게이트 판정 코드가 조용히 통과한다.
- 두 목록(금지 면제 / 이 테스트)은 손으로 맞춰야 하고, 어긋나면 이 테스트가 먼저 실패한다.

## Safety conclusion

- Safe edit boundary: 파일명 목록. 짝 테스트의 `mayNameTheBlock`과 함께 움직여야 한다.
- High-risk impact: yes(가드) — 유휴 면제는 가드에 뚫린 구멍이다.
