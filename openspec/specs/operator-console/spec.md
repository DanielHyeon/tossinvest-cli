# operator-console Specification

## Purpose
로컬 웹 콘솔의 안전 불변식(루프백 리스너·2경로 인증·CSRF·프로세스 재기동 외 상태변경 부재·무주문)과 read-only 운영 가시성(포지션·exit 라인·거래 이력·rate budget 보호) 계약.
## Requirements
### Requirement: 콘솔 안전 불변식

운영 콘솔은 127.0.0.1 리스너에만 서비스해야 한다(SHALL — Serve는 비루프백 리스너를 거부하고 닫는다). 인증 경로는 둘이다(SHALL — 2b 1.6·1.7·1.8 구현의 성문화이며 동작 변경이 아니다): ① **세션 토큰** — 기동 시 발급되어 프로세스 수명 동안 유효하며 URL→쿠키 1회 교환으로 전달된다(신뢰 근원은 터미널 점유), ② **재시작 핸드오프 토큰** — 웹 재시작 시 0600 파일로 전달되는 단발성 자격(소비 즉시 무효·재사용 거부·짧은 유효 시간(현행 2분) 경과 시 거부, 신뢰 근원은 사용자 데이터 디렉터리의 파일 소유권; 이미 인증된 세션만 발행할 수 있다). 상태를 바꾸는 모든 라우트는 세션+CSRF 이중 게이트를 요구한다(SHALL). 콘솔의 상태변경 행위는 검증 실행 제어(시작·승인·중단), **프로세스 기동·정지**(자기 재시작·soak 재시작·**엔진 시작/정지**), 그리고 **편입 설정 편집**(편입 설정 저장·종목 편입 지정 — 대상은 `engine.adoption` config 블록뿐이다)뿐이다(SHALL). 편입 설정 편집은 계좌·journal·브로커에 닿지 않는다(SHALL — 편입의 실행 주체는 엔진 대사 루프이며, 콘솔 저장은 다음 엔진 기동부터 읽히는 구성 기록이다; 반영 시점은 화면에 명시한다). automation gate(운영 게이트)·Guardian 한도·kill switch는 콘솔에서 편집할 수 없다(SHALL NOT — 게이트 ON은 §0.7 콘솔 밖 절차 유지). 콘솔 자신은 계좌를 건드리지 않는다(SHALL — 엔진 프로세스가 주문 능력을 갖는지는 §0.7로 승인된 게이트 설정과 기동 인터록이 결정하며, 콘솔 버튼은 그 구성의 프로세스를 켜고 끄거나 그 구성을 기록할 뿐이다). 엔진 상태(실행 여부·기동 거부 사유)는 대시보드에 표시한다(SHALL). 주문 발주·정정·취소·게이트 조작·자격증명 접근 라우트는 존재하지 않는다(SHALL NOT — 라우트 표 정적 검사 + 대표 경로 404 검사).

#### Scenario: 비루프백 리스너
- **WHEN** 루프백이 아닌 주소의 리스너로 Serve가 호출되면
- **THEN** 서비스가 거부되고 리스너가 닫힌다

#### Scenario: 핸드오프 토큰 재사용
- **WHEN** 이미 소비된 핸드오프 토큰으로 재접속하면
- **THEN** 인증이 거부된다

#### Scenario: 주문 라우트 부재
- **WHEN** 콘솔의 라우트 표를 검사하면
- **THEN** 주문·정정·취소·게이트·자격증명에 해당하는 라우트가 없고, 상태 변경 목록은 본문이 열거한 행위(검증 제어·프로세스 기동/정지·편입 설정 편집)뿐이며 그 외 신규 라우트는 전부 GET이다

#### Scenario: 인터록 미충족 상태의 엔진 시작 버튼
- **WHEN** ProtectionReady 미충족 상태에서 [엔진 시작]을 누르면
- **THEN** 엔진 프로세스가 기동 거부한 사유(미충족 항목)가 대시보드에 표시된다 — 콘솔이 인터록을 우회할 수 없다

#### Scenario: 엔진 정지 버튼
- **WHEN** 실행 중 엔진에 대해 [엔진 정지]를 누르면
- **THEN** 엔진은 시그널 종료 규율(루프 완주·journal 정합 close)로 정지하고 상태가 대시보드에 반영된다

