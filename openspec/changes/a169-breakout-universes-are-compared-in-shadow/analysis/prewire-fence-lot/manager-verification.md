# Manager 검증 — 선택적 관측 기준점과 전송 전 차단

2026-10-31. 고정 source/test/census `2764aa7016182eb6d62ced48871618900d09abdc`, 기준 main `b0f7f1d93`. 한정 구현 로트의 병합·푸시를 승인한다. a169 전체 완료나 archive, 운영 연결·측정·배포 승인은 아니다.

사용자 결정은 **현재 안정된 시점부터 시작하고 과거 gap은 제외**다. 기존 Snapshot에서 유효한 pending 없는 원장 기준점을 잡고 이전 event/gap/ring overrun을 제외한다. 기록 자체를 삭제하거나 건강함으로 바꾸지 않는다. 현재 pending·dirty/읽기·identity 오류는 거절하고, 기준점 뒤 새 event/gap·역행은 다음 전송 전에 거절한다. 정상 다른 요청의 generation 증가는 허용한다. no-pending은 모든 응답 body가 닫혔다는 뜻이 아니며 기존 wire 소유 수명은 유지된다.

Manager는 새 Linux/unsupported API, RoundTrip diff와 원본, release/body wrapper, 실제 official token/adopt/mint fixture 및 채널 기반 waiter fixture를 읽었다. Acquire 전 검사만으로 끝내지 않고 Acquire 후 같은 fence mutex 아래 다시 검사한다. wire를 보유한 동안 검사→Begin을 진행하며 Release 반환 오류를 기록하기 전 다음 요청이 fence 검사를 통과하지 못하게 직렬화한다. 이미 받은 HTTP 결과와 미전송 ErrNotSent 분류를 분리한다. fence 부재 경로의 기존 body/cleanup 순서를 손읽기로 대조했으며 HTTP/JSON 전체 바이트 동등성을 주장하지 않는다.

독립 reviewer가 고정 SHA의 원본 AST41/current68와 source SHA를 대조했다. Manager는 분기별 명명 시험과 미검증 표시가 있는 current 교차표를 읽었다. AST 구조·shape pin은 실행 coverage나 모든 간접 호출의 증명이 아니다. production 직접 fence 생성 호출이 없다는 독립 source 확인을 수락한다.

작성자의 정확 선택 normal 및 tagged race 원문, post-Acquire sticky·release 순서·event·gap 검증 우회 4개 변이의 행동 실패 원문을 대조했다. 독립 Terra의 별도 축소 선택 normal은 accountbudget/runtime 11.571/5.022초, tagged race는 12.589/8.508초 PASS다. 실행은 lo-only/no-route·가짜 자격·임시 manifest·local TLS에 한정했다. Manager는 시험을 직접 실행하지 않았다. 무태그 race나 전체 package/test/gate를 실행했다고 확대하지 않는다.

독립 최종 및 보조 source 리뷰에서 해당 범위 새 P0/P1 없음. gstack은 수동 범위 checklist로 잠금·오류·응답·취소·OFF·fixture를 검토했다. search-first는 현재 관측·lease·context 코드와 Go stdlib 재사용, 새 의존성 없음. grill-me/grilling의 미결 gap 정책은 사용자 결정으로 닫혔다. impeccable: N/A — backend/documentation-only; UI 변경 없음. 최신 CodeGraph index 미검증을 유지한다.

**절차 위반은 남는다.** 최초 source 편집 전 원본 AST와 요약만 있었고 완전 FLM/BTM은 사후 작성됐다. 중지·원본41분기 교차표 독립검토·`b80997c1c` 선커밋 후 재개했으며 current68 지도와 예외를 별도 보존한다. 이번 한정 제품 검증 통과를 사전 절차 준수로 소급하지 않는다.

Terra 종결자는 독립 리뷰/시험 원문을 이 change에 보존하고 문서·진행표를 최신화한다. pinned OpenSpec a166/a169 strict·선택 PM 검사·diff 검사를 수행하고, 검증된 source가 유지된 경우에만 정상 fast-forward 병합·푸시한다. source가 바뀌거나 충돌로 계약이 달라지면 새 고정 후보 검증이 필요하다. 무관한 main a128 변경과 미추적 상태는 보존한다.

남은 범위: 실제 a169 caller/session 하나의 fence 연결, 당일 abort 영속성과 재발급 방지, 활성 엔진/공유 root 결속, runtime/profile/ledger/report 및 운영 리허설. a087 활성화·a166 운영 검증도 별도다. 전체 tests/gate·실브로커·계좌·운영·배포·모든 비협조 writer 검증은 하지 않았다. 미구현 작업을 체크하거나 change를 archive하지 않는다.
