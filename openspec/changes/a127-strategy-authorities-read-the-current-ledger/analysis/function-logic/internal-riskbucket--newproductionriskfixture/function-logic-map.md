# Function Logic Map: `newProductionRiskFixture`

- Source: `internal/riskbucket/production_snapshot_authority_test.go` (`152`–`207`)
- Qualified: `newProductionRiskFixture`
- AST evidence: `ast.json` (`source_sha256` c578d9fd751b2112…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 8 · return 1 · 호출 35

**역할.** riskbucket 생산 위험 픽스처. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:158` `if err != nil {` |
| B2 | if | `:164` `if market == MarketUS {` |
| B3 | if | `:169` `if err != nil {` |
| B4 | if | `:174` `if market == MarketUS {` |
| B5 | if | `:178` `if err != nil {` |
| B6 | if | `:191` `if err != nil {` |
| B7 | if | `:197` `if err != nil {` |
| B8 | if | `:201` `if err := os.WriteFile(filePath, data, 0o400); err != nil {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: config 에 주입 값(시험 상수 1) 추가.
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
