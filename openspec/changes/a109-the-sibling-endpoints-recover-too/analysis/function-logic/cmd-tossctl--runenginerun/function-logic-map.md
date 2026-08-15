# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (183-328)
- AST evidence: `ast.json` — AST 분기 20 · return 13 · defer 9
  (source_sha256 `8111c1c9e20f501b6221e231836fb02d7d03d127b3592892175c1beb38788381`,
  a108 land 시점 판 = a109 base `016da624`)
- Risk scan: `risk-pattern-report.md`
- **a109 T2 편집 대상: B15·B16·B20** — 그 셋의 `return err`(:274 · :279 · :315)를 a108
  B18 이 이미 쓰는 강등 의례로 바꾼다(design D3). 나머지 17 분기는 건드리지 않는다.
- **편집하지 않는 이웃**: B19(`ectx.AlertOperations()` 실패, return :311)는 endpoint
  기동이 아니라 **in-process 표면 조립**이므로 design D3 표의 세 endpoint에 없다 —
  fatal 유지가 계약이다. B14(`NewPositionPolicyCommandService`) 도 같은 이유로 유지.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 `context.Background()` (B1) |
| `root.configDir` | 비면 기본 journal 경로 | `engineJournalDir` | 해석 실패는 즉시 return (B2) |
| journal 디렉터리 flock | **동시에 하나** | `enginelock.Acquire(dir)` (:196) | `ErrAlreadyRunning` return (B3). **엔진 싱글턴의 유일한 권위 — D3 강등 논증 전체가 여기에 기댄다** |
| 엔진 조립 | config→journal(RW)→official broker→obs→기동 인터록 | `engineAssemble` seam | 인터록 미충족이면 절 열거 후 `errEngineInterlockUnmet` (B4·B5·B6), 그 밖은 원인 그대로 |
| `ectx.Automation.Verified` | true 여야 루프가 있다 | 조립 결과 | false 면 `errEngineGateOff` (B7) |
| verify run lock | 신선하면 같은 계좌를 다투는 중 | `runlock.Fresh` | `errVerifyInProgress` (B8·B9) |
| proc instance 토큰 | `/proc` 없는 커널에서는 없을 수 있음 | `engineProcInstance` | 없으면 마커에 안 싣는다 (B10) |
| 활성 마커 | 자문(advisory) | `enginelock.Hold` | 실패는 **거절이 아니다** — note 한 줄 (B11/B12) |
| position policy command server | **오늘 fatal (B15)** — a109가 강등으로 바꾼다 | `engine.StartPositionPolicyCommandServer` (:272) | 오늘: `return err` :274 → autostart 영구 기동 루프. a109 이후: 강등 + 보고 |
| position policy runtime server | **오늘 fatal (B16)** — a109가 강등으로 바꾼다 | `engine.StartPositionPolicyRuntimeServer` (:277) | 오늘: `return err` :279. a109 이후: 강등 + 보고 |
| strategy projection endpoint | **없어도 된다 (a108)** | `engineStrategyProjectionStart` seam (:292) | 강등 + stderr note + obs Normal 이벤트 (B17·B18) |
| alert 운영 표면 조립 | 있어야 운영자가 승인한다 | `ectx.AlertOperations()` (:309) | fatal 유지 (B19) — endpoint 아님 |
| alert control server | **오늘 fatal (B20)** — a109가 강등으로 바꾼다 | `engine.StartAlertControlServer` (:313) | 오늘: `return err` :315. a109 이후: 강등 + 보고 |

**관통 불변식 1 (a108 이후, a109가 일반화)**: 조회·표면 전용 endpoint의 실패는 보호
루프의 기동을 막지 않는다. 배타성은 B3 의 flock 이고 어떤 control 디렉터리도 아니다.

