# Function Logic Map: `TestBootZeroDurationDoesNotHideInvalidClock`

- Source: `internal/clock/bootstamp_test.go` L58–69(편집 전)
- AST evidence: `ast.json`(편집 뒤 current), 편집 전 `../../time-namespace-lot/pre-edit-ast/TestBootZeroDurationDoesNotHideInvalidClock.json`(분기 B1 L62, B2 L66)
- Risk scan: `risk-pattern-report.md`
- 변경 이유: `RemainingBootHint` 에서 ns 신원 비교 제거(time namespace 로트, Manager 결정 2026-11-25). 통과 이유가 바뀐 기존 시험 — 지우지 않고 새 계약을 단언하게 고친다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake `BootStamp`(testBootA/B, ns {1,2} 등) | `validStamp` 만족 | 시험 상수 | 시험 실패(`t.Fatalf`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L62 | saved=now, duration 0 → `{0,false},nil` 단언. 유지 | 없음 | 편집 뒤: (무변) | 자기 자신 |
| B2 L66 | now ns inode+1, duration 0 — 편집 전: `ErrBootStamp` 단언(「zero masked mismatch」) | 없음 | 편집 뒤: 위 「바뀌는 절반」 참조 | 자기 자신 |
| B3(편집 뒤 신설) | 같은 boot uptime 역행 + duration 0 → `ErrBootStamp` 단언(남는 절반) | 없음 | 거절 단언 | 자기 자신 — 죽이는 변이 M1 · M8 |

바뀌는 절반: ns 차이 판을 수락(`{0,false}`)으로 뒤집고, 남는 절반(「0 이 무효 시계를 숨기지 않음」)은 같은 boot uptime 역행 + duration 0 → `ErrBootStamp` 단언을 새로 넣어 지킨다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RemainingBootHint` | 대상 산술 | 오류는 `errors.Is(e, ErrBootStamp)` 로 판정 | AST calls |

## State mutations and fallbacks

- 없음(지역 변수만).

## Safety conclusion

- Safe edit boundary: 이 시험 본문만. 남는 단언(uptime · anchor 순서 · 유효성)은 지우지 않는다.
- High-risk impact: 시험 전용 — 단 High-risk 함수의 계약 핀이라 편집 뒤 죽이는 변이를 원장에 적는다: M8(B5 duration=0 early return 을 B2 앞으로 이동) · M1(B3 uptime 항 삭제) — 새 역행 판; M4 — ns 수락 판.
