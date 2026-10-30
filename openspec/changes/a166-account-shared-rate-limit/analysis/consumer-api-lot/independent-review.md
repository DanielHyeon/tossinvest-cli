# a166 consumer API 독립 검토

후보 89c9a5b7eb9b1ea303e68fe60272ef84502eb557, base 6e0968222. 범위 내 생산 P0/P1 발견 없음. 인계 증거 표현에 아래 P2 명료화 필요. 소스 손읽기·기존 증거 확인과 정확 선택 시험이며 전체 배선 검증 아님.

`consumer_snapshot_linux.go:8-18`은 nil/context·closed/unobserved 검사 후 private store Snapshot에 위임한다. 기존 Snapshot pin이 Close와 경쟁하는 수명을 처리하며 view 사본·과거 cursor·wrong/future·closed/unobserved를 선택 fixture가 확인한다. snapshot은 prewire 권한·active engine 신원이 아니다. 지원 밖 stub은 fail-closed, Darwin compile은 실행·운영 검증이 아니다.

두 factory fixture는 서로 다른 Budget 포인터와 각 클라이언트 결속, 같은 group에서 Other 429 이벤트·cursor 전파를 확인한다. `TestA166ConsumerPriceLimits401EachWire`는 실제 PriceLimits의 token200→GET401→token200→GET200 네 plaintext Write, 350ms 간격·각 pending sequence/kind를 검증한다. 두 번 401은 이 fixture가 아니며 기존 generic official client 증거에 한정한다. Auxiliary 선거절 token/GET0과 token 전송 후 GET 거절을 분리하고, 별도 Body.Close/EOF/cancel 및 Critical 대기를 검증한다. Finish postcommit fault는 해당 fake 경계의 cleanup과 HTTP 429 유지, event를 확인한다. 실제 timestamp의 시장 의미나 실거래, a087 활성화/a169 독자 검증은 아니다.

P2: `consumer-api-handoff-plan.md:15`, `consumer-api-lot/review.md`의 'token/Other429' 표현은 두 종류의 cross-factory 429를 시사하지만 `TestA166ConsumerTwoFactoriesPriceLimits`에서는 token은 200, Other만 429다. `internal/accountbudgetruntime/v2_token_test.go:132-192`의 기존 token429 테스트는 자식 process에서 같은 manifest로 두 번째 factory를 열어 Token event를 기존 store helper Snapshot으로 확인한다. 따라서 기존 교차 프로세스·factory leaf 증거는 존재하지만 새 Budget.ObservationSnapshot wrapper 경유는 확인하지 않았으며 이번 선택 regex에도 포함되지 않는다. 인계 표에 각 fixture의 실제 범위를 분명히 기록하거나 cross-factory token429가 별도 요구라면 해당 fixture를 보강해야 한다. 이 결론은 코드 실패나 a169 운영 승인이 아니다.

Go1.25 지정 정확 regex normal/tagged race 두 실행 PASS. `/tmp/a166-consumer-independent-normal.log`, `/tmp/a166-consumer-independent-tag-race.log`에 각각 lo-only/no-route namespace와 원문을 보존. normal accountbudget/runtime/engine 2.305/7.150/3.653초; tagged race 3.388/8.304/5.057초. 14 gzip SHA 목록 일치. 위임 제거 및 nil Budget 제거 변이는 RED; nil context guard 제거는 기존 Snapshot guard로 동등변이여서 살아남았으며 검출 성공으로 세지 않는다.

a169 4개 문서는 draft 보존이며 설계 freeze·구현 PASS가 아니다. 전체 tests/gate, 브로커·실운영·활성화·archive 미실행. gstack은 작성자 수동 범위 검토다.

## 정정

초기 보고의 “기존 token429의 두 factory 교차 증거가 없다”는 판단은 잘못이었다. parent가 child를 실행해 새 factory의 기존 store Snapshot으로 Token event를 검증하는 원문을 확인했다. 새로운 wrapper를 경유하는 token429 두 factory fixture의 범위는 별도이며 작성자 delta 고정 후 검증한다. PriceLimits 기존 fixture는 한 논리 호출에서 401 한 번이다.

## delta 7609ff48d 독립 확인

parent 89c9a5b7e, 생산 source 변경 0. `TestA166ConsumerToken429CrossFactorySnapshot`은 서로 다른 factory Budget 중 한쪽 official token 429 후 다른 쪽 새 Budget.ObservationSnapshot의 Token event/cursor와 GET 0을 확인한다. `TestA166ConsumerPriceLimitsTwo401AdoptAndMint`는 한 논리 PriceLimits에서 cached GET401→adopted cache GET401→token 발급→GET200, GET3/token1, 총 4 first Write·350ms·pending 종류를 확인한다. 기존 generic 5 wire는 두 논리 호출 및 401 한 번이다. 두 fixture 손읽기와 인계표/delta 기록 정합, 이전 P2 검증 공백 해소. 범위 내 P0/P1 없음.

정확 두 이름 Go1.25 선택 normal 1.980초/tagged race 3.031초 PASS, lo-only/no-route 실행 원문 `/tmp/a166-consumer-delta-independent-normal.log`, `/tmp/a166-consumer-delta-independent-tag-race.log`. gzip SHA 목록 일치. 실제 운영·브로커·전체 tests/gate·activation/archive 미실행. 최종 main 결합 검증은 별도다.
