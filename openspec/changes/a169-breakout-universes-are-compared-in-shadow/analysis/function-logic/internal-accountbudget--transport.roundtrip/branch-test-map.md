# Branch Test Map: `Transport.RoundTrip`

사후 작성: 첫 source 수정 전 상세 BTM은 없었다. 원본 AST·source와 현재 초안 기준이며 빈 미시험 행을 PASS로 간주하지 않는다. `기존`은 저장소에 시험이 존재함을 뜻하며 이번 실행 증거가 아니다. `초안 GREEN`은 /tmp 원문에 해당 개별 실행이 있을 때만 표기한다. 전체 선택 normal/tagged race와 독립 검토는 미실행이다.

| ID | 시나리오 | 정확 테스트 | 현재 증거/남은 일 |
|---|---|---|---|
| G01 | pre-write vs 이미 write, stop drain | `TestAttemptClassificationAfterDrain` | 기존; 재실행 예정 |
| G02 | collector rate 후 같은 logical 요청 | `TestObservedTransportRateLimitAndPoison` | 기존; body-close 오류 별도 확인 필요 |
| G03 | 입장 전 거절·body | `TestTransportPreAdmissionIsNotSent` | 기존; priority/URL 개별 case 일반화 금지 |
| G04 | proxy URL/error | `TestProxySelectionRefusedBeforeAdmission` | 기존; 재실행 예정 |
| G05 | Acquire 실패/aux busy | `TestA169FenceAuxBusyAndCancellationNotSticky` | 초안 GREEN `/tmp/a169-fence-cleanup-green.log` |
| G06 | Poll POST 및 Begin fault | `TestObservedTransportPollRejectsPost`, `TestObservedTransportBeginFaultClosesBody` | 기존; fence 적용 Begin 오류 별도 추가 필요 |
| G07 | poll defer 및 cancel | `TestObservedTransportRateLimitAndPoison`, `TestObservedTransportCancelAfterBeginSettles` | 기존; 재실행 예정 |
| G08 | Begin 실패 vs 성공한 뒤 취소/settle, release failure | `TestObservedTransportBeginFaultClosesBody`, `TestObservedTransportPreSendReleaseError`, `TestA169FenceAuxBusyAndCancellationNotSticky` | 기존 둘은 재실행 예정, fence 취소 초안 GREEN |
| G09 | HTTP1 fresh/body ownership | `TestTransportOwnsFullBody` | 기존; 재실행 예정 |
| G10 | SplitHostPort 실패 | 해당 분기 개별 fixture 없음 | 미검증, 기존 경로 변경 없음 |
| G11 | TLS 설정 nil/clone 및 verified TLS | `TestTransportOwnsFullBody` | 기존; 선택 시험에서 조건별 상세 증명 미주장 |
| G12 | explicit dialer 및 실패 | `TestA169FenceAuxBusyAndCancellationNotSticky` | 초안 GREEN; nil default dial 분기 미검증 |
| G13 | handshake/gate 취소 경합 | `TestCancelBlockedTLSWriteDrains`, `TestAttemptClassificationAfterDrain` | 기존; 재실행 예정 |
| G14 | nil response, context 정상 settle·sent 분류 | `TestObservedTransportNoResponseAfterSent`, `TestA169FenceAuxBusyAndCancellationNotSticky` | 기존 전자 미실행, 후자 초안 GREEN |
| G15 | 429/Finish 오류가 기존 응답 보존 | `TestObservedTransportFinishFaultPreservesResponse`, `TestA169FencePostcommitFinishConcurrentWaiter`, `TestA169FenceTransportAfterEventAndOwn429` | 후자 둘 초안 GREEN `/tmp/a169-fence-finish-race.log`, `/tmp/a169-fence-green-first.log` |
| G16 | 검증 TLS state 보존 | `TestTransportOwnsFullBody` | 기존; 세부 nil 경계 미검증 |
| G17 | EOF/close/cancel 정확 종료 | `TestTransportOwnsFullBody`, `TestCancelUnreadResponseReleasesOwnership`, `TestA169FenceBodyCloseBeforeRelease` | 기존 재실행 예정; 후자 mock body 초안 GREEN, 실제 TLS close failure 아님 |
| F01 | nil·zero·pending 및 잘못된 Budget | `TestA169FenceInvalidOwnerAndPending` | nil/zero/busy/orphan 초안 GREEN; 다른 Budget HTTP·closed·dirty 추가 필요 |
| F02 | A unlock 뒤 오류와 B 획득 | `TestA169FenceReleaseFailureConcurrentWaiter` | 초안 GREEN `/tmp/a169-fence-release-race.log`: channel+TryLock, B wire0. mutation 및 race 재실행 필요 |
| F03 | 일반 Aux busy·Begin 성공 후 단순 취소 | `TestA169FenceAuxBusyAndCancellationNotSticky` | 초안 GREEN; Begin 실패 sticky 별도 fixture 필요 |
| F04 | release/stop/Finish 실패 게시 | `TestA169FenceReleaseFailureConcurrentWaiter`, `TestA169FencePostcommitFinishConcurrentWaiter` | 초안 GREEN, stop fault 별도 미검증 |
| F05 | 429, Finish postcommit, body close | `TestA169FenceTransportAfterEventAndOwn429`, `TestA169FencePostcommitFinishConcurrentWaiter`, `TestA169FenceBodyCloseBeforeRelease` | 초안 GREEN; body failure는 직접 wrapper, 실제 TLS close fault 아님 |
| F06 | 과거 event 대비 새 event / 다른 process | `TestA169FenceBaselineAndEvents`, `TestA169FenceCrossProcessToken429` | 초안 GREEN `/tmp/a169-fence-unit-green.log`, `/tmp/a169-fence-official-first.log`; historical gap/ring overrun/rollback 추가 필요 |
| F07 | actual official 401→mint→429→sticky | `TestA169FenceOfficial401TokenAndSticky` | 초안 GREEN `/tmp/a169-fence-official-first.log`; cached/adopt 별도 fixture 필요 |

