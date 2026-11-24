# a199 변이 원장 (2.7)

- 하네스 `run_mutations.py`, 원문 `run-1.txt`. 사본 = `git worktree add --detach` 로 만든 고정 커밋 `b856fdc08` 의 사본(시작 sha 단언), 끝나면 제거(`copy removed: True`, `git worktree list` 에 잔재 없음).
- 무변이 대조군: 네 패키지(engine · cmd/tossctl · execgw · attest)에서 `-run A199` rc 0 → 시작.
- 닿음: 변이마다 대상 문자열 1회 적중 · 바이트 변화 단언, 원복 뒤 바이트 동일 단언, 끝에 사본 `git status` 깨끗.
- **결과: 20 / 20 CAUGHT, SURVIVED 0, INVALID 0, BUILD 0.**

| 변이 | 잡은 시험 |
|---|---|
| M1 만료 시 계좌 확인 삭제 | RefusalsStayRefusals · TheWatchJudgesAnExpiredForeignAttestationAsTheInterlockDoes |
| M2 닫힌 집합 넓히기(불완전도 낡음으로) — design 의 「ErrMissing 추가」 대신: ErrMissing 은 Load 오류라 판정 함수에 닿지 않음(도달 불가), 같은 축(닫힌 집합 경계)을 ErrIncomplete 로 넓혀 잼 | RefusalsStayRefusals |
| M3 빈 endpoint 항목 가드 삭제 | RefusalsStayRefusals · TheWatchJudges… |
| M4 형식 하한 삭제 | AnInvalidFormatVersionDoesNotGoExitOnly |
| M5 뒤 조항 검사 생략 | Stale…Gateway… · Stale…Transport… · StaleWithAnotherClause… |
| M6 래치 세우기 삭제 | AStaleAttestationStartsExitOnly · ARestartRederives… · AnExitOnlyRuntime… |
| M7 Verified 를 true 로 | AStaleAttestationStartsExitOnly · AnExitOnlyRuntime… |
| M8 Guardian nil 유지 | AStaleAttestationStartsExitOnly · AnExitOnlyRuntime… |
| M9 감사 action 을 수락으로 | AStaleAttestationStartsExitOnly · ARestartRederives… |
| M10 감시 판정을 따로 두기 | TheWatchJudges… |
| M11 worker 조건에서 Verified 삭제 | TheStrategyWorkerStillRequiresAVerifiedGate |
| M12 감시 알림 조건을 파일 판정으로 | AHealthyFileOnAnExitOnlyEngineIsAnnouncedOncePerDay |
| M13 보호 수명 문장 조건에서 청산 전용 삭제 | TheExitOnlyStartSaysWhatItIs |
| M14 · M15 exit 관측 · 대사 생성을 Verified 로 | AStaleAttestationStartsExitOnly(+ 연결 시험) |
| M16 회복 안내 키 보존 삭제(R2-3) | AHealthyFile…OncePerDay |
| M17 낡음 표지 안 세움 | 4 시험 |
| M18 runEngineRun 이 청산 전용을 게이트 OFF 로 | AnExitOnlyContextReachesTheRuntimeAndSaysSo |
| M19 latchOrder 에서 새 사유 빼기 | TheStaleAttestationLatchReadsBeforeTheModeAndAfterStrategyCentral |
| M20 NamesAccount 를 가림 허용 규칙으로 | NamesAccountIsVerifysAccountComparison |

※ M11 의 시험은 구현 리뷰 r1 처분 1 로 `TestA199TheProjectionAssemblyWorkerConditionStillReadsVerified` 로 이름이 바뀌었다(투영용 조립의 worker 조건 — 실제 런타임 전략 worker 는 Verified 를 안 읽고 래치가 막는다).

## 추가분 (구현 리뷰 r1 처분 뒤)

- 하네스 `run_mutations_r1.py`, 원문 `run-r1.txt`. 같은 규율 — 시작 sha `723554957` 단언, 무변이 대조군 세 묶음(`-run A199` · `A199|TestEachInterlockClauseHasALine` · 태그 `tossos_testseams` 전략 dispatch 시험) rc 0, 원복 뒤 사본 깨끗, 사본 제거.

