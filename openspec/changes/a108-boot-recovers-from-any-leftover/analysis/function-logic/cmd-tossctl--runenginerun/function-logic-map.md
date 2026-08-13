# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (183-324)
- AST evidence: `ast.json` — AST branches 20 · returns 13 · defers 9 · calls 58
- Risk scan: `risk-pattern-report.md`
- 편집 대상: **B18** (겹2). 기준(base) 판은 분기 19개·return 14개였고, 그 14번째가
  `strategyprojectionrpc.Start` 오류의 `return err` 다 — 2026-08-13 사고가 죽은 줄.
  이번 편집이 그 return 을 없애고 B17(성공 경로 전용 defer)과 B18(강등)을 만들었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 `context.Background()` (B1) |
| `root.configDir` | 빈 문자열이면 기본 journal 경로 | `engineJournalDir` | 해석 실패는 즉시 return (B2) |
| journal 디렉터리 flock | **동시에 하나** | `enginelock.Acquire(dir)` | 이미 잡혀 있으면 `ErrAlreadyRunning` return (B3). **엔진 싱글턴의 유일한 권위** |
| 엔진 조립 | config→journal(RW)→official broker→obs→기동 인터록 | `engineAssemble` seam | 인터록 미충족이면 절 열거 후 `errEngineInterlockUnmet` (B4·B5·B6), 그 밖은 원인 그대로 (B4) |
| `ectx.Automation.Verified` | true 여야 루프가 있다 | 조립 결과 | false 면 `errEngineGateOff` (B7) |
| verify run lock | 신선하면 같은 계좌를 다투는 중 | `runlock.Fresh` | `errVerifyInProgress` (B8·B9) |
| proc instance 토큰 | `/proc` 없는 커널에서는 없을 수 있음 | `engineProcInstance` | 없으면 마커에 안 싣고(B10), 알림 key 는 기동 시각으로 대체 |
| 활성 마커 | 자문(advisory) | `enginelock.Hold` | 실패는 **거절이 아니다** — note 한 줄 (B11/B12) |
| position policy command/runtime 서버 | 있어야 콘솔이 정책을 바꾼다 | `engine.Start…Server` | 실패는 return (B14·B15·B16) |
| **strategy projection endpoint** | **없어도 된다 (a108)** | `engineStrategyProjectionStart` seam | **강등 + durable critical 알림 + 루프 계속 (B18)** |
| alert 운영 표면 | 있어야 운영자가 승인한다 | `ectx.AlertOperations` | 실패는 return (B19·B20) |

