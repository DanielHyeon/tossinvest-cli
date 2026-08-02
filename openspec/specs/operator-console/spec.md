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
콘솔의 대시보드·포지션·주문 화면은 운영자가 첫 화면에서 판단해야 하는 사실을 1차 정보로 표시해야 한다(SHALL). 대시보드와 포지션 화면은 동일한 read-only 보유 projection을 사용해야 하며(SHALL), 둘 중 하나가 lifecycle·desired/effective policy 또는 exit evidence를 다르게 해석해서는 안 된다(SHALL NOT).

각 보유의 1차 행은 종목명/심볼, 수량, 평균가, 현재가, 익절, 손절, 추적 회수, 기준선, 고점(high-water), 총금액, 미실현 PnL을 항상 표시해야 한다(SHALL). `익절`은 canonical exit projection의 다음 익절가, `손절`은 최초 손절, `추적 회수`는 다음 보호선, `기준`은 현재 보호선, `고점`은 high-water를 의미해야 한다(SHALL). canonical snapshot이 없거나 stale·lifecycle generation 불일치·runtime unknown이면 actionable 가격을 계산하거나 raw 원장값으로 대체하지 않고 `—`와 상태/사유를 표시해야 한다(SHALL NOT). 저장된 비실효 기준은 상세 증거로만 표시할 수 있다.

관리 상태와 차단·stale·미측정 verdict는 접지 않고 표시해야 한다(SHALL). 낮은 우선순위의 자격 provenance, 원장 수량·평단, position/policy generation, decision/snapshot/observation ID, 평가 시각, 원본 저장 기준은 키보드로 여는 native HTML 상세 영역에 둘 수 있다. 상세 영역을 열지 않아도 앞 문단의 1차 필드와 안전 verdict를 모두 확인할 수 있어야 한다(SHALL).

주문은 미체결 일반/조건/합계, 필터 결과, 각 주문의 시각·심볼/시장·방향·상태·수량·가격·발주 주체를 1차 정보로 표시해야 한다(SHALL). 낮은 우선순위의 원장 식별자·래칫 진단·주문번호·평균체결가는 native HTML 상세 영역에 둘 수 있다.

세 화면은 375 CSS pixel viewport에서 header, navigation, summary, table, 긴 식별자를 포함한 문서 전체의 수평 overflow 없이 핵심 정보를 읽고 조작할 수 있어야 한다(SHALL). 표는 접근 가능한 이름, 열/행 header 관계를 유지하고(SHALL), 현재 navigation은 `aria-current`를 제공해야 한다(SHALL). 버튼·summary·필터 링크는 keyboard focus가 보이고 모바일 조작 대상은 최소 44 CSS pixel 높이를 가져야 한다(SHALL). 이 반응형 표시는 JavaScript나 외부 asset을 요구하지 않아야 한다(SHALL NOT). 브로커·journal 미측정, 페이지 잘림/하한 건수, 조건주문 발주 주체 불명, 해석 불가능 상태, 파싱되지 않은 원본 시각, 캐시 실패·보류·stale 안내는 접지 않고 해당 수치와 함께 노출해야 한다(SHALL). 시각이 기록된 캐시·측정에는 그 시각을 표시하되, 시각이 없는 journal 실패에 화면이 시각을 만들어내지 않아야 한다(SHALL NOT).

#### Scenario: 대시보드와 포지션의 같은 보유
- **WHEN** 같은 캐시와 journal 상태에서 `/dashboard`와 `/positions`를 렌더하면
- **THEN** 두 화면의 해당 종목은 같은 관리 verdict와 같은 익절·손절·추적 회수·기준·고점 값을 표시한다

#### Scenario: 관리 포지션의 핵심 정보 스캔
- **WHEN** fresh canonical exit snapshot이 있는 관리 포지션을 렌더하면
- **THEN** 상세 영역을 열지 않은 같은 1차 행에서 종목명/심볼, 수량, 평균가, 현재가, 익절, 손절, 추적 회수, 기준, 고점, 총금액, 미실현 PnL을 확인할 수 있다

#### Scenario: canonical 근거가 없는 저장 기준
- **WHEN** seed-only raw exit state 또는 stale·generation mismatch 상태를 렌더하면
- **THEN** 1차 actionable 라인은 `—`와 상태/사유를 표시하고 저장된 raw 가격은 상세 증거로만 표시한다

#### Scenario: 보조 진단 접기
- **WHEN** 관리 포지션 행을 처음 렌더하면
- **THEN** 자격 provenance와 journal/snapshot 식별자는 native 상세 영역 안에 있고 관리·stale·미측정 verdict는 상세 영역 밖에 있다

#### Scenario: 주문 화면의 작은 viewport
- **WHEN** 주문이 있는 화면을 375 CSS pixel 폭으로 렌더하면
- **THEN** `documentElement.scrollWidth`가 viewport 너비를 넘지 않고 각 주문의 핵심 필드 라벨과 값을 읽을 수 있다

#### Scenario: 보유 화면의 작은 viewport
- **WHEN** KR 또는 US 보유가 있는 대시보드나 포지션 화면을 375 CSS pixel 폭으로 렌더하면
- **THEN** 1차 보유 필드는 라벨과 함께 card flow로 읽히고 문서 전체의 수평 overflow가 생기지 않으며 상세 summary는 최소 44 CSS pixel 높이다

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
콘솔은 편입 설정 화면을 제공해야 한다(SHALL): 상단 navigation은 **"설정"** 메뉴를 표시하고, 편입 설정은 그 안의 **상시(standing)** 하위 화면에서 열려야 한다(SHALL — 편입 규칙은 비가역이고 주 1회 미만으로 만지는 설정이므로 당일 조정과 같은 화면에 두지 않는다). 기존 `/settings#adoption` 경로는 그 하위 화면의 같은 섹션으로 리다이렉트해야 한다(SHALL — 이미 발행된 링크와 북마크가 바깥에 존재한다). 화면 제목과 첫 설명은 수동 매수한 외부 종목이 편입 완료 후 기존 공통 익절·보호선·손익 극대화 정책의 관리 대상이 된다는 결과와, 저장 자체가 편입·주문을 실행하지 않고 엔진 대사 루프가 다음 기동부터 수행한다는 경계를 명시해야 한다(SHALL).

화면은 `adoption.enabled`·`default_stop_pct`·`exclude_symbols`·`include_symbols`를 표시·편집한다(SHALL). 합성 손절폭은 기존 CSP 아래에서도 현재값과 편집값이 일치하도록 **명시적 백분율 숫자 입력**으로 표시한다(SHALL — 이전 마우스 슬라이더/타이핑 금지 결정은 CSP가 inline 갱신을 차단해 현재값을 거짓 표시하므로 이 요구가 대체한다). 입력 필드명은 `default_stop_percent`이고, 허용값은 유한한 2%..20%의 0.5 percentage-point 단위다(SHALL). 서버는 브라우저 검증을 신뢰하지 않고 empty·비수치·NaN·무한대·범위 밖·단위 밖·과거 fraction 형식 제출을 저장 전에 거부하며(SHALL), 허용 percentage tick만 100으로 나눠 기존 fractional `default_stop_pct`로 저장한다(SHALL). 렌더링은 불필요한 0이나 부동소수 artifact가 없는 결정적 10진 표기여야 한다(SHALL). 기존 파일에 엔진은 허용하지만 콘솔 0.5% grid에는 맞지 않는 값이 있으면 정확한 현재값과 보정 필요를 표시하고, 저장 시 침묵 반올림하지 않고 거부한다(SHALL).

목록 직접 기입은 고급 접힘 안에만 두고, 거부된 블록은 파일 원문 값과 거부 사유를 함께 표시한다(SHALL). 저장은 다른 편입 필드를 보존하고 audit을 남기며, 반영 시점("다음 엔진 기동부터 반영"), 편입 비가역 귀결, 지정의 상시성, 현재 엔진 실행 여부를 안내한다(SHALL). 저장은 엔진을 재시작하거나 기존 관리 포지션을 변경하지 않는다(SHALL NOT). 이 백분율 control에는 client-side script나 CSP 예외가 없어야 한다(SHALL NOT). 설정 화면의 어떤 폼도 inline event handler를 가져서는 안 된다(SHALL NOT — 배포 CSP에 `script-src`가 없어 `default-src 'none'`이 적용되므로 인라인 handler는 실행되지 않는다. 실행되지 않는 확인 대화는 확인이 아니라 확인의 외양이며, 그 자리는 적용 전 미리보기가 대신한다). automation gate 편집 표면은 존재하지 않는다(SHALL NOT). 포지션 화면의 관리 외 보유 행은 기존 종목 편입 지정/해제 행위를 유지하고, 그 행위는 include 목록만 바꾸며 편입을 직접 수행하지 않는다(SHALL NOT).

