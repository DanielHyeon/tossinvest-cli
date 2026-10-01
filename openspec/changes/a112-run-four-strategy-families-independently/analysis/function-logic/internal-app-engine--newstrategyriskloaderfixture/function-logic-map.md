# Function Logic Map: `newStrategyRiskLoaderFixture`

- Source: `internal/app/engine/strategy_risk_authority_test.go`
- Source SHA-256: `5e015333560303ac59eef1feab9098ef7b259696f80e9f370e1f292a0d00945c`
- Signature: `newStrategyRiskLoaderFixture(params=1, results=1)`
- Source range: `128:1`–`131:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 시험 코드 — 생산 동작 없음(비례 원칙: FLM 의무의 무거운 규율 대상 아님, 게이트 모양만 채움)

## Branches and early returns

- Exact AST return nodes: `130:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `t.Helper` | 129:2 |
| `newStrategyRiskLoaderFixtureWith` | 130:9 |

## State mutations and fallbacks

- 시험 임시 디렉터리에 stub 원장 · 서명 매니페스트를 쓴다(변경 없음).

## Safety conclusion

- High-risk 아님. 기존 호출자가 받는 fixture 는 바이트 단위로 같은 매니페스트(append 할 것이 없음).

a112 6.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
