# Function Logic Map: `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor`

- Source: `internal/clock/bootstamp_test.go` L43–56(편집 전)
- AST evidence: `ast.json`(편집 뒤 current), 편집 전 `../../time-namespace-lot/pre-edit-ast/TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor.json`(분기 B1 L48, B2 L53)
- Risk scan: `risk-pattern-report.md`
- 변경 이유: `RemainingBootHint` 에서 ns 신원 비교 제거(time namespace 로트, Manager 결정 2026-11-25). 통과 이유가 바뀐 기존 시험 — 지우지 않고 새 계약을 단언하게 고친다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| fake `BootStamp`(testBootA/B, ns {1,2} 등) | `validStamp` 만족 | 시험 상수 | 시험 실패(`t.Fatalf`) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 L48 | now=saved 의 ns inode+1, anchor 는 다른 boot(B) — 편집 전: `ErrBootStamp` 단언(B3 ns 항) | 없음 | 편집 뒤: 같은 입력을 **수락** 단언으로 뒤집음: B3 통과 → B7(now.boot≠anchor.boot) → `{10s, NeedsAnchor}` | 자기 자신 |
| B2 L53 | now=saved, uptime-1 — `ErrBootStamp` 단언(B3 uptime 항). 유지 | 없음 | 편집 뒤: 위 「바뀌는 절반」 참조 | 자기 자신 |

이름은 base 함수 신원 유지를 위해 그대로 둔다(이름 변경은 gate ⑤ 에서 삭제+신규가 된다). 본문 주석으로 현재 의미를 적는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RemainingBootHint` | 대상 산술 | 오류는 `errors.Is(e, ErrBootStamp)` 로 판정 | AST calls |

## State mutations and fallbacks

- 없음(지역 변수만).

## Safety conclusion

- Safe edit boundary: 이 시험 본문만. 남는 단언(uptime · anchor 순서 · 유효성)은 지우지 않는다.
- High-risk impact: 시험 전용 — 단 High-risk 함수의 계약 핀이라 편집 뒤 죽이는 변이를 원장에 적는다: M4(B3 에 ns 항 복원) · M5(B7 다른 boot 차감) — 수락 판 `{10s, NeedsAnchor}`; M1(B3 uptime 항 삭제) — 역행 판.
