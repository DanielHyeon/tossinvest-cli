# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (183-328)
- AST evidence: `ast.json` — AST branches 20 · returns 13 · defers 9
  (source_sha256 `8111c1c9e20f501b6221e231836fb02d7d03d127b3592892175c1beb38788381`,
  **Fix 라운드 6.7 편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- 편집 대상: **B18** (겹2). 기준(base) 판은 분기 19개·return 14개였고, 그 14번째가
  `strategyprojectionrpc.Start` 오류의 `return err` 다 — 2026-08-13 사고가 죽은 줄.
  이번 편집이 그 return 을 없애고 B17(성공 경로 전용 defer)과 B18(강등)을 만들었다.
- **Fix 라운드(6.7, design D3-2)**: B18 의 분기 **구조는 그대로**이고 그 안의 보고 수단만
  바뀌었다 — durable critical outbox 행이 사라지고 obs Normal 이벤트가 그 자리를 받는다.
  AST 상 사라진 호출은 `clk.Now`(강등 인자) 하나이며 분기 수는 20 으로 불변이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 `context.Background()` (B1) |
| `root.configDir` | 빈 문자열이면 기본 journal 경로 | `engineJournalDir` | 해석 실패는 즉시 return (B2) |
| journal 디렉터리 flock | **동시에 하나** | `enginelock.Acquire(dir)` | 이미 잡혀 있으면 `ErrAlreadyRunning` return (B3). **엔진 싱글턴의 유일한 권위** |
| 엔진 조립 | config→journal(RW)→official broker→obs→기동 인터록 | `engineAssemble` seam | 인터록 미충족이면 절 열거 후 `errEngineInterlockUnmet` (B4·B5·B6), 그 밖은 원인 그대로 (B4) |
| `ectx.Automation.Verified` | true 여야 루프가 있다 | 조립 결과 | false 면 `errEngineGateOff` (B7) |
| verify run lock | 신선하면 같은 계좌를 다투는 중 | `runlock.Fresh` | `errVerifyInProgress` (B8·B9) |
| proc instance 토큰 | `/proc` 없는 커널에서는 없을 수 있음 | `engineProcInstance` | 없으면 마커에 안 싣는다(B10). **a102 마커 전용이다** — 6.7 이 강등 보고의 dedup 토큰 사용을 지웠다 |
| 활성 마커 | 자문(advisory) | `enginelock.Hold` | 실패는 **거절이 아니다** — note 한 줄 (B11/B12) |
| position policy command/runtime 서버 | 있어야 콘솔이 정책을 바꾼다 | `engine.Start…Server` | 실패는 return (B14·B15·B16) |
| **strategy projection endpoint** | **없어도 된다 (a108)** | `engineStrategyProjectionStart` seam | **강등 + stderr 경고 + obs Normal 이벤트 + 루프 계속 (B18)**. 원장 outbox 에는 **아무것도 쓰지 않는다** (D3-2) |
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
| **B18** | **projection 기동 실패** | stderr note + **obs Normal 이벤트 로그** (원장 행 0) | **없음 — 루프는 계속 돈다** | `TestAFailedStrategyProjectionDoesNotStopTheEngine` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` · `TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched` · `TestTheDegradedBootStillHoldsTheJournalFlock` · `TestTheDegradedBootLeavesReadyToTheRuntimeSeam` |
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
| **`engineStrategyProjectionStart` (seam)** | **조회 전용 export** | **실패는 강등 (a108 D3-2)** | AST · a108 T2 테스트 |
| `ectx.Notifier.Notify` | 강등의 obs 기록 | **Normal 등급** — 반환 오류는 critical 경로 전용이라 여기서는 발생하지 않는다(`obs.Notifier.notifyCritical`). **gstack 라운드부터 goroutine 에서 돈다** — 발행 한 번이 최대 `obs.DefaultPublishTimeout`(10s)이고 이 줄은 `rt.Run` 앞이다 | `reportStrategyProjectionDegraded` |
| `engine.StartAlertControlServer` | a098 운영자 승인 경로 | 실패는 return | AST |
| `rt.Run(runCtx)` | 루프 감독 | 첫 정지가 전부를 내린다 | AST · internal/app/engine |

## State mutations and fallbacks

- flock·마커·세 소켓(policy command TCP · policy runtime unix · alert control unix)은
  defer 로 정리된다. **projection 소켓의 defer 는 B17 안에서만 등록된다** — 강등 경로에는
  닫을 것이 없다.
- **`reportStrategyProjectionDegraded` 는 영속 상태를 하나도 만들지 않는다** (6.7 이후).
  stderr 한 줄과 obs 로그 한 줄이 전부다. 원 D3 은 여기에 durable critical outbox 행을
  뒀는데, 그 행이 `UndeliveredCount`(Type 무필터, `Journal.UndeliveredCount`)에 잡혀
  `restoreAlertEntryLatch`(`restoreAlertEntryLatch`)로 **다음 부팅의 진입 게이트를 잠갔다**.
  publisher 미설정 배포에서는 영구 PENDING 이므로 해제 수단이 운영자 ack 뿐이었다.
- 그래서 이 분기의 부작용 목록은 「없음」이 계약이다. 뮤테이션으로도 그것을 잰다:
  행을 하나라도 쓰면 `TestTheDegradedBootWritesNoUndeliveredOutboxRow` 가 죽는다.
- 잃은 것을 상시로 말하는 표면은 이 함수 밖에 있다 — 콘솔·httpapi 의 전략 화면이
  dormant/unavailable 로 뜬다. 「stderr 는 1회 유실형」이라는 걱정의 답은 그쪽이다.

## Safety conclusion

- Safe edit boundary: B17·B18 과 `reportStrategyProjectionDegraded`. Fix 라운드는 그중
  후자의 **본문만** 바꿨다(분기 구조 불변). B1~B16·B19·B20 의
  판정과 **순서**는 건드리지 않았다 — 특히 B3(flock) → B7(게이트) → B18(projection)의
  선후는 테스트로 고정돼 있다.
- High-risk impact: **yes.** 엔진 기동 경로다. 다만 변경 방향은 「기동을 더 자주 성공시키는」
  쪽이고, 그것이 보호 루프를 세우는 방향이다. 반대 방향(주문·손절 즉시성)에는 닿지 않는다.

## gstack Fix 라운드가 이 함수에서 바꾼 것 (2026-08-14)

세 가지이고 **판정은 하나도 없다.**

1. **B17 의 `defer strategyRuntime.Close()` 에 불변식 주석을 붙였다.** 이 Close 는
   journal flock 을 쥔 채로 돈다 — defer 는 LIFO 이고 1단계의 `lock.Release()` 가 가장
   먼저 등록됐으므로 가장 나중에 돈다. 회수 함수가 flock 을 자기 방어로 인용하는
   근거가 이 순서이고, 그 근거가 두 파일에 흩어져 있어서 한쪽에 적었다.
2. **`token, terr :=` 를 원래 if-스코프로 되돌렸다.** 함수 스코프로 끌어올린 것은
   철회된 proc-instance dedup 기계(D3-2 가 지웠다)의 흉터였다. 분기·return 수 불변.
3. 하드코딩 line-range 인용을 심볼 인용으로 바꿨다. 줄 번호는 편집마다 썩고, 썩은
   좌표는 다음 독자를 **다른 코드**로 데려간다.

강등 보고 자체의 실행 자리는 `reportStrategyProjectionDegraded`(이 change 가 만든
함수) 안에서 바뀌었다 — 동기 `Notify` → detached goroutine. 그 함수는 base 에 없던
것이라 자기 FLM 을 요구하지 않지만, 이 함수의 **시간** 계약이 그것으로 바뀐다:
7단계에서 `rt.Run` 까지의 지연이 알림 transport 의 속도와 무관해졌다.
뮤테이션 M13(동기 복원)·M14(보고 삭제)·M15(부모 ctx 상속)가 셋 다 죽는다.
