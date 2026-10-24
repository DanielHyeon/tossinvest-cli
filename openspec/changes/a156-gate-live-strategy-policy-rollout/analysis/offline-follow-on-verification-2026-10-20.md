# a156 기존 후속 leg 순수 경로 확인 — 2026-10-20

## 목적

새 후속 진입 계산기나 가상 broker 입력을 만들지 않고, 현재 lane별 순수 proposal 경로와 관련 선택 시험이 후속 leg의 계획 잔량·부분 체결·취소·위험 예산·예약 상태를 다루는지 확인했다. 이 문서는 a156 전체 구현 또는 live-ready 판정이 아니다.

## 입력 고정

- HEAD: `3415cf89f713ccdf61604963f29da9e0d5116294`
- 실행 원문, 종료 코드, 파일 hash는 아래에 그대로 포함한다. 임시 원문도 `/tmp/tossos-a156-offline-follow-on-20261020/`에 보존했다.

### 입력 manifest

```text
head=3415cf89f713ccdf61604963f29da9e0d5116294
7f309e45426d1cb740f2a76925d210ae74c88e2df4e05ce99fe9de5ad52d8b9b  internal/continuationlane/evaluator.go
a61222af6ec0de94234f1172779f3fe16cc15afa05ad6535483f4bd99ab27505  internal/continuationlane/allocation_risk_test.go
49d7c8fa46d17a4a7e8baa85d717d1b3e86ee32bbe4680d2087d7613c82b37d4  internal/continuationlane/proposal_test.go
525db732e564b987d6cf8d1d5ab29a2f54a18e8ed63d0600abf76c9dac692e58  internal/reversallane/evaluate.go
2a0dd0ee9f6701c74e6b60a2a2161218048b688c64428982966c7d172083ee91  internal/reversallane/allocation_risk_test.go
cf97509f7e54b2a72fae6057e3f00e8d1e3c61972aa8bb39140b83c4f5ed2fee  internal/reversallane/proposal_test.go
678e87a72f784bd0e924b1bc2a256ea6c31a3c7ae857d9d2a93b154a5ece17ed  internal/weeklyvaluelane/evaluate.go
dc40f1d82a43e821f2d74bfe819a15f19a6a3bc5d1aecadbbc3e0356b01edca9  internal/weeklyvaluelane/plan_risk_rr_test.go
6c9947be0620d42497574dbfeccd1d1c32053d3f733ea1b73bfecc5f8dca9af1  internal/weeklyvaluelane/reservation_test.go
c401f2b0c3aa5372c18f3f14973af85d7a4dd32e5d33339b6b5ee3f5ee989a66  internal/weeklyvaluelane/proposal_test.go
64b919d8dabdd93a5164d98759ba4d1865127d0b66beda3604dfda735822ec06  internal/reversallane/review_regression_test.go
098f0b9d9a078221b41ce1b0f1240e3930b5ccbb687690ddf5139e14a16676d4  internal/weeklyvaluelane/hardening_test.go
```

### 원문 출력과 종료 코드

```text
ok   github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane  0.014s
EXIT_CODE=0
ok   github.com/JungHoonGhae/tossinvest-cli/internal/reversallane  0.029s
EXIT_CODE=0
ok   github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane  0.024s
EXIT_CODE=0
```

### 원문 hash

```text
63426fff22fddc6e960988e7f9736e79a1795956b108bfdac32113435a95a687  continuation.log
76ca0ecff974ad514d24e641e23a23d5b594c5920e4d9bb3fad0fc65b0267b5f  reversal.log
566a01d202cf1b5c728d5d22a804d7e9a734ffbfa73e288a9bd056cdc680fdda  weekly.log
b44b6dc4520b20159b491cea99e97cf7eb73eeacdfd8c6deff4edb562adb430c  input-manifest.txt
```

## 기존 재사용 API와 선택 시험 범위

| lane | 순수 proposal API | 후속 leg 계약 |
|---|---|---|
| continuation | `internal/continuationlane/evaluator.go`의 `ProposeKR`/`ProposeUS` | cap을 소비하지 않는 `q_candidate`; 현재 leg ceiling의 미체결 잔량을 대상으로 한다. native plan/risk/evidence/stop/terms가 필요하다. |
| reversal | `internal/reversallane/evaluate.go`의 `ProposeKR`/`ProposeUS` | 동일하게 cap-free `q_candidate`; native plan/risk/evidence/terms 입력이 필요하다. |
| weekly value | `internal/weeklyvaluelane/evaluate.go`의 `ProposeKR`/`ProposeUS` | cap-free `q_candidate`이지만 durable reservation 상태와 7-leg ceiling이 필요하다. |

