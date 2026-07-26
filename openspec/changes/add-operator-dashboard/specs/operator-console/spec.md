# operator-console Specification (delta)

## ADDED Requirements

### Requirement: 콘솔 안전 불변식
운영 콘솔은 127.0.0.1 전용으로 바인딩하고 비루프백 요청을 거부해야 한다(SHALL). 인증은 기동 시 발급되는 1회성 세션 토큰이며(SHALL — 터미널 점유가 신뢰 근원), 상태를 바꾸는 모든 라우트는 세션+CSRF 이중 게이트를 요구한다(SHALL). 주문 발주·정정·취소·게이트 조작·자격증명 접근에 해당하는 라우트는 존재하지 않는다(SHALL NOT — 라우트 표 정적 검사 + 대표 경로 404 검사). 이 요구사항은 기존 콘솔 구현(2b 1.6)의 성문화이며 동작 변경이 아니다.

#### Scenario: 비루프백 접속
- **WHEN** 루프백이 아닌 주소에서 콘솔에 접속하면
- **THEN** 요청이 거부된다

#### Scenario: 주문 라우트 부재
- **WHEN** 콘솔의 라우트 표를 검사하면
- **THEN** 주문·정정·취소·게이트·자격증명에 해당하는 라우트가 없고, 대시보드가 추가하는 라우트는 전부 GET이며 상태 변경 목록에 추가되지 않는다

### Requirement: 포지션 가시성
콘솔은 계좌의 보유·포지션 상태를 read-only로 표시해야 한다(SHALL): 브로커 보유 스냅샷(수량·평균단가·현재가 — holdings 응답의 lastPrice·평가손익)과 journal 투영(positions·exit_states)을 심볼 기준으로 조인한다. exit 관리 자격이 있는 포지션(entry 결정 또는 편입 기록)은 exit 상태 — t0 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부 — 를 함께 표시한다(SHALL). 자격 없는 보유는 **"관리 외(미편입)"**로 구분하고 exit 라인이 없는 이유(엔진 미관리 — 편입은 adopt-external-positions가 정의)를 명시한다(SHALL — 라벨은 이 하나로 통일). 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다(SHALL — 조인 실패가 화면 실패가 되어서는 안 된다). journal 스키마가 바이너리보다 새로우면(ErrSchemaTooNew) 빈 상태가 아니라 명시적 안내로 표시한다(SHALL).

#### Scenario: 엔진 관리 포지션의 exit 라인 표시
- **WHEN** journal에 exit_states 행이 있는 포지션을 포지션 화면이 렌더하면
- **THEN** 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부가 해당 심볼 행에 표시된다

#### Scenario: 관리 외 보유의 정직한 구분
- **WHEN** 브로커 보유에는 있으나 exit 관리 자격이 없는 심볼을 렌더하면
- **THEN** 해당 행은 "관리 외(미편입)"로 표시되고 exit 라인 없음이 엔진 미관리 때문임이 안내된다

#### Scenario: 엔진 미가동 상태의 대시보드
- **WHEN** journal 파일이 없거나 포지션 테이블이 비어 있으면
- **THEN** 화면은 브로커 보유만으로 렌더되고 "엔진 미가동/관리 포지션 없음"이 안내되며 오류로 처리되지 않는다

#### Scenario: 스키마 불일치
- **WHEN** journal 스키마 버전이 콘솔 바이너리보다 새로우면
- **THEN** 빈 결과 대신 "콘솔 업데이트 필요" 안내가 표시되고 브로커 측 표시는 계속 동작한다

### Requirement: 거래 이력 가시성
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 표시는 journal에 동결된 값과 명시 조인(positions의 심볼, exit_states의 진입가)만 사용하며 fills 재계산을 하지 않는다(SHALL NOT). 스키마에 없는 값(청산가 등)은 표시하지 않는다(SHALL NOT — 동결 계약 우선).

#### Scenario: 동결된 왕복 결과 표시
- **WHEN** trade_outcomes에 행이 있는 상태에서 이력 화면을 열면
- **THEN** 각 왕복의 심볼(positions 조인)·비용 차감 실현손익·실현 R·초기 수량·보유 시간·도달 exit 단계·청산 시각이 동결 값 그대로 표시된다

### Requirement: read-only 불변식
대시보드는 어떤 mutation도 수행해서는 안 된다(SHALL NOT): journal은 `OpenReadOnly`(디렉터리·파일 생성 없음·마이그레이션 없음·`mode=ro`)로만 열고 쓰기 연결 부재를 가드 테스트로 고정한다(SHALL). 콘솔이 주입받는 브로커 인터페이스는 **조회 메서드만 선언**하고(SHALL — holdings 계열), mutation 메서드가 없음을 정적 테스트로 고정한다(SHALL — verifylive.Broker 같은 광폭 인터페이스 주입 금지). 기존 콘솔의 게이트·주문 라우트 부재 가드는 새 라우트 표에서도 유지된다(SHALL).

#### Scenario: journal 쓰기 시도 차단
- **WHEN** 콘솔 코드가 journal 쓰기 경로를 얻으려는 변경이 들어오면
- **THEN** RO 접근 가드 테스트가 실패한다

#### Scenario: 광폭 브로커 인터페이스 주입 차단
- **WHEN** 콘솔의 브로커 인터페이스에 mutation 메서드가 추가되면
- **THEN** 정적 테스트가 실패한다

### Requirement: rate budget 보호
브로커 스냅샷은 요청 시 lazy 갱신이며 백그라운드 폴러를 두지 않는다(SHALL NOT). 갱신 1회의 브로커 호출은 holdings 1콜로 한정하고(SHALL — 현재가는 응답의 lastPrice 사용, 심볼별 시세 fan-out 금지), 서버측 캐시 TTL은 15초 이상이다(SHALL). TTL 내 재요청·다중 탭은 추가 브로커 호출을 유발하지 않으며 캐시 시각이 화면에 표시된다(SHALL). 검증 실행 중에는 갱신을 보류한다(SHALL): 이 콘솔 프로세스의 실행 중 run은 in-process 신호로, 다른 프로세스의 run은 runlock 마커의 mtime 신선도(5분 상한)로 판단한다.

#### Scenario: 새로고침 연타
- **WHEN** TTL 내에 포지션 화면이 여러 번 새로고침되면
- **THEN** 브로커 호출은 추가로 발생하지 않고 캐시 시각이 표시된다

#### Scenario: 검증 실행 중 — 캐시 있음
- **WHEN** 검증 run이 진행 중일 때 캐시가 있는 상태로 포지션 화면을 열면
- **THEN** 새 브로커 호출 없이 캐시 값과 "검증 중 — 갱신 보류" 안내가 표시된다

#### Scenario: 검증 실행 중 — 콜드 캐시
- **WHEN** 검증 run이 진행 중이고 캐시가 비어 있으면
- **THEN** 브로커 데이터 영역은 "검증 중 — 데이터 없음"으로 렌더되고 journal 측 표시는 계속 동작한다
