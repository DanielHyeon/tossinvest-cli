# Function Logic Map: `newEngineAlertsAckCmd`

- Source: `cmd/tossctl/engine_alerts.go`
- AST evidence: `ast.json` — **편집 전**, :74–108, 분기 0 · 반환 2 · 호출 5, source_sha256 `3f3ec4cec7aa…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): C17 — `Annotations` 에 `mutating: "true"`(원장 쓰기 명령 분류).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | — | 한 식 반환 | 호출자 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StringVar`, `cmd.Flags`, `cmd.MarkFlagRequired`, `runEngineAlertsAck`, `strings.TrimSpace` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: Low — 명령 분류 표지.