| 변이 | 결과 |
|---|---|
| M21 nil 진입 게이트 가드 삭제(처분 2) | CAUGHT — NoEntryGateMeansNoExitOnlyStart |
| M22 청산 전용 감사 실패 무시(처분 3) | CAUGHT — AnUnrecordableExitOnlyStartIsRefused |
| M23 gate_exit_only 를 모든 줄에(처분 5) | CAUGHT — TheExitOnlyFieldIsAbsentFromOtherStarts |
| M24 래치 세우기 삭제 — 전략 dispatch 시험만 돌림 | **SURVIVED(예상)** — 그 시험은 `exitOnlyStart` 를 거치지 않고 자기 EntryGate 에 래치를 직접 세운다. 그 시험이 재는 것은 「래치가 서 있으면 실제 dispatch 경로가 브로커 주문 0」이고(래치 없는 대조 사례는 주문 1), 「청산 전용 기동이 래치를 세운다」는 M25 가 잰다 |
| M25 래치 세우기 삭제 — A199 시험 전체 | CAUGHT — ARestartRederivesTheLatchFromTheFile · AScopeLostRuntimeRefusesABuyAndSubmitsTheStopLoss · AStaleAttestationStartsExitOnly · AnExitOnlyRuntimeRecoversAndSubmitsTheStopLoss |
| M26 지운 `ErrExpired` 조항 줄을 되살림(대조) | SURVIVED(예상) — 거절 사슬에 그 sentinel 이 실리는 입력이 0 이라(처분 6, `TestA199NoRefusalCarriesAStaleSentinel` 9 사례) 줄의 유무가 행동을 안 바꾼다는 확인 |

## 추가분 (gstack review G2 · G3 · G9 — Manager 처분 2026-11-24)

- 하네스 `run_mutations_g2.py`, 원문 `run-g2.txt`(전체) · `run-g2-m27.txt`(M27 재실행). 시작 sha `fadd2fc4c` 단언, 무변이 대조군 넷(engine `A199` · engine 조립 G3 시험 · 태그 dispatch 시험 · journal `A199`) rc 0, 원복 뒤 사본 깨끗, 사본 제거. 빌드 실패는 CAUGHT 로 세지 않음.
- 첫 판의 M27(`if false`)은 `proven` 미사용으로 **빌드 실패 — CAUGHT 로 세지 않음**. 컴파일되는 같은 뜻(`(readErr != nil || !proven) && false`)으로 재실행.
- 중간에 접두 인자 없는 옛 하네스로 잘못 띄운 재실행 하나를 중단하고 그 사본을 `git worktree remove --force` 로 지움(원장에 결과 없음).

| 변이 | 결과 |
|---|---|
| M27 증거 검사 삭제(낡으면 늘 내림) | CAUGHT — AStaleAttestationWithoutAnEngineOrderIsRefused · AnInstallWithoutAnEngineOrderRefusesAStaleStartUntilAnOrderIsRecorded · TheWatchTellsARefusalForAStaleFileWithoutAnEngineOrder |
| M28 SQL 계좌 조건 삭제 | CAUGHT — HasConfirmedPlaceCountsOnlyAConfirmedPlaceOfThatSourceOnThatAccount |
| M29 SQL source 조건 삭제 | CAUGHT — 같은 시험 |
| M30 SQL 상태 조건을 CONFIRMED 밖으로 | CAUGHT — 같은 시험 |
| M31 SQL kind 조건 삭제(CANCEL 도 셈) | CAUGHT — 같은 시험 |
| M32 읽기 실패 시 내림 | CAUGHT — AStaleAttestationWithoutAnEngineOrderIsRefused · TheWatchTellsARefusal… |
| M33 유효 경로에서도 원장 읽기 | CAUGHT — AValidOrGateOffStartReadsNoLedger · TheWatchForecast…AndAHealthyFileReadsNoLedger |
| M34 감시의 기한 뒤 예측을 증거와 무관하게 | CAUGHT — TheWatchForecastFollowsTheEvidence… |
| M35 G9 거절 갈래의 표지 지우기 삭제 | CAUGHT — NoEntryGateMeansNoExitOnlyStart |
| M24' 래치 세우기 삭제 — 전략 dispatch 시험만(이제 exitOnlyStart 를 거침) | **CAUGHT** — TheStaleLatchStopsTheRealStrategyDispatchBeforeTheBroker(r1 의 M24 SURVIVED 를 메움) |
| M36 래치 세우기 삭제 — 조립된 게이트웨이 시험만(G3) | CAUGHT — TheBootLatchIsTheGateTheAssembledGatewayChecks |
