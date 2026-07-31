# operator-console Specification

## Purpose
로컬 웹 콘솔의 안전 불변식(루프백 리스너·2경로 인증·CSRF·프로세스 재기동 외 상태변경 부재·무주문)과 read-only 운영 가시성(포지션·exit 라인·거래 이력·rate budget 보호) 계약.
## Requirements
### Requirement: 콘솔 안전 불변식

운영 콘솔의 접근 경계는 배포 모드와 일치해야 한다(SHALL): 로컬/토큰 모드는 기존 세션 토큰과 재시작 핸드오프 토큰을 사용하고, 명시적으로 구성된 trusted-network 모드는 인증된 VPN membership을 애플리케이션 접근 경계로 사용하며 별도 애플리케이션 로그인을 요구하지 않는다(SHALL NOT). 상태를 바꾸는 모든 라우트는 공통 `session0` 접근 모드 gate와 CSRF gate를 요구한다(SHALL — trusted-network에서 `session0`는 허용 peer의 VPN membership을 소비한다). 콘솔의 상태변경 행위는 검증 실행 제어(시작·승인·중단), **프로세스 기동·정지**(자기 재시작·soak 재시작·**엔진 시작/정지**), 그리고 **쓰기 전용 Open API 자격증명 설정**뿐이다(SHALL). Open API 설정 라우트는 직접 TLS·허용 peer·정확한 Host/origin·공통 접근 모드·CSRF 뒤에서만 key/secret을 받아야 하고(SHALL), 기존 자격증명을 읽어 응답할 수 없으며(SHALL NOT), 공식 조회 probe로 검증된 key/secret을 보호 저장한 뒤 soak을 시작하는 좁은 capability만 가진다(SHALL). 콘솔 자신은 주문으로 계좌를 건드리지 않는다(SHALL — 자격증명 검증의 공식 계좌 목록 조회는 명시된 read-only 예외이며, 엔진 프로세스가 주문 능력을 갖는지는 §0.7로 승인된 게이트 설정과 기동 인터록이 결정한다). 엔진 상태(실행 여부·기동 거부 사유)는 대시보드에 표시한다(SHALL). 주문 발주·정정·취소·게이트 우회·자격증명 조회 라우트는 존재하지 않는다(SHALL NOT — 라우트 표 정적 검사 + 대표 경로 404 검사).

#### Scenario: 비루프백 리스너
- **WHEN** 루프백이 아닌 주소의 리스너로 Serve가 호출되면
- **THEN** 서비스가 거부되고 리스너가 닫힌다

#### Scenario: 핸드오프 토큰 재사용
- **WHEN** 이미 소비된 핸드오프 토큰으로 재접속하면
- **THEN** 인증이 거부된다

#### Scenario: 주문 라우트 부재
- **WHEN** 콘솔의 라우트 표를 검사하면
- **THEN** 주문·정정·취소·게이트 우회·자격증명 조회 라우트가 없고, Open API 자격증명 설정 POST를 포함한 모든 상태 변경 라우트는 명시적으로 열거되어 공통 접근 모드+CSRF 게이트 뒤에 있다

#### Scenario: trusted-network 자격증명 설정
- **WHEN** 허용 VPN peer가 trusted-network HTTPS 콘솔의 Open API 설정 폼을 제출하면
- **THEN** 별도 애플리케이션 로그인 없이 기존 VPN membership 접근 경계·정확한 Host/origin·CSRF를 통과해야 하며, 어느 경계라도 틀리면 저장 seam 전에 거부된다

#### Scenario: Open API 자격증명 비노출
- **WHEN** 콘솔의 Open API 설정 GET·오류 응답·redirect를 검사하면
- **THEN** 저장되었거나 제출된 key와 secret은 어느 응답 표면에도 존재하지 않는다

#### Scenario: 인터록 미충족 상태의 엔진 시작 버튼
- **WHEN** ProtectionReady 미충족 상태에서 [엔진 시작]을 누르면
- **THEN** 엔진 프로세스가 기동 거부한 사유(미충족 항목)가 대시보드에 표시된다 — 콘솔이 인터록을 우회할 수 없다

