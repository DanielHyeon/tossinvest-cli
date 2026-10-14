# Function Logic Map: `quotedOrUnset`

- Source: `cmd/tossctl/limitsettings.go` (108-113, 편집 뒤)
- Qualified: `quotedOrUnset`
- AST evidence: `ast.json` (revision: base) — 본문 무편집, 바로 뒤에 새 함수가 붙어 검사기가 교체된 base 함수로 분류.
- Risk scan: `risk-pattern-report.md`

**역할.** audit 이전 값 표기.

**a136 편집.** 본문 무편집 — 바로 뒤에 새 함수를 붙여 검사기가 수정으로 분류.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if v == "" \|\| v == "0" {` (if 109) | — | 분기 본문 참조 | `TestTheLimitSeamSavesAuditsAndLeavesTheSwitchAlone` |

Exact AST return positions: 110 · 112

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 설정 파일/audit(해당 시) — 기존과 같음.

## Safety conclusion

- Safe edit boundary: 칸 · 문장 하나 추가 — 기존 분기 무변경.
- High-risk impact: no
