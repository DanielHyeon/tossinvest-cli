# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` — **편집 뒤**, :636–724, 분기 6 · 반환 7 · 호출 14, source_sha256 `aeefd5dcc3dd…`, 추출 커밋 `e55102f0`. 편집 전 번들은 `analysis/pre-edit/unit5/`(없으면 단위 ④ 번들이 편집 전).
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ⑤ `e55102f0`): `Auxiliary` 에 `ectx.NormalAlertRelayExecutor()` 추가(C8). 분기 불변.
- 재번호: difflib 정렬(편집 전 `b01e0cd0` 소스 대비): 분기 좌표 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:639) | — | — | (미실행) |
| B2 | `if err != nil` (:648) | — | — | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B3 | `if err != nil` (:659) | — | — | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B4 | `if err != nil` (:664) | — | — | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B5 | `if err != nil` (:668) | — | — | (미실행) |
| B6 | `if err != nil` (:680) | — | — | (미실행) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `costs.DefaultModel`, `ectx.AlertDeliverer`, `ectx.ExitObserver`, `ectx.NewRefreshingPairedStrategyEntrySupervisor`, `ectx.NormalAlertRelayExecutor`, `ectx.ReconcileDriver`, `ectx.Recovery`, `ectx.SnapshotCollector`, `engine.NewRuntime`, `engineFillDetector`, `engineRecoveryObserver`, `engineRecoverySequence`, `recoverThenReady`, `strategyEntry.SupervisedLoop` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(조립) — 보조 실행자 하나 추가.
