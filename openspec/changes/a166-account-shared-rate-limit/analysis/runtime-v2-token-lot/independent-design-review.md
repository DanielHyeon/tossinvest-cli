# 독립 설계 검토 — runtime v2 / token

판정: 조건부 PASS. P0/P1/P2 반례 없음, P3 1건. 설계 작성자는 Manager이며 검토자는 선행 observed 구현 작성자여서 구현 독립성은 별도 한계다. 이번 검토는 문서·현재 main 소스 읽기만 했고 코드와 시험을 실행하지 않았다.

- P3 `analysis/runtime-v2-token-plan.md:3`: 기준이 main 42009와 미배선 후보 aee3로 남아 있다. observed는 main 2938622a8bb8ef43e0f11e410e7affaf784bcab1에 게시됐다. 최소 수정: 기준 main을 2938622로 바꾸고 aee3는 통합된 선행 후보로 명시한다. 이후 착수 시 최신 main/source SHA를 다시 고정한다. 사용자가 지정한 판정 범위와 기능에는 영향 없다.
- v2 신뢰 경계: 9–13행은 event.lock을 외부 manifest의 device/inode에 결속하고 v1·누락·교체·dirty를 거절한다. 현재 OpenObserved는 OpenObservation 실패 시 직접 Budget.Close를 합치며, 성공 후 VerifyIdentity 실패는 runtime이 Budget.Close를 합치는 구분이 맞다. 자동 provisioning·event pin 자기 승인 금지와 client 공개 전 root/queue/wire 검증은 유지된다. 동일 디바이스/아이노드 관계가 신뢰된 배포에 의해 모든 참가자에 전달된다는 전제는 이미 선행 설계의 범위다.
- token: 17–19행의 scope 복제는 현재 observedScope에서 collector를 그대로 유지하고 kind만 Token, key zero로 덮는 최소 API로 구현 가능하다. context.WithValue 파생은 부모 취소/deadline/priority를 계승한다. 실제 token.exchange는 캐시 미스/401 refresh에서만 호출되며 현재 *accountbudget.Transport 검사 가지를 이미 갖는다. 구현 시 kind 변경이 token request 한정인지, cache hit과 OFF 기존 분기가 변하지 않는지 시험해야 한다.
- 23–29행의 v2 parsing·fault cleanup·same Budget/store·token/GET/401·local TLS·pin·clone/replay/probe 및 두 process 증거는 요청 범위에 맞다. 단 fixture와 값 흐름을 실제로 연결하기 전에는 완료 근거가 아니며 필요한 새 test 결과는 없다.
- 33행은 canonical design.md:73과 일치한다. 같은 RequiredQuery의 보호 및 일반 가격 read가 공유 hint로 대기할 수 있고 token·주문 본체·flatten·별도 walk에는 해당 hint를 적용하지 않는다. Retrier 계좌/selected-seq/client 결속, freshness·재시도·AuthFatal은 후속 범위에 남는다.

ai-memory CLI project-scoped 검색은 결과 없음. 외부 자료 검색·신규 의존성·실행 시험·운영/배포는 수행하지 않았다. 현재는 설계 검토이며 production 활성화나 전체 a166 완료 승인으로 확대하지 않는다.
