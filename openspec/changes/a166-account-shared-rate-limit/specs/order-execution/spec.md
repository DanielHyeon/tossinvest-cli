> **4판 proposal-freeze 후보.** 개발 진행은 승인됐으며 Manager READY 및 로트별 검증 뒤 구현한다. 운영·실계좌·활성화 승인은 별도다. 현재 연구 GREEN을 제품 PASS로 해석하지 않는다.

## MODIFIED Requirements

### Requirement: Retry Matrix 산출물
재시도 정책은 endpoint×method×오류 클래스 표로 스펙 산출물화해야 한다(SHALL). 최소 규정: **주문 mutation은 어떤 오류에도 자동 재시도 금지 — 단, 멱등 재생은 재시도가 아니라 정체 회수이고, 취소의 `409 already-processing` 은 「취소의 일시적 거절 처리」를 따른다**(같은 키·같은 본문의 재요청은 유효 창 안에서 새 주문을 만들 수 없다 — openapi; IN_DOUBT 해소 요구의 조건·상한·비활성 규칙을 따른다). 이 구분은 재시도 정책 구현부의 문서화된 근거에도 반영한다(SHALL). 조회는 재시도 예산(횟수·총 시간)과 bounded jitter, 401/403은 즉시 신규 진입 차단 + 알림, 필수 조회의 staleness 임계 초과 시 신규 진입 차단. 표의 수치는 구현 시 확정하고 표 없이 구현하지 않는다(SHALL NOT).

엔진 폴링 조회가 429 또는 계좌 공유 admission 거절을 받으면 같은 조회 안에서 다시 보내서는 안 된다(SHALL NOT). 그 조회는 실패로 끝나고 기존 staleness 판정을 유지한다(SHALL). 해당 429 응답에 결속된 유효 `Retry-After`가 30초 이하이면 지정한 엔진 폴링 `RequiredQuery` 종류의 후속 시작을 제한해야 한다(SHALL). 누락·invalid·30초 초과이면 그 조회를 포기하며 30초로 잘라 재전송하거나 기본 대기 후 같은 조회를 재전송해서는 안 된다(SHALL NOT). 기존 유효 미만료 hint를 누락 응답으로 지워서는 안 된다(SHALL NOT). 새 계좌 전역 cooldown을 만들지 않는다. soak·verify·IN_DOUBT 스캔 같은 비지연 walk의 기존 429 백오프(추가 2회, 15s → 30s)는 유지한다.

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
- **THEN** 그 조회 안에서 다시 보내지 않고, 조회 실패가 staleness 임계를 넘으면 신규 진입이 차단되며, 그 429는 공유 이벤트로 기록된다. 누락된 hint로 계좌 전역 패널티를 새로 만들지 않는다

## ADDED Requirements

### Requirement: 계좌 공유 요청 admission
계좌 공유 예산 enforcing 모드에서 동일한 중재에 연결된 참여자의 공식 API 요청 시작은 어떤 rolling 1초에도 합계 3건을 넘어서는 안 된다(SHALL NOT). 시작 경계는 직접 TLS 연결에 HTTP 요청 plaintext를 처음 Write하는 지점이다. 보호·token 교환·401·redirect 및 허용된 내부 재전송도 예외 없이 계수해야 한다(SHALL). 서버 도착률, TLS handshake와 TCP 재전송률로 이 값을 해석해서는 안 된다(SHALL NOT).

보호 우선은 공유 대기열에서 선택을 확정하는 순서이며 이미 선택된 소유자를 선점하지 않는다(SHALL). 공유 상태 또는 동일 중재 경계를 신뢰할 수 없으면 모든 등급의 새 전송을 fail-closed로 막아야 한다(SHALL). 필수 예산이 누락/무효인 경우를 의도된 OFF로 바꾸거나 독립 로컬 예산으로 우회해서는 안 된다(SHALL NOT). OFF는 기존 upstream 동작을 보존한다(SHALL). 운영 활성화와 배포는 별도 사람 승인이 필요하다. 유한 wall-time 보호 대기 상한은 약속하지 않는다.