#### Scenario: 엔진 정지 버튼
- **WHEN** 실행 중 엔진에 대해 [엔진 정지]를 누르면
- **THEN** 엔진은 시그널 종료 규율(루프 완주·journal 정합 close)로 정지하고 상태가 대시보드에 반영된다

### Requirement: 포지션 가시성
콘솔은 계좌의 보유·포지션 상태를 read-only로 표시해야 한다(SHALL): 브로커 보유 스냅샷(수량·평균단가·현재가 — holdings 응답의 lastPrice·평가손익)과 journal 투영(positions·exit_states)을 심볼 기준으로 조인한다. exit 관리 자격이 있는 포지션(자격의 정의는 exit-policy·position-ledger 스펙이 소유한다)은 exit 상태 — t0 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부 — 를 함께 표시한다(SHALL). 자격 없는 보유는 **"관리 외(미편입)"**로 구분하고 exit 라인이 없는 이유를 명시한다(SHALL — 라벨은 이 하나로 통일하며 행마다 표시한다). 사유의 명시 위치는 스코프를 따른다(SHALL): 같은 상태의 모든 행에 공통인 사유 — 원장 미판독(관리 여부 불명), 원장에 포지션 없음 — 는 페이지 수준 안내 1회로 명시하고 같은 문장을 행마다 반복하지 않는다(SHALL NOT — 반복 안내는 표를 읽을 수 없게 만든다); 행 고유 사유(자격 기록 없는 원장 포지션, 자격은 있으나 exit 미개설)는 해당 행에 명시한다. 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다(SHALL — 조인 실패가 화면 실패가 되어서는 안 된다). journal 스키마 불일치는 방향별로 구분 안내한다(SHALL): 바이너리보다 새로우면 "콘솔 업데이트 필요", 오래됐으면(필요 테이블 부재) "엔진 기동으로 마이그레이션 필요" — 어느 쪽도 빈 상태로 위장하지 않는다.

#### Scenario: 엔진 관리 포지션의 exit 라인 표시
- **WHEN** journal에 exit_states 행이 있는 포지션을 포지션 화면이 렌더하면
- **THEN** 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부가 해당 심볼 행에 표시된다

#### Scenario: 관리 외 보유의 정직한 구분
- **WHEN** 브로커 보유에는 있으나 exit 관리 자격이 없는 심볼을 렌더하면
- **THEN** 해당 행은 "관리 외(미편입)"로 표시되고 exit 라인 없음이 엔진 미관리 때문임이 안내된다

#### Scenario: 같은 사유의 수동 보유 다수
- **WHEN** 원장에 포지션이 없는 보유가 여러 종목 렌더되면
- **THEN** 각 행은 "관리 외(미편입)" 라벨을 달고, 원장 부재 사유 안내는 페이지에 1회만 나타난다

#### Scenario: 엔진 미가동 상태의 대시보드
- **WHEN** journal 파일이 없거나 포지션 테이블이 비어 있으면
- **THEN** 화면은 브로커 보유만으로 렌더되고 "엔진 미가동/관리 포지션 없음"이 안내되며 오류로 처리되지 않는다

#### Scenario: 스키마 불일치 — 양방향
- **WHEN** journal 스키마가 콘솔 바이너리보다 새롭거나, 마이그레이션 전이라 필요 테이블이 없으면
- **THEN** 각각 "콘솔 업데이트 필요"/"엔진 기동 필요" 안내가 표시되고 브로커 측 표시는 계속 동작한다

### Requirement: 거래 이력 가시성
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 표시는 journal에 동결된 값과 명시 조인(positions의 심볼, exit_states의 진입가)만 사용하며 fills 재계산을 하지 않는다(SHALL NOT). 스키마에 없는 값(청산가 등)은 표시하지 않고(SHALL NOT), nullable 필드(보유 시간 등)의 NULL은 "—"로 렌더한다(SHALL). 성과 행이 생기지 않는 종결(외부 매도로 닫힌 포지션 — adopt-external-positions design A7)은 이 화면의 한계로 명시하고 exit_events 표시가 그 공백을 보완한다(SHALL 명시).

