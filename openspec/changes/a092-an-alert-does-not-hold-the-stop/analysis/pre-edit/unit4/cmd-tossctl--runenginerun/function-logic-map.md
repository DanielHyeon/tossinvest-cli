# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` — **편집 전**, :186–359, 분기 23 · 반환 10 · 호출 64, source_sha256 `9dd4f4532837…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 모드 전용 제어 엔드포인트(`.mode-control`)를 알림 제어 엔드포인트 옆에서 띄운다(C12 — 다른 힘이면 다른 엔드포인트). 실패는 알림 제어와 같이 「엔드포인트 강등」 보고.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if ctx == nil` (:188) | — | — | (미실행) |
| B2 | `if err != nil` (:194) | — | — | (미실행) |
| B3 | `if err != nil` (:200) | — | — | (미실행) |
| B4 | `if err != nil` (:210) | — | — | (미실행) |
| B5 | `if clauses := engine.UnmetInterlockClauses(err); clauses != nil` (:211) | — | — | (미실행) |
| B6 | `for _, clause := range clauses` (:213) | — | — | (미실행) |
| B7 | `if !ectx.Automation.Verified` (:222) | — | — | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` |
| B8 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil` (:232) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B9 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh` (:233) | — | — | (미실행) |
| B10 | `if token, terr := engineProcInstance(os.Getpid()); terr == nil` (:249) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B11 | `if merr != nil` (:252) | — | — | (미실행) |
| B12 | `} else` (:256) | — | — | (미실행) |
| B13 | `if err != nil` (:268) | — | — | (미실행) |
| B14 | `if err != nil` (:272) | — | — | (미실행) |
| B15 | `if policyControl != nil` (:295) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B16 | `if policyControlErr != nil` (:298) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B17 | `if policyRuntime != nil` (:303) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B18 | `if policyRuntimeErr != nil` (:306) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B19 | `if strategyRuntime != nil` (:321) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B20 | `if projErr != nil` (:331) | — | — | `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B21 | `if err != nil` (:338) | — | — | (미실행) |
| B22 | `if alertControl != nil` (:342) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B23 | `if alertControlErr != nil` (:345) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `(함수 리터럴)`, `Format`, `alertControl.Close`, `at.UTC`, `cancel`, `clk.Now`, `clock.System`, `cmd.Context`, `cmd.ErrOrStderr`, `cmd.OutOrStdout`, `context.Background`, `context.WithCancel`, `ectx.AlertOperations`, `ectx.Automation.MaskedAccount` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: 중간 — 운영자 표면 배선. 루프 · 손절 경로 무관.
