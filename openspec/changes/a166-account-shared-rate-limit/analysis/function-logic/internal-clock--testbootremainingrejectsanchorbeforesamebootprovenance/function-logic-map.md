# Function Logic Map: `TestBootRemainingRejectsAnchorBeforeSameBootProvenance`

- Source: `internal/clock/bootstamp_test.go` L118–132(편집 전)
- AST evidence: `ast.json`(편집 뒤 current), 편집 전 `../../time-namespace-lot/pre-edit-ast/TestBootRemainingRejectsAnchorBeforeSameBootProvenance.json`(분기 B1 L124, B2 L129)
- Risk scan: `risk-pattern-report.md`
- 변경 이유: `RemainingBootHint` 에서 ns 신원 비교 제거(time namespace 로트, Manager 결정 2026-11-25). 통과 이유가 바뀐 기존 시험 — 지우지 않고 새 계약을 단언하게 고친다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake `BootStamp`(testBootA/B, ns {1,2} 등) | `validStamp` 만족 | 시험 상수 | 시험 실패(`t.Fatalf`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L124 | anchor=saved 이고 uptime 0(앞섬) — `ErrBootStamp` 단언(B2 uptime 항). 유지 | 없음 | 편집 뒤: (무변) | 자기 자신 |
| B2 L129 | anchor=saved 의 ns inode+1 — 편집 전: `ErrBootStamp` 단언(「conflicting same boot anchor」, B2 ns 항) | 없음 | 편집 뒤: 위 「바뀌는 절반」 참조 | 자기 자신 |
| B3(편집 뒤 신설) | 다른 ns 의 같은 boot anchor 가 앞섬(uptime 0) → `ErrBootStamp` 단언 | 없음 | 거절 단언 | 자기 자신 — 죽이는 변이 M3 |

바뀌는 절반: 다른 ns 의 같은 boot anchor(순서 맞음)를 수락하고 anchor 기준 차감 값 `{9.5s,false}` 를 단언한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RemainingBootHint` | 대상 산술 | 오류는 `errors.Is(e, ErrBootStamp)` 로 판정 | AST calls |

## State mutations and fallbacks

- 없음(지역 변수만).

## Safety conclusion

- Safe edit boundary: 이 시험 본문만. 남는 단언(uptime · anchor 순서 · 유효성)은 지우지 않는다.
- High-risk impact: 시험 전용 — 단 High-risk 함수의 계약 핀이라 편집 뒤 죽이는 변이를 원장에 적는다: M3(B2 uptime 항 삭제) — 앞 판; M4a · M4b · M4c(B2 · B4 · B8 에 ns 항 복원) — 뒤 판(수락 `{9.5s,false}`, anchor 를 saved+0.5s 로 둬 base 를 가름); M9(anchor 무시).