필요 비시험 증거: `shapes.json`의 기존 RoundTrip 및 새 fence 함수 shape pin은 값 흐름 의미 검증이 아니다. 시험용 private hooks 두 개는 Manager 승인된 컴파일 seam, non-test 할당 0 AST 확인 예정. No-fence는 G01–G17 동일 동작을 기존 선택 회귀로 확인한다. 실제 first-write timing, 비협조 프로그램·운영 namespace·영속 abort는 범위 밖.

## 원본 AST 분기 41개 교차표

G는 위 흐름 묶음이며 AST B 번호와 다르다. named test는 관련 선택 후보이며 개별 분기의 실행·충족을 자동으로 증명하지 않는다. 미검증 명시 외의 기존 시험도 이 작업에서 아직 재실행하지 않았다.

| AST ID | kind | 원본 행 | 흐름 | named test 또는 미검증 |
|---|---|---:|---|---|
| B1 | if | 192 | G01 | TestAttemptClassificationAfterDrain |
| B2 | if | 198 | G02 | TestObservedTransportRateLimitAndPoison |
| B3 | if | 200 | G02 | TestObservedTransportRateLimitAndPoison |
| B4 | if | 203 | G02 | TestObservedTransportRateLimitAndPoison |
| B5 | if | 209 | G03 | TestTransportPreAdmissionIsNotSent |
| B6 | if | 210 | G03 | TestTransportPreAdmissionIsNotSent |
| B7 | else | 212 | G03 | TestTransportPreAdmissionIsNotSent |
| B8 | if | 217 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B9 | if | 219 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B10 | if | 221 | G04 | TestProxySelectionRefusedBeforeAdmission |
| B11 | if | 228 | G05 | TestTransportPreAdmissionIsNotSent |
| B12 | if | 229 | G05 | TestTransportPreAdmissionIsNotSent |
| B13 | if | 232 | G05 | TestTransportPreAdmissionIsNotSent |
| B14 | if | 240 | G06 | TestObservedTransportPollRejectsPost / TestObservedTransportBeginFaultClosesBody |
| B15 | if | 242 | G06 | TestObservedTransportPollRejectsPost / TestObservedTransportBeginFaultClosesBody |
| B16 | else | 244 | G06 | TestObservedTransportPollRejectsPost / TestObservedTransportBeginFaultClosesBody |
| B17 | if | 247 | G07 | TestObservedTransportRateLimitAndPoison / TestObservedTransportCancelAfterBeginSettles |
| B18 | if | 250 | G07 | TestObservedTransportRateLimitAndPoison / TestObservedTransportCancelAfterBeginSettles |
| B19 | if | 255 | G08 | TestObservedTransportBeginFaultClosesBody / TestObservedTransportPreSendReleaseError |
| B20 | if | 257 | G08 | TestObservedTransportBeginFaultClosesBody / TestObservedTransportPreSendReleaseError |
| B21 | if | 260 | G08 | TestObservedTransportBeginFaultClosesBody / TestObservedTransportPreSendReleaseError |
| B22 | if | 265 | G08 | TestObservedTransportBeginFaultClosesBody / TestObservedTransportPreSendReleaseError |
| B23 | if | 283 | G10 | 미검증: SplitHostPort 오류 |
| B24 | if | 288 | G11 | TestTransportOwnsFullBody; TLS nil/clone 개별 미검증 |
| B25 | else | 290 | G11 | TestTransportOwnsFullBody; TLS nil/clone 개별 미검증 |
| B26 | if | 296 | G12 | TestA169FenceAuxBusyAndCancellationNotSticky; default dial 미검증 |
| B27 | if | 300 | G12 | TestA169FenceAuxBusyAndCancellationNotSticky; default dial 미검증 |
| B28 | if | 306 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B29 | if | 310 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B30 | if | 314 | G13 | TestCancelBlockedTLSWriteDrains / TestAttemptClassificationAfterDrain |
| B31 | if | 320 | G14 | TestObservedTransportNoResponseAfterSent / TestCancelBlockedTLSWriteDrains |
| B32 | if | 321 | G14 | TestObservedTransportNoResponseAfterSent / TestCancelBlockedTLSWriteDrains |
| B33 | if | 328 | G14 | TestObservedTransportNoResponseAfterSent / TestCancelBlockedTLSWriteDrains |
| B34 | if | 333 | G14 | TestObservedTransportNoResponseAfterSent / TestCancelBlockedTLSWriteDrains |
| B35 | if | 342 | G15 | TestObservedTransportRateLimitAndPoison / TestObservedTransportFinishFaultPreservesResponse |
| B36 | if | 345 | G15 | TestObservedTransportRateLimitAndPoison / TestObservedTransportFinishFaultPreservesResponse |
| B37 | if | 348 | G15 | TestObservedTransportRateLimitAndPoison / TestObservedTransportFinishFaultPreservesResponse |
| B38 | if | 355 | G15 | TestObservedTransportRateLimitAndPoison / TestObservedTransportFinishFaultPreservesResponse |
| B39 | if | 361 | G16 | TestTransportOwnsFullBody; nil TLS 보완 개별 미검증 |
| B40 | select | 369 | G17 | TestTransportOwnsFullBody / TestCancelUnreadResponseReleasesOwnership |
| B41 | if | 371 | G17 | TestTransportOwnsFullBody / TestCancelUnreadResponseReleasesOwnership |