#### Scenario: 동결된 왕복 결과 표시
- **WHEN** trade_outcomes에 행이 있는 상태에서 이력 화면을 열면
- **THEN** 각 왕복의 심볼(positions 조인)·비용 차감 실현손익·실현 R·초기 수량·보유 시간(NULL은 "—")·도달 exit 단계·청산 시각이 동결 값 그대로 표시된다

### Requirement: read-only 불변식
대시보드는 어떤 mutation도 수행해서는 안 된다(SHALL NOT): journal은 `OpenReadOnly`로만 연다 — DB 파일·디렉터리를 생성하지 않고, 마이그레이션을 실행하지 않으며, DB에 쓰지 않는다(SHALL — `mode=ro`; WAL 공유 인덱스(`-shm`/`-wal`) 접근은 SQLite WAL 읽기의 전제로서 명시된 예외다). 쓰기 연결 부재를 가드 테스트로 고정한다(SHALL). 콘솔이 주입받는 브로커 인터페이스는 **조회 메서드만 선언**하고(SHALL — holdings 계열), mutation 메서드가 없음을 정적 테스트로 고정한다(SHALL — verifylive.Broker 같은 광폭 인터페이스 주입 금지). 기존 콘솔의 게이트·주문 라우트 부재 가드는 새 라우트 표에서도 유지된다(SHALL).

#### Scenario: journal 쓰기 시도 차단
- **WHEN** 콘솔 코드가 journal 쓰기 경로를 얻으려는 변경이 들어오면
- **THEN** RO 접근 가드 테스트가 실패한다

#### Scenario: 광폭 브로커 인터페이스 주입 차단
- **WHEN** 콘솔의 브로커 인터페이스에 mutation 메서드가 추가되면
- **THEN** 정적 테스트가 실패한다

### Requirement: rate budget 보호
브로커 스냅샷은 요청 시 lazy 갱신이며 서버측 백그라운드 폴러를 두지 않는다(SHALL NOT). 갱신 1회의 브로커 호출은 holdings 1콜로 한정하고(SHALL — 현재가는 응답의 lastPrice 사용, 심볼별 시세 fan-out 금지), 서버측 캐시 TTL은 15초 이상이다(SHALL). TTL 내 재요청·다중 탭은 추가 브로커 호출을 유발하지 않으며 캐시 시각이 화면에 표시된다(SHALL). 포지션 화면은 브라우저 재로드 지시(meta refresh)를 포함할 수 있으며 그 주기는 캐시 TTL 이상이어야 한다(SHALL — 각 재로드는 요청 시 lazy 갱신을 그대로 타므로 열린 탭 하나의 비용 상한은 holdings 1콜/TTL이다; 이 지시는 서버측 폴러가 아니다). 검증 실행 중에는 갱신을 보류한다(SHALL): 이 콘솔 프로세스의 실행 중 run은 in-process 신호로, 다른 프로세스의 run은 runlock 마커의 mtime 신선도(5분 상한 — stale은 무시)로 판단한다. 자동 재로드는 이 보류를 우회하지 못한다(SHALL — 보류 중 재로드는 캐시를 서빙할 뿐 브로커 호출을 만들지 않는다).

#### Scenario: 새로고침 연타
- **WHEN** TTL 내에 포지션 화면이 여러 번 새로고침되면
- **THEN** 브로커 호출은 추가로 발생하지 않고 캐시 시각이 표시된다

#### Scenario: 자동 재로드의 비용 상한
- **WHEN** 포지션 화면이 재로드 지시 주기에 따라 스스로 다시 열리면
- **THEN** 서버는 요청 시 lazy 판정을 그대로 적용하며, 브로커 호출은 TTL당 1콜을 넘지 않는다

#### Scenario: 검증 실행 중 — 캐시 있음
- **WHEN** 검증 run이 진행 중일 때 캐시가 있는 상태로 포지션 화면을 열면
- **THEN** 새 브로커 호출 없이 캐시 값과 "검증 중 — 갱신 보류" 안내가 표시된다

#### Scenario: 검증 실행 중 — 콜드 캐시
- **WHEN** 검증 run이 진행 중이고 캐시가 비어 있으면
- **THEN** 브로커 데이터 영역은 "검증 중 — 데이터 없음"으로 렌더되고 journal 측 표시는 계속 동작한다

