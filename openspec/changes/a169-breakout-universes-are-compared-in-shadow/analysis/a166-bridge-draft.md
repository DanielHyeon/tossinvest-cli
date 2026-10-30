# a169 재개 계약 정합화 초안

2026-10-31 Manager 작성. **아직 design freeze나 구현 승인이 아니다.** 기준 source cf0f0919e 및 읽기 전용 조사 `/tmp/a169-a166-integration-research.md`. 기존 a169 §10 FAIL과 보류 결정을 유지한다. a166의 일부 코드 착지를 리허설·측정 승인으로 바꾸지 않는다.

## 그대로 유지하는 사람 결정

- 공통 group 합계 3건/초 우선, 보호 요청 대기 허용. 별도 앱 키도 같은 예산을 쓴다. 별도 키는 토큰 격리 목적이며 한도 분리 근거가 아니다.
- 별도 credential/token 프로필과 네트워크 0 preflight, top=30, U0/U1/U2 선정·유형 필터·귀속·수익성 결론 금지 계약을 유지한다.
- U2 마감 09:30:55, 미완이면 그 세션 U2 전체 late. 후보 일부로 정상 U2를 만들거나 마감을 늦추지 않는다.
- 같은 날 abort한 세션 자동 재개와 abort 세션의 판 합계 포함을 금지한다. 첫 429 자체를 예방한다는 주장은 하지 않는다.
- 운영자 전환·프로필 확인·리허설·실측은 별도 사람 단계다. 시크릿·엔진 로그 원문·원장 계좌 정보를 개발 fixture로 읽지 않는다.

## 기존 설계의 정정 대상

1. a169 design §7.5의 키별 독립 예산 가정과 4 GET/s 상한을 폐기한다. 이미 게시된 a140 tool factory 배선은 required 입력일 때만 중재한다. a169 모드는 required/observed v2와 검증된 같은 Budget을 필수로 해야 하며 OFF fallback을 허용하지 않는다.
2. U2의 25초는 두 랭킹부터 최종 ready까지다. 후보 일봉 25는 요청 시간이 아니라 봉 수다. 350ms만 고려한 이상화 전체 시작 상한 72회(랭킹 제외 70회)는 성공 보장치가 아니다. 현 250ms pause를 매 요청 그대로 더한다는 조건에서 전체 41회(보충 39회)이며 네트워크·token/401·다른 참여자·처리 시간이 추가된다. 실제 보장 최소는 0이다. 숫자를 목표 처리량으로 쓰지 않는다.
3. 모든 보충 request context를 남은 마감에 결속하고 완료 시각도 재검사한다. 마감에 한번 들어간 요청이 완료됐다는 이유로 late 판을 정상으로 바꾸지 않는다. 취소 뒤 이미 시작한 wire를 없었던 것으로 기록하지 않는다.
4. engine.log 문자열 감시는 숨은429·기록실패·활성 엔진 신원을 증명하지 못한다. 이미 a166이 기록한 그룹 전체 event/gap을 소비하는 명시 계약으로 대체해야 한다. 실제 event에는 engine ID가 없으므로 '엔진429만 정확 필터'라고 쓰지 않는다. 더 보수적으로 그룹 전체 사건을 abort 대상으로 삼는 상세안은 아래 검토 대상이다.

## 최소 소비자 경계 — 독립 검토 전 제안

기존 ObservationStore.Snapshot은 event/gap cursor, overrun, pending Busy/Orphan을 제공한다. required Factory가 검증한 Budget 안의 동일 store로 좁은 read 메서드를 제공하는 확장을 우선한다. manifest를 다시 파싱해 실제 event inode를 스스로 신뢰 입력으로 채우거나 mutable store를 노출하지 않는다. borrowed lifetime, nil/closed/unobserved와 instance/cursor 오류는 fail-closed다.