**관통 불변식 2 (defer LIFO — D5a 가 정적으로 핀한다)**: `lock.Release()` defer 는
:200 에서 **가장 먼저** 등록되므로 LIFO 상 **가장 나중에** 실행된다. 그래서 모든
endpoint Close(:276 policyControl · :281 policyRuntime · :301 strategyRuntime ·
:317 alertControl)는 **journal flock 을 쥔 채로** 돈다 — 이 프로세스의 잔재 제거와
다음 엔진의 회수·발행이 겹칠 수 없다. 회수 함수가 flock 을 자기 1차 방어로 인용하는
근거가 이 순서다(design D2). 강등으로 defer 등록을 조건부로 바꿀 때 **nil 가드는
필요하지만 등록 순서는 바뀌면 안 된다.**

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:185) | `ctx == nil` | 없음 | 없음 (배경 ctx 대입) | 간접 (모든 CLI 경로) |
| B2 (:191) | `engineJournalDir` 실패 | 없음 | 원인 그대로 (:192) | 미고정 — 경로 해석 실패는 환경 이상 |
| B3 (:197) | `enginelock.Acquire` 실패 | 없음 | `ErrAlreadyRunning` 등 (:198) | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` |
| B4 (:207) | `engineAssemble` 실패 | 없음 | 인터록이면 B5, 아니면 원인 (:215) | `TestANonInterlockFailureIsReportedAsItself` |
| B5 (:208) | 미충족 절이 있다 | stderr 열거 | `errEngineInterlockUnmet` (:213) | `TestAnUnmetInterlockIsEnumerated` |
| B6 (:210) | 절 순회 | stderr 한 줄씩 | 없음 | 위와 같음 |
| B7 (:219) | 게이트 미검증 | 없음 | `errEngineGateOff` (:220) | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` |
| B8 (:229) | verify lock 경로 해석 성공 | 없음 | 없음 | `TestAFreshVerifyRunLockRefusesTheStart` |
| B9 (:230) | verify lock 이 신선 | stderr | `errVerifyInProgress` (:233) | 위 / `TestAStaleVerifyRunLockDoesNotRefuse` |
| B10 (:246) | proc instance 를 읽었다 | `marker.Identify` | 없음 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` |
| B11 (:249) | 마커를 못 잡았다 | stderr note | **없음 — 거절 아님** | 미고정 (기존 강등, 이번 편집 밖) |
| B12 (:253) | 마커를 잡았다 | stdout 한 줄 | 없음 | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` |
| B13 (:265) | 루프 조립 실패 | 없음 | 원인 그대로 (:266) | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B14 (:269) | policy command **service** 실패 | 없음 | 원인 그대로 (:270) — **fatal 유지** | 미고정 (endpoint 아님) |
| **B15 (:273)** | **policy command server 기동 실패** | 오늘 없음 | **오늘 `return err` (:274)** — a109가 강등으로 바꾼다 | a109 §2.1 RED → §2.2 GREEN |
| **B16 (:278)** | **policy runtime server 기동 실패** | 오늘 없음 | **오늘 `return err` (:279)** — a109가 강등으로 바꾼다 | a109 §2.1 RED → §2.2 GREEN |
| B17 (:293) | projection 서버가 섰다 | `defer Close` 등록 | 없음 | `TestASucceedingProjectionIsStillServedAndClosed` |
| B18 (:303) | projection 기동 실패 | stderr note + obs Normal 이벤트 (원장 행 0) | **없음 — 루프는 계속 돈다** | `TestAFailedStrategyProjectionDoesNotStopTheEngine` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` · `TestTheDegradedBootDoesNotWaitForTheNotifier` |
| B19 (:310) | alert 운영 표면 조립 실패 | 없음 | 원인 그대로 (:311) — **fatal 유지** | 미고정 (endpoint 아님) |
| **B20 (:314)** | **alert control server 기동 실패** | 오늘 없음 | **오늘 `return err` (:315)** — a109가 강등으로 바꾼다 | a109 §2.1 RED → §2.2 GREEN |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire` | 1단계 배타성 | 실패는 fail-closed return | AST calls · engine.go:196 |
| `engineAssemble` (seam) | 2~4단계 조립 | `engine.UnmetInterlockClauses` 로 분류 | AST · `engine_test.go` |
| `runlock.Fresh` | 5단계 rate 예산 | 오래된 lock 은 거절하지 않는다 | AST · engine.go:230 |
| `enginelock.Hold` / `Identify` / `Release` | 6단계 자문 마커 | 실패는 note (B11) | AST · a102 |
| `engineRuntimeFactory` (seam) | 7단계 루프 집합 | 실패는 return (B13) | AST · `engine_runtime_branch_test.go` |
| `engine.NewPositionPolicyCommandService` | 정책 명령 서비스 조립 | 실패는 return (B14) | AST · engine.go:268 |
| `engine.StartPositionPolicyCommandServer` | Preview/Apply·**격리 해제** 표면 | **오늘 fatal (B15)** — a109 강등 대상. 격리 해제는 격리된 포지션의 미판정을 푸는 유일한 장중 경로다(design D3) | AST · engine.go:272 |
| `engine.StartPositionPolicyRuntimeServer` | 콘솔·httpapi 관리 런타임 화면 | **오늘 fatal (B16)** — a109 강등 대상. 조회 전용 | AST · engine.go:277 |
| `engineStrategyProjectionStart` (seam) | 조회 전용 export | 실패는 강등 (a108 D3-2) | AST · a108 T2 테스트 |
| `reportStrategyProjectionDegraded` | B18 의 보고 | 동기 stderr + 비동기 Notify (자체 맵 참조) | AST · engine.go:304 |
| `ectx.AlertOperations` | 운영자 ack 표면 조립 | 실패는 return (B19) — **endpoint 아님** | AST · engine.go:309 |
| `engine.StartAlertControlServer` | 운영자 ack endpoint | **오늘 fatal (B20)** — a109 강등 대상. `AlertOperations` 는 gateway 에 닿지 않는다 | AST · engine.go:313 |
| `rt.Run(runCtx)` | 보호 루프 셋의 실행 | 이 호출에 닿는 것이 「부팅 성공」의 정의다 | AST · engine.go:327 |