#### Scenario: CSRF 없는 편입 설정 저장
- **WHEN** 세션은 유효하나 CSRF 토큰이 없거나 틀린 요청이 편입 설정 저장·종목 편입 지정 라우트에 도달하면
- **THEN** 요청이 거부되고 config는 변경되지 않는다

### Requirement: 포지션 가시성
콘솔은 계좌의 보유·포지션 상태를 read-only로 표시해야 한다(SHALL): 브로커 보유 스냅샷(수량·평균단가·현재가 — holdings 응답의 lastPrice·평가손익)과 journal 투영(positions·exit_states)을 심볼 기준으로 조인한다. exit 관리 자격이 있는 포지션(자격의 정의는 exit-policy·position-ledger 스펙이 소유한다)은 exit 상태 — t0 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부 — 를 함께 표시한다(SHALL). 자격 없는 보유의 판정 라벨은 편입 지정 상태를 반영한다(SHALL — 사용자 UX 결정 2026-07-27): `include_symbols`에 지정된(체크된) 행은 **"관리 편입"**, 지정되지 않은 행은 **"관리 외(미편입)"** — 각 라벨의 철자는 하나로 통일하며 행마다 표시한다. "관리 편입" 라벨은 편입 예약 상태의 표시일 뿐 보호 성립을 의미하지 않는다(SHALL — "편입 예약됨"과 실행 주체(엔진 대사 루프) 안내를 병기하고, 원장을 읽지 못한 행은 지정 여부와 무관하게 "관리 여부 불명"을 유지한다 — 콘솔이 관측하지 않은 보호를 단정하지 않는다). exit 라인이 없는 이유의 명시 위치는 스코프를 따른다(SHALL): 같은 상태의 모든 행에 공통인 사유 — 원장 미판독(관리 여부 불명), 원장에 포지션 없음 — 는 페이지 수준 안내 1회로 명시하고 같은 문장을 행마다 반복하지 않는다(SHALL NOT — 반복 안내는 표를 읽을 수 없게 만든다); 행 고유 사유(자격 기록 없는 원장 포지션, 자격은 있으나 exit 미개설)는 해당 행에 명시한다. 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다(SHALL — 조인 실패가 화면 실패가 되어서는 안 된다). journal 스키마 불일치는 방향별로 구분 안내한다(SHALL): 바이너리보다 새로우면 "콘솔 업데이트 필요", 오래됐으면(필요 테이블 부재) "엔진 기동으로 마이그레이션 필요" — 어느 쪽도 빈 상태로 위장하지 않는다.

#### Scenario: 엔진 관리 포지션의 exit 라인 표시
- **WHEN** journal에 exit_states 행이 있는 포지션을 포지션 화면이 렌더하면
- **THEN** 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부가 해당 심볼 행에 표시된다

#### Scenario: 관리 외 보유의 정직한 구분
- **WHEN** 브로커 보유에는 있으나 exit 관리 자격이 없고 include에도 지정되지 않은 심볼을 렌더하면
- **THEN** 해당 행은 "관리 외(미편입)"로 표시되고 exit 라인 없음이 엔진 미관리 때문임이 안내된다

#### Scenario: 지정된 행의 라벨
- **WHEN** include에 지정된 미편입 보유 행을 렌더하면
- **THEN** 판정 라벨이 "관리 편입"으로 표시되고 "편입 예약됨"과 실행 주체 안내가 병기되며, 같은 심볼이라도 원장을 읽지 못한 상태에서는 "관리 여부 불명"이 유지된다

#### Scenario: 같은 사유의 수동 보유 다수
- **WHEN** 원장에 포지션이 없는 지정되지 않은 보유가 여러 종목 렌더되면
- **THEN** 각 행은 "관리 외(미편입)" 라벨을 달고, 원장 부재 사유 안내는 페이지에 1회만 나타난다

#### Scenario: 엔진 미가동 상태의 대시보드
- **WHEN** journal 파일이 없거나 포지션 테이블이 비어 있으면
- **THEN** 화면은 브로커 보유만으로 렌더되고 "엔진 미가동/관리 포지션 없음"이 안내되며 오류로 처리되지 않는다