API 선언과 아래 선택 시험이 확인하는 범위는 현재 leg의 계획 잔량, 부분 체결/취소, 위험 예산, continuation의 손절 비후퇴, weekly reservation이다. 이 문서는 evaluator 내부의 모든 분기를 AST로 주장하지 않는다.

continuation의 `TestPartialCancelDoesNotReallocateAndA066CapStillBinds`, reversal의 `TestPartialCancelRetryCannotReallocateAndA066CapBinds`는 사용하지 않은 이전 leg 수량을 후속 leg에 재배분하지 않는 계약을 선택 시험으로 검증한다. 세 package의 `dependency_test.go`는 broker, journal, gateway, exit, toggle authority 의존을 금지한다. 아래 실행한 시험은 해당 순수 package 안의 계산·상태 전이 시험이며 live API나 주문 경로를 호출하지 않는다.

## 선택 실행

실행 명령은 정확히 12개 함수 이름을 regex로 지정했다. `-v` 없이 실행했으므로 원문은 개별 12건의 RUN/PASS가 아니라 이 12개 이름을 포함한 filter와 3개 package PASS만 보여준다. 전체 package, 전체 suite, race, gate는 실행하지 않았다.

```text
go test ./internal/continuationlane -count=1 -run '^(TestPairedContinuationProposalIsCapFreePlannedRemaining|TestPartialCancelDoesNotReallocateAndA066CapStillBinds|TestScaleInStopNeverRetreatsAndInvalidStopRefuses|TestRiskAdmissionUsesFilledHeldProposedCheckedBudget)$'
go test ./internal/reversallane -count=1 -run '^(TestPairedReversalProposalIsCapFreePlannedRemaining|TestPartialCancelRetryCannotReallocateAndA066CapBinds|TestFirstEntryEstablishesStopAndLaterScaleInCannotRetreat|TestRiskAdmissionUsesCheckedFilledHeldProposedBudget)$'
go test ./internal/weeklyvaluelane -count=1 -run '^(TestPairedWeeklyProposalIsCapFreeButKeepsDurableReservation|TestImmutableSevenLegAllocationAndNoUpwardReallocation|TestReservationCASUniquenessCorrectionReplayConsumeAndZeroRelease|TestApplyPositiveFillAtomicCommitsReservationAndRiskTogether)$'
```

| package | 선택 이유 | 결과 |
|---|---|---|
| continuationlane | proposal, leg 2 partial/cancel, stop monotonicity, filled+held+proposed budget | 지정 regex 4개 포함 package PASS, exit 0 |
| reversallane | proposal, leg 2 partial/cancel, ordinal 1 최초 stop authority와 future/unversioned stop 거절, filled+held+proposed budget | 지정 regex 4개 포함 package PASS, exit 0 |
| weeklyvaluelane | proposal/reservation, seven-leg immutable allocation, reservation replay, positive fill의 reservation+risk 원자 전이 | 지정 regex 4개 포함 package PASS, exit 0 |

## 결론과 미검증 범위

기존 순수 evaluator가 native `CampaignPlan`, `RiskState`, evidence, execution terms와 stop 후보를 받은 경우 후속 leg `q_candidate` 계산을 이미 제공한다. 별도 pure next-leg module, 새 allocator, 가상 broker/current-risk DTO는 중복이며 authority mint 위험이 있다.

기존 보고서가 미구현으로 지적한 부분은 후속 leg 평가 함수 자체가 아니다. 실제 권위 입력을 복원하고 새 `qFinal`을 산정한 뒤 reservation, lease, atomic admission과 dispatch를 묶는 통합 계약이다.

A161 detached evidence는 inspection 복사본이다. detached DTO를 native plan, FX seal, execution terms 또는 risk authority로 복원·mint해서 위 evaluator에 넣는 것은 이 확인 범위 밖이다. A160 read-only snapshot도 stored fact와 `FULL_PLAN_PREIMAGE_UNAVAILABLE`만 제공하며 full native plan을 복원하지 않는다.

다음은 이번에 실행하지 않았고 a156 전체 계약에도 여전히 필요하다.

- live 또는 durable journal에서 actual fill, pending reservation, sellable을 권위 있게 복원하는 계약
- 새 qFinal 계산과 atomic reservation/lease/admission
- restart/reconciliation과 unknown/cancel 처리의 운영 보장
- intraday deadline, auction, partial exit, halt, authoritative flat 증명
- a156의 별도 사람 승인과 live 실증

따라서 현재 확인은 기존 오프라인 순수 계산의 회귀 확인일 뿐이며, a156 task, readiness, archive 또는 운영 승인 상태를 변경하지 않는다.

research-first: 기존 lane source와 시험 재사용; grill-me: 기존 승인 계약 변경 없음; impeccable: N/A — UI 영향 없음.