#### Scenario: 상단 메뉴에서 편입 설정 열기
- **WHEN** 운영자가 어느 콘솔 화면에서든 상단의 "설정"을 선택하고 상시 하위 화면을 열면
- **THEN** 자동 편입 ON/OFF·합성 보호폭·제외·지정 설정과 기존 정책 적용 설명이 표시된다

#### Scenario: 기존 편입 설정 링크
- **WHEN** 기존 `/settings#adoption` 링크로 접속하면
- **THEN** 상시 하위 화면의 같은 섹션으로 리다이렉트되고 404가 되지 않는다

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

#### Scenario: 설정 화면의 인라인 handler 부재
- **WHEN** 설정 화면 전체의 렌더 결과를 검사하면
- **THEN** `on[a-z]+=` 형태의 inline event handler가 하나도 없다

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

### Requirement: positions는 현재와 다음 exit 동작을 함께 표시한다
콘솔은 `/positions`에서 entry, initial stop, current protection, next target, next protection, rung, projected exit quantity와 evaluated-at을 권위 snapshot 그대로 표시해야 한다 (SHALL).

#### Scenario: 관리 중 포지션
- **WHEN** 완전한 exit snapshot이 있는 포지션을 조회한다
- **THEN** 운영자는 현재 보호와 다음 가격 도달 시 기준선·수량 동작을 한 화면에서 읽을 수 있다

#### Scenario: 1주 포지션
- **WHEN** 보유 수량이 1주이고 다음 rung이 partial이다
- **THEN** 화면은 `중간 매도 없음 · 보호선 승격`과 최종/손절 시 1주 전량을 표시한다

#### Scenario: stale snapshot
- **WHEN** snapshot이 stale 또는 일부 unknown이다
- **THEN** 값은 0이 아니라 `—`와 stale/unknown 사유로 표시된다

### Requirement: orders는 exit 주문 근거를 결정적으로 연결한다
콘솔은 broker order의 명시적 mutation-attempt intent lineage로 exit event의 decision ID와 기준선 snapshot을 연결하고 trigger line, observation, policy, rung과 reason을 표시해야 한다 (SHALL).

#### Scenario: 연결된 손절 주문
- **WHEN** broker order의 mutation attempt intent가 protection breach exit event를 참조한다
- **THEN** 주문 상세는 당시 현재가와 보호선, 전량 사유를 표시한다

#### Scenario: 연결 식별자 부재
- **WHEN** broker order에 결정적 snapshot 링크가 없다
- **THEN** 화면은 근거 미연결로 표시하고 symbol/time으로 추정하지 않는다

### Requirement: 거래 화면과 설정 화면의 역할은 분리된다
`/positions`와 `/orders`는 exit 상태를 읽기 전용으로 설명해야 하며 (SHALL), 정책 설정 control을 복제해서는 안 된다 (MUST NOT). 설정이 필요한 문맥에는 a050의 canonical category deep link를 제공해야 한다 (SHALL).

#### Scenario: positions에서 종목 정책 확인
- **WHEN** 운영자가 포지션의 정책을 변경하려 한다
- **THEN** 화면은 `/optimization?category=position-management` 링크를 제공하고 현재 표 안에서 즉시 변경하지 않는다

#### Scenario: orders에서 보호주문 확인
- **WHEN** 운영자가 exit 주문의 보호 설정을 확인하려 한다
- **THEN** 화면은 `/optimization?category=exit-protection` 링크를 제공한다

#### Scenario: 입력 없는 거래 화면
- **WHEN** 운영자가 `/positions` 또는 `/orders`를 연다
- **THEN** 화면에는 form/input/textarea/select/button/contenteditable이 없고 해당 경로의 POST는 405로 거부된다

### Requirement: 종목별 정책 쓰기는 명확한 범위와 결과를 표시한다
콘솔은 a050의 `position-management` 카테고리에서 override/release/re-adopt 전 대상 계좌·시장·종목·현재 정책·기본 정책·실효 정책·다음 정책·적용 generation과 적용 시점/restart 필요 여부를 표시해야 한다 (SHALL).

#### Scenario: 정책 승인 화면
- **WHEN** 운영자가 종목 정책 변경을 연다
- **THEN** 기존 포지션 재해석 여부와 LIVE toggle이 바뀌지 않음을 승인 전에 표시한다

#### Scenario: version conflict
- **WHEN** write가 412로 거부된다
- **THEN** 현재 version을 다시 불러오도록 안내하고 자동 재시도하지 않는다

### Requirement: 종목 관리 설정은 기본값과 동작을 가까이 설명한다
`position-management`는 `종목별 정책`과 `외부 매수 자동편입`을 구분하고 모든 설정에 label, 쉬운 설명, 기본값, desired/effective 값, 단위·범위와 적용 시점을 표시해야 한다 (SHALL).

#### Scenario: 초기 자동편입 설정
- **WHEN** 저장 설정이 없는 새 설치가 화면을 연다
- **THEN** 자동편입 OFF, 합성 손절폭 5%, 허용범위 2~20%와 0.5% 조정 단위, 빈 include/exclude 및 exclude 우선을 표시한다

#### Scenario: 종목별 정책 미지정
- **WHEN** 한 종목에 override가 없다
- **THEN** 기본값과 실효값은 `공통 정책 상속`으로 표시되고 특정 preset을 복제 저장하지 않는다

#### Scenario: 1주 외부 보유
- **WHEN** 자동편입 preview 대상 보유가 1주다
- **THEN** `중간 익절 없음 · 보호선 승격 · 최종/손절 시 1주 전량` 설명을 표시한다

#### Scenario: release 요청
- **WHEN** 운영자가 자동관리 해제를 선택한다
- **THEN** 일반 설정 저장과 분리된 danger confirmation에서 보호 공백과 active exit 충돌 여부를 설명하고 3초 대기·명시 checkbox/button을 제공하되 문구 입력을 요구하지 않는다

#### Scenario: 입력 없는 정책 제어
- **WHEN** 운영자가 종목별 정책 또는 자동편입을 변경한다
- **THEN** server preset·OFF/ON·0.5% stop option·현재 보유 행 action·server reason만 선택할 수 있고 자유 text/number/symbol/reason 입력이나 typed confirmation은 없다

### Requirement: 브로커 보호 설정과 상태는 한 카테고리에서 설명된다
콘솔은 a050의 `exit-protection` 카테고리에서 capability, activation, current effective trigger, protected quantity, broker identifier, updated-at과 reconcile reason을 표시해야 한다 (SHALL). 각 항목은 label, 쉬운 설명, 기본값, desired/effective 값, 적용 시점과 provenance를 가져야 한다 (SHALL).

#### Scenario: attestation 미완료
- **WHEN** 현재 시장의 조건주문 capability가 확인되지 않았다
- **THEN** 화면은 `OFF / 지원 확인 전 사용 불가`를 기본·실효 상태로 표시하고 주문 유형이나 trigger 기본값을 임의 생성하지 않는다

#### Scenario: capability 확인 완료
- **WHEN** SINGLE+MARKET capability가 attested됐다
- **THEN** 지원 유형과 근거를 표시하되 activation 기본값은 OFF이고 운영자 승인 전 자동 활성화하지 않는다

#### Scenario: 활성 보호주문
- **WHEN** protection saga가 ACTIVE다
- **THEN** effective trigger, 수량, broker ID, 마지막 확인 시각과 다음 reconcile 설명을 한 section에서 읽을 수 있다

### Requirement: 보호 약화는 강화와 구분해 확인한다
콘솔은 trigger 하향, 보호 해제 또는 수량 감소처럼 보호를 약화하는 변경을 분류하고 before/after, 영향 포지션·수량, 보호 공백 가능성과 적용 시점을 표시한 뒤 3초 지연 확인을 요구해야 한다 (SHALL).

#### Scenario: trigger 하향 요청
- **WHEN** 운영자가 ACTIVE trigger보다 낮은 값을 preview한다
- **THEN** 위험 확인을 표시하더라도 domain contract에 따라 apply는 거부되고 현재 보호를 유지한다

#### Scenario: 보호 강화
- **WHEN** 새 trigger가 더 안전한 방향이고 모든 capability가 유효하다
- **THEN** 변경 범위와 적용 시점을 표시하되 약화 전용 경고 문구를 사용하지 않는다

### Requirement: 후보 필터는 판정 의미와 근거를 함께 표시한다
콘솔은 a050의 `candidate-filters` 카테고리에서 `seen_late`, `extended`, `near_high`를 시장·세션별로 구분하고 각 필터의 쉬운 설명, 판정 방향, 단위, 기본 상태, desired/effective 값, 범위, 표본과 evidence provenance를 표시해야 한다 (SHALL).

#### Scenario: 최초 미승인 상태
- **WHEN** 승인된 threshold set이 없다
- **THEN** 화면은 `미승인 · passed 구조적 0 · verdict 비활성`을 표시하고 숫자 0을 기본 threshold처럼 보여주지 않는다