#### Scenario: 스키마 불일치 — 양방향
- **WHEN** journal 스키마가 콘솔 바이너리보다 새롭거나, 마이그레이션 전이라 필요 테이블이 없으면
- **THEN** 각각 "콘솔 업데이트 필요"/"엔진 기동 필요" 안내가 표시되고 브로커 측 표시는 계속 동작한다

### Requirement: 거래 화면 핵심 정보 계층과 반응형 표시
콘솔의 포지션과 주문 화면은 운영자가 첫 화면에서 판단해야 하는 사실을 1차 정보로 표시해야 한다(SHALL).
포지션은 각 보유의 관리 상태·평가손익/수익률·진입가·현재 보호선·익절 진행을 항상 보이는
행에 두고, 현재 보호선은 journal `exit_states.baseline`, 최초 손절은 `initial_stop`으로 구분해야
한다(SHALL NOT — 화면이 정책을 재계산하거나 둘을 같은 값으로 명명하지 않는다). 주문은 미체결
일반/조건/합계, 필터 결과, 각 주문의 시각·심볼/시장·방향·상태·수량·가격·발주 주체를 1차 정보로
표시해야 한다(SHALL). 낮은 우선순위의 원장 식별자·래칫 진단·주문번호·평균체결가는 native HTML
상세 영역에 둘 수 있다.

두 화면은 375 CSS pixel viewport에서 header, navigation, summary, table, 긴 식별자를 포함한 문서
전체의 수평 overflow 없이 핵심 정보를 읽고 조작할 수 있어야 한다(SHALL). 표는 접근 가능한 이름,
열/행 header 관계를 유지하고(SHALL), 현재 navigation은 `aria-current`를 제공해야 한다(SHALL).
버튼·summary·필터 링크는 keyboard focus가 보이고 모바일 조작 대상은 최소 44 CSS pixel 높이를
가져야 한다(SHALL). 이 반응형 표시는 JavaScript나 외부 asset을 요구하지 않아야 한다(SHALL NOT).
브로커·journal 미측정, 페이지 잘림/하한 건수, 조건주문 발주 주체 불명, 해석 불가능 상태,
파싱되지 않은 원본 시각, 캐시 실패·보류·stale 안내는 접지 않고 해당 수치와 함께 노출해야 한다
(SHALL). 시각이 기록된 캐시·측정에는 그 시각을 표시하되, 시각이 없는 journal 실패에 화면이 시각을
만들어내지 않아야 한다(SHALL NOT).

#### Scenario: 관리 포지션의 보호 상태 스캔
- **WHEN** exit 상태가 있는 관리 포지션을 렌더하면
- **THEN** 같은 1차 행에서 관리 상태, 평가손익/수익률, 진입가, 현재 보호선, 최초 손절과 익절 진행을 확인할 수 있다

#### Scenario: 주문 화면의 작은 viewport
- **WHEN** 주문이 있는 화면을 375 CSS pixel 폭으로 렌더하면
- **THEN** `documentElement.scrollWidth`가 viewport 너비를 넘지 않고 각 주문의 핵심 필드 라벨과 값을 읽을 수 있다

#### Scenario: 미측정 상태의 진실성
- **WHEN** 브로커 또는 journal 읽기가 실패하거나 캐시가 stale이면
- **THEN** 해당 사유와 기록된 시각이 접히지 않은 안내로 표시되고, 기록되지 않은 시각을 만들거나 빈 목록/현재 측정값으로 위장하지 않는다

### Requirement: CSP 안전한 포지션 관리 조작
포지션 화면의 관리 외 보유에 대한 편입 지정/해제와 자동관리 제외/해제 조작은 배포 CSP에서 동작해야 한다(SHALL).
조작은 현재 상태에서 발생할 변경을 동사형 문구로 표시하는 same-origin POST
form이어야 하고 세션과 CSRF 게이트를 그대로 통과해야 한다(SHALL). inline event handler,
client-side script, `javascript:` URL, CSP 완화에 의존하지 않아야 한다(SHALL NOT). submit은 기존
include/exclude 설정만 멱등 갱신하며 편입 실행, 기존 보호선 변경, 엔진 기동 또는 주문을 수행하지
않아야 한다(SHALL NOT). 기존 관리 판정 라벨은 동작 버튼과 별도로 항상 표시해야 하며(SHALL),
반복되는 버튼의 접근 가능한 이름에는 대상 심볼과 동작이 모두 포함되어야 한다(SHALL).

