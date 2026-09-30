# Function Logic Map: `newEngineAlertsAckCmd`

- Source: `cmd/tossctl/engine_alerts.go`
- AST evidence: `ast.json` — **편집 뒤**, :74–109, 분기 0 · 반환 2 · 호출 5, source_sha256 `0b1df1ebe19b…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): `mutating: "true"` 표지(C17).
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): 분기 좌표 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| — | 분기 없음 | — | — | 호출 시험 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `StringVar`, `cmd.Flags`, `cmd.MarkFlagRequired`, `runEngineAlertsAck`, `strings.TrimSpace` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: Low.