**관통 불변식(a108 이후):** 조회 전용 export 표면의 실패는 보호 루프의 기동을 막지 않는다.
배타성은 B3 의 flock 이고 projection 디렉터리가 아니다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ctx == nil` | 없음 | 없음 (배경 ctx 대입) | 간접 (모든 CLI 경로) |
| B2 | `engineJournalDir` 실패 | 없음 | 원인 그대로 | 미고정(경로 해석 실패는 환경 이상) |
| B3 | `enginelock.Acquire` 실패 | 없음 | `ErrAlreadyRunning` 등 | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` |
| B4 | `engineAssemble` 실패 | 없음 | 인터록이면 B5 로, 아니면 원인 | `TestANonInterlockFailureIsReportedAsItself` |
| B5 | 미충족 절이 있다 | stderr 열거 | `errEngineInterlockUnmet` | `TestAnUnmetInterlockIsEnumerated` |
| B6 | 절 순회 | stderr 한 줄씩 | 없음 | 위와 같음 |
| B7 | 게이트 미검증 | 없음 | `errEngineGateOff` | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` |
| B8 | verify lock 경로 해석 성공 | 없음 | 없음 | `TestAFreshVerifyRunLockRefusesTheStart` |
| B9 | verify lock 이 신선 | stderr | `errVerifyInProgress` | 위와 같음 / `TestAStaleVerifyRunLockDoesNotRefuse` |
| B10 | proc instance 를 읽었다 | `marker.Identify` | 없음 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` |
| B11 | 마커를 못 잡았다 | stderr note | **없음 — 거절 아님** | 미고정(기존 강등, 이번 편집 밖) |
| B12 | 마커를 잡았다 | stdout 한 줄 | 없음 | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` |
| B13 | 루프 조립 실패 | 없음 | 원인 그대로 | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B14 | policy command service 실패 | 없음 | 원인 그대로 | 미고정 |
| B15 | policy command server 실패 | 없음 | 원인 그대로 | 미고정 |
| B16 | policy runtime server 실패 | 없음 | 원인 그대로 | 미고정 |
| **B17** | **projection 서버가 섰다** | `defer Close` 등록 | 없음 | `TestASucceedingProjectionIsStillServedAndClosed` |
| **B18** | **projection 기동 실패** | stderr note + **durable critical 알림 1행** | **없음 — 루프는 계속 돈다** | `TestAFailedStrategyProjectionDoesNotStopTheEngine` 외 4건 |
| B19 | alert 운영 표면 실패 | 없음 | 원인 그대로 | 미고정 |
| B20 | alert control server 실패 | 없음 | 원인 그대로 | 미고정 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire` | 1단계 배타성 | 실패는 fail-closed return | AST calls · engine.go:196 |
| `engineAssemble` (seam) | 2~4단계 조립 | `engine.UnmetInterlockClauses` 로 분류 | AST · engine_test.go |
| `runlock.Fresh` | 5단계 rate 예산 | 오래된 lock 은 거절하지 않는다 | AST · engine.go:230 |
| `enginelock.Hold` / `Identify` | 6단계 자문 마커 | 실패는 note (B11) | AST · a102 |
| `engineRuntimeFactory` (seam) | 7단계 루프 집합 | 실패는 return (B13) | AST · engine_runtime_branch_test.go |
| `engine.StartPositionPolicy{Command,Runtime}Server` | 콘솔 정책 제어 | 실패는 return | AST |
| **`engineStrategyProjectionStart` (seam)** | **조회 전용 export** | **실패는 강등 (a108 D3)** | AST · a108 T2 테스트 |
| `ectx.Journal.EnqueueAlert` | 강등의 durable 기록 | 실패해도 return 하지 않는다 — note 한 줄 더 | `reportStrategyProjectionDegraded` |
| `engine.StartAlertControlServer` | a098 운영자 승인 경로 | 실패는 return | AST |
| `rt.Run(runCtx)` | 루프 감독 | 첫 정지가 전부를 내린다 | AST · internal/app/engine |

## State mutations and fallbacks

- flock·마커·세 소켓(policy command TCP · policy runtime unix · alert control unix)은
  defer 로 정리된다. **projection 소켓의 defer 는 B17 안에서만 등록된다** — 강등 경로에는
  닫을 것이 없다.
- `reportStrategyProjectionDegraded` 가 만드는 상태는 원장 alert outbox 행 하나다.
  event key 는 `type|dir|실행토큰` 이라 **실행마다 갈리고 한 실행 안에서는 접힌다**.
  토큰이 없으면 기동 시각(RFC3339Nano)이 그 자리를 대신한다.
- 알림 쓰기 실패는 강등을 되돌리지 않는다. 되돌리면 「알림을 못 썼다」가 다시
  엔진을 죽이는 이유가 되고, 그것은 이 change 가 지우려는 모양 그 자체다.

## Safety conclusion

- Safe edit boundary: B17·B18 과 `reportStrategyProjectionDegraded`. B1~B16·B19·B20 의
  판정과 **순서**는 건드리지 않았다 — 특히 B3(flock) → B7(게이트) → B18(projection)의
  선후는 테스트로 고정돼 있다.
- High-risk impact: **yes.** 엔진 기동 경로다. 다만 변경 방향은 「기동을 더 자주 성공시키는」
  쪽이고, 그것이 보호 루프를 세우는 방향이다. 반대 방향(주문·손절 즉시성)에는 닿지 않는다.