#### Scenario: 제외되지 않은 보유를 자동관리에서 제외
- **WHEN** 운영자가 포지션 행의 “자동관리 제외”를 누르면
- **THEN** inline handler 없이 `/settings/exclude`로 CSRF 보호 POST가 전송되고 exclude 목록만 갱신된다

#### Scenario: 이미 제외된 보유의 제외 해제
- **WHEN** 운영자가 제외된 포지션 행의 “제외 해제”를 누르면
- **THEN** 같은 form이 remove 의도를 전송하고 해당 심볼만 exclude 목록에서 제거된다

#### Scenario: CSP 회귀 검사
- **WHEN** 포지션 페이지의 렌더 결과와 응답 CSP를 검사하면
- **THEN** 렌더 결과 전체에 `on[a-z]+=` inline handler, `<script>`, `javascript:` URL이 없고 응답 CSP의 `default-src 'none'`과 `form-action 'self'`가 유지된다

### Requirement: 거래 이력 가시성
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 표시는 journal에 동결된 값과 명시 조인(positions의 심볼, exit_states의 진입가)만 사용하며 fills 재계산을 하지 않는다(SHALL NOT). 스키마에 없는 값(청산가 등)은 표시하지 않고(SHALL NOT), nullable 필드(보유 시간 등)의 NULL은 "—"로 렌더한다(SHALL). 성과 행이 생기지 않는 종결(외부 매도로 닫힌 포지션 — adopt-external-positions design A7)은 이 화면의 한계로 명시하고 exit_events 표시가 그 공백을 보완한다(SHALL 명시).

#### Scenario: 동결된 왕복 결과 표시
- **WHEN** trade_outcomes에 행이 있는 상태에서 이력 화면을 열면
- **THEN** 각 왕복의 심볼(positions 조인)·비용 차감 실현손익·실현 R·초기 수량·보유 시간(NULL은 "—")·도달 exit 단계·청산 시각이 동결 값 그대로 표시된다

### Requirement: read-only 불변식
대시보드는 계좌·원장에 대한 어떤 mutation도 수행해서는 안 된다(SHALL NOT): journal은 `OpenReadOnly`로만 연다 — DB 파일·디렉터리를 생성하지 않고, 마이그레이션을 실행하지 않으며, DB에 쓰지 않는다(SHALL — `mode=ro`; WAL 공유 인덱스(`-shm`/`-wal`) 접근은 SQLite WAL 읽기의 전제로서 명시된 예외다). 쓰기 연결 부재를 가드 테스트로 고정한다(SHALL). 콘솔이 주입받는 브로커 인터페이스는 **조회 메서드만 선언**하고(SHALL — holdings 계열), mutation 메서드가 없음을 정적 테스트로 고정한다(SHALL — verifylive.Broker 같은 광폭 인터페이스 주입 금지). **config에 대한** 콘솔의 유일한 쓰기 표면은 주입된 편입 설정 seam이며(SHALL — 대상은 config 파일의 `engine.adoption` 블록만; 검증 증거 기록·핸드오프 토큰 파일 등 기존 주입 writer의 계약은 무변경), 이 seam은 다른 config 키를 유실하지 않고(SHALL — 구조체 왕복이 아니라 해당 블록만 교체·블록 밖 바이트 보존) 유일 임시파일과 잠금 아래 원자적으로 기록한다(SHALL — 동시 기록의 lost-update 금지). seam의 Load는 파일의 `engine.adoption` 블록 **원문**을 반환하고 검증 판정을 별도로 병기한다(SHALL — 거부된 블록의 목록이 화면 왕복으로 유실되어서는 안 된다). 파싱할 수 없는 config 파일에 대한 저장은 거부된다(SHALL — 골격 생성은 파일 부재에 한정). seam은 Load·Save 두 메서드만 선언하며(SHALL — 정적 검사) internal/console은 config 서비스 타입을 직접 명명하지 않는다(SHALL NOT — 정적 검사). seam이 배선되지 않은 빌드에서 설정 화면은 저장 불가를 안내하고 나머지 화면은 영향받지 않는다(SHALL). Save 성공은 audit 로그에 저장 시점 엔트리를 남긴다(SHALL — §0.5; 엔진 기동 시 diff 기록과 이중이며, 기동 없는 flip도 기록에 남는다). 기존 콘솔의 게이트·주문 라우트 부재 가드는 새 라우트 표에서도 유지된다(SHALL).

