# a087 변이 원장 — 2026-10-29

## 범위와 원본 확인

- 실행 방식: `go test -overlay`로 사본만 대체했으며 production 파일은 바꾸지 않음. 구조 시험 두 건은 원본을 `apply_patch`로 잠시 변이한 뒤 즉시 원복함.
- 원본 SHA-256: `exit_lower_limit.go` `8c858f055b04c8fe22e1b994e1bd3f76d8b21db74202bafedbf921e3979d8dad`, `exitloop.go` `27ae828542d30f31271138d62866a933ee582d3b92d87c8e47ae1298e4e7627d`, `exit_order_log.go` `419352e976848c9e210a08cf897d3495f84029e8ca54ed2bfcb79fbc2c2bd7b1`, `a087_protective_exit_attempt.go` `5776c7ba6edafd8fe481bc1535714e4a0dd3813991a2efd9b332d23d3dd96f0d`, `a087_price_limits_raw.go` `15324c6283dd447beeec1432ea9b60498fd72e62df8c176b9153422e4ef635a7`.
- 무변이 대조군: `go test ./internal/app/engine ./internal/journal ./internal/official -run '^TestA087' -count=1` 통과.

## 실행 결과

| 설계/작업 변이 | 변이 | 결과 | 잡은 시험 |
|---|---|---|---|
| D11·2.2 | 정규장 창 조건 제거 | CAUGHT | `TestA087ExitLimitPriceUsesOnlyEligibleKoreanProtectiveLowerLimit` |
| D11·2.2 | 15:30을 16:00으로 연장 | CAUGHT | 같은 시험 close boundary |
| D11·2.2 | 익절에도 하한가 허용 | CAUGHT | 같은 시험 take profit |
| D11·2.2 | US에도 하한가 허용 | CAUGHT | 같은 시험 us |
| D12.1·2.8 | timestamp 같은 날 대조 제거 | CAUGHT | stale·previous day 행 |
| D13.4·2.2 | 하한가 ≤ 관측가 대조 제거 | CAUGHT | observed lower·invalid 행 |
| D12.1·D13.4 | 0·선행 0 허용 | CAUGHT | zero·`050400` 행 |
| D13.1·2.5 | Place 뒤 첫 시도 표시 제거 | CAUGHT | `TestA087FirstProtectiveAttemptIsMonotonic` |
| D13.1·2.5 | 하한가 출처일 때만 첫 시도 표시 | CAUGHT | `TestA087FallbackProtectivePlaceStillClosesFirstAttempt` |
| D13.1·2.5 | 표시를 Place 앞으로 이동 | CAUGHT | `TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace` |
| D13.1·2.9 | 계좌 scope를 고정값으로 무시 | CAUGHT | journal other account 행 |
| D14.1·2.9 | STOP_LOSS_LADDER action 제외 | CAUGHT | `TestA087ProtectiveActionValuesFollowTypedActionSet` |
| D13.1·2.9 | 익절 action을 보호 집합에 추가 | CAUGHT | journal take-profit 행 |
| D12.2·2.4 | 한 주기에서 publish 뒤 다음 종목도 읽음 | CAUGHT | `TestA087LowerLimitRefresherReadsAtMostOneSymbolPerCycle` |
| D12.2·2.4 | 만료된 positions 사전 읽기 뒤 GET 계속 | CAUGHT | `TestA087ExpiredPrefetchContextSkipsLowerLimitWire` |
| D12.2·2.4 | 만료된 attempt 사전 읽기 뒤 GET 계속 | CAUGHT | `TestA087ExpiredAttemptReadSkipsLowerLimitWire` |
| D11·2.4 | 장외 갱신 창 대조 제거 | CAUGHT | `TestA087LowerLimitRefresherSkipsOutsideKoreanLimitWindow` |
| D14.2·2.4 | 시작 정숙 5분을 0으로 | CAUGHT | `TestA087LowerLimitRefresherStartsQuietForFiveMinutes` |
| D15.1·2.4 | admission nil도 갱신 가능 | CAUGHT | `TestA087NilAdmissionMakesNoLowerLimitWire` |
| D15.1·2.4 | 401 재전송 admission을 첫 wire만 통과 | CAUGHT | `TestA087Every401WireRequiresAdmission` |
| D15.4·2.4 | 생성자 literal nil을 typed nil 식으로 변경 | CAUGHT | `TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil` |
| D12.6·3.1 | 로그 price를 baseline으로 분리 | CAUGHT | `TestA087ExitOrderLineKeepsChosenPriceAndSource` |

## 미실행·제외와 살아남은 변이

- 살아남은 유효 변이: 없음.
- a166 소유 wire admission의 실제 `doRequest` 재전송·429·공유 예산 변이는 a087 범위 밖임. a087은 test double의 401 세 wire 계수만 고정하며, a166 착지 전 production admission은 literal nil임.
- 5.4 사람 실측과 5.5 활성화는 지시대로 실행하지 않음.
