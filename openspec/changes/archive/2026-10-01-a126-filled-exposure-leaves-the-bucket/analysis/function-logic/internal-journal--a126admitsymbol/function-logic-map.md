# Function Logic Map: `a126AdmitSymbol`

- Source: `internal/journal/a126_filled_exposure_leaves_the_bucket_test.go` (`292`–`296`)
- Qualified: `a126AdmitSymbol`
- AST evidence: `ast.json` (`source_sha256` d2df830b7ee0d70b…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 0 · return 1 · 호출 2

**역할.** US fresh admission fixture. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. base 재고정(review 2.1) 뒤 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 본문을 a126AdmitOwner(owner 키 넷을 호출자가 정함) 위임으로 바꿈 — 같은 입력에서 같은 plan(1.5 R6).
- **High-risk impact**: no — 시험. 생산 동작 변화 0(`64ad3058` 비시험 `.go` 변경 0 — codex 재리뷰 확인).