응답 header 반환만으로 다음 소유자에게 예산을 넘겨서는 안 된다(SHALL NOT). 응답 EOF/Close/취소의 정리와 실제 요청 writer의 revoke/drain을 완료하고 잠금을 정확히 한 번 반환해야 한다(SHALL). 기존 caller 시한을 연장해서는 안 되며, 대기 중 종료가 확인된 요청은 새 Write가 없어야 한다(SHALL). 전송 전 거절의 판단은 현재 주문 본체 attempt에 결속하며, 이미 전송한 결과 또는 a165의 최초 AMBIGUOUS를 후속 거절로 지워서는 안 된다(SHALL NOT).

공유 상태를 갱신하는 중에는 내구화된 dirty 표지를 유지해야 한다(SHALL). 불확실한 갱신 실패 이후 다른 참여자가 그 state를 새 허가로 사용해서는 안 된다(SHALL NOT). dirty를 자동 초기화하거나 살아 있는 waiter를 PID/TTL만으로 지워서는 안 된다(SHALL NOT). namespace/계좌 scope/동일 inode/지원 환경은 실제 활성화 전에 검증해야 한다(SHALL).

#### Scenario: 합산 상한과 보호 대기
- **WHEN** 동일 계좌의 여러 참여자가 일반 요청과 보호 요청을 동시에 전송하려 하면
- **THEN** 모든 실제 wire 시도는 합산 상한을 지키며 보호 요청도 필요한 만큼 기다린다

#### Scenario: 공유 상태 장애
- **WHEN** 공유 파일이나 잠금의 무결성 또는 동일성을 확인할 수 없으면
- **THEN** 보호를 포함한 모든 등급의 새 전송을 막고 로컬 예산으로 우회하지 않는다

#### Scenario: 내부 재전송
- **WHEN** 인증 또는 HTTP 전송 계층에서 재전송을 시도하면
- **THEN** 각 실제 wire 시도는 같은 계좌 상한에 포함되며 이를 보장하지 못하는 경로는 활성화하지 않는다

#### Scenario: 의도된 OFF는 upstream
- **WHEN** 의도된 OFF 경로에서 요청하면
- **THEN** 기존 upstream 동작을 유지한다

#### Scenario: 상한 필수 배포의 예산 누락
- **WHEN** 상한 적용이 필요한 배포에 예산이 없거나 유효하지 않으면
- **THEN** 보호를 포함한 모든 등급의 새 전송을 거부한다

#### Scenario: 느린 응답과 다음 보호 요청
- **WHEN** 일반 요청의 응답 본문이 아직 열려 있고 보호 요청이 대기하면
- **THEN** 보호 요청은 현재 요청의 종료와 최소 간격을 기다리며, 응답 header 반환만으로 다음 Write를 시작하지 않는다

#### Scenario: 상태 갱신의 불확실한 실패
- **WHEN** dirty를 내구화한 뒤 state 갱신 또는 state rename을 내구화하는 directory sync가 실패하여 state의 내구성을 확정하지 못하면
- **THEN** 선택자는 전송하지 않고, 다른 참여자의 critical/normal/auxiliary도 남은 dirty를 확인해 거부한다

#### Scenario: 내구화된 state 이후 dirty 삭제의 동기화 실패
- **WHEN** state와 그 rename의 directory sync가 성공한 뒤 dirty를 삭제했지만 그 삭제의 마지막 directory sync만 실패하면
- **THEN** runtime에서 dirty가 없고 state가 유효한 경우와 재기동 뒤 dirty가 재등장한 경우를 구분하며, 후자는 모든 등급의 새 전송을 거부한다

#### Scenario: 계좌 공유 밖의 요청
- **WHEN** 같은 계좌를 사용하는 다른 호스트나 외부 프로그램이 중재에 연결되지 않았으면
- **THEN** 그 전송을 제한했다고 주장하지 않으며 배포 전 참여 범위를 확인하기 전 계좌 전체 보장을 선언하지 않는다

#### Scenario: 유효한 폴링 hint와 프로세스 재시작
- **WHEN** 유효한 30초 이하 폴링 hint가 남은 상태에서 엔진 프로세스가 재시작하면
- **THEN** 해당 종류의 남은 제한을 복원하며, 무관한 주문이나 모든 계좌 요청에 새 30초/5분 기동 패널티를 추가하지 않는다