#### Scenario: evidence 불완전
- **WHEN** sample count 또는 evidence digest가 누락됐다
- **THEN** 관련 입력은 read-only이고 누락 항목과 승인에 필요한 다음 행동을 설명한다

#### Scenario: 승인된 시장별 값
- **WHEN** KR regular-session threshold set이 승인됐다
- **THEN** 각 metric의 값·단위·방향·표본·version을 표시하고 US에는 같은 값을 기본값으로 복제하지 않는다

### Requirement: threshold 승인은 변경 영향 preview를 선행한다
콘솔은 threshold apply 전에 before/after, 대상 시장·세션, 예상 verdict count 변화, evidence version과 적용 시점을 preview해야 한다 (SHALL).

#### Scenario: 승인 preview
- **WHEN** 운영자가 완전한 threshold set을 승인하려 한다
- **THEN** 후보 판정만 활성화되고 주문·RiskIntent·LIVE 상태는 바뀌지 않음을 확인 전에 설명한다

### Requirement: 전략 설정과 실행 권한은 분리해 표시된다
콘솔은 a050의 `strategy-runtime` 카테고리에서 전략 파라미터, lane desired/effective 상태, 자동 기동과 LIVE 주문 승인을 별도 section과 별도 action으로 제공해야 하며 (SHALL), 이를 한 번에 활성화하는 control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: 새 설치
- **WHEN** lane 설정이 처음 표시된다
- **THEN** lane desired와 auto-start 기본값은 OFF이고 LIVE 주문은 별도 미승인 상태다

#### Scenario: 전략 설정 저장
- **WHEN** 운영자가 lane 파라미터만 저장한다
- **THEN** lane, auto-start와 LIVE approval 상태는 바뀌지 않고 적용 시점과 restart 필요 여부를 설명한다

#### Scenario: lane ON 요청
- **WHEN** 운영자가 lane desired state를 ON으로 바꾼다
- **THEN** Guardian, protection, reconciliation과 LIVE approval 결과에 따라 effective 상태와 첫 refusal reason을 별도로 표시한다

### Requirement: 확정되지 않은 lane 값은 기본값으로 꾸미지 않는다
모든 lane field는 label, 쉬운 설명, default, desired/effective, 단위·범위, source/version과 적용 시점을 가져야 한다 (SHALL). proposal-freeze가 끝나지 않은 field는 `미구성 / 읽기 전용`이어야 한다 (MUST).

#### Scenario: 첫 lane 미동결
- **WHEN** StockOS source policy·시장·상수가 아직 동결되지 않았다
- **THEN** UI는 숫자 0이나 추정값을 표시하지 않고 미구성 사유와 선행 문서를 안내한다

#### Scenario: a047 dormant 상태 card
- **WHEN** a045 protection, a046 provenance, a048 scheduler/calendar 또는 source manifest가 미완성이다
- **THEN** UI는 blocker별 desired/effective를 읽기 전용으로 보여주고 값 입력·일괄 활성화·LIVE action을 제공하지 않는다

### Requirement: 시장 스케줄은 desired와 effective를 구분해 설명한다
콘솔은 a050의 `strategy-runtime > 시장·일정` section에서 scheduler와 auto-start의 desired/effective 상태, 시장·세션 범위, calendar version/updated-at, 다음 전환 시각과 typed decision reason을 표시해야 한다 (SHALL).
시장·세션 범위와 운영 reason은 server-defined option으로만 선택해야 하며 (SHALL), 임의 문자열·시간·휴장일 입력 control을 제공해서는 안 된다 (MUST NOT).
시장 선택지는 `none`, `KR`, `US`만 제공해야 하며 (SHALL), 정확한 per-market calendar/activation binding이 없는 결합 시장을 광고해서는 안 된다 (MUST NOT).

#### Scenario: 새 설치
- **WHEN** scheduler 저장값이 없다
- **THEN** scheduler OFF, auto-start OFF, 선택 시장 없음, 정규장만 허용을 기본값과 쉬운 설명으로 표시한다

#### Scenario: 장 닫힘
- **WHEN** desired는 ON이지만 시장이 휴장이다
- **THEN** effective는 WAIT_MARKET이고 다음 세션과 함께 exit/reconcile은 계속됨을 설명한다

#### Scenario: API 예산 대기
- **WHEN** scheduler decision이 BUDGET_DEFERRED다
- **THEN** 사용자 OFF와 구분된 대기 사유 및 safety budget을 침범하지 않는다는 설명을 표시한다

### Requirement: exchange calendar는 운영 근거로 읽기 전용이다
calendar version, source와 updated-at은 표시해야 하지만 (SHALL), 최초 범위에서 사용자가 임의 휴장일을 입력하는 control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: stale calendar
- **WHEN** calendar freshness gate가 실패한다
- **THEN** entry effective 상태는 fail-closed이고 stale reason과 갱신 방법을 표시한다

### Requirement: 성과와 이력은 읽기 전용 카테고리에서 설명된다
콘솔은 a050의 `performance-history` 카테고리에서 lane/policy 성과와 설정 변경 이력을 읽기 전용으로 제공하고 각 metric의 쉬운 정의, 단위, 기간, 표본 수와 provenance를 표시해야 한다 (SHALL).

#### Scenario: 최초 조회
- **WHEN** 운영자가 별도 filter 저장 없이 카테고리를 연다
- **THEN** 최근 30일, 전체 시장, 전체 lane, 완전한 lineage만 포함하는 조회 기본값을 표시한다

#### Scenario: 누락 결과
- **WHEN** 한 거래의 결정적 lineage 또는 markout이 없다
- **THEN** 0으로 표시하지 않고 각각 `link_missing` 또는 `not_measured` 설명을 표시한다

#### Scenario: 표본 부족
- **WHEN** 선택 구간의 표본이 승인 최소치 미만이다
- **THEN** 관측값과 표본은 보여주되 `insufficient_sample · 추천 근거로 사용 불가`를 표시한다

### Requirement: 성과 화면은 거래나 설정 권한을 갖지 않는다
`performance-history`는 주문, lane toggle, LIVE approval 또는 설정 apply control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: 성과 비교
- **WHEN** 운영자가 lane 두 개를 비교한다
- **THEN** 비교 결과만 표시하고 더 좋은 lane를 자동 활성화하거나 저장하지 않는다

### Requirement: optimization 화면은 근거와 변경 수명주기를 표시한다
콘솔은 parameter registry, current/effective version, lane performance evidence, candidate diff, apply history와 rollback 동작을 표시해야 한다 (SHALL).

#### Scenario: 추천 불가
- **WHEN** 필수 성과 근거가 부족하다
- **THEN** 화면은 구체적인 누락 사유를 표시하고 apply control을 활성화하지 않는다

#### Scenario: 적용 승인
- **WHEN** 운영자가 preview와 version을 확인하고 apply한다
- **THEN** 결과 version과 restart 필요 여부를 표시하며 LIVE toggle은 별도 상태로 남는다

### Requirement: optimization 설정은 카테고리와 설명으로 탐색된다
콘솔은 `overview`, `exit-protection`, `position-management`, `candidate-filters`, `strategy-runtime`, `performance-history` 여섯 category를 고정 순서로 제공해야 하며 (SHALL), 모든 설정에 한국어 설명, parameter key, 단위, registry 기본값, desired/effective 값, 범위와 적용 시점을 표시해야 한다 (SHALL).

#### Scenario: 기본값과 현재값 구분
- **WHEN** 운영자가 설정 가능한 field를 연다
- **THEN** placeholder가 아닌 별도 label로 기본값·현재 desired·현재 effective와 적용 시점을 구분해 표시한다

#### Scenario: 모바일 category 탐색
- **WHEN** 360px viewport에서 optimization을 연다
- **THEN** 동일한 여섯 category와 deep link를 사용할 수 있고 페이지 전체의 수평 overflow 없이 설정과 설명을 읽을 수 있다

#### Scenario: category-scoped save
- **WHEN** 두 category에 미저장 draft가 있고 한 category에서 저장한다
- **THEN** 해당 category changed subset만 preview/apply하며 다른 draft와 LIVE 상태를 변경하지 않는다

### Requirement: 위험 설정은 before/after 확인을 요구한다
손절폭 확대, 보호 약화, lane 또는 LIVE 권한 변화는 일반 설정 저장과 구분해야 하며 (SHALL), before/after·적용 대상·restart 여부를 표시한 명시적 확인 없이는 적용해서는 안 된다 (MUST NOT).

콘솔은 StockOS lane-console의 화면 단위 navigation·partial save·effective mismatch 패턴을 따라야
하며 (SHALL), 운영자에게 자유 텍스트·숫자·symbol·확인 문구 입력을 요구해서는 안 된다 (MUST NOT).
모든 변경은 preset/radio/select/chip/toggle/discrete step과 server-defined reason code로 수행해야 한다
(SHALL).

