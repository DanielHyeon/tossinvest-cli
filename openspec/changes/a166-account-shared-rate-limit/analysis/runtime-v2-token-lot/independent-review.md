# runtime v2/token 독립 검토

대상 288358373 (부모 21b9bc85b), 기준 e5c9d0256 및 최신 tools afc8d3486. 제한된 runtime v2/token 구현에 P0/P1/P2 미발견. 전체 a166·Retrier poll·운영 활성화 완료 판정이 아니다.

소스 손읽기: required manifest v2와 event.lock 외부 pin을 OpenObserved에 넘기고 root/queue/wire VerifyIdentity 실패 시 Close 오류를 결합한다. OFF는 기존 official.New 경로다. token의 observed scope 전환은 collector·priority·deadline을 보존하고 Token/zero key를 사용한다. nil ctx P2는 token.go의 helper 호출에 ctx != nil을 추가하여 기존 http.NewRequestWithContext 오류 경로로 보존한다. fixture의 a112 unsafe reflection은 local TLS 검증용이며 실제 운영 authority 증명이 아니다. 현재 구현의 생명주기·branch는 수동 검토이며 pre AST 네 개는 e5 source SHA와 일치한다.

Go1.25 독립 선택 normal/tagged race: 승인 selected-tests.sh의 11개 package·정확 이름, 별도의 추가 tools 7개 시험. 원문 /tmp/a166-runtime-v2-independent-{normal,tag-race}.log 및 /tmp/a166-runtime-v2-independent-tools-{normal,tag-race}.log. 각 실행은 별도 user/network namespace, lo-only/no-route이며 local TLS fixture는 주소를 고정한다. 실제 official token429·401/cache, a112 clone, budgeted replay/probe와 두 process factory 해당 테스트 실행을 원문에서 확인했다. 외부 브로커 및 실제 주문은 실행하지 않았다.

nil P2 delta에서 TestA166TokenNilContextMatchesOff와 TestA166AccountsNilContextMatchesOff 독립 normal/tagged race PASS (/tmp/a166-v2-independent-nil-{normal,tag-race}.log). 기존 token429·401/cache·Finish fault 세 선택 재실행 normal/tagged race PASS (/tmp/a166-v2-independent-delta-{normal,tag-race}.log). writer의 nil guard 제거 시 Accounts panic RED를 포함해 version·event pin·duplicate·Token scope 변이 원문 확인; gzip 원문 14개 SHA 일치.

실제 filesystem 전원 장애, unsupported 플랫폼 실행, 실 broker·운영 cutover, 전체 tests/gate 및 Retrier poll 소비는 미검증이다. 별도 engine poll 설계/구현과 모든 참여자 value-flow 검증이 후속이다. 테스트 중 source 편집 없음.
