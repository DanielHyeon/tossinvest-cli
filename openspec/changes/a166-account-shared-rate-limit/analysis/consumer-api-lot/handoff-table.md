# a166 → a087/a169 인계 증거 표 (초안)

대상 consumer 제품 후보 `89c9a5b7eb9b1ea303e68fe60272ef84502eb557` (base `6e0968222993756a41065044996f674b35e20a27`), 아직 독립 검증/최신 main 결합 전. 엔진 poll 착지 `bc8ae32ee`와 결합 후 선택 검증 및 최종 착지 SHA를 추가해야 한다. 아래는 주어진 고정 SHA 범위의 fixture이며 운영·실 브로커·인간 실측이 아니다.

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

## 작업표 판정 제안

`/tmp/a166-completion-crosswalk.md`의 cf0 시점 손읽기·lot별 Manager 증거를 재사용하되 최종 완성으로 확대하지 않는다. 2.1–2.3은 각 lot 원문 및 FLM/BTM 대조 뒤 개별 판정; 2.4의 모든 공식 wire/caller 구조 census는 진행 중이므로 pending, 2.5의 최종 통합 검증도 pending. 3.1–3.2의 전체 프로필·호출자 census와 결합 검증 pending. 3.3 engine poll 구현은 bc8ae32ee 이후 최신 결합 증거 확인 후 판정; 3.4 matrix와 이 인계표를 최종 SHA에 동기화한 후 판정하며 후속 소비자 구현을 완료로 표시하지 않는다. 3.5 최종 통합 및 독립 검증 pending. 4.2 사람 namespace·참여자 확인과 a087 5.4/5.5는 미완이며 archive 금지. 소스/test/운영 실행 없음.

후속 test delta `7609ff48d`: `TestA166ConsumerToken429CrossFactorySnapshot`은 두 factory 간 Token429 event와 GET0; `TestA166ConsumerPriceLimitsTwo401AdoptAndMint`는 실제 한 logical 요청 cached GET401→adopted GET401→token 발급→GET200, GET3+token1=4 wire, 각 pending 및 첫 Write 간격을 검증했다. 부모 89c9a5b7e의 테스트 보완이며 source 변경0. 작성자 및 독립 Go1.25 선택 normal/tagged race PASS. enginepoll 이후 최종 통합 SHA와 선택 결과는 게시 기록에 추가한다.
