# a087·a169에 넘길 공통 예산 API와 통합 증거

독립 설계 검토의 정본 범위 P1과 cursor·factory identity 명료화를 아래에 반영했다. Manager는 이 좁은 로트를 승인하며 독립 구현 검증은 별도다. 기준 main 6e0968222. a166 design/tasks의 마지막 통합 로트는 후속 소비자가 검증된 같은 Budget·event를 읽을 수 있도록 하고 실제 official.Client 수명 시험 SHA를 제공한다. a087 5.5나 a169 도구 구현을 이 로트에서 켜지 않는다.

## 제공할 최소 경계

기존 Factory.Budget은 borrowed Budget, Factory.NewClient는 그 객체의 guarded client다. official.Client.UsesAccountBudget은 정확 포인터 결속을 확인하며 계좌 권한·readiness·활성화 허가가 아니다. 기존 Priority context의 Auxiliary와 실제 token/401 admission을 재사용한다. no-op Admit 구현이나 bool만으로 a087 production nil을 해제하지 않는다.

Budget.ObservationSnapshot(ctx, cursor) 하나를 추가해 생성 시 검증되어 private로 결속된 ObservationStore.Snapshot에 위임한다. mutable store/manifest/신뢰 입력 setter나 원시 state file 경로를 노출하지 않는다. nil Budget·nil ctx·closed·unobserved·unsupported는 오류; source의 기존 pin/lifetime을 그대로 쓴다. cursor scope/instance 불일치·현재보다 미래 cursor 오류, 과거 cursor 재조회 허용, ring overrun·pending 상태 반환과 generation-race 오류 의미를 바꾸지 않는다. wrapper가 이 상태를 자동 abort로 변환하지 않는다. 반환 slice/pointer는 caller가 바꾸더라도 store state에 영향을 주지 않는 기존 view 사본이어야 한다. 소비자는 객체를 Close하지 않고 factory 소유자가 닫는다.

이 API는 snapshot이다. 다음 wire와 원자적인 '이벤트 없음' 허가가 아니고 엔진 생존·same-day abort·감시기 정상·프로필 동등성을 증명하지 않는다. a169의 prewire 중단 경쟁은 해당 change의 `analysis/a166-bridge-draft.md`에 별도 blocker로 남긴다. a087의 이벤트 기반 갱신 전용 정지와 epoch/만료 토큰도 D16/5.5에서 결합한다. 공통 API 증거를 이 소비자 동작의 PASS로 확대하지 않는다.

## 실제 소비자 경로 인계 fixture

required v2 factory 두 개가 같은 root/pins를 열고 각자의 official client를 생성한다. 서로 다른 factory의 Budget 포인터가 같다고 가정하지 않고 각 client를 자기 factory Budget에 정확 결속하며 공통 root/pins/instance와 이벤트 공유로 두 factory의 공유성을 검증한다. 하나의 local TLS 요청에서 기록한 token/Other429를 다른 factory의 borrowed Budget.ObservationSnapshot이 읽고 cursor가 전진하는지 확인한다. bootstrap0, 성공 noevent, Retry-After 헤더가 없는429 및 Finish오류를 분리한다. Finish 실패 단계에 따라 event 또는 pending/gap이 남을 수 있으므로 무조건 event 기록 성공을 주장하지 않는다. 기존 store ring/boot leaf 시험 전체를 중복하지 않고 wrapper의 same-store·closed/unobserved·view 사본을 확인한다. 오류를 감지했다고 원 HTTP 응답을 NOT_SENT로 바꾸지 않는다.

a087 인계는 실제 official.Client.ProtectivePriceLimits를 Auxiliary ctx로 호출하는 격리 fixture다. 검증한 같은 Budget, token 및 각401 재전송별 pending/attempt, 로컬 plaintext 첫 Write 간격, 응답 EOF/Close/cancel에 따른 wire 소유권을 기존 fixture와 결합한다. 보호 Critical read와 경쟁해 최초 Aux 허가가 거부되면 새 token/GET Write0; token이 이미 나간 뒤 후속 GET 허가가 거부된 경우 token까지 wire0이라고 하지 않는다. token cache hit/miss를 구별한다. 이미 선택된 Aux는 선점하지 않으며 보호는 그 수명을 기다린다. 과거 '관측 지연0'이나 finite wall-time 상한을 재도입하지 않는다. fixture가 가격 결과를 받은 것은 실측 timestamp 의미·실거래·첫 보호 주문 정책 검증이 아니다.

a087 production nil을 검사하는 기존 선택 구조 시험을 유지하고, a169 production 측정은 실행하지 않는다. a087의 실제 board/new token/5분정지 통합은 5.5, a169의 event 소비/abort는 해당 후속 코드의 책임임을 handoff 표에서 분리한다. 여기서 제공하는 API/시험명/착지 SHA를 그 단계가 반드시 인용하도록 한다.

## 검증과 제한

신규 helper는 기존 Snapshot의 source를 재사용하므로 범용 event bus나 dependency를 추가하지 않는다. 기존 함수 변경이 필요하면 사전 AST/FLM/BTM을 먼저 고정한다. missing-method RED→GREEN, same-store 우회·nil/closed guard 제거의 유효 변이, 정확 선택 normal/tagged race, scoped gstack·독립 리뷰를 수행한다. fixture는 fake credentials·명시 exact local DialContext·loopback-only/no-route만 쓴다. 실제 broker·운영 root·전체 test/gate는 금지다.

search-first: source상 existing ObservationStore.Snapshot·Factory.Budget/NewClient·UsesAccountBudget·Auxiliary 및 official raw-price reader로 요구 대부분을 충족한다. 좁은 wrapper와 인계 fixture를 확장하고 새 framework/dependency는 만들지 않는다. grill-me/grilling: 승인 정책을 바꾸지 않고 snapshot과 송신 허가의 한계를 구별했다. impeccable N/A — backend/API·문서만 영향. 아직 이 제안의 AST·시험·독립 검증은 미실행이다.
