# Function Logic Map: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`

- Source: `internal/app/engine/a163_source_age_audit_test.go` (16-54)
- Qualified: `TestA163TheReportToggleIsTheLastAuditedSettingAndOnlyItsChangeIsWritten`
- AST evidence: `ast.json` (source_sha256 `2b521a3b04c0…`) — AST branches 8 · returns 1
- Risk scan: `risk-pattern-report.md`(설정 규칙 일치 0)
- 작성 시점: 편집 **전**(a208 tasks 1.1, 2026-11-28, HEAD `810aec6f7` — base `baeddfe31` 와 본문 동일).

**역할.** a163 시험 — 기동 audit 표에서 보고 토글 행의 위치(끝)와 값 변경만 기록되는지 잰다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recordGateSettings` 표 | 끝에만 덧붙임 | interlock.go | — |
| 불변식 | 새 행은 끝에만(끼워 넣기 금지) | a074 · a163 시험 | — |

## Branches and early returns

| Branch | Condition | At | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if err := recordGateSettings(...); err != nil` | 21:3 | — | t.Fatalf | 자기 자신 |
| B2 | `if err != nil` — Entries | 25:3 | — | t.Fatalf | 자기 자신 |
| B3 | `if last.Setting != … || …`(끝 행 = a163) | 33:2 | — | t.Fatalf | 자기 자신 |
| B4 | `if prev := first[len(first)-2]; …`(앞 행 = a139) | 36:2 | — | t.Fatalf | 자기 자신 |
| B5 | `if audit.ActionExitSourceAgeReport != …` | 39:2 | — | t.Fatalf | 자기 자신 |
| B6 | `if got := len(turnedOn) - len(first); got != 1` | 44:2 | — | t.Fatalf | 자기 자신 |
| B7 | `if e := turnedOn[len(turnedOn)-1]; …` | 47:2 | — | t.Fatalf | 자기 자신 |
| B8 | `if again := record(true); …` | 51:2 | — | t.Fatalf | 자기 자신 |

Exact AST return positions: 28:3

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordGateSettings` · `log.Entries` | 측정 대상 | — | AST |

Live config binding: 없음(시험).

## State mutations and fallbacks

- 임시 audit 로그.

## Safety conclusion

- Safe edit boundary(a208): B3 · B4 가 보는 첨자를 한 칸씩 앞으로(`len-1` → `len-2`, `len-2` → `len-3`) — a208 행이 끝에 붙으므로. 분기 · 조건 모양 무변. B6~B8 은 그대로 성립(토글 켬은 a163 행 하나만 바뀌어 끝에 붙음).
- High-risk impact: no(시험).
