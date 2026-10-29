> design 1판(2026-10-29) delta. 기존 요구 「Retry Matrix 산출물」을 MODIFIED 로 다시 쓰고(429 문장), 계좌 공유 admission 을 새 요구로 더한다.
> 수치(350ms·15s·30s)와 critical 예외는 design §8 의 사람 결정 H2·H4 를 받기 전 권장값이다.

## MODIFIED Requirements

### Requirement: Retry Matrix 산출물
재시도 정책은 endpoint×method×오류 클래스 표로 스펙 산출물화해야 한다(SHALL). 최소 규정: **주문 mutation은 어떤 오류에도 자동 재시도 금지 — 단, 멱등 재생은 재시도가 아니라 정체 회수이고, 취소의 `409 already-processing` 은 「취소의 일시적 거절 처리」를 따른다**(같은 키·같은 본문의 재요청은 유효 창 안에서 새 주문을 만들 수 없다 — openapi; IN_DOUBT 해소 요구의 조건·상한·비활성 규칙을 따른다). 이 구분은 재시도 정책 구현부의 문서화된 근거에도 반영한다(SHALL). 조회는 재시도 예산(횟수·총 시간)과 bounded jitter, 401/403은 즉시 신규 진입 차단 + 알림, 필수 조회의 staleness 임계 초과 시 신규 진입 차단. 표의 수치는 구현 시 확정하고 표 없이 구현하지 않는다(SHALL NOT).

엔진 폴링 조회가 429 또는 계좌 공유 admission 거절을 받으면 같은 조회 안에서 다시 보내서는 안 된다(SHALL NOT). 그 조회는 실패로 끝나고 staleness 규칙이 진입을 막는다(SHALL). `Retry-After` 는 30초 이하일 때만 그 길이로 존중하고, 30초를 넘으면 30초로 깎으며, 없으면 기본 길이를 쓴다 — 존중은 대기가 아니라 계좌 공유 패널티로 한다(SHALL, 「계좌 공유 요청 admission」). soak·verify·IN_DOUBT 스캔 같은 비지연 walk 의 429 백오프(추가 2회, 15s → 30s)는 그대로다.

공식 클라이언트가 401 에 토큰을 갱신해 같은 요청을 다시 보내는 기존 동작(최대 2회, 같은 본문 바이트)은 브로커가 그 요청을 받지 않은 경우로 본다(SHALL — 근거는 시험으로 고정한다; 실측 근거는 없다). 이 재전송은 클라이언트 안의 일이라 attempt 기록 대상이 아니다. 그 재전송 요청도 각각 계좌 공유 admission 을 지난다(SHALL).

#### Scenario: 재생과 재시도의 구분
- **WHEN** 멱등키 없이 전송된 mutation이 모호한 오류로 끝나면
- **THEN** 어떤 자동 재전송도 발생하지 않는다 (재생은 영속된 키·본문이 있는 attempt의 해소 절차에서만; 취소의 `409 already-processing` 은 모호한 오류가 아니라 「취소의 일시적 거절 처리」가 다루는 확정 거절이다)

#### Scenario: 필수 조회 장기 실패
- **WHEN** 잔고 조회가 재시도 예산을 소진하고 staleness 임계를 초과하면
- **THEN** 신규 진입이 차단되고 조회 복구 후 자동 해제된다

#### Scenario: 401 재전송은 같은 본문이다
- **WHEN** 주문 접수 POST 가 401 을 받고 토큰이 갱신되면
- **THEN** 다시 보내는 요청의 본문은 첫 요청과 바이트 동일하고, 401 이 아닌 오류에는 다시 보내지 않는다

#### Scenario: 엔진 폴링 GET 의 429 — Retry-After 없음
- **WHEN** 엔진의 미체결 조회가 `Retry-After` 없는 429 를 받으면
- **THEN** 그 조회 안에서 다시 보내지 않고, 조회 실패가 staleness 임계를 넘으면 신규 진입이 차단되며, 그 429 는 계좌 공유 패널티로 기록된다

## ADDED Requirements

### Requirement: 계좌 공유 요청 admission
같은 계좌로 공식 Open API 를 치는 모든 TossOS 프로세스(엔진·콘솔·httpapi·soak·verify·CLI·측정 도구)는 wire 요청 하나마다 — `official.Client` 의 401 재전송과 토큰 교환을 포함해 — 하나의 계좌 공유 예산의 admission 을 받은 뒤에만 보내야 한다(SHALL). admission 은 HTTP 전송 직전에 걸리고, 거절이면 그 요청은 보내지 않으며 그 오류는 브로커 429 와 구별된다(SHALL — 주문 mutation 의 admission 거절은 미전송으로 분류된다). 예산의 키는 프로세스·config 디렉터리·앱 키가 아니라 계좌여야 한다(SHALL); 한 예산 디렉터리는 한 계좌를 나타낸다. 손절·비상 청산·취소·주문 mutation·보호 판단 조회·토큰 교환(critical)을 제외한 요청의 합은 어떤 1초 창에서도 3건을 넘어서는 안 된다(SHALL NOT). critical 요청은 admission 때문에 기다려서는 안 된다(SHALL NOT — 안전 불변식 3). 모든 wire 429 는 `Retry-After` 유무와 호출자의 재시도 결과와 무관하게 비밀 없는 정규 이벤트로 남고 프로세스 재시작을 견디는 공유 패널티가 된다(SHALL). 판정 시각은 주입 clock 하나로 정하고 시계 역행이 대기를 30초 넘게 만들어서는 안 된다(SHALL NOT). 예산을 쓸 수 없으면 엔진은 보호를 계속하되 신규 진입을 막는다(SHALL). admission 이 없는 구성은 upstream 과 같은 요청을 보낸다(SHALL).

#### Scenario: 두 프로세스의 합산
- **WHEN** 엔진과 soak 이 같은 계좌로 동시에 normal 조회를 내면
- **THEN** 두 프로세스의 normal 요청 합은 어떤 1초 창에서도 3건을 넘지 않는다

#### Scenario: 401 재전송도 관문을 지난다
- **WHEN** 한 요청이 401 → 401 → 200 으로 끝나면
- **THEN** admission 은 3번 일어나고, 둘째 admission 이 거절되면 둘째·셋째 HTTP 전송은 일어나지 않는다

#### Scenario: 헤더 없는 429 뒤 재시작
- **WHEN** 한 프로세스가 `Retry-After` 없는 429 를 받고 다른 프로세스가 곧바로 재시작해 normal 조회를 내면
- **THEN** 그 조회는 패널티 안이라 보내지 않고 거절되며, critical 요청은 기다리지 않고 보내진다

#### Scenario: 보호 청산은 굶지 않는다
- **WHEN** normal 조회가 예산을 다 쓰고 대기 중일 때 엔진이 보호 매도를 내면
- **THEN** 보호 매도는 admission 대기 없이 보내진다

#### Scenario: 닫은 뒤 새 전송 없음
- **WHEN** 측정 도구가 자기 admission 을 닫으면
- **THEN** 닫기가 반환된 뒤 새로 admission 을 얻는 요청은 없고, 닫기는 그 순간 이미 admission 을 받은 요청 수를 돌려준다