#### Scenario: LIVE 보호 약화
- **WHEN** LIVE 상태에서 draft가 허용 손실폭이나 유예를 확대한다
- **THEN** 3초 대기, 확인 checkbox와 별도 위험 승인 전에는 저장 control이 활성화되지 않는다

#### Scenario: 입력 없는 위험 확인
- **WHEN** 위험 확대 candidate를 확인한다
- **THEN** 3초 대기, 영향 범위 확인 checkbox와 승인 button을 제공하되 typed phrase나 자유 reason 입력은 요구하지 않는다

#### Scenario: 자유 입력 control 회귀
- **WHEN** optimization HTML과 handler contract를 검사한다
- **THEN** text, textarea, number, contenteditable control이 0개이고 제출값은 registry option ID에 한정된다

### Requirement: 콘솔 공통 상태 표시줄

콘솔의 모든 화면은 같은 자리에 같은 상태 표시줄을 렌더해야 한다(SHALL): 엔진 상태, 시장 세션 advisory, 이 화면이 보여주는 데이터의 시각, 지금 걸려 있는 자동 재로드 주기. 표시줄은 화면마다 다른 문구로 같은 사실을 말해서는 안 된다(SHALL NOT — 지금 `TakenAt`·`AgeSeconds`·`NowText`가 설명 문단 안에 흩어져 화면마다 다른 문장으로 나타난다).

엔진 칸은 세 상태를 구분해야 한다(SHALL): 마커 경로가 주입되지 않은 **미배선**, 마커가 신선도 창 밖인 **정지**, 창 안인 **실행 중**. 미배선을 정지로 표시해서는 안 된다(SHALL NOT). 실행 중인 엔진이 설치된 바이너리와 다른 실행 파일에서 기동됐으면 그 사실을 함께 표시해야 한다(SHALL).

시장 세션 칸은 advisory임을 표시해야 한다(SHALL — 세션 판정은 요일과 시각만 읽으며 휴장일을 알지 못한다). 콘솔은 이 칸을 어떤 행위의 판단 근거로 삼아서는 안 된다(SHALL NOT).

데이터 시각 칸은 세 상태 중 하나여야 한다(SHALL): ① **캐시 기반** 화면은 기록된 캐시 시각과 경과 시간과 톤을 표시한다, ② **요청 시점 읽기** 화면은 렌더 시각과 그 값이 요청 시 읽은 것임을 표시하고 톤을 붙이지 않는다, ③ **읽지 못한** 경우 사유를 표시하고 마지막 성공 시각은 그 기록이 있을 때만 표시한다. 기록되지 않은 시각을 표시줄이 만들어내서는 안 된다(SHALL NOT). 엔진 마커의 갱신 시각을 화면 데이터의 시각으로 표시해서는 안 된다(SHALL NOT — 둘은 다른 사실이다).

톤 임계는 그 화면 자신의 캐시 TTL에서 유도해야 하며 새 상수를 도입해서는 안 된다(SHALL NOT): 경과가 TTL 미만이면 정상, TTL 이상 2×TTL 미만이면 주의, 2×TTL 이상이거나 읽기가 실패했으면 경고.

**갱신하지 않는 것이 정상인 상태에는 톤을 붙여서는 안 된다**(SHALL NOT): 갱신 보류 사유가 알려져 있으면 — 검증 run 진행 중이라 `rate budget 보호`가 갱신 보류를 의무화한 경우, 또는 스캔 주기가 돌지 않는 것이 알려진 경우 — 표시줄은 경과에 톤을 붙이는 대신 **그 사유**를 표시한다(SHALL). 규정대로 동작하는 시스템에 경고를 붙이면 경고가 늘 켜져 있게 되고, 늘 켜진 경고는 아무도 보지 않는다.

자동 재로드 주기 정책은 이 요구사항으로 바뀌지 않는다(SHALL — 화면별 주기와 "검증 run이 작업 중일 때만" 조건은 그대로다). 표시줄은 지금 걸려 있는 주기를 말할 뿐이며, 재로드가 걸려 있지 않은 화면에서는 걸려 있지 않다고 말한다(SHALL).

승인을 기다리는 검증 run이 있으면 표시줄은 그 사실과 그 화면으로 가는 직접 링크를 표시해야 한다(SHALL — 승인 창은 짧고 소진 사고 기록이 있다. 표시줄은 모든 화면에 있으므로 이 표시가 승인 창의 발견성을 화면 위치와 무관하게 만든다). 이 표시는 승인 행위를 표시줄에 두는 것이 아니다(SHALL NOT — 표시줄에는 폼이 없고, 승인은 그 화면에서 한다).

상태 표시줄은 브로커 호출을 추가해서는 안 된다(SHALL NOT — 표시줄이 읽는 것은 엔진 마커 파일과 설치 경로 stat과 화면이 이미 계산한 값이다). 상태 표시줄은 콘솔의 상태변경 행위를 늘려서는 안 된다(SHALL NOT — 표시줄에는 폼이 없다).

#### Scenario: 모든 화면의 같은 표시줄
- **WHEN** 콘솔의 각 화면을 차례로 열면
- **THEN** 같은 자리에 엔진 상태·시장 세션·데이터 시각·재로드 주기가 같은 문구 형식으로 표시된다

#### Scenario: 엔진 미배선과 정지의 구분
- **WHEN** 엔진 마커 경로가 주입되지 않은 빌드에서 아무 화면을 열면
- **THEN** 엔진 칸은 "미배선"을 말하고 "정지"라고 말하지 않는다

#### Scenario: 캐시 기반 화면의 데이터 시각
- **WHEN** 브로커 캐시가 채워진 상태로 포지션 화면을 열면
- **THEN** 데이터 시각 칸에 기록된 캐시 시각과 경과 시간이 표시되고, 경과가 그 화면의 TTL을 넘었으면 주의 톤이 붙는다

#### Scenario: 요청 시점 읽기 화면의 데이터 시각
- **WHEN** 캐시를 갖지 않는 화면(설정·거래 이력 등)을 열면
- **THEN** 데이터 시각 칸은 렌더 시각과 요청 시 읽었다는 사실을 표시하고, 존재하지 않는 캐시 시각을 만들지 않는다

#### Scenario: 읽기 실패
- **WHEN** 브로커 조회가 실패한 상태로 화면을 열면
- **THEN** 데이터 시각 칸은 실패 사유를 표시하고, 마지막 성공 시각은 그 기록이 있을 때만 함께 표시된다

#### Scenario: 알려진 갱신 보류
- **WHEN** 검증 run이 진행 중이라 브로커 갱신이 보류된 상태로 캐시 기반 화면을 열면
- **THEN** 데이터 시각 칸은 경과에 경고 톤을 붙이는 대신 갱신 보류 사유를 표시한다

#### Scenario: 승인 대기 중인 검증 run
- **WHEN** 승인을 기다리는 검증 run이 있는 동안 아무 화면을 열면
- **THEN** 상태 표시줄이 승인 대기 사실과 그 화면으로 가는 직접 링크를 표시하고, 표시줄 자체에는 승인 폼이 없다

#### Scenario: 표시줄이 브로커 비용을 늘리지 않는다
- **WHEN** 상태 표시줄이 있는 화면을 TTL 안에서 여러 번 재로드하면
- **THEN** 브로커 호출은 TTL당 holdings 1콜 상한을 넘지 않는다

#### Scenario: 화면별 재로드 주기 보존
- **WHEN** 공용 표시줄을 도입한 뒤 각 화면의 재로드 지시를 검사하면
- **THEN** 검증 화면은 run이 작업 중일 때만 짧은 주기를 쓰고, 캐시 기반 화면은 각자의 캐시 TTL을 그대로 쓴다

### Requirement: 화면 이름·경로·제목은 한 화면을 가리킨다

콘솔의 각 화면에서 경로, navigation 라벨, 문서 제목(`<h1>`)은 같은 화면을 같은 이름으로 가리켜야 한다(SHALL — 현재 루트 경로는 navigation에서 "검증 콘솔"이고 제목은 "대시보드"이며, `/dashboard`의 제목은 "개요"다. 한 화면에 이름이 셋이면 운영자는 매번 어느 화면에 있는지 다시 판단한다). 이 일치는 **렌더 결과 비교**로 고정해야 한다(SHALL — 각 화면을 렌더해 `aria-current`가 붙은 navigation 항목의 텍스트와 `<h1>` 텍스트를 대조한다. 세 값은 Go 문자열 템플릿 안의 HTML과 핸들러의 문자열 필드에 흩어져 있어 정적 파싱 대상이 아니며, 렌더 비교가 더 쉽고 더 강하다).

