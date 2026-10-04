# Function Logic Map: `newStrategyRiskLoaderFixture`

- Source: `internal/app/engine/strategy_risk_authority_test.go`
- Source SHA-256: `88a27a5175907ac6e3239a4d07a64096b28e72fb41822ddf37b55d6f23b28f23`
- Signature: `newStrategyRiskLoaderFixture(params=1, results=1)`
- Source range: `130:1`–`133:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 동작 없음(비례 원칙: FLM 의무의 무거운 규율 대상 아님, 게이트 모양만 채움)

## Branches and early returns

- Exact AST return nodes: `132:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 131:2 |
| `newStrategyRiskLoaderFixtureWith` | 132:9 |

## State mutations and fallbacks

- 시험 임시 디렉터리에 stub 원장 · 서명 매니페스트를 쓴다(변경 없음).

## Safety conclusion

- High-risk 아님. 기존 호출자가 받는 fixture 는 바이트 단위로 같은 매니페스트(append 할 것이 없음).

a112 6.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
