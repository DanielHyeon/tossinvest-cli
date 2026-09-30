# Function Logic Map: `Journal.RestoreOperatingModeProjection`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :574–589, 분기 2 · 반환 2 · 호출 3, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 기동 복원의 투영 레코드에 최신 행의 rowid 를 싣는다(울타리). 오류는 호출자(`buildGateway`)가 모드 사유 래치로 처리(C15 · K8).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil` (:576) | — | — | (미실행) |
| B2 | `if p := j.modeProjectorRef(); p != nil` (:579) | — | — | `TestTheModeIsRestoredAfterARestart` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.CurrentOperatingMode`, `j.modeProjectorRef`, `p.ProjectOperatingMode` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: High-risk.