루트 경로는 계좌 개요를 답해야 한다(SHALL). 검증 콘솔은 자기 경로를 가져야 하며 그 화면의 컨트롤(검증 시작·승인·중단, 엔진 시작·정지, 자기 재시작·soak 재시작)은 이동으로 줄어들어서는 안 된다(SHALL NOT).

루트 경로는 404를 답해서는 안 되고 개요로 리다이렉트해야 한다(SHALL — 북마크와 이미 발행된 링크가 바깥에 존재한다).

재시작 핸드오프 토큰은 이 이동 후에도 **정확히 한 번 소비되고 브라우저는 렌더된 화면에 착지해야 한다**(SHALL). 이 요구는 리다이렉트가 `handoff` 파라미터를 실어 나르는 것으로 충족되어서는 안 된다(SHALL NOT — 토큰은 세션 미들웨어가 핸들러보다 먼저 소비하고 그 파라미터를 의도적으로 제거한 뒤 리다이렉트한다. 소비된 토큰을 주소창에 남기는 것은 이 설계에 반한다). 재시작·soak 재기동 결과 안내는 **그 컨트롤이 있는 화면**으로 돌아가야 하며(SHALL — 운영자가 누른 버튼의 결과를 다른 화면에서 읽게 해서는 안 된다), 그 안내 쿼리는 보존해야 한다(SHALL).

원격 로그인 후 리다이렉트 경로는 존재하는 화면을 가리켜야 한다(SHALL). 세션 쿠키의 경로 스코프는 루트로 유지해야 한다(SHALL NOT — 쿠키 스코프를 특정 화면으로 좁히면 다른 화면에서 세션이 사라진다. 쿠키 경로는 라우트가 아니다).

이 요구사항은 콘솔의 상태변경 행위 목록을 늘리지 않는다(SHALL — 추가되는 라우트는 리다이렉트와 화면 이동뿐이고 둘 다 GET이며 폼이 없다).

#### Scenario: 루트 경로
- **WHEN** 운영자가 루트 경로를 열면
- **THEN** 계좌 개요로 리다이렉트되고, 개요 화면의 navigation 라벨과 `<h1>`이 같은 이름을 쓴다

#### Scenario: 재시작 핸드오프 왕복
- **WHEN** `?handoff=<token>`이 붙은 루트 URL로 접속하면
- **THEN** 토큰이 정확히 한 번 소비되고 브라우저는 렌더된 화면에 착지하며, 소비된 토큰이 주소창에 남지 않는다

#### Scenario: 소비된 토큰의 재사용
- **WHEN** 같은 핸드오프 토큰으로 다시 접속하면
- **THEN** 이동 전과 동일하게 거부된다

#### Scenario: 재시작 안내의 착지
- **WHEN** 자기 재시작 또는 soak 재기동을 수행하면
- **THEN** 그 컨트롤이 있는 화면으로 돌아가고 결과 안내가 함께 표시된다

#### Scenario: 검증 콘솔의 컨트롤 보존
- **WHEN** 검증 콘솔을 새 경로에서 열면
- **THEN** 검증 시작·승인·중단과 엔진 시작·정지와 재시작 컨트롤이 이동 전과 같이 존재한다

#### Scenario: 세션 쿠키 스코프
- **WHEN** 루트 경로가 옮겨진 뒤 개요에서 포지션 화면으로 이동하면
- **THEN** 세션이 유지된다 — 쿠키 경로 스코프는 좁아지지 않았다

#### Scenario: 이름 일치 정적 검사
- **WHEN** 어떤 화면의 `<h1>`이 그 화면의 navigation 라벨과 다른 이름으로 바뀌면
- **THEN** 정적 검사가 실패한다

### Requirement: 포지션·주문 외 화면의 반응형 표시

`거래 화면 핵심 정보 계층과 반응형 표시`가 포지션·주문 화면에 요구하는 반응형 계약은 콘솔의 **나머지 모든 화면**에도 적용되어야 한다(SHALL): 개요, 검증 콘솔, 검증, 발굴 신호, 거래 이력, 설정, 최적화, 성과 이력, 포지션 정책, 리포트.

이 계약의 준수는 **렌더 결과에서 판정 가능한 조건**으로 고정해야 한다(SHALL): ① viewport meta가 존재한다, ② 모든 표가 반응형 클래스를 달거나 자체 가로 스크롤 영역 안에 있다, ③ viewport보다 넓은 고정 px 폭이 없다, ④ 좁은 viewport 미디어 쿼리가 적용된다. 실제 레이아웃 측정(`documentElement.scrollWidth`)을 자동 검사의 조건으로 삼아서는 안 된다(SHALL NOT — 이 저장소에는 그 측정을 수행하는 테스트 하니스가 없고, 없는 하니스를 전제한 요구는 검증되지 않은 채 통과한다. 브라우저 실측은 자동 검사가 아니라 사람이 한 번 수행해 증거로 남긴다).

반응형 클래스도 스크롤 래퍼도 갖지 않은 표가 있어서는 안 된다(SHALL NOT — 현재 그런 표가 다수이고, 그 표들은 카드 변환도 스크롤 격리도 받지 못한다). 표가 어느 방식을 썼는지는 렌더 결과에서 판별 가능해야 한다(SHALL).

문서의 제목 단계는 크기로 구분 가능해야 한다(SHALL — 본문과 섹션 제목이 같은 크기면 화면이 무엇이 중요한지 말하지 않는다).

같은 의미의 상태 표시에 서로 다른 클래스 이름을 두어서는 안 된다(SHALL NOT). 통합의 방향은 **실제 사용이 많은 쪽을 남기는 것**이어야 한다(SHALL — 적게 쓰이는 이름을 남기면 많이 쓰이는 쪽의 모든 사용처와 그것을 검사하는 기존 테스트를 함께 고치는 작업이 되고, 그 작업은 이 요구사항이 얻으려는 것과 무관하다).

이 반응형 표시는 JavaScript나 외부 asset을 요구해서는 안 된다(SHALL NOT). 색만으로 의미를 전달해서는 안 되며(SHALL NOT), 배경 대비 기준을 낮추는 색 교체를 해서는 안 된다(SHALL NOT — 현재 값이 대비 기준을 통과하고 있으면 통과하지 못하는 값으로 바꾸지 않는다).

#### Scenario: 클래스 없는 표
- **WHEN** 어떤 화면의 표가 반응형 클래스도 스크롤 래퍼도 갖지 않은 채 렌더되면
- **THEN** 검사가 실패한다

#### Scenario: 좁은 viewport 조건
- **WHEN** 콘솔의 각 화면 렌더 결과를 검사하면
- **THEN** viewport meta가 있고, 좁은 viewport 미디어 쿼리가 적용되며, viewport보다 넓은 고정 px 폭이 없다

#### Scenario: 제목 단계
- **WHEN** 한 화면의 문서 제목·섹션 제목·본문을 비교하면
- **THEN** 각 단계의 크기가 서로 구분된다

#### Scenario: 상태 표시 클래스 통합
- **WHEN** 통합 후 스타일시트와 모든 화면의 렌더 결과를 검사하면
- **THEN** 같은 의미의 상태 표시 클래스가 하나만 남아 있고, 남은 이름은 통합 전 사용이 더 많던 쪽이며, 그 이름을 검사하던 기존 테스트가 계속 통과한다

### Requirement: 개요는 상단 요약에서 답한다

개요 화면은 "지금 무엇이 잘못됐는가"를 세로 스크롤 없이 답해야 한다(SHALL): 상세 표들 앞에 요약 영역을 두고, 각 칸은 한 가지 사실과 그 사실의 상세를 소유한 화면으로 가는 링크를 갖는다(SHALL). 요약 칸은 상세 화면의 값을 재계산해서는 안 되며(SHALL NOT) 상세 화면이 이미 답한 값을 그대로 옮긴다.

값을 얻지 못한 칸은 0이나 빈 값이 아니라 미측정과 사유를 표시해야 한다(SHALL — 0은 "측정했고 없었다"는 뜻이고 미측정은 다른 사실이다). 요약 영역은 브로커 호출을 추가해서는 안 된다(SHALL NOT).

기존 상세 섹션은 요약 도입으로 제거되지 않는다(SHALL — 요약은 상세 앞에 서는 것이지 상세를 대체하는 것이 아니다).

#### Scenario: 개요 첫 화면
- **WHEN** 개요를 열면
- **THEN** 상세 표 앞의 요약 영역에서 엔진·계좌·오늘·미체결·안전에 해당하는 사실을 읽을 수 있고 각 칸이 해당 상세 화면으로 링크한다

#### Scenario: 미측정 칸
- **WHEN** 요약 칸의 원본 값을 읽지 못하면
- **THEN** 그 칸은 0이 아니라 미측정과 사유를 표시한다

