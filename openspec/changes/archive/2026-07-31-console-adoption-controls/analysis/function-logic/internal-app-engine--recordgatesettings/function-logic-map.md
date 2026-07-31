# Function Logic Map: `recordGateSettings`

- Source: `internal/app/engine/interlock.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

High-risk(§0.5 audit). include_symbols 항목 1행 추가(P1-2) — 다른 항목·dedupe 계약 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | TestIncludeListIsAudited(RED: 항목 부재)·TestAdoptionToggleIsAudited green |
| B2 | range | 없음 | — | TestIncludeListIsAudited(RED: 항목 부재)·TestAdoptionToggleIsAudited green |
| B3 | if | 없음 | — | TestIncludeListIsAudited(RED: 항목 부재)·TestAdoptionToggleIsAudited green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- audit append(기존 RecordChange 경유)

## Safety conclusion

- Safe edit boundary: 항목 목록 1행 추가만
- High-risk impact: yes — reconciliation/audit 경로, Pre-Edit 선언·적대적 리뷰 하에 수정
