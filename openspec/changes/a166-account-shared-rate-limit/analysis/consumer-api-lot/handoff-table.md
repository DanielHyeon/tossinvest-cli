# a166 → a087/a169 개발 인계 증거 표

consumer 후보 `89c9a5b7e` + test delta `7609ff48d`를 engine poll `bc8ae32ee`와 합친 게시 SHA는 `d9827f31661e60209b75f8fa0370e9525e48e225`다. 합성 후 정확 선택 normal/tagged race의 accountbudget/runtime/engine 세 패키지가 PASS했고 scoped a166/a169 strict·PM·diff도 통과했다. 원문 `/tmp/a166-consumer-merged-normal.log`, `/tmp/a166-consumer-merged-tag-race.log`. 각 표 항목은 고정 fixture 범위이며 운영·실 브로커·인간 실측이 아니다.

| 대상 | 정확 근거 | 증명 범위와 한계 |
|---|---|---|
| Budget snapshot | `internal/accountbudget/consumer_snapshot_linux.go` `Budget.ObservationSnapshot`; `TestA166ConsumerSnapshotBoundary`, `TestA166ConsumerSnapshotCopyAndPastCursor`, `TestA166ConsumerSnapshotPinsDuringClose` at 89c9a5b7e | nil/context/closed/unobserved·wrong/future/past cursor, view 사본, 소유자 Close 경쟁. 다음 wire에 대한 원자 허가가 아님. |
| 공유 Other429 | `TestA166ConsumerTwoFactoriesPriceLimits` at 89c9a5b7e | 두 Factory의 서로 다른 Budget 포인터 각각에 client 결속, token 성공 뒤 Other429를 다른 factory의 wrapper로 조회, 성공 noevent. Token429 교차 wrapper는 아래 delta의 별도 시험이다. |
| a087 실제 PriceLimits | `TestA166ConsumerPriceLimits401EachWire` at 89c9a5b7e | 한 logical PriceLimits 호출: token200→GET401→token200→GET200, 총 첫 plaintext Write 4회, pending 종류/순서 및 ≥350ms. 단일 요청에서 GET401 두 번을 시험한 것은 아님. |
| 우선순위·수명 | `TestA166ConsumerTwoFactoriesPriceLimits`, `TestA166ConsumerTokenSentThenAuxGETRefused`, `TestA166ConsumerPriceLimitsBodyOwnership`, `TestA166ConsumerObservedBodyCloseReleasesWire` at 89c9a5b7e | 최초 Aux 거절 token/GET0; 이미 token이 나간 경우 token1/GET0; EOF/cancel 및 별도 응답 Close 후 wire 소유권 해제. 이미 시작한 wire의 소급 취소나 유한 wall-time 상한은 아님. |
| 오류 보존 | `TestA166ConsumerFinishFaultPreservesHTTP429` at 89c9a5b7e | postcommit Finish fsync 오류 보고, 해당 429 결과를 NOT_SENT로 바꾸지 않음. 다른 failure 단계가 모두 event를 남긴다는 뜻은 아님. |
| 선행 generic 401/cache | `internal/accountbudgetruntime/v2_token_test.go::TestA166RuntimeToken401AndCached` at `cf0f0919e19bd384995b8bfb7fb8ee44a42031a7` | 실제 Accounts 두 논리 호출에서 token2+GET3=5 wire. 첫 GET 한 번만 401, 다음 GET 성공 및 별도 cache-hit GET. 단일 논리 요청 401→401→200을 이 fixture가 검증한다고 주장하지 않는다. |
| 별도 프로세스 wire | `internal/accountbudget/runtime_wire_external_test.go::TestA166RuntimeFactoryTwoProcessesTLSStarts` at cf0f0919e | 두 프로세스 각각 token2+GET2=4 wire, 한 차례 401; 총 8 plaintext Write 간격 확인. 가격 데이터나 실제 다른 namespace 배포 검증은 아님. |
| dormant | `internal/app/engine/a087_admission_dataflow_test.go::TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil`; `a087_lower_limit_refresher_test.go::TestA087NilAdmissionMakesNoLowerLimitWire` at 89c9a5b7e | 생산 admission nil 및 nil 갱신 GET0. board 활성화 아님. |
| clock/ring | 기존 `TestA166ObservationCursorAndRing`, `TestA166ObservationSnapshotGenerationRace` 및 선행 clock leaf 증거 | 공통 leaf·generation 동작, 새 wrapper에 대한 선택 시험과 결합. 운영 reboot/power-loss 실측 아님. |
| a087 후속 | a087 D16 및 tasks 5.4/5.5 | 사람 측정 5.4, board/epoch/만료 token/429 정지 및 활성화 5.5 미완. 현재 admission nil. |
| a169 후속 | `analysis/a166-bridge-draft.md`, 독립 연구·검토 | draft, prewire fence/영속 abort/실제 engine identity 미구현·미승인; snapshot API만으로 증명 불가. |

## 현재 작업표 판정

최신 source census와 합성 검증은 `9dab3dda4`, 상태표 게시본은 `7d68c9ef6`이다. 현재 판정은 `../development-status.md`와 tasks를 따른다. 2.2–2.5·3.2–3.5는 문서화한 개발 범위 완료이며, 2.1/3.1은 사전 FLM 절차 예외를 숨기지 않기 위해 완전 준수 체크를 하지 않았다. 구조 검사는 83 함수 형태·116 selector와 독립 손읽기 및 지정 행동 회귀를 결합하며 임의 동적 호출·실행 중 참여자의 전체 증명은 아니다. 운영 절차 4.1은 초안, 실환경 결속4.2·운영4.5 및 a0875.4/5.5·a169 소비자 연결은 미완이다. archive하지 않는다.

후속 test delta `7609ff48d`: `TestA166ConsumerToken429CrossFactorySnapshot`은 두 factory 간 Token429 event와 GET0; `TestA166ConsumerPriceLimitsTwo401AdoptAndMint`는 실제 한 logical 요청 cached GET401→adopted GET401→token 발급→GET200, GET3+token1=4 wire, 각 pending 및 첫 Write 간격을 검증했다. 부모 89c9a5b7e의 테스트 보완이며 source 변경0. 작성자 및 독립 Go1.25 선택 normal/tagged race PASS. enginepoll 이후 최종 통합 SHA와 선택 결과는 위 게시 기록과 같다.