초기 baseline cursor와 이후 abort 판정은 분리한다. 기존 감시기의 '시작 전 기록은 세지 않음'을 유지할 수 있는 정확한 안정 snapshot 조건, 과거 gap/overrun 처리, 시작 시 Busy인 경우를 먼저 정한다. 이후 cursor 손실·새 event/gap·orphan·읽기 오류·감시 중단/기한 초과는 성공 수집으로 세지 않는다. Busy만으로 원래 프로세스가 죽었다고 판단하지 않는다. 과거 cursor를 새 instance로 조용히 옮기지 않는다.

감시 고루틴의 사전 확인만으로 다음 wire를 차단했다고 증명할 수 없다. 같은 그룹 wire 소유권을 얻은 뒤 실제 송신 전 shared event 경계를 검사하는 좁은 transport 결속이 필요하다. token/401도 포함하며, 다른 참여자가 429를 기록하고 lease를 놓은 뒤 도구의 다음 wire가 나가는 반례를 시험해야 한다. Begin이 orphan을 gap으로 전환하는 경우도 포함한다. 자기 Begin의 pending을 Busy 장애로 오인해 모든 요청을 막는 설계를 피한다. 이 절차의 정확한 순서·신원·context 전파·원래 HTTP 결과 보존 계약은 **미확정**이며 임의 callback framework를 먼저 만들지 않는다.

엔진 poll 로트의 Query collector latch는 같은 논리 호출의 자기 관측 오류만 차단한다. 이 기능을 다른 프로세스의 shared429 감시로 오인하지 않는다. 감시 취소와 동시의 이미 시작한 wire, gate의 첫 Write 선형화 경계를 별도로 설명하고 '취소 시각 뒤 절대0'을 wall-clock 문구로 주장하지 않는다.

## 남은 설계 blocker

- 시작 baseline의 안정성과 과거 gap/overrun·pending 처분, fail-closed prewire 검사와 shared event 원자성.
- 동일 budget identity와 실제 활성 엔진·검증된 프로필의 결속은 별개다. 기존 G2 파일 존재만으로 활성 엔진 결속을 증명하지 않는다. 운영 전환 영수증과 런타임 신호 중 필요한 최소 근거를 source로 정하고, 자기 주장 파일만으로 authority를 만들지 않는다.
- abort 기록의 실패·동시 프로세스·재시작에도 같은 날 수집이 다시 열리지 않는 저장 계약.
- 기존 a140 기본 모드 바이트/판 digest 보존과 a169 전용 오류 처리 분리. a140의 기존 retrying 429 재시도를 그대로 쓰지 않는다.

## 수용 증거 계획

정확한 local TLS official.Client/401/token 경로에서 shared group429→후속 wire0, own Finish fault→후속 wire0, orphan/gap/overrun/instance mismatch, cancel-vs-firstWrite, 감시기 중단, same-day abort 재시작을 분리 검증한다. profile guard는 fake private temp 파일만 사용하고 secret 값 출력 금지를 확인한다. fake clock으로 랭킹 지연·보충 deadline·선정시각 양측·late 전체제외, 순수 U0/U1/U2 계산과 report 제외를 검증한다. 후보200개를 실제 브로커에 조회하지 않는다.

search-first: existing a140 factory/pacedSource, accountbudget store/collector/transport, official token 및 Go context 재사용. 읽기 전용 연구와 Manager의 Snapshot/transport/profile/tasks 원문 대조를 사용했다. 외부 package 조사 없이 충분한 기존 부분 해법을 확인했으며 새 의존성 도입은 없다. grill-me/grilling: 확정 정책은 재질문하지 않고 위 구현 계약 반례를 source·독립 검토로 먼저 해소한다. impeccable N/A — CLI/backend·문서만 영향. AST·시험·CodeGraph 최신 검증은 이 초안에서 수행하지 않았다.

## 2026-10-31 기준점 결정 반영

사용자가 현재 안정 시점부터 시작하고 과거 gap 제외를 선택했다(issues §7). 위 baseline 과거 gap 선택 미결 부분은 이 결정으로 대체한다. 현재 pending·dirty·손상·읽기 오류는 계속 거절한다. lease 아래 실제 전송 검사는 `prewire-fence-plan.md`의 독립 검토 대상이며 기존 Snapshot 자체를 원자 전송 허가로 취급하지 않는다.