## State mutations and fallbacks

- **디스크**: journal flock 파일(B3), 활성 마커(B12), 그리고 endpoint 넷의 control
  디렉터리·descriptor·socket. 후자는 각 Start 가 만들고 대응 defer Close 가 지운다.
- **defer 스택(등록 순서)**: `lock.Release`(:200) → `ectx.Close`(:217) →
  `marker.Release`(:240) → `policyControl.Close`(:276) → `policyRuntime.Close`(:281) →
  `strategyRuntime.Close`(:301, **조건부**) → `alertControl.Close`(:317) →
  `cancel`(:323) → `stopWatching`(:325). 실행은 역순.
- **fallback 은 오늘 하나뿐이다**: B18 의 강등. B15·B16·B20 은 fallback 없이 죽고,
  그 죽음이 autostart 앞에서 **영구 기동 루프**가 된다 — a109가 닫는 구멍.
- 강등 후에도 `defer …Close()` 를 등록하려면 **nil 수신자 가드**가 필요하다: 세 서버
  타입 모두 `Close` 가 nil 수신자에서 안전하지만(각 Close 의 `if s == nil`), a108 이
  B17 에서 택한 규율은 「강등 경로는 Close 를 부를 이유가 없다」이므로 같은 형태를 쓴다.

## Safety conclusion

- Safe edit boundary: **B15·B16·B20 의 `return err` 세 줄과 그 자리에 들어가는 강등
  블록**. 다른 분기·다른 return·defer **등록 순서**는 건드리지 않는다.
- High-risk impact: **yes** — 엔진 기동 경로다. 다만 방향은 보수적이다: 오늘의 결과가
  「전 포지션 무보호(영구 기동 루프)」이고 강등 후 결과는 「그 표면 하나 없음」이다.
  policy command 강등이 잃는 것은 격리 해제 표면이므로 **격리된 포지션의 무보호가
  유지된다** — fatal 대비 개선으로만 성립하는 논증이다(design D3, freeze P0-4).
- 금지: 강등 보고에 critical 등급·obs 등급표 등재·원장 outbox 적재 중 어느 것도 쓰지
  않는다(`engineStrategyProjectionDegradedEvent` 주석 :330–359 가 정본).