#### Scenario: journal 쓰기 시도 차단
- **WHEN** 콘솔 코드가 journal 쓰기 경로를 얻으려는 변경이 들어오면
- **THEN** RO 접근 가드 테스트가 실패한다

#### Scenario: 광폭 브로커 인터페이스 주입 차단
- **WHEN** 콘솔의 브로커 인터페이스에 mutation 메서드가 추가되면
- **THEN** 정적 테스트가 실패한다

#### Scenario: 편입 설정 저장의 외과적 기록
- **WHEN** 콘솔이 편입 설정을 저장할 때 config 파일에 이 스키마가 모르는 키가 존재하면
- **THEN** 저장 후에도 그 키는 보존되고 `engine.adoption` 블록만 바뀐다

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

### Requirement: 편입 설정 화면
콘솔은 편입 설정 화면을 제공해야 한다(SHALL): 상단 navigation은 **"외부 종목 자동관리"** 메뉴를 표시하고 `/settings#adoption`으로 기존 편입 설정의 첫 섹션을 직접 열어야 한다(SHALL). 화면 제목과 첫 설명은 수동 매수한 외부 종목이 편입 완료 후 기존 공통 익절·보호선·손익 극대화 정책의 관리 대상이 된다는 결과와, 저장 자체가 편입·주문을 실행하지 않고 엔진 대사 루프가 다음 기동부터 수행한다는 경계를 명시해야 한다(SHALL).

화면은 `adoption.enabled`·`default_stop_pct`·`exclude_symbols`·`include_symbols`를 표시·편집한다(SHALL). 합성 손절폭은 기존 CSP 아래에서도 현재값과 편집값이 일치하도록 **명시적 백분율 숫자 입력**으로 표시한다(SHALL — 이전 마우스 슬라이더/타이핑 금지 결정은 CSP가 inline 갱신을 차단해 현재값을 거짓 표시하므로 이 요구가 대체한다). 입력 필드명은 `default_stop_percent`이고, 허용값은 유한한 2%..20%의 0.5 percentage-point 단위다(SHALL). 서버는 브라우저 검증을 신뢰하지 않고 empty·비수치·NaN·무한대·범위 밖·단위 밖·과거 fraction 형식 제출을 저장 전에 거부하며(SHALL), 허용 percentage tick만 100으로 나눠 기존 fractional `default_stop_pct`로 저장한다(SHALL). 렌더링은 불필요한 0이나 부동소수 artifact가 없는 결정적 10진 표기여야 한다(SHALL). 기존 파일에 엔진은 허용하지만 콘솔 0.5% grid에는 맞지 않는 값이 있으면 정확한 현재값과 보정 필요를 표시하고, 저장 시 침묵 반올림하지 않고 거부한다(SHALL).

목록 직접 기입은 고급 접힘 안에만 두고, 거부된 블록은 파일 원문 값과 거부 사유를 함께 표시한다(SHALL). 저장은 다른 편입 필드를 보존하고 audit을 남기며, 반영 시점("다음 엔진 기동부터 반영"), 편입 비가역 귀결, 지정의 상시성, 현재 엔진 실행 여부를 안내한다(SHALL). 저장은 엔진을 재시작하거나 기존 관리 포지션을 변경하지 않는다(SHALL NOT). 이 백분율 control에는 client-side script나 CSP 예외가 없어야 한다(SHALL NOT — 같은 template의 다른 legacy confirm handler는 이 change 범위 밖이다). automation gate 편집 표면은 존재하지 않는다(SHALL NOT). 포지션 화면의 관리 외 보유 행은 기존 종목 편입 지정/해제 행위를 유지하고, 그 행위는 include 목록만 바꾸며 편입을 직접 수행하지 않는다(SHALL NOT).

