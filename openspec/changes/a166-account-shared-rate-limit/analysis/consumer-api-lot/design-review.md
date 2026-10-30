# a166 소비자 API 인계 초안 독립 설계 검토

대상 `analysis/consumer-api-handoff-plan.md` (worktree `/home/daniel/tossos-work/a166-runtime`), source 기준 local origin/main 6e0968222993756a41065044996f674b35e20a27. 손읽기; AST/시험/운영 미실행. 결론: 범위를 유지하면 좁은 후속 API와 인계 fixture로 타당. 단 a166 design:103의 완료 의미는 명세에서 먼저 정렬해야 한다.

## P1 — 정본 §103 인계 경계
현재 `design.md:103`은 마지막 통합 로트에서 'a087 후속 활성화 API, a169 공유 이벤트 독자를 연결한다'고 명시한다. 초안 :9–11·19의 Snapshot wrapper는 공유 이벤트 **읽기 기반 API**를 제공하지만 실제 a169 독자/abort/prewire fence를 연결하지 않는다. 또한 실제 a087 board의 활성 token/epoch/보호 관측 동시성은 검증하지 않는다. 따라서 이를 문자 그대로 §103 전체가 완료됐다고 판정하면 범위 불일치다. Manager가 정본·tasks 3.4에 a166의 소비자용 snapshot 및 실제 guard/PriceLimits 인계 증거와 a087 5.5/a169의 실제 독자·fence·activation 후속 작업을 명시적으로 분리한 뒤에만 해당 부분을 평가해야 한다. a087 production nil 및 a169 측정 보류 유지. 이 정렬은 기존 safety 정책을 완화하지 않으며 소비자 동작을 생략한 채 PASS를 주장하지 않는다.

## Snapshot wrapper와 생명주기
`runtime_linux.go:95-102`은 신뢰 manifest event_lock pin으로 OpenObserved한 Budget을 VerifyIdentity 뒤 반환하고 `:121-139` Factory가 Close를 소유한다. `accountbudget/observation_linux.go:39-52,616-673` 기존 Snapshot은 Budget pin/닫힘 확인 후 store lock을 쓰고 cursor scope·instance 검증 및 generation-race를 반환한다. wrapper는 반드시 nil ctx를 먼저 거절해야 한다(기존 호출이 모두 nil을 안전하게 처리한다고 추정 금지). 관측 store를 여는 기존 저수준 API가 따로 있어도 wrapper는 `b.observation`의 바로 그 객체에만 위임하고 임의 FileID나 root를 수용하지 않아야 한다. pin과 Close의 경합에서 성공은 pin이 먼저 확보된 경우에만, Close 실패/closed 오류를 숨기지 않는다는 시험을 유지한다.

`observationView`는 read한 state에서 Last/Pending과 slice를 구성하며 events/gaps는 새 view에 append한다(`observation_linux.go:675-695`); 읽은 state는 파일 복호화된 로컬 값이므로 공개 view가 파일 state를 직접 참조하지 않는다는 손읽기 근거다. 포인터와 슬라이스를 caller가 바꾼 뒤 같은 store의 새 Snapshot 결과가 불변임을 fixture로 고정한다. cursor 역행은 범위 오류가 아니라 기존 API가 과거 이력을 다시 읽는 데 허용할 수 있으므로 초안 :9의 '역행'은 기존 Snapshot의 미래 cursor·scope 불일치 오류와 구분해야 한다. 무조건 역행 거절이라는 새 계약을 만들지 말 것. overrun과 pending Busy/Orphan은 반환 상태이지 wrapper 자체가 자동 abort하는 것이 아니다.

## 실제 official PriceLimits fixture
`internal/official/a087_price_limits_raw.go:20-28`의 실제 client 호출과 required factory client/Budget pointer 결속, Auxiliary context를 이용하는 접근은 기존 guard 재사용 면에서 적절하다. Busy 경합은 다른 process 또는 같은 Budget의 실제 lease로 만들고 token cache hit/miss를 구분해야 한다. 각 401 attempt·token과 GET의 wire 관측을 분리하고 exact local dial·namespace에서만 실행한다. 보조가 선택되기 전 Critical/active wire에 막혀 거절됐을 때만 새 wire0을 단언하고, 이미 Aux가 선택된 후에는 보호 대기를 검증한다. token이 이미 전송된 뒤 후속 GET이 거절되는 경우 전체 논리 호출을 token wire0이라고 부르면 안 된다. 서로 다른 Factory Budget 포인터는 다르므로 UsesAccountBudget은 각 client와 자기 factory의 포인터로 확인하고, 두 factory는 공통 manifest inode/instance 및 cross-process event로 공유를 입증한다.

Finish 실패가 응답을 덮지 않는 계약은 collector/local error와 함께 검증해야 한다. event 기록 실패가 어느 단계인지에 따라 event 또는 pending/gap이 남을 수 있으므로 무조건 429 event가 남는다고 가정하지 않는다. `429 header 없는`은 HTTP 응답 헤더 전체가 없는 것인지 Retry-After만 없는 것인지 명세를 명확히 한다. 실제 시장 price timestamp·보호 주문의 지연·a087 5.5·a169 중단 검증으로 확대하지 않는다.

## 판정
설계의 기능 축소는 별도 사용자 정책 결정을 요구하지 않는 구현 경계 분리로 가능하지만, §103 정본문구/체크 항목을 먼저 명확히 고쳐야 한다. 현재 초안은 proposal이며 구현·테스트 승인이나 완결 판정이 아니다. 전체 tests/gate, 운영·브로커·live는 미실행이다.

## 재검토 — 수정 초안

대상 `/home/daniel/tossos-work/a166-consumer-api`의 미커밋 design.md:103, tasks.md:36–37 및 analysis/consumer-api-handoff-plan.md:3–19 손읽기. 이전 P1 해소: 정본과 task가 a166의 Snapshot/실제 official fixture 인계와 a087 5.5 및 a169 독자·fence·abort·엔진 신원 후속 책임을 분리하고 소비자 미완을 명시한다. 과거 cursor 재조회/미래 거절, factory별 포인터, token 선행 wire 경계도 반영됐다. 새 P1 없음. source 구현·독립 구현 검증·시험·운영은 아직 없으며 정본문구 검토를 소비자 기능 완료로 확대하지 않는다.