#### Scenario: 요약이 비용을 늘리지 않는다
- **WHEN** 요약 영역이 있는 개요를 열면
- **THEN** 브로커 호출 수는 요약 도입 전과 같다

### Requirement: 포지션 관리는 실제 adoption desired/effective를 구분한다
`/position-management`는 registry 기본값, config file의 desired adoption 값, running engine이 시작 때 로드한 effective 값을 별도 label로 표시해야 한다 (SHALL). engine runtime을 읽지 못한 경우 effective를 기본값이나 desired로 대체해서는 안 된다 (MUST NOT).

#### Scenario: 저장 설정과 실행 설정이 다르다
- **WHEN** config desired는 ON·3%이고 running engine effective는 OFF·5%다
- **THEN** 화면은 두 값을 각각 표시하고 어느 하나를 다른 값으로 덮지 않는다

#### Scenario: engine control plane을 읽지 못한다
- **WHEN** config desired는 읽히지만 running engine runtime은 unavailable이다
- **THEN** desired는 표시하고 effective는 `알 수 없음`으로 표시한다

### Requirement: 편입 보조 상태는 candidate와 reconcile 차단을 함께 설명한다
`/positions`와 `/position-management`는 기존 관리 판정 라벨 옆에 stable adoption status를 표시해야 한다 (SHALL). projector는 running effective settings가 known일 때 `candidate=(globalEnabled || included) && !excluded`를 계산하되 `journal unknown > already managed > operator released > excluded > candidate and covering reconcile block > candidate pending > unselected` 순서로 평가해야 한다 (SHALL). runtime unavailable인 non-managed 행은 desired를 effective로 위장하지 않고 `UNKNOWN`과 runtime unavailable 이유를 표시해야 한다 (SHALL). 미국 시장이라는 이유만으로 `편입 불가`라고 단정해서는 안 된다 (MUST NOT).

#### Scenario: include된 미국 보유분이 account-wide 차단을 만난다
- **WHEN** 미국 보유분에 adoption include가 있고 entry/adoption record는 아직 없으며 account-wide quantity-mismatch block이 active다
- **THEN** 두 화면은 기존 `관리 편입` 판정과 `대사 차단으로 대기` 보조 상태, block 사유를 표시하고 미국 시장 미지원으로 설명하지 않는다

#### Scenario: managed와 exclude가 함께 있다
- **WHEN** 이미 entry/adoption 근거로 managed인 symbol이 장래 candidate exclude에도 있다
- **THEN** 현재 보호 상태는 `MANAGED`로 유지되고 exclude가 기존 관리 lifecycle을 해제한 것으로 표시되지 않는다

#### Scenario: 미지정 행과 account block이 함께 있다
- **WHEN** global adoption이 OFF이고 include되지 않은 미편입 symbol에 account-wide block이 존재한다
- **THEN** 행은 `UNMANAGED`이며 candidate가 아니므로 `RECONCILE_BLOCKED`로 표시되지 않는다

#### Scenario: include와 exclude가 함께 있다
- **WHEN** 미편입 symbol이 include와 exclude에 모두 있다
- **THEN** candidate는 false이고 `EXCLUDED`가 표시된다

#### Scenario: runtime은 unavailable이고 desired만 include다
- **WHEN** journal은 readable이고 미편입 symbol이 desired config include에 있으나 running engine runtime을 읽지 못한다
- **THEN** desired include 사실은 설정 요약에 보존되지만 행의 effective status는 `UNKNOWN`이며 `ADOPTION_PENDING`으로 승격되지 않는다

#### Scenario: 운영자가 external lifecycle을 release했다
- **WHEN** adoption ID는 남아 있지만 authoritative position-policy lifecycle이 `RELEASED`다
- **THEN** 두 화면은 `UNMANAGED`, `OPERATOR_RELEASED`, `관리 외(운영자 해제)`를 표시하고 candidate 또는 account block보다 release를 우선한다

### Requirement: 저장 exit 근거와 실효 보호선을 구분한다
canonical persisted effective snapshot이 없는 exit state는 현재 실효 보호선이나 다음 익절 가격을 만들어 내서는 안 된다 (MUST NOT). 다만 journal에 저장된 t0 entry, initial stop, baseline과 high-water는 `원장 기록 · 실효 미확인` 증거로 별도 표시해야 하며 (SHALL), actionable effective line과 동일한 필드/라벨로 표시해서는 안 된다 (MUST NOT).

#### Scenario: legacy seed-only exit state
- **WHEN** exit state에 entry/initial-stop/baseline은 있으나 canonical effective snapshot이 없다
- **THEN** 화면은 current protection과 next target을 `—`로 유지하고 저장된 가격들을 `원장 기록 · 실효 미확인` 상세로 표시한다

#### Scenario: canonical effective snapshot이 있다
- **WHEN** exit state에 유효한 canonical effective snapshot이 있다
- **THEN** 기존 operatorview freshness와 effective-source 판정을 그대로 사용해 실효 보호선과 다음 익절을 표시한다

### Requirement: a052 운영 상태 표면은 읽기 전용이다
a052는 reconcile preview/apply route, capability, free-text field 또는 journal mutation을 console에 추가해서는 안 된다 (MUST NOT). 기존 position-policy lifecycle mutation surface와 a052 runtime read endpoint는 별도 권한으로 유지해야 한다 (SHALL). Compose sidecar용 shared Unix endpoint는 authenticated GET runtime만 제공하고 lifecycle Preview/Apply를 제공해서는 안 된다 (MUST NOT).

#### Scenario: 정적 route와 HTML 검사
- **WHEN** a052 route table과 `/positions`, `/position-management` HTML을 검사한다
- **THEN** reconcile resolution POST route와 text/textarea/number/contenteditable 입력이 없고 shared runtime surface에는 authenticated GET 외의 command가 없다

### Requirement: positions는 모든 시장에서 기준선 근거 상태를 직접 설명한다
`/positions`는 시장과 무관하게 실효 snapshot, 저장 원장 근거, 미생성 상태를 서로 다른 라벨로 표시해야 한다 (SHALL). 저장 원장 근거는 표의 주 정보로 읽을 수 있어야 하지만 actionable `ExitLine` 가격으로 복사되어서는 안 된다 (MUST NOT).

#### Scenario: KR legacy managed position
- **WHEN** KR 관리 포지션에 baseline과 initial stop은 있으나 canonical effective snapshot이 없다
- **THEN** 표는 `원장 기준선 · 실효 미확인`, 저장 baseline과 최초 손절을 접지 않은 상태에서 표시하고 current effective/next target을 만들지 않는다

#### Scenario: US legacy managed position
- **WHEN** US 관리 포지션에 같은 legacy 원장 근거가 있다
- **THEN** KR과 동일한 증거 상태와 필드를 표시하며 시장별로 숨기거나 다른 가격을 계산하지 않는다

#### Scenario: stale canonical snapshot
- **WHEN** canonical snapshot이 stale이다
- **THEN** 기존 `오래된 평가`와 사유를 유지하고 actionable 및 raw 가격을 표의 주 기준선으로 표시하지 않는다

#### Scenario: 다른 lifecycle generation의 저장 근거
- **WHEN** 현재 position-policy generation과 exit state 또는 snapshot generation이 다르다
- **THEN** 과거 가격과 snapshot identity를 모두 숨기고 세대 불일치 사유를 표시한다

#### Scenario: 손상되었거나 검증되지 않은 저장 근거
- **WHEN** snapshot 상태가 partial/invalid/corrupt이거나 시도된 lifecycle lookup이 현재 generation을 확인하지 못한다
- **THEN** 저장 가격을 숨기고 손상 또는 관리 세대 확인 불가 사유를 표시한다

### Requirement: 미편입 후보는 기준선이 아직 없는 이유와 정책 폭을 표시한다
KR 또는 US 포지션이 `ADOPTION_PENDING` 또는 `RECONCILE_BLOCKED`이고 exit state가 없으면 `/positions`는 `기준선 미생성`과 편입 후 생성된다는 설명을 표시해야 한다 (SHALL). running effective adoption 설정을 아는 경우 최초 손절폭을 percentage로 표시해야 하며 (SHALL), broker average/current price에서 보호 가격을 계산해서는 안 된다 (MUST NOT).

#### Scenario: reconcile-blocked US candidate
- **WHEN** include된 US 보유분은 effective initial stop 3%를 사용하지만 account reconcile block 때문에 아직 편입되지 않았다
- **THEN** 행은 `기준선 미생성`, `편입 후 확정`, `effective 최초 손절폭 3%`를 표시하고 숫자 가격선은 만들지 않는다

#### Scenario: pending KR candidate
- **WHEN** KR 보유분이 편입 후보이고 runtime effective 설정을 읽을 수 있다
- **THEN** US와 동일한 미생성 설명과 effective percentage를 표시한다