#### Scenario: 상단 메뉴에서 외부 종목 자동관리 설정 열기
- **WHEN** 운영자가 어느 콘솔 화면에서든 상단의 "외부 종목 자동관리"를 선택하면
- **THEN** `/settings#adoption`이 열리고 자동 편입 ON/OFF·합성 보호폭·제외·지정 설정과 기존 정책 적용 설명이 첫 섹션에 표시된다

#### Scenario: 현재 비기본값 표시
- **WHEN** 저장된 adoption fraction이 `0.075`이면
- **THEN** 화면은 inline event handler나 CSP 예외 없이 편집 가능한 `7.5%`를 표시한다

#### Scenario: 유효한 비기본값 저장
- **WHEN** 운영자가 `default_stop_percent=7.5`를 제출하면
- **THEN** `default_stop_pct=0.075`가 저장·audit되고 다른 adoption 필드는 보존되며 실행 중 엔진에는 재시작이 필요하다고 안내한다

#### Scenario: 전체 허용 grid
- **WHEN** 2%부터 20%까지 0.5 percentage-point 단위의 어느 값이든 제출하면
- **THEN** 정확히 100으로 나눈 fraction이 저장되고 재로드한 config가 같은 fraction을 반환한다

#### Scenario: 잘못된 백분율
- **WHEN** empty·비수치·NaN·무한대·2% 미만·20% 초과·0.5 단위 밖 값 또는 과거 `0.05` fraction 형식을 제출하면
- **THEN** 구체적 안내와 함께 저장이 거부되고 adoption 블록은 변경되지 않는다

#### Scenario: 기존 grid 밖 값
- **WHEN** 기존 config에 `default_stop_pct=0.076`이 있으면
- **THEN** 화면은 `7.6%`를 정확히 표시하고 허용 단위로 보정해야 함을 안내하며, 그대로 저장해도 침묵 반올림 없이 거부한다

#### Scenario: deny-by-default CSP
- **WHEN** remote settings 응답이 배포 CSP와 함께 렌더되면
- **THEN** adoption percentage control에는 inline script가 없고 값은 표시·편집 가능하다

#### Scenario: 종목별 편입 지정과 해제
- **WHEN** 포지션 화면에서 관리 외 보유의 지정 또는 기존 지정의 해제를 수행하면
- **THEN** include 목록만 멱등 갱신되고 이미 편입된 포지션과 엔진 프로세스에는 영향이 없다

#### Scenario: seam 미배선 빌드
- **WHEN** 편입 설정 seam이 주입되지 않은 콘솔에서 설정 화면을 열면
- **THEN** 저장 불가와 그 이유가 안내되고 다른 화면은 정상 동작한다

### Requirement: active profile journal identity
The operator console SHALL resolve its read-only journal from the same active
profile rule as the engine: when an explicit `--config-dir` is present it SHALL
read `<config-dir>/journal.db`, and without that override it SHALL use the
default data-directory journal. The console SHALL continue to open the selected
file through the read-only journal API and SHALL NOT create, migrate, copy,
merge, modify, or fall back to another journal. A selected journal that lacks
the required v9 `exit_states.policy_id` column SHALL be classified as too old
before account queries run.

#### Scenario: explicit profile has a newer journal than the default data directory
- **WHEN** the console starts with `--config-dir P`, `P/journal.db` is a readable current-schema journal, and the default data-directory journal is stale or incompatible
- **THEN** the positions screen reads management state only from `P/journal.db` and does not query or modify the unrelated default journal

#### Scenario: no explicit profile
- **WHEN** the console starts without `--config-dir`
- **THEN** it resolves the same default journal path as before and opens it read-only

#### Scenario: selected profile journal is unavailable
- **WHEN** the selected profile journal is missing, too new, or too old
- **THEN** the corresponding existing failure state is rendered without falling back to another journal

#### Scenario: selected v8 journal lacks policy identity
- **WHEN** the selected journal has the v8 tables but lacks `exit_states.policy_id`
- **THEN** the console renders the too-old migration guidance before any positions query and never exposes a raw `no such column: policy_id` error
