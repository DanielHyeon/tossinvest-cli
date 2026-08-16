# Function Logic Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go` (184-357)
- AST evidence: `ast.json` — AST 분기 23 · return 10 · defer 9
  (source_sha256 `3d941a4616837f73ce2a2acf87ed4bfada5eaccdf4b77fa966b46ac4cd972209`,
  **a109 §2.1–§2.2 편집 후 재생성**)
- Risk scan: `risk-pattern-report.md`
- **편집 요약**: 기준(base) 판은 분기 20 · return 13 이었다. 형제 셋의 fatal
  `return err`(:274 · :279 · :315)가 사라지면서 **return 이 13 → 10 으로 줄었고**, 그 셋이
  각각 「Close 등록(nil 가드)」+「강등 보고」 두 분기로 갈라져 **분기가 20 → 23 으로**
  늘었다. 늘어난 셋은 B15·B17·B22(nil 가드)이고, 그 짝인 B16·B18·B23 이 강등 보고다.
- **편집하지 않은 이웃**: B14(`NewPositionPolicyCommandService`, return :271)와
  B21(`ectx.AlertOperations()`, return :337)은 endpoint 기동이 아니라 **in-process 조립**
  이므로 design D3 표에 없다 — fatal 유지가 계약이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cmd.Context()` | nil 허용 | cobra | nil 이면 `context.Background()` (B1) |
| `root.configDir` | 비면 기본 journal 경로 | `engineJournalDir` | 해석 실패는 즉시 return (B2) |
| journal 디렉터리 flock | **동시에 하나** | `enginelock.Acquire(dir)` (:197) | `ErrAlreadyRunning` return (B3). **엔진 싱글턴의 유일한 권위 — D3 강등 논증 전체가 여기에 기댄다** |
| 엔진 조립 | config→journal(RW)→official broker→obs→기동 인터록 | `engineAssemble` seam | 인터록 미충족이면 절 열거 후 `errEngineInterlockUnmet` (B4·B5·B6) |
| `ectx.Automation.Verified` | true 여야 루프가 있다 | 조립 결과 | false 면 `errEngineGateOff` (B7) |
| verify run lock | 신선하면 같은 계좌를 다투는 중 | `runlock.Fresh` | `errVerifyInProgress` (B8·B9) |
| proc instance 토큰 | `/proc` 없는 커널에서는 없을 수 있음 | `engineProcInstance` | 없으면 마커에 안 싣는다 (B10) |
| 활성 마커 | 자문(advisory) | `enginelock.Hold` | 실패는 **거절이 아니다** — note 한 줄 (B11/B12) |
| position policy command server | **없어도 부팅은 계속된다 (a109)** | `enginePositionPolicyCommandStart` seam | 강등 + 보고 (B15·B16). 잃는 것: Preview/Apply·**격리 해제** |
| position policy runtime server | **없어도 된다 (a109)** | `enginePositionPolicyRuntimeStart` seam | 강등 + 보고 (B17·B18). 잃는 것: 관리 런타임 화면 |
| strategy projection endpoint | **없어도 된다 (a108)** | `engineStrategyProjectionStart` seam | 강등 + 보고 (B19·B20) |
| alert 운영 표면 조립 | 있어야 운영자가 승인한다 | `ectx.AlertOperations()` | **fatal 유지** (B21) — endpoint 아님 |
| alert control server | **없어도 된다 (a109)** | `engineAlertControlStart` seam | 강등 + 보고 (B22·B23). 잃는 것: 운영자 ack 표면 |

**관통 불변식 1 (a108 이 세우고 a109 가 일반화)**: 엔진이 소유한 표면 endpoint 의 기동
실패는 보호 루프의 기동을 막지 않는다. 배타성은 B3 의 flock 이고 어떤 control
디렉터리도 아니다.

**관통 불변식 2 (defer LIFO — §2.4 가 정적으로 핀한다)**: `lock.Release()` defer 는
:201 에서 **가장 먼저** 등록되므로 LIFO 상 **가장 나중에** 실행된다. 그래서 네 endpoint
Close(:294 policyControl · :302 policyRuntime · :327 strategyRuntime · :341 alertControl)는
**journal flock 을 쥔 채로** 돈다 — 이 프로세스의 잔재 제거와 다음 엔진의 회수·발행이
겹칠 수 없다. 회수 함수가 flock 을 자기 1차 방어로 인용하는 근거가 이 순서다(design D2).

**관통 불변식 3**: 강등 경로는 Close 를 부르지 않는다. 세 서버 타입 모두 nil 수신자에
안전하지만(각 `Close` 의 `if s == nil`), 조건을 여기 두면 그 **판단이 이 파일 안에**
남는다 — a108 이 B19 에서 택한 규율 그대로다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (:186) | `ctx == nil` | 없음 | 없음 | 간접 (모든 CLI 경로) |
| B2 (:192) | `engineJournalDir` 실패 | 없음 | 원인 그대로 (:193) | 미고정 — 환경 이상 |
| B3 (:198) | `enginelock.Acquire` 실패 | 없음 | `ErrAlreadyRunning` 등 (:199) | `TestTheJournalDirectoryIsLockedBeforeAnythingIsAssembled` |
| B4 (:208) | `engineAssemble` 실패 | 없음 | 인터록이면 B5, 아니면 원인 (:216) | `TestANonInterlockFailureIsReportedAsItself` |
| B5 (:209) | 미충족 절이 있다 | stderr 열거 | `errEngineInterlockUnmet` (:214) | `TestAnUnmetInterlockIsEnumerated` |
| B6 (:211) | 절 순회 | stderr 한 줄씩 | 없음 | 위와 같음 |
| B7 (:220) | 게이트 미검증 | 없음 | `errEngineGateOff` (:221) | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` |
| B8 (:230) | verify lock 경로 해석 성공 | 없음 | 없음 | `TestAFreshVerifyRunLockRefusesTheStart` |
| B9 (:231) | verify lock 이 신선 | stderr | `errVerifyInProgress` (:234) | 위 / `TestAStaleVerifyRunLockDoesNotRefuse` |
| B10 (:247) | proc instance 를 읽었다 | `marker.Identify` | 없음 | `TestTheReadySignalReachesTheMarkerThroughTheRuntimeSeam` |
| B11 (:250) | 마커를 못 잡았다 | stderr note | **없음 — 거절 아님** | 미고정 (기존 강등) |
| B12 (:254) | 마커를 잡았다 | stdout 한 줄 | 없음 | `TestTheMarkerIsHeldWhileTheLoopsRunAndRemovedAfter` |
| B13 (:266) | 루프 조립 실패 | 없음 | 원인 그대로 (:267) | `TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess` |
| B14 (:270) | policy command **service** 실패 | 없음 | 원인 그대로 (:271) — **fatal 유지** | 미고정 (endpoint 아님) |
| **B15 (:293)** | **policy command server 가 섰다** | `defer Close` 등록 | 없음 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` |
| **B16 (:296)** | **policy command 기동 실패** | stderr 세 줄 + obs Normal 이벤트 | **없음 — 부팅 계속** | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestADegradedSiblingSaysWhichSurfaceIsMissing` · `TestADegradedSiblingBootWritesNoUndeliveredOutboxRow` |
| **B17 (:301)** | **policy runtime server 가 섰다** | `defer Close` 등록 | 없음 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` |
| **B18 (:304)** | **policy runtime 기동 실패** | stderr + obs Normal | **없음 — 부팅 계속** | 위 셋과 같음 |
| B19 (:319) | projection 서버가 섰다 | `defer Close` 등록 | 없음 | `TestASucceedingProjectionIsStillServedAndClosed` |
| B20 (:329) | projection 기동 실패 | stderr + obs Normal (원장 행 0) | **없음 — 루프는 계속 돈다** | `TestAFailedStrategyProjectionDoesNotStopTheEngine` · `TestTheDegradedBootWritesNoUndeliveredOutboxRow` |
| B21 (:336) | alert 운영 표면 조립 실패 | 없음 | 원인 그대로 (:337) — **fatal 유지** | 미고정 (endpoint 아님) |
| **B22 (:340)** | **alert control server 가 섰다** | `defer Close` 등록 | 없음 | `TestSucceedingSiblingEndpointsAreStillServedAndClosed` |
| **B23 (:343)** | **alert control 기동 실패** | stderr + obs Normal | **없음 — 부팅 계속** | `TestAFailedSiblingEndpointDoesNotStopTheEngine` · `TestEveryEndpointCanFailAtOnceAndTheLoopsStillRun` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `enginelock.Acquire` | 1단계 배타성 | 실패는 fail-closed return | AST · engine.go:197 |
| `engineAssemble` (seam) | 2~4단계 조립 | `engine.UnmetInterlockClauses` 로 분류 | AST · `engine_test.go` |
| `runlock.Fresh` | 5단계 rate 예산 | 오래된 lock 은 거절하지 않는다 | AST · engine.go:231 |
| `enginelock.Hold` / `Identify` / `Release` | 6단계 자문 마커 | 실패는 note (B11) | AST · a102 |
| `engineRuntimeFactory` (seam) | 7단계 루프 집합 | 실패는 return (B13) | AST · `engine_runtime_branch_test.go` |
| `engine.NewPositionPolicyCommandService` | 정책 명령 서비스 조립 | 실패는 return (B14) | AST · engine.go:269 |
| `enginePositionPolicyCommandStart` (seam) | Preview/Apply·격리 해제 표면 | **실패는 강등** (B16) | AST · engine.go:292 · a109 T2 테스트 |
| `enginePositionPolicyRuntimeStart` (seam) | 관리 런타임 화면 | **실패는 강등** (B18) | AST · engine.go:300 |
| `engineStrategyProjectionStart` (seam) | 조회 전용 export | 실패는 강등 (B20, a108 D3-2) | AST · engine.go:318 |
| `reportEngineEndpointDegraded` | 네 강등의 공통 보고 | 동기 stderr + 비동기 Notify (자체 맵 참조) | AST · :297·305·330·344 |
| `ectx.AlertOperations` | 운영자 ack 표면 조립 | 실패는 return (B21) — **endpoint 아님** | AST · engine.go:335 |
| `engineAlertControlStart` (seam) | 운영자 ack endpoint | **실패는 강등** (B23) | AST · engine.go:339 |
| `rt.Run(runCtx)` | 보호 루프 셋의 실행 | 이 호출에 닿는 것이 「부팅 성공」의 정의다 | AST · engine.go:356 |

## State mutations and fallbacks

- **디스크**: journal flock 파일(B3), 활성 마커(B12), 그리고 endpoint 넷의 control
  디렉터리·descriptor·socket. 후자는 각 Start 가 만들고 대응 defer Close 가 지운다.
- **defer 스택(등록 순서)**: `lock.Release`(:201) → `ectx.Close`(:218) →
  `marker.Release`(:241) → `policyControl.Close`(:294) → `policyRuntime.Close`(:302) →
  `strategyRuntime.Close`(:327) → `alertControl.Close`(:341) → `cancel`(:352) →
  `stopWatching`(:354). 실행은 역순.
- **fallback 이 넷으로 늘었다**: B16·B18·B20·B23. 남은 fatal 은 조립 실패(B13·B14·B21)와
  부팅 전제(B2·B3·B4·B7·B9)뿐이다.

## Safety conclusion

- Safe edit boundary: **B15–B18·B22–B23 의 여섯 분기**. 다른 분기·다른 return·defer
  **등록 순서**는 건드리지 않았다(§2.4 의 정적 핀이 그것을 고정한다).
- High-risk impact: **yes** — 엔진 기동 경로다. 방향은 보수적이다: 편집 전 결과가
  「전 포지션 무보호(영구 기동 루프)」이고 편집 후 결과는 「그 표면 하나 없음」이다.
  policy command 강등이 잃는 것은 격리 해제 표면이므로 **격리된 포지션의 무보호가
  유지된다** — fatal 대비 개선으로만 성립하는 논증이다(design D3, freeze P0-4).
- 금지: 강등 보고에 critical 등급·obs 등급표 등재·원장 outbox 중 어느 것도 쓰지 않는다
  (engine.go:358–397 주석이 정본). B14·B21 을 함께 강등하는 것(표면이 아니라 서비스다).
