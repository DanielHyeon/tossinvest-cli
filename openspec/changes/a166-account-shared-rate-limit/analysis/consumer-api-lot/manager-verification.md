# 소비자 API 인계 로트의 Manager 검증

2026-10-31, 후보 89c9a5b7e 및 test-only delta7609ff48d, base6e0968222. **개발 로트 승인**. a166 전체 종결·a087 활성화·a169 설계 freeze/측정 승인이 아니다.

Manager는 새 Linux Snapshot wrapper와 unsupported stub 전부, pin/사본/cursor fixture, 실제 PriceLimits의 token/401 delta source와 handoff 표를 대조했다. wrapper는 생성 시 검증되어 immutable로 결속된 store만 빌려 읽고 nil/closed/unobserved를 거절한다. 기존 store pin·state 검증·generation/cursor 규칙을 바꾸거나 mutable store/신뢰입력을 노출하지 않는다. 같은 root의 별도 factory는 다른 Budget 포인터이며 각 client가 자기 Budget에 결속되고 cross-factory event를 관측한다.

독립 Terra는 89c의 정확 선택 목록을 Go1.25 normal/tagged race로 accountbudget·runtime·engine에서 재실행했고, delta 두 시험만 다시 normal 1.980초/tagged race 3.031초로 확인했다. fake credentials·exact local DialContext·loopback-only/no-route만 사용했다. gzip SHA 일치와 위임/정당한 nil Budget guard 제거 변이 RED를 확인했다. nil ctx guard 단독 제거는 하위 검사로 동일 오류인 동등변이이며 유효 검출로 세지 않는다. Darwin compile-only는 운영 지원 증거가 아니다.

증거 해석을 명료화했다. cf0의 기존 token429는 child process가 새 factory로 기존 store를 읽는 증거가 이미 있었고, 새 delta는 공개 wrapper 경유를 추가했다. 선행 generic 5 wire는 Accounts 두 논리 호출·401 한 번이다. 단일 PriceLimits의 401 두 번은 delta에서 cache adoption 후 mint를 거쳐 GET3+token1=4 wire, pending sequence/종류와 ≥350ms first plaintext Write를 따로 확인했다. 이 둘을 합쳐 더 강한 과거 증거라고 주장하지 않는다.

Aux 최초 허가 거절 token/GET0과 token이 이미 나간 뒤 GET0은 분리된다. 기존 응답 EOF/Close/cancel까지의 소유권과 postcommit Finish 오류에서 원 HTTP 의미 보존은 지정 fixture 범위다. 전체 board/epoch/429 정지, 실제 시장 timestamp/실거래, 다음 wire와 원자적인 snapshot 허가는 검증하지 않았다.

정본 design/tasks는 a166 API·actual client 인계와 a087 D16/5.5·a169의 실제 소비자 연결을 분리했다. 독립 설계 P1 해소 후 구현했으며 소비자 후속 작업을 삭제하지 않았다. 함께 보존한 a169 draft/research/review는 미해결 fence·신원·영속 abort를 명시하며 freeze 전이다.

작성자 gstack은 수동 API/오류/수명 범위 checklist, 독립 리뷰는 별도다. 새 의존성·기존 함수 본문 변경 없음, impeccable N/A — backend/documentation-only. 전체 tests/gate·실제 broker/계좌·운영 전환·최신 CodeGraph 검증은 미실행. 최신 main enginepoll과 통합한 뒤 정확 영향 선택과 scoped a166/a169 strict·PM·diff를 확인해 Terra가 정상 게시한다. 최종 wire census와 사람4.2는 남아 archive하지 않는다.

게시 후 갱신: consumer는 `d9827f316`로 enginepoll과 합성한 정확 선택 normal/tagged race·scoped strict/PM/diff를 통과해 게시됐다. 구조 census도 `7d68c9ef6`로 게시됐다. 당시 남았던 개발 단계가 해소됐음을 인계표에 반영하며 운영·활성화·사람 확인 경계는 그대로다. 새 코드나 시험 재실행은 없다.
