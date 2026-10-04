# Function Logic Map: `Lane.Health`

- Source: `internal/strategyworker/lane.go`
- Source SHA-256: `b869921551c2058b80fa03d2bc9c601f3ea306234d918ee41cde019312d529ab`
- Signature: `Lane.Health(params=0, results=1)`
- Source range: `185:1`–`189:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.5).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Health 와 Status 의 건강 판정은 같은 함수다 — 변이 R06(Status 판정 분리) CAUGHT.

## Branches and early returns

- Exact AST return nodes: `188:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `lane.mu.Lock` | 186:2 |
| `lane.mu.Unlock` | 187:8 |
| `lane.healthLocked` | 188:9 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기.

## Safety conclusion

- High-risk 아님 — 레인 건강 읽기(값 불변).