#### Scenario: runtime unavailable
- **WHEN** desired include에 지정된 KR 또는 US 보유분이 있지만 candidate 여부나 effective stop percentage를 증명할 runtime commander를 읽지 못한다
- **THEN** 화면은 편입 요청 저장과 실행 상태 미확인을 구분하고, desired/default 값을 대신 사용하지 않으며 기준선/percentage를 `알 수 없음`으로 유지한다

#### Scenario: managed 또는 released 상태와 desired include가 충돌한다
- **WHEN** 이미 엔진이 관리 중이거나 운영자가 해제한 종목이 desired include 목록에도 남아 있다
- **THEN** 현재 lifecycle 상태가 편입 예약보다 우선하며 해당 행을 `편입 예약됨` 또는 runtime-unknown candidate로 표시하지 않는다

### Requirement: 기준선 복원 화면은 입력과 mutation을 추가하지 않는다
`/positions`는 form, visible input, button, contenteditable 또는 reconcile/order mutation route를 추가해서는 안 된다 (MUST NOT).

#### Scenario: read-only responsive view
- **WHEN** 375px viewport와 정적 route 계약을 검사한다
- **THEN** 세 증거 상태는 읽을 수 있고 visible input이나 POST capability는 없다

### Requirement: 콘솔 내비게이션은 운영 흐름을 따른다

상단 navigation은 엔진의 운영 흐름 순서로 배열된 소수의 최상위 항목이어야 한다(SHALL — 발굴 → 발주 → 보유 → 종결 → 설정. 학습 가능한 순서가 기억할 필요 없는 순서다). 각 항목은 라벨과 함께 그 화면이 답하는 질문을 한 줄로 표시해야 한다(SHALL — 라벨만으로는 "검증 콘솔"과 "검증", "거래 이력"과 "성과 이력", "포지션"과 "포지션 정책"이 구분되지 않는다).

navigation 라벨은 그 화면 전체를 가리켜야 하며 그 화면의 한 섹션 이름이어서는 안 된다(SHALL NOT — 한 섹션 이름을 라벨로 쓰면 같은 화면의 나머지 섹션은 navigation에서 보이지 않는다).

어떤 화면도 다른 화면 내부의 링크로만 도달 가능해서는 안 된다(SHALL NOT — 현재 세 화면이 그렇다). 최상위 항목에 들어가지 않는 화면은 상위 항목의 하위 화면에서 명시적 진입점을 가져야 한다(SHALL).

navigation 재편은 기존 경로를 제거해서는 안 된다(SHALL NOT — 항목이 navigation에서 빠지는 것과 라우트가 사라지는 것은 다르다. 문서·북마크·기존 링크가 계속 동작해야 한다). 현재 화면에는 `aria-current`를 제공해야 한다(SHALL).

#### Scenario: navigation 항목의 설명
- **WHEN** 아무 화면의 navigation을 렌더하면
- **THEN** 각 항목이 라벨과 그 화면이 답하는 질문을 함께 표시한다

#### Scenario: 고아 화면 부재
- **WHEN** 콘솔의 화면 목록과 navigation 도달 경로를 대조하면
- **THEN** 최상위 항목 또는 그 하위 화면의 진입점을 통해 모든 화면에 도달할 수 있다

#### Scenario: 재편 후 기존 경로
- **WHEN** navigation에서 빠진 화면의 기존 URL로 직접 접속하면
- **THEN** 그 화면이 그대로 렌더된다

### Requirement: 설정은 가역성과 빈도로 분류된다

설정 화면은 기능 이름이 아니라 **변경의 가역성과 빈도**로 분류된 하위 화면으로 나뉘어야 한다(SHALL): 비가역·저빈도(상시), 가역·일 단위(당일), 전략 규칙, 진단 도구. 기본 진입은 가역·일 단위 화면이어야 한다(SHALL — 가장 자주 여는 화면이 가장 적은 클릭을 갖는다).

한 설정 **컨트롤**은 정확히 하나의 하위 화면에만 나타나야 한다(SHALL NOT — 같은 컨트롤을 두 화면에 복제하면 어느 쪽이 진짜인지 판단해야 한다). 이 금지는 컨트롤에만 적용된다(SHALL — 다른 탭이 소유한 값의 **읽기 전용 표시와 그 탭으로 가는 링크는 오히려 요구된다**. 지금 Guardian 한도 섹션이 게이트 ON/OFF 상태를 보여주고 스위치는 운영 섹션에 있다고 안내하는 패턴이 그것이며, 탭을 나눈 뒤 그 안내가 사라지면 운영자는 두 값의 관계를 화면에서 읽을 수 없게 된다).

각 하위 화면은 제목과 한 줄 설명과 현재 운영 상태 요약을 머리에 표시해야 한다(SHALL — 설정을 바꾸는 사람은 지금 게이트가 켜져 있는지, 엔진이 도는지, 한도가 설정돼 있는지를 그 자리에서 알아야 한다).

전략 규칙과 진단 도구 하위 화면은 이미 자기 경로를 가진 화면을 흡수하지 않고 **진입점**을 제공해야 한다(SHALL — a050이 고정한 canonical category deep link와 그 계약을 따르는 요구사항들이 그 경로를 참조한다). 각 진입점은 그 화면의 현재 desired/effective 요약 한 줄을 함께 표시해야 하며(SHALL — 링크만 있는 화면은 또 하나의 빈 화면이다), 그 요약을 재계산해서는 안 된다(SHALL NOT — 대상 화면이 이미 계산한 값을 옮긴다).

이 분류는 설정 저장의 의미를 바꾸지 않아야 한다(SHALL NOT — 제출 경로·필드명·검증·audit 기록·CSRF 게이트 무변경). 신설되는 하위 화면 라우트는 전부 GET이어야 한다(SHALL). automation gate·Guardian 한도·kill switch의 편집 권한 경계는 이 분류로 바뀌지 않아야 한다(SHALL NOT).

기존 설정 경로는 기본 하위 화면으로 리다이렉트해야 한다(SHALL).

#### Scenario: 기본 진입
- **WHEN** 운영자가 설정을 열면
- **THEN** 가역·일 단위 하위 화면이 열린다

#### Scenario: 컨트롤의 단일 위치
- **WHEN** 설정 하위 화면 넷의 렌더 결과를 대조하면
- **THEN** 같은 설정 컨트롤이 둘 이상의 화면에 나타나지 않는다

#### Scenario: 다른 탭 값의 읽기 전용 표시
- **WHEN** 당일 탭에서 Guardian 한도를 보면
- **THEN** automation gate의 현재 상태가 읽기 전용으로 표시되고 스위치가 있는 상시 탭으로 가는 링크가 함께 있으며, 그 탭에 스위치가 복제되지는 않는다

#### Scenario: 하위 화면 머리의 운영 상태
- **WHEN** 설정 하위 화면을 열면
- **THEN** 제목·한 줄 설명과 함께 게이트 상태·엔진 실행 여부·한도 설정 여부를 그 자리에서 읽을 수 있다

#### Scenario: 전략 진입점
- **WHEN** 전략 하위 화면을 열면
- **THEN** 최적화·종목 정책·전략 lane으로 가는 진입점이 각각 현재 요약 한 줄과 함께 표시되고, 각 링크는 기존 canonical 경로를 가리킨다

#### Scenario: 저장 계약 무변경
- **WHEN** 재분류 후 어떤 설정 폼을 제출하면
- **THEN** 제출 경로·필드명·검증 판정·audit 기록이 재분류 전과 같다

#### Scenario: 신설 라우트의 메서드
- **WHEN** 라우트 표를 검사하면
- **THEN** 신설된 설정 하위 화면 라우트는 전부 GET이고 콘솔의 상태변경 행위 목록은 늘지 않았다

### Requirement: 설정 폼은 변경 전후와 차단 사유를 표시한다

모든 설정 폼은 같은 형식으로 네 가지를 표시해야 한다(SHALL): ① 현재값과 제출하려는 값, ② 적용 후 무엇이 바뀌고 언제 반영되는지, ③ 저장할 수 없거나 주의가 필요하면 그 사유의 이름, ④ 저장 결과.

저장할 수 없거나 저장이 거부될 상태라면 화면은 **이름 붙은 사유**를 저장 표면이 있어야 할 자리에 표시해야 한다(SHALL). 이 요구는 저장 표면이 **비활성인 경우와 아예 렌더되지 않는 경우 모두**에 적용된다(SHALL — 이 콘솔은 seam이 없으면 폼을 비활성화하는 것이 아니라 렌더하지 않는다. `disabled` 속성만 찾는 검사는 0건을 찾고 통과하므로 도달 불가능한 요구가 된다). 사유 없이 저장 표면이 사라지거나 비활성이어서는 안 된다(SHALL NOT — 이유를 말하지 않는 빈자리는 운영자에게 아무 다음 행동도 주지 않는다).

