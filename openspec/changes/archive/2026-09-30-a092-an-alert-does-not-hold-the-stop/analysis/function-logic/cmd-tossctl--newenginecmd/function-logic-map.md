# Function Logic Map: `newEngineCmd`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` — **편집 뒤**, :114–129, 분기 0 · 반환 1 · 호출 8, source_sha256 `aeefd5dcc3dd…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): 단위 ⑤ 무편집 — 줄 이동(재추출). 단위 ④ 편집은 `AddCommand(newEngineModeReleaseCmd)` 한 줄.
- 재번호: 편집 전 번들 없음(이 단위에서 새로 만든 번들).

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
| `cmd.AddCommand`, `newEngineAlertsCmd`, `newEngineModeReleaseCmd`, `newEngineReconcileResolveCmd`, `newEngineRiskRelaxationCmds`, `newEngineRunCmd` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: Low.
