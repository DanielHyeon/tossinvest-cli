# 수동 review — runtime v2 / token

기준 e5c9d0256; 사전 AST 4개와 분기·시험 지도는 pre-edit.md와 function-logic에 기록했다. store version 1은 유지하고 외부 manifest만 version 2로 이행했다. 기존 OpenObserved/VerifyIdentity·scope·stdlib 재사용; 신규 의존성 없음. impeccable N/A — backend-only.

수동 gstack review 체크리스트 중 경계·정리·동시성·context·순서·기존 OFF 및 오류 의미를 source로 검토했다. 자동 specialist·업그레이드·telemetry는 실행하지 않았다. 독립 검증을 대체하지 않는다. v1 거절 RED→GREEN, version/event pin/duplicate 검사와 token scope 제거 4개 변이 RED 원문을 보존했다.

Go 1.25 정확 선택 normal/tagged race를 selected-tests.sh의 11개 package·이름에 한정했다. 모든 네트워크 fixture는 user/net namespace lo-only/no-route, fake credentials 및 exact DialContext local TLS다. 기존 두 process 첫 Write 검증도 v2 manifest·주소 pin으로 실행했다. 실제 official token 429·401·cached GET의 pending kind, token의 poll collector/hint, body 실패·저장 실패, 별도 child event 조회, 복수 client, a112 clone·replay·probe의 같은 store를 확인했다. a112는 test-only unsafe로 private 기존 transport의 RootCAs·DialContext만 설정, base/authority/pointer 미변경이다. fake hostname 인증서는 운영 authority 증명이 아니다.

fd 정리는 invalid event/root identity의 반복 Open 후 /proc/self/fd 대상으로 확인했다. 테스트 목록의 바깥 기능·실제 broker·실원장·운영 설정·전원 장애·전체 tests/gate 미실행. Retrier poll 소비와 전체 참여자 value-flow 보증은 후속이다. 새 main 병합 뒤 추가된 fixture는 별도로 재확인해야 한다.

최신 main afc8d3486 통합 시 runtime/official 제품 소스는 변경되지 않았다. 추가된 a112-mb-us-source 두 manifest 및 a087 probe 한 manifest fixture를 v2 임시 event provision으로 이행했다. 정확 7개 시험(TestA166MBCachedOff, TestA166MBCachedRequiredMissing, TestA166MBCachedBusyNoWire, TestA166MBCachedRequiredOneGET, TestA166ProbeRequiredUsesNormalPriority, TestA166ProbeRequiredMissingAfterLocalGuards, TestA166ProbeFactoryLocalTokenAndPost)은 Go 1.25 normal 각각 패키지 0.857초·0.910초, tagged race 1.905초·1.920초 PASS. 원문 a166-v2-tools-last-{normal,tag-race}.log.raw.gz. 다른 패키지 대상은 통합 전 source 동일하고 위 선택 검증이 유효하다.