사유는 화면이 새로 판정한 것이 아니라 이미 존재하는 게이트에서 유도해야 한다(SHALL — 저장 seam 미배선, 설정 파일 읽기 실패, 엔진이 거부할 블록, 한도 미설정/부분 설정, 단위 grid 불일치, 기동 인터록 미충족, 엔진 실행 중으로 인한 반영 지연). 새 판정을 도입해서는 안 된다(SHALL NOT).

저장 결과는 그 결과를 만든 폼 옆에 표시해야 한다(SHALL — 페이지 수준 한 줄로는 여러 폼 중 어느 것이 저장됐는지 알 수 없다).

Guardian 한도 변경은 방향을 구분해 표시해야 한다(SHALL): 한도를 좁히는 변경과 넓히는 변경이 같은 표시를 받아서는 안 된다(SHALL NOT). 방향 판정은 **같은 통화 안에서만** 성립한다(SHALL — 한도 통화가 바뀌면 숫자의 대소는 강화·완화를 뜻하지 않는다). 통화 변경은 강화도 완화도 아닌 **별개의 축**으로 표시하고 어느 시장의 자동 진입이 닫히는지를 함께 말해야 한다(SHALL — 한도 통화는 하나이고 Guardian 체인은 다른 통화의 진입을 거부하므로, 통화를 바꾸는 것은 반대편 시장을 닫는 것이다).

이 표시는 판정이 아니다 — 허용 여부는 서버의 기존 검증이 소유하며 화면은 그 결과를 옮길 뿐, 완화를 차단하거나 정책을 재계산해서는 안 된다(SHALL NOT).

설정 화면은 자유 문구 입력이나 typed confirmation을 요구해서는 안 된다(SHALL NOT — 설정 저장은 이미 서버에서 audit되므로 추적성을 위한 입력 강제는 근거가 없고, 콘솔 UI에 확인 문구 입력 마찰을 두지 않는다는 결정이 있다). 필수 사유 입력을 도입해서는 안 된다(SHALL NOT).

#### Scenario: 변경 전후
- **WHEN** 어떤 설정 폼에 현재값과 다른 값이 들어 있으면
- **THEN** 폼 머리가 현재값과 제출값을 함께 표시하고, 적용 후 무엇이 바뀌며 언제 반영되는지를 저장 전에 읽을 수 있다

#### Scenario: 저장 seam 미배선
- **WHEN** 저장 seam이 주입되지 않은 빌드에서 설정 폼을 열면
- **THEN** 저장 컨트롤 옆에 "저장 미배선"이 이름으로 표시되고 다른 폼은 영향받지 않는다

#### Scenario: 사유 없는 저장 표면 부재
- **WHEN** 설정 화면에서 저장 표면이 비활성이거나 렌더되지 않은 자리를 찾으면
- **THEN** 그 자리에 이름 붙은 사유가 있으며, 사유 없이 사라지거나 비활성인 저장 표면이 있으면 검사가 실패한다

#### Scenario: 폼별 저장 결과
- **WHEN** 여러 폼이 있는 설정 화면에서 하나를 저장하면
- **THEN** 결과가 그 폼 옆에 표시되어 어느 폼이 저장됐는지 식별된다

#### Scenario: 한도 완화
- **WHEN** 운영자가 같은 통화에서 현재보다 넓은 Guardian 한도를 입력하면
- **THEN** 화면은 그것이 완화임을 강화와 구분해 표시하되 제출을 차단하지 않으며, 허용 여부는 서버의 기존 검증이 답한다

#### Scenario: 한도 통화 변경
- **WHEN** 운영자가 한도 통화를 바꾸는 값을 입력하면
- **THEN** 화면은 그것을 강화·완화가 아닌 별개의 변경으로 표시하고 어느 시장의 자동 진입이 닫히는지를 함께 말하며, 숫자의 대소를 방향으로 해석하지 않는다

#### Scenario: 입력 마찰 부재
- **WHEN** 설정 화면의 모든 폼을 검사하면
- **THEN** 확인 문구 타이핑을 요구하는 입력이나 필수 사유 입력란이 없다

### Requirement: 설명은 상태와 분리해 접는다

콘솔 화면의 설명문은 두 종류로 나뉘고 다르게 표시되어야 한다(SHALL): **지금 무엇이 참인가**와 **이 컨트롤을 누르면 무엇이 일어나는가**는 항상 보여야 하고(SHALL), **왜 이렇게 설계했는가**·**출처와 전례**·**경계 조건 해설**은 native HTML 상세 영역으로 접어야 한다(SHALL). 접힌 요소의 요약문은 그 안에 무엇이 있는지 예고해야 한다(SHALL).

다음은 어떤 경우에도 접어서는 안 된다(SHALL NOT): ① 실계좌에 요청이 나간다는 경고, ② 저장 seam 미배선 경고, ③ 설정 파일 읽기 실패, ④ 엔진이 지금 거부한다는 판정과 미충족 항목, ⑤ 잔여물·캐시 stale·갱신 보류 안내, ⑥ 반영 시점, ⑦ 한도 통화가 반대편 시장의 진입을 닫는다는 귀결, ⑧ 사전 판정 통과가 기동을 보장하지 않는다는 경계.

이 목록은 **문구 매칭이 아니라 클래스로 고정해야 한다**(SHALL — 한국어 문장 조각을 찾는 검사는 문구가 한 글자만 바뀌어도 침묵으로 통과한다. 검사가 죽어도 알 수 없는 검사는 검사가 아니다): 위 여덟 종류는 경고·주의 클래스를 달아야 하고(SHALL — 대부분 이미 그렇게 렌더되고 있으며, 그렇지 않은 항목은 승격한다), **그 클래스를 가진 요소는 상세 영역 안에 나타나서는 안 된다**(SHALL NOT). 검사는 클래스의 위치만 본다.

접힘 상태는 **자동 재로드가 걸리는 화면에서** 재로드를 넘어 유지되어야 한다(SHALL — 접힘 상태가 매 주기 초기화되면 그 화면에서는 접기를 쓸 수 없다). 그 화면의 상세 영역은 URL 질의 파라미터로 여닫아야 하며(SHALL — JavaScript 없이 성립하고 딥링크·뒤로가기가 따라온다), 그 화면에서 URL을 바꾸지 않는 여닫기 수단을 제공해서는 안 된다(SHALL NOT — 열리는 것처럼 보이다가 다음 재로드에 닫히는 것은 접기가 없는 것보다 나쁘다).

자동 재로드가 걸리지 않는 화면에는 이 요구를 적용하지 않는다(SHALL NOT — 설정 화면은 자동 재로드가 없으므로 native 상세 영역으로 충분하고, 불필요한 URL 상태를 도입하지 않는다).

접힘 파라미터는 표시 전용이어야 한다(SHALL NOT — 저장·판정·audit 어디에도 도달하지 않는다). 알 수 없는 값은 오류가 아니라 무시하고 접힌 상태로 렌더해야 한다(SHALL).

콘솔 전체의 문체는 하나여야 한다(SHALL — 현재 한 화면만 다른 종결어미를 쓴다. 화면을 옮길 때마다 다른 사람이 말하는 인상을 준다).

#### Scenario: 근거의 접힘
- **WHEN** 설명이 많은 화면을 열면
- **THEN** 현재 상태와 컨트롤의 결과는 펼쳐진 채이고, 설계 근거와 출처 설명은 요약문이 붙은 상세 영역 안에 있다

#### Scenario: 접지 않는 항목의 클래스 위치
- **WHEN** 경고·주의 클래스를 가진 요소가 상세 영역 안에 렌더되면
- **THEN** 검사가 실패한다 — 문구가 아니라 클래스의 위치로 판정한다

#### Scenario: 자동 재로드 화면의 접힘 상태
- **WHEN** 자동 재로드가 걸린 화면에서 상세 영역을 연 뒤 재로드가 일어나면
- **THEN** 같은 상세 영역이 열린 채로 다시 렌더된다

#### Scenario: 자동 재로드가 없는 화면
- **WHEN** 설정 화면의 상세 영역을 검사하면
- **THEN** native 상세 영역으로 렌더되고 접힘 상태를 위한 질의 파라미터를 요구하지 않는다

#### Scenario: 알 수 없는 접힘 파라미터
- **WHEN** 존재하지 않는 상세 영역 식별자가 질의 파라미터로 들어오면
- **THEN** 화면은 오류 없이 전부 접힌 상태로 렌더된다

#### Scenario: 표시 전용 파라미터
- **WHEN** 접힘 파라미터가 붙은 상태로 설정을 저장하면
- **THEN** 저장 판정과 audit 기록은 그 파라미터가 없을 때와 동일하다

#### Scenario: 문체 일관성
- **WHEN** 콘솔 전 화면의 설명문을 검사하면
- **THEN** 종결어미 문체가 하나로 통일돼 있다

