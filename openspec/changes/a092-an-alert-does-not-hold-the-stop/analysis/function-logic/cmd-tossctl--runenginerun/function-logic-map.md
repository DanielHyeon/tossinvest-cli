# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json` — **편집 뒤**, :188–376, 분기 26 · 반환 10 · 호출 69, source_sha256 `ef56c3613d41…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 알림 제어 뒤에 모드 제어 엔드포인트 기동 — 표면 생성 · 기동 실패는 강등 보고(엔진 계속). 분기 셋 추가.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B23 → B1~B23(같은 분기, 줄 이동); 새 분기 B24(:356), B25(:359), B26(:363).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if ctx == nil` (:190) | — | — | (미실행) |
| B2 | `if err != nil` (:196) | — | — | (미실행) |
| B3 | `if err != nil` (:202) | — | — | (미실행) |
| B4 | `if err != nil` (:212) | — | — | (미실행) |
| B5 | `if clauses := engine.UnmetInterlockClauses(err); clauses != nil` (:213) | — | — | (미실행) |
| B6 | `for _, clause := range clauses` (:215) | — | — | (미실행) |
| B7 | `if !ectx.Automation.Verified` (:224) | — | — | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` |
| B8 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil` (:234) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B9 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.StaleAfter); fresh` (:235) | — | — | (미실행) |
| B10 | `if token, terr := engineProcInstance(os.Getpid()); terr == nil` (:251) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B11 | `if merr != nil` (:254) | — | — | (미실행) |
| B12 | `} else` (:258) | — | — | (미실행) |
| B13 | `if err != nil` (:270) | — | — | (미실행) |
| B14 | `if err != nil` (:274) | — | — | (미실행) |
| B15 | `if policyControl != nil` (:297) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B16 | `if policyControlErr != nil` (:300) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B17 | `if policyRuntime != nil` (:305) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B18 | `if policyRuntimeErr != nil` (:308) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B19 | `if strategyRuntime != nil` (:323) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B20 | `if projErr != nil` (:333) | — | — | `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B21 | `if err != nil` (:340) | — | — | (미실행) |
| B22 | `if alertControl != nil` (:344) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |
| B23 | `if alertControlErr != nil` (:347) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine` |
| B24 | `if modeControlErr == nil` (:356) | — | — | (미실행) |
| B25 | `if modeControl != nil` (:359) | — | — | (미실행) |
| B26 | `if modeControlErr != nil` (:363) | — | — | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `(리터럴)`, `Format`, `alertControl.Close`, `at.UTC`, `cancel`, `clk.Now`, `clock.System`, `cmd.Context`, `cmd.ErrOrStderr`, `cmd.OutOrStdout`, `context.Background`, `context.WithCancel`, `ectx.AlertOperations`, `ectx.Automation.MaskedAccount`, `ectx.Close`, `ectx.ModeOperations` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: 중간 — 운영자 표면.
