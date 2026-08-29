# operator-console Specification

## Purpose
로컬 웹 콘솔의 안전 불변식(루프백 리스너·2경로 인증·CSRF·프로세스 재기동 외 상태변경 부재·무주문)과 read-only 운영 가시성(포지션·exit 라인·거래 이력·rate budget 보호) 계약.
## Requirements
### Requirement: 콘솔 안전 불변식

운영 콘솔은 두 mode 중 하나로만 서비스해야 한다 (SHALL): ① 기본 native local mode는
127.0.0.1 listener와 기존 terminal possession session을 사용하고, ② 명시적
trusted-network mode는
`vpn-console-access` spec의 host loopback/VPN publish·CIDR·TLS·canonical Host/Origin
계약을 모두 사용하며 application login을 요구하지 않는다. remote mode가
완전히 구성되지 않은 non-loopback listener는 거부하고 닫아야 한다 (SHALL).

재시작 handoff는 프로세스 교체 무결성을 위한 내부 단발성 자격으로만 유지하며,
사용자 application login으로 사용해서는 안 된다 (SHALL NOT).

상태를 바꾸는 모든 route는 CSRF gate를 요구하고 trusted-network에서는
same-origin 검사도 요구한다 (SHALL). 현재 허용된 상태 변경은 검증 제어, 자기/soak/
engine process 제어, 편입·Guardian 한도·trading policy·automation gate·공통 exit
policy 설정, 검증된 system update 작업뿐이다(SHALL). 이 change는 그 목록이나 각
seam의 기존 쓰기 범위를 넓히지 않는다 (SHALL NOT). console에는 direct 주문 발주·
정정·취소 또는 credential 표시 route가 없어야 하며(SHALL NOT), engine/verify가
계좌를 변경할 수 있는 조건은 기존 사람 승인, audit, startup interlock과 공식 API
경로가 계속 결정한다 (SHALL).

route table 정적 검사는 package 모든 파일의 `HandleFunc`/`Handle`을 검사하고,
모든 state-changing route가 mode와 무관하게 CSRF와 해당 mode의 origin gate chain
뒤에 있음을 보장해야 한다 (SHALL). fixed health 외 별도의 public login/logout
lifecycle route는 필요하지 않으며 direct account mutation capability를 추가해서는
안 된다 (SHALL NOT).

#### Scenario: 기본 local listener
- **WHEN** remote option 없이 console을 시작한다
- **THEN** 127.0.0.1만 bind하고 기존 terminal session URL로 인증한다

#### Scenario: 완전한 remote listener
- **WHEN** 유효한 remote access 구성을 가진 console이 non-loopback listener로 Serve된다
- **THEN** TLS/network/origin gate 아래에서 application login 없이 local console과 같은 handler 기능을 제공한다

#### Scenario: 불완전한 비루프백 listener
- **WHEN** remote access 구성이 없거나 불완전한 console이 non-loopback listener를 받는다
- **THEN** service가 거부되고 listener가 닫힌다

#### Scenario: 사용자 credential 부재
- **WHEN** loopback 또는 trusted VPN browser가 console을 연다
- **THEN** token 파일 조회, login form 또는 login cookie 교환 없이 화면을 제공한다

#### Scenario: remote 상태 변경 gate
- **WHEN** CSRF 또는 same-origin 증거가 틀린 설정/engine/verify POST가 도착한다
- **THEN** 요청은 handler seam 전에 거부되고 config, process, account는 바뀌지 않는다

#### Scenario: 주문 route 부재
- **WHEN** console route table과 capability closure를 검사한다
- **THEN** remote login 기능이 direct order/broker credential capability를 추가하지 않았고 기존 direct 주문 route 금지가 유지된다

#### Scenario: engine interlock 유지
- **WHEN** remote session에서 ProtectionReady 미충족 engine start를 누른다
- **THEN** engine process의 기존 interlock refusal이 그대로 표시되고 console이 우회하지 않는다

### Requirement: 포지션 가시성

콘솔은 계좌의 보유·포지션 상태를 read-only로 표시해야 한다(SHALL): 브로커 보유 스냅샷(수량·평균단가·현재가 — holdings 응답의 lastPrice·평가손익)과 journal 투영(positions·exit_states)을 심볼 기준으로 조인한다. exit 관리 자격이 있는 포지션(자격의 정의는 exit-policy·position-ledger 스펙이 소유한다)은 exit 상태 — t0 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부 — 를 함께 표시한다(SHALL). 자격 없는 보유의 판정 라벨은 편입·제외 지정 상태를 반영한다(SHALL — 사용자 UX 결정 2026-07-27, 2026-07-30): `exclude_symbols`에 있는 행은 **"관리 제외"**, `include_symbols`에 지정된(체크된) 행은 **"관리 편입"**, 어느 목록에도 없는 행은 **"관리 외(미편입)"** — 각 라벨의 철자는 하나로 통일하며 행마다 표시한다. 제외는 편입보다 우선하므로 라벨 판정도 제외를 먼저 본다(SHALL — 화면의 우선순위가 엔진의 우선순위와 어긋나면 라벨이 엔진의 행동을 잘못 예고한다). "관리 제외"는 운영자가 의도적으로 보호에서 뺐다는 사실의 표시이며 지정되지 않아 편입되지 않은 것과 구별된다(SHALL — 엔진의 미관리 알림도 둘을 구별해 사유를 말한다). "관리 편입" 라벨은 편입 예약 상태의 표시일 뿐 보호 성립을 의미하지 않는다(SHALL — "편입 예약됨"과 실행 주체(엔진 대사 루프) 안내를 병기하고, 원장을 읽지 못한 행은 지정 여부와 무관하게 "관리 여부 불명"을 유지한다 — 콘솔이 관측하지 않은 보호를 단정하지 않는다). exit 라인이 없는 이유의 명시 위치는 스코프를 따른다(SHALL): 같은 상태의 모든 행에 공통인 사유 — 원장 미판독(관리 여부 불명), 원장에 포지션 없음 — 는 페이지 수준 안내 1회로 명시하고 같은 문장을 행마다 반복하지 않는다(SHALL NOT — 반복 안내는 표를 읽을 수 없게 만든다); 행 고유 사유(자격 기록 없는 원장 포지션, 자격은 있으나 exit 미개설)는 해당 행에 명시한다. 어느 쪽 데이터 소스가 없거나 비어 있어도 다른 쪽만으로 렌더한다(SHALL — 조인 실패가 화면 실패가 되어서는 안 된다). journal 스키마 불일치는 방향별로 구분 안내한다(SHALL): 바이너리보다 새로우면 "콘솔 업데이트 필요", 오래됐으면(필요 테이블 부재) "엔진 기동으로 마이그레이션 필요" — 어느 쪽도 빈 상태로 위장하지 않는다.

#### Scenario: 엔진 관리 포지션의 exit 라인 표시
- **WHEN** journal에 exit_states 행이 있는 포지션을 포지션 화면이 렌더하면
- **THEN** 진입가·최초 손절·기준선·워터마크·래칫 단계·ladder rung·부분익절 여부가 해당 심볼 행에 표시된다

#### Scenario: 관리 외 보유의 정직한 구분
- **WHEN** 브로커 보유에는 있으나 exit 관리 자격이 없고 어느 목록에도 없는 심볼을 렌더하면
- **THEN** 해당 행은 "관리 외(미편입)"로 표시되고 exit 라인 없음이 엔진 미관리 때문임이 안내된다

#### Scenario: 지정된 행의 라벨
- **WHEN** include에 지정된 미편입 보유 행을 렌더하면
- **THEN** 판정 라벨이 "관리 편입"으로 표시되고 "편입 예약됨"과 실행 주체 안내가 병기되며, 같은 심볼이라도 원장을 읽지 못한 상태에서는 "관리 여부 불명"이 유지된다

#### Scenario: 제외된 행의 라벨
- **WHEN** exclude에 있는 미편입 보유 행을 렌더하면
- **THEN** 판정 라벨이 "관리 제외"로 표시되고, 같은 심볼이 include에도 있더라도 "관리 제외"가 유지되며, 원장을 읽지 못한 상태에서는 "관리 여부 불명"이 유지된다

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
콘솔의 대시보드·포지션·주문 화면은 운영자가 첫 화면에서 판단해야 하는 사실을 1차 정보로 표시해야 한다(SHALL). 대시보드와 포지션 화면은 동일한 read-only 보유 projection과 동일한 shared holdings template을 사용해야 하며(SHALL), 둘 중 하나가 lifecycle·desired/effective policy 또는 exit evidence를 다르게 해석하거나 다른 열 구조로 표시해서는 안 된다(SHALL NOT).

각 보유의 desktop 1차 표는 `종목, 수량, 평균가, 현재가, 라인, 총금액, 미실현 PnL` 일곱 핵심 값 열을 이 순서로 표시하고, 마지막에 좁은 `상세` action 열을 두어야 한다(SHALL). header와 body는 같은 명시적 column grid를 사용하고(SHALL), 수량·가격·금액·손익 header와 값은 같은 오른쪽 정렬 축과 tabular numerals를 사용해야 한다(SHALL). holdings table의 desktop header/body typography와 cell spacing은 StockOS reference의 compact scan density(약 10/12 CSS px, 15/18 CSS px line-height, 8×12 CSS px padding)에 맞춰야 한다(SHALL).

`라인` 열은 익절, 손절, 추적 회수, 기준, 고점을 이 순서의 compact stack으로 표시해야 한다(SHALL). `익절`은 canonical exit projection의 다음 익절가, `손절`은 최초 손절, `추적 회수`는 다음 보호선, `기준`은 현재 보호선, `고점`은 high-water를 의미해야 한다(SHALL). canonical snapshot이 없거나 stale·lifecycle generation 불일치·runtime unknown이면 actionable 가격을 계산하거나 raw 원장값으로 대체하지 않고 `—`를 표시해야 한다(SHALL NOT).

관리·차단·pending·excluded 및 exit evidence의 concise verdict는 접지 않고 표시해야 한다(SHALL). 반복되거나 긴 reconciliation detail, management explanation, exit reason, 자격 provenance, 원장 수량·평단, generation과 decision/snapshot/observation ID, 평가 시각, 원본 저장 기준은 마지막 `상세` 열에서 키보드로 여는 native HTML disclosure popup에 두어야 한다(SHALL). popup은 화면 중앙의 독립된 non-modal 표면으로 표시되고 행 높이나 PnL 열을 확장해서는 안 된다(SHALL NOT). 상세 영역을 열지 않아도 일곱 핵심 값과 concise safety verdict를 확인할 수 있어야 한다(SHALL).

375 CSS pixel viewport에서는 같은 일곱 핵심 필드와 마지막 상세 action을 label-value card flow로 표시하고 문서 전체에 수평 overflow가 생기지 않아야 한다(SHALL). 상세 summary는 최소 44 CSS pixel 높이와 visible keyboard focus를 유지해야 한다(SHALL). popup은 viewport 안에서 스크롤 가능해야 하고 JavaScript, 입력 control, 외부 asset을 요구하지 않아야 한다(SHALL NOT).

#### Scenario: desktop 일곱 열 정렬
- **WHEN** 대시보드 또는 포지션 화면을 desktop viewport에서 렌더하면
- **THEN** 일곱 header와 모든 보유 값은 같은 explicit column grid에 놓이고 수량·평균가·현재가·총금액·미실현 PnL은 오른쪽 축으로 정렬된다

#### Scenario: StockOS형 compact line stack
- **WHEN** 보호 근거가 있는 KR 또는 US 보유를 렌더하면
- **THEN** `라인` 열은 concise verdict 다음에 익절·손절·추적 회수·기준·고점을 12 CSS px 수준의 compact stack으로 표시하고 verbose reason은 primary row에 반복하지 않는다

#### Scenario: 근거가 없는 보유의 진실성
- **WHEN** canonical exit evidence가 없거나 stale 또는 generation mismatch이면
- **THEN** visible concise verdict와 다섯 개 `—` 값이 상태를 알리고 긴 설명과 raw evidence는 상세 영역에만 표시된다

#### Scenario: 작은 viewport card flow
- **WHEN** 같은 표를 375 CSS pixel 폭에서 렌더하면
- **THEN** 일곱 필드가 label-value card flow로 읽히고 문서의 수평 overflow가 없으며 상세 summary는 최소 44 CSS pixel 높이다

#### Scenario: 마지막 열의 상세 popup
- **WHEN** desktop 또는 mobile 보유 행에서 마지막 `상세` action을 열면
- **THEN** 긴 진단 내용은 화면 중앙 popup으로 표시되고 원래 행의 PnL 및 높이를 확장하지 않으며 같은 summary를 다시 작동해 닫을 수 있다

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
콘솔은 완결된 왕복 거래(trade_outcomes)와 exit 이벤트(exit_events)를 시간순으로 표시해야 한다(SHALL). 각 행의 종목은 journal의 market+symbol을 공식 종목 메타데이터와 batch로 보강해 `심볼 · 종목명`을 함께 표시해야 한다(SHALL). 종목명 조회가 실패하거나 해당 심볼의 메타데이터가 없으면 journal 행과 심볼은 그대로 표시하고 종목명을 만들지 않아야 한다(SHALL NOT). 표시는 journal에 동결된 값과 명시 조인(positions의 심볼, exit_states의 진입가)만 사용하며 종목명 보강을 포함해 fills 재계산을 하지 않는다(SHALL NOT). 스키마에 없는 값(청산가 등)은 표시하지 않고(SHALL NOT), nullable 필드(보유 시간 등)의 NULL은 "—"로 렌더한다(SHALL). 성과 행이 생기지 않는 종결(외부 매도로 닫힌 포지션 — adopt-external-positions design A7)은 이 화면의 한계로 명시하고 exit_events 표시가 그 공백을 보완한다(SHALL 명시).

종목명 조회는 `(market, symbol)`을 정규화해 중복 제거하고 한 logical lookup당 400개로 제한해야 하며(SHALL), 첫 live-data 화면에서 계좌를 한 번 확인해 생성한 공식 client와 OAuth token manager를 history 및 다른 account read 화면이 공유해야 한다(SHALL). 공식 endpoint의 200-symbol 한도에 맞춰 chunk하고 client 생성부터 모든 chunk까지 하나의 10초 total timeout을 적용해야 한다(SHALL). 결과는 24시간 TTL·최대 2048개 bounded cache에 저장하고 같은 키의 동시 요청은 single-flight로 직렬화해야 한다(SHALL). 실패하거나 일부 key가 누락된 lookup은 1분 동안 다시 시도하지 않아 공식 API 장애·429 때 새로고침이 rate budget을 증폭하지 않아야 하며(SHALL NOT), 일부 응답에서는 검증된 이름만 장기 cache하고 누락 key는 backoff 뒤 다시 요청해야 한다(SHALL). 실계좌 검증 중에는 새 메타데이터 요청을 보내지 않고 cached name 또는 symbol-only로 표시해야 한다(SHALL NOT). 이를 위해 metadata는 active profile journal directory의 cross-process rate-budget lease를 non-blocking으로 얻어야 하고(SHALL), CLI verify run·console verify·verify abort는 기존 execution exclusion을 얻고 profile run-intent marker를 게시한 뒤 broker 생성 전에 같은 lease를 얻어 작업 종료까지 보유해야 한다(SHALL). active verifier가 있으면 verify abort는 즉시 거부되어 같은 marker를 공동 소유해서는 안 되며(SHALL NOT), 배타 admission 뒤 evidence record를 다시 읽어 최신 취소 대상만 사용해야 한다(SHALL). `--record` override가 profile rate-budget lease 위치를 바꾸어서는 안 된다(SHALL NOT). 원격 name은 길이·control/bidi 문자를 검증한 plain string으로 `html/template`의 escaping을 거쳐야 하며(SHALL), 모호하거나 상충하는 결과를 다른 market/symbol에 붙여서는 안 된다(SHALL NOT). `/history`는 GET/HEAD만 허용하고 다른 method는 메타데이터 요청 전에 405로 거부해야 한다(SHALL).

#### Scenario: 동결된 왕복 결과와 종목명 표시
- **WHEN** KR 또는 US trade_outcomes 행이 있고 공식 종목 메타데이터 조회가 성공한 상태에서 이력 화면을 열면
- **THEN** 각 왕복은 `심볼 · 종목명`과 비용 차감 실현손익·실현 R·초기 수량·보유 시간(NULL은 "—")·도달 exit 단계·청산 시각을 표시하며 동결 성과 값은 바뀌지 않는다

#### Scenario: exit 이벤트의 동일한 종목 라벨
- **WHEN** KR 또는 US exit_events 행의 종목명이 조회되면
- **THEN** exit 이벤트 표도 완결 왕복 표와 같은 `심볼 · 종목명` 형식으로 표시한다

#### Scenario: 종목명 조회 실패
- **WHEN** 공식 종목 메타데이터 조회가 실패하거나 특정 심볼의 이름이 비어 있으면
- **THEN** 해당 journal 행과 심볼은 계속 표시되고 화면은 조회 실패를 이름 부재로 위장하거나 종목명을 추측하지 않는다

#### Scenario: 검증 중 또는 잘못된 method
- **WHEN** 실계좌 검증이 진행 중이거나 인증된 사용자가 `/history`에 POST를 보내면
- **THEN** 공식 종목 메타데이터 요청은 0건이고 각각 symbol-only 안내 또는 405 응답을 반환한다

### Requirement: read-only 불변식

대시보드는 계좌·원장에 대한 어떤 mutation도 수행해서는 안 된다(SHALL NOT): journal은 `OpenReadOnly`로만 연다 — DB 파일·디렉터리를 생성하지 않고, 마이그레이션을 실행하지 않으며, DB에 쓰지 않는다(SHALL — `mode=ro`; WAL 공유 인덱스(`-shm`/`-wal`) 접근은 SQLite WAL 읽기의 전제로서 명시된 예외다). 쓰기 연결 부재를 가드 테스트로 고정한다(SHALL). 콘솔이 주입받는 브로커 인터페이스는 **조회 메서드만 선언**하고(SHALL — holdings 계열), mutation 메서드가 없음을 정적 테스트로 고정한다(SHALL — verifylive.Broker 같은 광폭 인터페이스 주입 금지). **config에 대한** 콘솔의 유일한 쓰기 표면은 주입된 편입 설정 seam이며(SHALL — 대상은 config 파일의 `engine.adoption` 블록만; 검증 증거 기록·핸드오프 토큰 파일 등 기존 주입 writer의 계약은 무변경), 이 seam은 다른 config 키를 유실하지 않고(SHALL — 구조체 왕복이 아니라 해당 블록만 교체·블록 밖 바이트 보존) 유일 임시파일과 잠금 아래 원자적으로 기록한다(SHALL — 동시 기록의 lost-update 금지). seam의 Load는 파일의 `engine.adoption` 블록 **원문**을 반환하고 검증 판정을 별도로 병기한다(SHALL — 거부된 블록의 목록이 화면 왕복으로 유실되어서는 안 된다). 파싱할 수 없는 config 파일에 대한 저장은 거부된다(SHALL — 골격 생성은 파일 부재에 한정). seam은 Load·Save 두 메서드만 선언하며(SHALL — 정적 검사) internal/console은 config 서비스 타입을 직접 명명하지 않는다(SHALL NOT — 정적 검사). seam이 배선되지 않은 빌드에서 설정 화면은 저장 불가를 안내하고 나머지 화면은 영향받지 않는다(SHALL). Save 성공은 audit 로그에 저장 시점 엔트리를 남긴다(SHALL — §0.5; 엔진 기동 시 diff 기록과 이중이며, 기동 없는 flip도 기록에 남는다). 기존 콘솔의 게이트·주문 라우트 부재 가드는 새 라우트 표에서도 유지된다(SHALL). 브로커·설정 능력은 **용도별로 분리된 seam**으로 주입하며 하나로 합치지 않는다(SHALL — 합치면 그 능력을 쓰지 않는 화면도 그것을 쥔다). Guardian 한도 읽기는 편입 설정 seam의 세 번째 메서드가 아니라 **별개의 한 메서드 seam**이다(SHALL — 편입 seam의 "Load·Save 두 메서드뿐" 규칙은 그대로 유지된다). 주입 능력의 정적 검사는 **`Options`가 선언한 필드**를 단위로 하며(SHALL), 인터페이스뿐 아니라 **func 타입 seam**도 대상이다(SHALL — 현행 주입 seam 다수가 func 타입이고, 인터페이스만 보는 검사는 그것들을 전혀 보지 못한다). 검사는 **패키지의 모든 파일**을 대상으로 하고 파일 이름·타입 이름에 고정되지 않는다(SHALL NOT). allowlist에 등록되지 않은 새 `Options` 필드는 그 자체로 검사를 실패시킨다(SHALL — 실제 불변식은 "콘솔이 받는 능력이 전부 열거되어 있다"이다). 브로커 캐시는 **갱신하지 않고 읽는 접근자**를 제공해야 한다(SHALL — 갱신 보류 플래그를 대신 쓰면 검증이 돌고 있지 않은데도 "검증 중 — 갱신 보류"라는 지어낸 사유가 렌더된다).

#### Scenario: journal 쓰기 시도 차단
- **WHEN** 콘솔 코드가 journal 쓰기 경로를 얻으려는 변경이 들어오면
- **THEN** RO 접근 가드 테스트가 실패한다

#### Scenario: 광폭 브로커 인터페이스 주입 차단
- **WHEN** 콘솔의 브로커 인터페이스에 mutation 메서드가 추가되면
- **THEN** 정적 테스트가 실패한다

#### Scenario: 편입 설정 저장의 외과적 기록
- **WHEN** 콘솔이 편입 설정을 저장할 때 config 파일에 이 스키마가 모르는 키가 존재하면
- **THEN** 저장 후에도 그 키는 보존되고 `engine.adoption` 블록만 바뀐다
#### Scenario: 새 파일에 선언된 광폭 seam
- **WHEN** 기존 검사 대상이 아니던 파일에 mutation 메서드를 가진 seam이 선언되고 `Options`에 주입되면
- **THEN** 정적 테스트가 실패한다 — 검사는 패키지 전체를 걷는다

#### Scenario: func 타입으로 주입된 mutation 능력
- **WHEN** 인터페이스가 아니라 func 타입으로 주문 능력이 `Options`에 주입되면
- **THEN** 정적 테스트가 실패한다 — 검사의 단위는 인터페이스가 아니라 주입된 능력이다

#### Scenario: 열거되지 않은 새 능력
- **WHEN** allowlist에 없는 새 필드가 `Options`에 추가되면
- **THEN** 정적 테스트가 실패한다 — 능력은 열거된 것만 주입될 수 있다

### Requirement: rate budget 보호

브로커 스냅샷은 요청 시 lazy 갱신이며 서버측 백그라운드 폴러를 두지 않는다(SHALL NOT). 갱신 1회의 브로커 호출은 holdings 1콜로 한정하고(SHALL — 현재가는 응답의 lastPrice 사용, 심볼별 시세 fan-out 금지), 서버측 캐시 TTL은 15초 이상이다(SHALL). TTL 내 재요청·다중 탭은 추가 브로커 호출을 유발하지 않으며 캐시 시각이 화면에 표시된다(SHALL). 포지션 화면은 브라우저 재로드 지시(meta refresh)를 포함할 수 있으며 그 주기는 캐시 TTL 이상이어야 한다(SHALL — 각 재로드는 요청 시 lazy 갱신을 그대로 타므로 열린 탭 하나의 비용 상한은 holdings 1콜/TTL이다; 이 지시는 서버측 폴러가 아니다). 검증 실행 중에는 갱신을 보류한다(SHALL): 이 콘솔 프로세스의 실행 중 run은 in-process 신호로, 다른 프로세스의 run은 runlock 마커의 mtime 신선도(5분 상한 — stale은 무시)로 판단한다. 자동 재로드는 이 보류를 우회하지 못한다(SHALL — 보류 중 재로드는 캐시를 서빙할 뿐 브로커 호출을 만들지 않는다). 브로커 호출 상한은 **화면별**로 정한다(SHALL): 포지션 화면의 갱신 1회는 holdings 1콜, **주문 화면의 갱신 1회는 미체결 주문 1콜·종결 주문 1콜·조건주문 1콜로 모두 3콜**이며, 개요 화면은 브로커 호출을 0콜로 하고 이미 있는 캐시를 갱신 없이 읽기만 한다(SHALL). 조건주문을 세지 않는 미체결 건수는 측정으로 표시해서는 안 된다(SHALL NOT — 조건주문도 노출 상한을 채우는 잔여물이며 프로세스 종료를 넘어 존속한다; 한쪽만 세고 "0건"이라 적으면 다음 검증을 막고 있는 잔여물이 화면에서 사라진다). 두 조회 중 하나만 실패한 경우 그 부분은 **미측정**으로 표시하고 성공한 쪽의 건수와 합산하지 않는다(SHALL NOT). 화면의 필터는 **한 번 가져온 캐시 위에서 in-process로** 적용한다(SHALL — 필터를 브로커 파라미터로 넘기면 필터 조합마다 캐시가 갈라져 상한이 배로 늘어난다).

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

#### Scenario: 조건주문만 실패한 주문 조회
- **WHEN** 일반 주문 조회는 성공하고 조건주문 조회가 실패하면
- **THEN** 화면은 일반 주문 건수와 "조건주문 미측정"을 함께 표시하며, 하나의 합계로 합치지 않는다

#### Scenario: 필터를 바꿔 가며 여는 주문 화면
- **WHEN** TTL 안에서 상태·시장·방향 필터를 바꿔 가며 주문 화면을 여러 번 열면
- **THEN** 추가 브로커 호출이 발생하지 않는다 — 필터는 캐시 위에서 적용된다

### Requirement: 편입 설정 화면

콘솔은 편입 설정 화면을 제공해야 한다(SHALL): `adoption.enabled`(전 종목 자동 편입)·`default_stop_pct`·`exclude_symbols`·`include_symbols`(종목별 편입 — exit-policy 스펙이 의미를 소유한다)를 표시·편집하되, 손절폭은 마우스 조절 슬라이더로 기본값(5%)과 현재값을 함께 표시하고(SHALL — 사용자 UX 결정 2026-07-27: 타이핑 요구 금지), 목록 직접 기입은 고급 접힘 안에만 둔다(SHALL — 1차 경로는 포지션 화면의 행별 버튼이며, 이는 **두 목록 모두**에 해당한다 — 사용자 결정 2026-07-30), 거부된 블록은 **파일 원문 값**과 거부 사유를 함께 표시한다(SHALL — 침묵 무시 금지·목록 유실 금지). 저장은 서버측에서 exit-policy의 검증 규칙과 동일하게 재검증하며 위반 시 저장을 거부하고 사유를 표시한다(SHALL — 클라이언트 폼을 신뢰하지 않는다; 엔진이 zeroing할 블록을 기록하는 것은 금지된다). 화면은 반영 시점("다음 엔진 기동부터 반영" — 가동 중 엔진은 기동 시점 설정으로 동작함을 포함)과 편입 비가역 귀결(include 제거·exclude 추가는 이미 편입된 포지션에 무효 — 편입 해제 기능은 존재하지 않는다), 지정의 상시성(CLOSED 후 재매수 재편입 포함)을 명시하고(SHALL), 저장 응답은 현재 엔진 실행 여부를 함께 안내한다(SHALL — 엔진 마커 판독). automation gate의 ON/OFF 편집 표면은 존재하지 않는다(SHALL NOT — Guardian 한도의 편집 표면은 별도 요구사항이 소유하며, 그 표면도 `enabled`를 쓰지 않는다). 포지션 화면의 관리 외(미편입) 보유 행은 종목 편입 지정 행위와 **종목 제외 지정 행위**를 제공하며(SHALL — 각각 확인 대화상자 1회 외에 어떤 입력도 요구하지 않는다; 지정된 행과 제외된 행은 각각 해제 행위를 제공한다), 그 행위는 목록 추가·제거일 뿐 편입을 직접 수행하지 않는다(SHALL NOT — 안내 문구는 실행 주체가 엔진 대사 루프임을 유지한다). 이미 include 또는 exclude에 있는 심볼의 행은 그 상태를 표시한다(SHALL). 종목 제외 지정은 `default_stop_pct`를 기록해서는 안 된다(SHALL NOT — 제외 목록만으로는 블록이 손절폭 검증을 부르지 않으므로, 제외가 손절폭을 발명하면 운영자가 고르지 않은 숫자가 파일에 남는다; 편입 지정의 기본값 채움은 검증이 요구하기 때문이며 제외에는 그 근거가 없다). 화면은 두 지정이 동시에 참인 상태를 만들어서는 안 된다(SHALL NOT — 엔진은 제외를 편입보다 우선하므로 동시 등재는 아무 효과도 없는 편입 지정을 화면이 성공으로 보고하는 상태다): 제외 지정은 같은 저장에서 그 심볼의 편입 지정을 함께 해제하고 무엇이 해제됐는지 안내하며(SHALL — 보수 방향이므로 추가 확인을 요구하지 않는다), 제외된 심볼의 행에는 편입 지정 행위를 제공하지 않고 제외 해제가 선행함을 안내한다(SHALL). 화면을 우회해 도달한 편입 지정 요청이 제외된 심볼을 가리키면 지정은 기록하되 제외가 우선하여 편입되지 않음을 안내해야 한다(SHALL — 화면이 컨트롤을 감춘 것은 강제가 아니며, 침묵으로 성공을 보고해서는 안 된다). 이미 엔진이 관리 중인 포지션의 행에는 제외 행위를 제공하지 않는다(SHALL NOT — exclude 추가가 이미 편입된 포지션에 무효라는 규칙과 같은 사실이고, 효과 없는 행위를 그리는 것은 그 규칙을 화면에서 부정하는 것이다; 그 편집은 고급 접힘의 목록에 남는다).

#### Scenario: 자동 편입 켜기
- **WHEN** 설정 화면에서 enabled를 켜고 유효한 default_stop_pct와 함께 저장하면
- **THEN** config의 adoption 블록이 갱신되고, 반영 시점 안내가 표시되며, 다음 엔진 기동의 audit 기록에 flip이 남는다

#### Scenario: 종목별 편입 지정
- **WHEN** 포지션 화면에서 관리 외(미편입) 보유의 편입 지정 행위를 수행하면
- **THEN** 그 심볼이 include_symbols에 추가되고(중복은 1회로 정규화), 편입은 수행되지 않으며, 다음 대사 주기(엔진 가동 시)에 편입 후보가 됨이 안내된다

#### Scenario: 종목별 제외 지정
- **WHEN** 포지션 화면에서 관리 외 보유의 제외 지정 행위를 수행하면
- **THEN** 그 심볼이 exclude_symbols에 추가되고(중복은 1회로 정규화), config의 다른 값은 바뀌지 않으며, 반영 시점이 안내된다

#### Scenario: 제외 지정은 손절폭을 발명하지 않는다
- **WHEN** default_stop_pct가 미설정인 상태에서 제외 지정을 수행하면
- **THEN** 저장된 블록의 default_stop_pct는 여전히 미설정이고, 저장은 거부되지 않는다

#### Scenario: 편입 지정된 심볼의 제외
- **WHEN** include_symbols에 있는 심볼에 대해 제외 지정을 수행하면
- **THEN** 같은 저장에서 그 심볼이 include_symbols에서 빠지고 exclude_symbols에 들어가며, 편입 지정이 함께 해제됐음이 안내된다

#### Scenario: 제외된 심볼에 도달한 편입 지정 요청
- **WHEN** 제외된 심볼을 가리키는 편입 지정 요청이 화면을 우회해 도달하면
- **THEN** 안내는 제외가 우선하여 편입되지 않는다고 말하고, 편입 예약이 성립했다고 보고하지 않는다

#### Scenario: 무효 설정 저장 거부
- **WHEN** enabled 또는 비어 있지 않은 include와 함께 범위 밖 default_stop_pct를 저장하려 하면
- **THEN** 저장이 거부되고 사유가 표시되며 config는 변경되지 않는다

#### Scenario: 손절폭 미설정 상태의 종목 편입 지정
- **WHEN** default_stop_pct가 미설정·무효인 상태에서 포지션 화면의 종목 편입 지정을 수행하면
- **THEN** 콘솔 기본값(5%)이 블록에 명시적으로 기록되어 지정이 즉시 성립하고, 기본값이 적용됐음이 안내된다 (엔진이 zeroing할 블록을 쓰지 않는다는 규칙은 유지 — 기록되는 블록은 항상 유효하다)

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

톤 임계는 그 화면 자신의 캐시 TTL에서 유도해야 하며 새 상수를 도입해서는 안 된다(SHALL NOT): 경과가 TTL 미만이면 정상, TTL 이상 2×TTL 미만이면 주의, 2×TTL 이상이거나 읽기가 실패했으면 경고. 톤의 근거인 캐시 TTL과 그 화면의 재로드 주기는 별개의 값이다(SHALL — 재로드가 잦아지는 것은 캐시 도달 빈도만 바꾸고 톤 임계를 바꾸지 않는다).

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
- **THEN** 검증 화면은 run이 작업 중일 때만 짧은 주기를 쓰고, 나머지 화면은 각자 자기 주기의 출처를 그대로 유지한다 — 표시줄은 그 주기를 말할 뿐 정하지 않는다

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

Guardian 한도 변경은 방향을 구분해 표시해야 한다(SHALL): 한도를 좁히는 변경과 넓히는 변경이 같은 표시를 받아서는 안 된다(SHALL NOT). `limit_currency`는 account base currency이며, 그 통화를 바꾸는 것은 모든 account-wide money limit의 단위를 바꾸는 별도 정책 migration이다(SHALL). 화면은 이를 단순 숫자 대소나 한 시장의 ON/OFF로 표현해서는 안 되고(MUST NOT), 현재 KR identity 및 US official FX authority 준비 상태와 재기동 후 반영을 함께 말해야 한다(SHALL). 한 시장의 안정화가 다른 시장 설계·구현의 선행조건이라는 안내를 표시해서는 안 된다(MUST NOT).

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
- **WHEN** 운영자가 같은 account base currency에서 현재보다 넓은 Guardian 한도를 입력하면
- **THEN** 화면은 그것이 완화임을 강화와 구분해 표시하되 제출을 차단하지 않으며, 허용 여부는 서버의 기존 검증이 답한다

#### Scenario: account base currency 변경
- **WHEN** 운영자가 `limit_currency`를 바꾸는 값을 입력하면
- **THEN** 화면은 account-wide limit 단위 migration, KR/US FX authority 준비 상태와 재기동 반영을 표시하고 특정 시장이 선행 안정화되어야 한다고 안내하지 않는다

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

### Requirement: 콘솔이 띄우는 프로세스는 콘솔의 프로필로 뜬다

콘솔이 spawn하는 자식 프로세스는 콘솔 자신이 실행 중인 프로필(`--config-dir`·`--session-file`)을 물려받아야 한다(SHALL — 콘솔은 그 프로필로 자격증명·기록·로그 경로를 계산해 화면에 그리므로, 자식이 다른 프로필로 뜨면 운영자가 보고 있는 것과 자식이 만지는 것이 다른 파일이 된다). 이 규칙은 엔진과 조회 전용 서베이 양쪽에 같이 적용된다(SHALL).

콘솔이 화면에 표시하는 산출물 경로와 그 화면의 버튼이 만들어 내는 산출물 경로는 같아야 한다(SHALL — 다르면 버튼은 조용히 실패하거나 운영자가 볼 수 없는 곳에 쓴다).

프로세스를 찾는 패턴은 콘솔이 실제로 spawn하는 명령줄에 일치해야 하며(SHALL), 다른 하위 명령의 명령줄에 일치해서는 안 된다(SHALL NOT — 서베이를 찾는 패턴이 엔진을 잡으면 정지 시그널이 엔진에 간다).

발견된 프로세스에 시그널을 보내기 전에 그것이 이 콘솔이 소유한 인스턴스인지 판정해야 한다(SHALL). 소유의 기준은 그 프로세스가 쓰는 산출물이다(SHALL — 엔진은 journal 디렉터리, 서베이는 기록 경로). 판정은 콘솔 자신이 그 경로를 구할 때와 같은 해석 함수를 거쳐야 하며(SHALL — 기본 경로를 명시한 콘솔과 생략한 autostart는 같은 인스턴스다), 소유를 증명할 수 없는 프로세스에는 시그널을 보내서는 안 된다(SHALL NOT).

이 요구사항은 서베이의 조회 전용 성질·판정 기준·주기를 바꾸지 않으며 새 상태변경 라우트를 만들지 않는다(SHALL NOT).

#### Scenario: 격리 프로필 콘솔이 서베이를 재시작한다
- **WHEN** `--config-dir`를 지정해 실행 중인 콘솔에서 서베이 재시작을 요청하면
- **THEN** spawn된 서베이가 그 프로필의 자격증명을 읽고 그 프로필의 기록에 append한다

#### Scenario: 화면이 가리키는 기록과 버튼이 만드는 기록이 같다
- **WHEN** 콘솔이 "아직 기록이 없다"며 기록 경로를 표시한 상태에서 재시작 버튼을 누르면
- **THEN** 그 경로에 기록이 생긴다

#### Scenario: 서베이 패턴은 엔진을 잡지 않는다
- **WHEN** 엔진 명령줄을 서베이 패턴으로 검사하면
- **THEN** 일치하지 않는다

#### Scenario: 다른 기록의 서베이는 건드리지 않는다
- **WHEN** 다른 기록 경로로 실행 중인 서베이가 함께 관측되는 상태에서 재시작을 요청하면
- **THEN** 그 프로세스에는 시그널이 가지 않는다

#### Scenario: 열거 실패는 부재가 아니다
- **WHEN** 프로세스 열거가 오류를 반환하면
- **THEN** 부재로 단정하지 않는다

### Requirement: 콘솔의 엔진 읽기는 화면 수와 재로드 주기에서 독립이다

콘솔이 표시를 위해 엔진 프로세스에 거는 읽기는 렌더 1회당 1벌이어서는 안 된다(SHALL NOT — 엔진의 journal 쓰기 핸들은 커넥션이 하나이고 그 커넥션은 exit 판정 트랜잭션이 쓴다. 표시 경로가 렌더마다 거기에 닿으면 열린 화면 수와 재로드 주기가 손절 판정 간격에 더해진다). 표시 경로의 엔진 읽기는 서버측 캐시가 상한을 강제해야 하며(SHALL), 그 상한은 캐시 간격당 각 읽기 1회다(SHALL — 콘솔이 스스로 수행한 mutation에 의한 무효화는 이 상한의 예외이며, 그 mutation은 이미 엔진에 닿은 행위다). 동시에 도착한 렌더 여러 건은 읽기 1벌만 만들어야 한다(SHALL).

캐시 간격은 그 읽기가 다투는 자원과 그 읽기가 실어 나르는 값이 함께 정해야 한다(SHALL). 엔진의 쓰기 커넥션을 다투는 읽기와 다투지 않는 읽기는 같은 간격을 공유해서는 안 된다(SHALL NOT — 전자에는 안전 상한이 필요하고 후자에는 필요 없다). 살아 있는 대사 차단처럼 운영자에게 낙관적으로 틀릴 수 있는 값을 실어 나르는 읽기의 간격은 엔진의 관측 주기를 넘어서는 안 된다(SHALL — 차단이 걸린 보유를 "편입 예약됨"으로 표시하는 것은 보호가 진행 중이라는 거짓 안심이다). 어떤 간격도 브로커 캐시 TTL에서 파생되어서는 안 된다(SHALL NOT — 서로 다른 자원의 서로 다른 결정이다).

캐시는 마지막 **성공**이 아니라 마지막 **시도의 결과**를 서빙해야 한다(SHALL — 두 읽기 모두에 적용된다). 엔진 읽기가 실패하면 그 실패가 간격 동안 서빙되어야 하며 직전 성공이 되살아나서는 안 된다(SHALL NOT — 응답하지 않는 엔진의 옛 성공을 계속 보여주면 화면은 아무도 유지하지 않는 보호 설정을 effective로 주장하게 되고, lifecycle 쪽에서는 한 걸음 더 나아가 그 포지션이 **엔진 관리 중**이라고 주장하게 된다. 관리 판정은 effective 설정보다 먼저 평가되므로 runtime 쪽이 정직해도 막지 못한다). 실패 시도 역시 간격에 계상되어야 한다(SHALL — 엔진이 답하지 못하는 순간이 렌더마다 다시 연결하기에 가장 나쁜 순간이다).

캐시 갱신은 그것을 요청한 HTTP 요청의 취소에 영향받아서는 안 된다(SHALL NOT — 렌더를 버린 브라우저 하나가 공유 읽기에 실패를 기록하면 건강한 엔진인데도 모든 화면의 보호선이 한 간격 동안 사라지고, 재로드로 회복할 수 없다. "엔진이 답하지 못했다"와 "이 요청이 사라졌다"는 다른 사실이고 공유 캐시에는 앞의 것만 들어간다). 갱신에는 자체 시간 상한이 있어야 한다(SHALL — 답을 멈춘 엔진이 HTTP 핸들러를 붙들면 안 된다).

캐시된 읽기가 화면에 반영되는 나이는 그 읽기의 간격을 넘어서는 안 된다(SHALL). 엔진 상태를 읽지 못하는 행이 `UNKNOWN`으로 표시되어야 한다는 기존 요구사항은 이 상한 안에서 성립한다(SHALL — 엔진이 사라진 뒤 늦어도 한 간격 안에 화면이 `UNKNOWN`으로 돌아간다).

lifecycle 목록은 그것이 join되는 journal 행보다 오래될 수 있다(SHALL 인정 — journal은 렌더마다 읽히므로 셋은 결코 한 시점이 아니다). 그 불일치는 판정을 **보류할 수 있을 뿐 만들어낼 수 없어야 한다**(SHALL — 목록이 모르는 행은 보호선 없이 `관리 여부 불명`으로 렌더되며, 보호받고 있다고 표시되어서는 안 된다(SHALL NOT)).

콘솔이 수행해 성공한 정책 mutation은 이 캐시를 즉시 무효화해야 한다(SHALL). 실패한 mutation은 무효화해서는 안 된다(SHALL NOT — 거부된 요청 반복만으로 캐시를 무력화할 수 있게 된다). 이 캐시를 지나지 않는 상태를 바꾸는 행위는 무효화 대상이 아니다(SHALL NOT — 무효화 목록은 실제 data flow를 따라야 하며, 닿지 않는 상태를 위해 무효화하는 것은 엔진 읽기를 사고 존재하지 않는 흐름을 주장하는 주석을 남긴다).

**자동 재로드가 없는 화면**은 이 캐시를 거치지 않고 직접 읽어야 한다(SHALL — 상한이 필요한 이유는 화면이 사람 없이 스스로 반복하기 때문이다. 사람이 여는 만큼만 읽는 화면에는 그 이유가 없고, 엔진에 명령을 발행하는 화면에서는 운영자 행동의 근거가 그 행동을 위해 방금 취한 읽기여야 한다는 요구가 더해진다).

읽기가 신선할 때 화면이 표시하는 값은 캐시 도입 전과 같아야 한다(SHALL).

#### Scenario: 재로드가 잦아져도 엔진 도달은 늘지 않는다
- **WHEN** 보호선을 그리는 화면이 캐시 간격 동안 여러 번 다시 그려지면
- **THEN** 엔진 lifecycle 목록 읽기는 1회를 넘지 않는다

#### Scenario: 두 화면이 함께 열려 있다
- **WHEN** 두 보호선 화면이 같은 간격 안에서 각각 렌더되면
- **THEN** 엔진 읽기는 화면마다가 아니라 간격마다 1벌 발생하고, 두 화면 모두 그 판정을 렌더한다

#### Scenario: 동시에 도착한 렌더
- **WHEN** 여러 렌더가 같은 순간에 캐시에 도달하면
- **THEN** 엔진 읽기는 각 읽기당 1회만 발생한다

#### Scenario: 쓰기 커넥션을 다투지 않는 읽기의 주기
- **WHEN** 대사 차단이 엔진에 걸린 뒤 보호선 화면을 다시 열면
- **THEN** 늦어도 엔진 관측 주기 안에 그 행은 "편입 예약됨"이 아니라 대사 차단 대기로 표시되고, 그 사이 lifecycle 목록은 다시 읽히지 않는다

#### Scenario: 엔진이 runtime을 답하지 않는다
- **WHEN** 직전 간격에 성공한 읽기가 있고 이번 runtime 시도가 실패하면
- **THEN** 화면은 effective를 `알 수 없음`으로 표시하고 직전 성공값을 되살리지 않는다

#### Scenario: 엔진이 lifecycle 목록을 답하지 않는다
- **WHEN** 직전 간격에 성공한 목록이 있고 이번 시도가 실패하면
- **THEN** 그 포지션은 `엔진 관리`가 아니라 `관리 여부 불명`으로 표시된다

#### Scenario: 답하지 않는 엔진을 반복해서 부르지 않는다
- **WHEN** 엔진 읽기가 실패한 뒤 같은 간격 안에 화면이 여러 번 다시 그려지면
- **THEN** 엔진에 다시 연결하지 않고 같은 실패 결과가 서빙된다

#### Scenario: 렌더를 버린 브라우저
- **WHEN** 한 렌더가 엔진 읽기 도중 요청 취소로 중단되면
- **THEN** 그 취소는 캐시에 기록되지 않고, 뒤이은 정상 렌더는 엔진에 대한 답을 서빙받는다

#### Scenario: journal보다 오래된 lifecycle 목록
- **WHEN** journal에는 있으나 캐시된 목록에는 없는 포지션을 렌더하면
- **THEN** 그 행은 `관리 여부 불명`으로 표시되고 보호선을 갖지 않는다

#### Scenario: 운영자가 정책을 바꾼다
- **WHEN** 콘솔을 통한 정책 mutation이 성공한 직후 보호선 화면을 열면
- **THEN** 캐시 간격을 기다리지 않고 바뀐 lifecycle 상태가 보인다

#### Scenario: 거부된 mutation
- **WHEN** 콘솔을 통한 정책 mutation이 거부되면
- **THEN** 캐시는 그대로 유지되고 다음 렌더는 엔진에 다시 닿지 않는다

#### Scenario: 자동 재로드가 없는 화면의 읽기
- **WHEN** 자동 재로드가 없는 화면이 엔진 상태를 표시하면
- **THEN** 그 값은 캐시가 아니라 그 요청을 위해 취한 직접 읽기다

### Requirement: 보호선 화면은 스크립트 없이 엔진 주기를 따라간다

보호선을 그리는 화면은 갱신 주기를 얻기 위해 client-side script를 도입하거나 배포 CSP를 완화해서는 안 된다(SHALL NOT). 콘솔의 화면 표면은 `html/template`와 inline CSS만으로 성립해야 하며(SHALL — `2026-07-31-streamline-trading-views` design의 Non-Goal "JavaScript, CSP nonce, `unsafe-inline` script, 외부 CSS/폰트 도입"의 성문화), 응답 CSP는 `default-src 'none'`을 유지하고 `script-src`를 추가해서는 안 된다(SHALL NOT — `script-src`의 부재는 템플릿에 섞여 들어온 inline handler가 반드시 죽어 있게 만드는 장치이며, 그 장치는 스크립트를 실제로 쓰는지와 별개로 유지된다).

전체 재로드가 이 화면들에서 감당 가능한 근거는 명시되어야 한다(SHALL): 접힘 상태는 URL이 보존하고, 이 화면들에는 편집 중 잃을 form·input·button이 없으며, 브로커 비용 상한은 캐시가 지킨다. 셋 중 하나라도 성립하지 않게 되면 주기를 올리기 전에 그 사실을 먼저 해소해야 한다(SHALL).

#### Scenario: 보호선 화면의 렌더 표면
- **WHEN** 보호선을 그리는 화면의 렌더 결과와 응답 CSP를 확인하면
- **THEN** `<script>`가 없고 응답 CSP에 `script-src`가 없다

#### Scenario: 재로드가 접힘을 잃지 않는다
- **WHEN** 열린 상세를 가진 보호선 화면이 재로드 주기에 따라 다시 열리면
- **THEN** 그 상세는 URL이 보존한 상태 그대로 열려 있다

### Requirement: Settings exposes one fixed staged system update
The authenticated settings page SHALL display the running executable and the
single staged candidate at `<running-path>.candidate`. The console SHALL accept
no path, URL, command, or uploaded executable from an HTTP request. It SHALL
display file size, UTC modification time, SHA-256, Go/module metadata, and
GOOS/GOARCH for an inspectable candidate, and SHALL explain why a missing or
invalid candidate cannot be installed. The displayed SHA-256 SHALL bind the
operator's click to the bytes later prepared for installation.

#### Scenario: Candidate is absent
- **WHEN** no sibling `.candidate` file exists
- **THEN** settings reports that no update is staged and renders no enabled install action

#### Scenario: Valid candidate is staged
- **WHEN** the fixed sibling candidate is a valid tossctl executable for the current platform
- **THEN** settings shows its identity and an authenticated install action

#### Scenario: Request attempts to select another path
- **WHEN** a client adds a path, URL, or command field to the install request
- **THEN** the console ignores it and the installer can reach only its pre-bound sibling candidate

#### Scenario: Candidate changes after review
- **WHEN** the candidate bytes no longer match the SHA-256 displayed in the submitted settings form
- **THEN** installation is refused and neither current nor rollback bytes change

### Requirement: Candidate validation does not execute candidate code
Before installation the updater SHALL open with no-follow semantics, copy from
that one descriptor into a same-directory prepared file, and inspect file
metadata, SHA-256, and Go build information on the prepared bytes without
executing the candidate. It SHALL reject symlinks, non-regular files, files
without executable bits, wrong module identities, main-package identities other
than `github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl`, GOOS/GOARCH
mismatches, and a prepared hash that differs from the operator-reviewed hash.

#### Scenario: Symlink candidate
- **WHEN** the sibling candidate is a symlink
- **THEN** inspection refuses it and no installed or rollback byte changes

#### Scenario: Wrong module or platform
- **WHEN** build information names another module, GOOS, or GOARCH
- **THEN** inspection refuses it and no candidate code runs

#### Scenario: Different command from the same module
- **WHEN** build information names this repository as the module but names a main package other than `cmd/tossctl`
- **THEN** inspection refuses it and the running executable cannot be replaced by another repository tool

### Requirement: Install is idle-only, authenticated, and recoverable
`POST /settings/system-update/install` SHALL require the normal console session
and CSRF token. It SHALL refuse while a verification run is unfinished, while
the real engine exclusion cannot be held, while external verification evidence
is fresh or unreadable, or when same-port relaunch is unavailable. The
engine-start, verification-start, and update handlers SHALL share an in-process
exclusion, and a successful commit SHALL refuse new starts until the old console
exits. Immediately before commit, the installed target SHALL still be the same
regular executable fingerprinted at console startup.

The engine, updater, standalone verification command, and console verification
starter SHALL also share the same kernel-enforced journal-directory flock.
Verification SHALL acquire it before record/account/broker work and hold it
through complete runner cleanup. The updater SHALL hold it through executable
replacement and relaunch request. Advisory verification evidence SHALL NOT be
the sole exclusion.

A successful install SHALL prepare and sync a same-directory temporary file,
atomically publish and sync a rollback copy while the current path remains
intact, atomically replace the running path, sync the directory, record old/new
hashes and rollback path in console output, and then request the existing
authenticated same-port relaunch. A failed replacement SHALL leave or restore
the previous executable and keep the current console serving.

#### Scenario: Engine is running
- **WHEN** the engine holds its journal-directory exclusion and the operator posts a valid install request
- **THEN** installation is refused, the engine is not stopped, and executable files are unchanged

#### Scenario: Verification is running
- **WHEN** a verification run is unfinished and the operator posts a valid install request
- **THEN** installation is refused and executable files are unchanged

#### Scenario: External verification races with replacement
- **WHEN** update holds the real execution flock and an external or console verification attempts to start before replacement finishes
- **THEN** verification refuses before account resolution or any order-capable broker construction

#### Scenario: Advisory verification marker cannot be written
- **WHEN** a live verification owns the real execution flock but its advisory marker is missing or unwritable
- **THEN** installation is still refused by the real flock

#### Scenario: Replacement succeeds
- **WHEN** the console is idle, the candidate is valid, and replacement completes
- **THEN** the old executable is the rollback file, the candidate bytes occupy the running path, and the browser returns through the existing same-port handoff

#### Scenario: Final replacement fails
- **WHEN** candidate rename fails or the post-replacement directory sync fails
- **THEN** the current path is never absent, the updater leaves or restores the previous executable, reports restoration status, and does not request relaunch

#### Scenario: Installed target drifted
- **WHEN** the current executable path or bytes differ from the console startup fingerprint at commit time
- **THEN** installation is refused and the operator is told to restart the console before reviewing the update again

### Requirement: Development staging has a stable target
The repository SHALL provide a staging target that builds the current source and
writes the result to `<install-path>.candidate` with executable mode. It SHALL
not overwrite the installed executable or trigger a restart.

#### Scenario: Agent stages a gated build
- **WHEN** a developer or agent runs the documented staging target
- **THEN** only the fixed candidate is written and the operator can inspect and install it from settings

### Requirement: Settings stages only an authenticated signed release
The authenticated settings page SHALL display a “check and stage signed
release” action when production release download is wired.
`POST /settings/system-update/download` SHALL require the normal console session
and CSRF token and SHALL ignore request fields that attempt to choose a
repository, URL, tag, asset, destination, trust root, or signer. The action
SHALL invoke the fixed latest-release verifier and candidate publisher, then
redirect to settings with either an actionable refusal or the verified tag,
asset, archive SHA-256, signer workflow, and staged candidate SHA-256.

Download and signature verification SHALL NOT install an executable, request a
restart, stop an engine or verification, access an account or credentials, or
execute candidate code. Only the final local candidate publication window SHALL
share the existing in-process update/start exclusion; network work SHALL NOT
hold the engine/verification flock. The full signed-release operation SHALL use
a separate single-flight mutex so discovery/TUF verification cannot run
concurrently and an older slow result cannot overwrite a newer result.
Installation remains the separate existing operator-reviewed action.

#### Scenario: Unauthenticated download request
- **WHEN** a client posts the download route without the console session and matching CSRF token
- **THEN** the request is rejected and no release or candidate operation starts

#### Scenario: Request supplies another repository and path
- **WHEN** an authenticated request includes repository, URL, tag, path, command, or destination fields
- **THEN** those fields cannot change the constructor-bound release source or sibling candidate destination

#### Scenario: Signed release is staged
- **WHEN** the latest platform archive passes provenance, extraction, and executable validation
- **THEN** settings reports the verified identities and renders the existing separate candidate install review

#### Scenario: Verification fails
- **WHEN** release discovery, download, provenance verification, extraction, or publication fails
- **THEN** settings explains the failure, preserves the running executable and previous candidate, and requests no relaunch

#### Scenario: Download runs while engine is active
- **WHEN** an engine already holds its execution flock and the operator checks and stages a release
- **THEN** bounded network and verification work may proceed without stopping the engine, but no executable installation occurs

#### Scenario: Install commit races final staging
- **WHEN** an install commit or process relaunch becomes committed before a verified download can publish
- **THEN** candidate publication is refused and the old console cannot stage bytes after committing its own replacement

#### Scenario: Two release checks overlap
- **WHEN** one signed-release request is still discovering or verifying and another request arrives
- **THEN** the second request waits on the release-only single-flight lock without blocking engine start, and reverse completion cannot publish an older tag last

#### Scenario: Candidate survived a console restart
- **WHEN** settings renders a local candidate without process-local verified release provenance
- **THEN** it labels the provenance as unknown and does not call the candidate signed until a new signed download succeeds

### Requirement: 지속 관측되는 flat managed position은 actionable exit line을 유지한다
콘솔의 `/positions`와 `/position-management`는 shared freshness 판정을 사용해야 한다 (SHALL): engine stopped가 확정되면 즉시 stale이며, running·unavailable·unwired에서는 canonical snapshot integrity와 마지막 성공 관측의 30초 age bound를 함께 적용한다. 유효한 flat 관측이 계속 영속되면 current protection, next target과 next protection을 계속 actionable하게 표시해야 하며 (SHALL), `SEED`, corrupt, generation mismatch 또는 실제 stale 증거를 raw 값으로 보충해서는 안 된다 (MUST NOT).

#### Scenario: unchanged first quote 뒤 관리 화면
- **WHEN** 새 관리 포지션이 t0와 같은 첫 공식 가격으로 `EVALUATED` snapshot을 얻는다
- **THEN** 두 console 화면은 `not_evaluated_yet` 대신 canonical current/next line과 evaluated-at을 표시한다

#### Scenario: 30초 이상 가격이 움직이지 않는다
- **WHEN** 가격과 policy state는 30초 이상 같지만 성공한 공식 관측 refresh가 age bound 안에서 계속 영속된다
- **THEN** 두 console 화면은 `오래된 평가`로 강등하지 않고 최신 canonical line을 표시한다

#### Scenario: 실제로 관측이 끊긴다
- **WHEN** engine liveness와 무관하게 한 position의 마지막 성공 snapshot 관측이 30초를 넘으며 그 뒤 성공한 refresh가 없다
- **THEN** 두 console 화면은 actionable 가격을 `—`로 숨기고 typed stale 사유를 표시한다

#### Scenario: age 경계와 engine liveness
- **WHEN** running·unavailable·unwired snapshot을 29.999초, 정확히 30초, 30초 초과에서 읽거나 engine을 stopped로 확정해 읽는다
- **THEN** 앞의 두 age는 fresh이고 초과만 stale이며, stopped는 age와 무관하게 즉시 `engine_not_running`이다

#### Scenario: console blocking read 중 freshness 경계를 지난다
- **WHEN** 실제 `/position-management` 요청의 journal·runtime·quarantine 또는 단일 engine-marker read가 진행되는 동안 snapshot age가 30초를 넘거나 marker가 stopped 경계를 지난다
- **THEN** console은 모든 blocking read 뒤의 한 response clock으로 판정해 즉시 stale 사유와 dash를 표시하고 추가 marker read를 만들지 않는다

#### Scenario: stopped marker 뒤 wall clock이 rollback한다
- **WHEN** marker read는 engine을 stopped로 판정했지만 그 직후 response clock이 뒤로 움직여 marker가 다시 fresh처럼 보인다
- **THEN** console은 stopped를 running으로 승격하지 않고 `engine_not_running`과 dash를 유지한다

#### Scenario: running engine에서 한 symbol만 invalid다
- **WHEN** engine은 running이고 같은 batch의 valid sibling은 계속 평가되지만 이 position의 quote만 invalid/missing이다
- **THEN** 이 position의 timestamp는 전진하지 않아 30초 초과 뒤 stale로 숨겨지며 sibling liveness가 대신 freshness를 만들지 않는다

#### Scenario: 저장 snapshot이 손상됐다
- **WHEN** observed-at, identity 또는 JSON/flattened tuple 무결성 검증이 실패한다
- **THEN** engine runtime이 running이어도 화면은 freshness를 추정하지 않고 unknown/corrupt로 fail-closed한다

### Requirement: flat refresh는 운영 event나 주문으로 표시되지 않는다
콘솔은 의미가 동일한 observation refresh를 새 exit transition, proposal, intent 또는 broker order로 표시해서는 안 된다 (MUST NOT). 기존 exit history는 실제 first evaluation, state/action transition과 arming decision만 유지해야 한다 (SHALL).

#### Scenario: 동일 가격 refresh가 여러 번 영속된다
- **WHEN** 한 evaluated position이 동일한 line으로 여러 번 refresh된다
- **THEN** 현재 line의 evaluated-at은 전진하지만 exit event와 order history 개수는 늘지 않는다

### Requirement: 검증 배치 승인의 형식

콘솔에서의 배치 승인은 **표시된 계획에 대한 명시적 단일 클릭**으로 성립해야 한다(SHALL — 사용자 결정 2026-07-27). 성립 요건은 셋이다: ① 유효한 세션(또는 소비되지 않은 핸드오프) 자격, ② CSRF 토큰, ③ 그 계획을 표시하고 있는 승인 화면에서의 POST. 콘솔은 승인을 위해 **확인 문자열 타이핑을 요구해서는 안 된다**(SHALL NOT — 단일 사용자의 루프백 콘솔에서 타이핑은 세션 탈취자에게 비용이 아니고(문자열이 화면에 표시된다) 정당한 사용자에게만 비용이며, 2026-07-27 장중 측정 창이 승인 이전 마찰로 소모된 실측이 근거다).

승인 화면은 이 실행이 보낼 수 있는 **모든 라이브 요청의 완전한 목록**을 승인 전에 표시해야 한다(SHALL — 배치 승인 모델은 목록의 완전성에 걸려 있다; 목록을 표시하지 않는 승인 경로는 존재해서는 안 된다(SHALL NOT)). 화면이 표시하는 목록과 터미널이 출력하는 목록은 같은 원천에서 렌더되어야 하며(SHALL — 두 번째 요약을 만들지 않는다), 화면에는 그 화면에서 실제로 성립하지 않는 승인 방식(타이핑 지시·확인 문자열)을 표시해서는 안 된다(SHALL NOT).

승인 창 만료(현행 5분)와 만료 시 "아무것도 전송되지 않았다" 종결은 유지된다(SHALL). 계획에 라이브 mutation이 0건이면 승인을 묻지 않는다(SHALL — 승인할 것이 없다). 자동 승인·비대화 승인 경로·승인을 대신하는 플래그나 환경변수는 신설되지 않는다(SHALL NOT — §0.1: 실계좌 mutation 직전에는 반드시 사람의 행위 하나가 있어야 하며, 스케줄러나 에이전트의 동작은 그 행위가 아니다). 승인 이후의 레일 — 계획 인가(목록 밖 요청은 전송되지 않고 실행이 멈춘다)·수량 상한·즉시 취소·프로세스 경계 — 는 무변경이다(SHALL).

승인 증거 기록은 실제 승인 채널을 말해야 한다(SHALL — 클릭 승인 run이 "타이핑된 문자열"로 기록되어서는 안 된다(SHALL NOT)). CLI TTY 경로(`tossctl verify run`)의 타이핑 확인과 자동화 플래그 부재는 이 요구의 영향을 받지 않는다(SHALL — 콘솔 밖에서 자동화가 승인을 흉내낼 수 없어야 한다는 요구는 그대로다).

#### Scenario: 클릭 승인

- **WHEN** 계획을 표시한 승인 화면에서 세션과 CSRF를 갖춘 승인 POST가 도착하면
- **THEN** 배치가 승인되고 실행이 계획된 요청만 진행하며, 확인 문자열 입력은 요구되지 않는다

#### Scenario: CSRF 없는 승인 시도

- **WHEN** CSRF 토큰 없이 승인 POST가 도착하면
- **THEN** 승인이 거부되고 아무것도 전송되지 않는다

#### Scenario: 만료된 승인 창

- **WHEN** 승인 창이 만료된 뒤 승인 POST가 도착하면
- **THEN** 승인이 거부되고 아무것도 전송되지 않으며, 만료가 사유로 안내된다

#### Scenario: 화면 문구와 승인 방식의 일치

- **WHEN** 콘솔 승인 화면이 렌더되면
- **THEN** 계획 목록은 터미널과 같은 원천으로 표시되고, 확인 문자열과 타이핑 지시는 화면에 나타나지 않는다

#### Scenario: 승인 채널의 기록

- **WHEN** 콘솔 클릭으로 승인된 run의 승인 엔트리를 읽으면
- **THEN** 승인 채널이 클릭 승인으로 기록되어 있고, 계획 digest·요청 수·단계 목록은 종전과 같이 남는다

### Requirement: 검증 시작 화면의 무동작 방지

검증 시작 화면은 **아무 단계도 실행하지 않을 동작을 기본 동작으로 제시해서는 안 된다**(SHALL NOT — 2026-07-27 실측: 판정이 모두 terminal인 기록에서 [이어하기]가 두 번 눌려 `0 step(s) recorded`로 끝났고 장중 측정 창이 소모됐다). 이어하기의 대상(판정이 없는 단계)이 비어 있으면 이어하기는 비활성으로 그 사유와 함께 표시되어야 하고(SHALL), 재측정 대상(판정이 `fail`·`skipped`인 단계)이 있으면 재측정이 기본 동작으로 제시되어야 한다(SHALL). 두 대상 집합은 계속 **증거 기록**에서 계산되며 폼 입력에서 받지 않는다(SHALL — 이미 `pass`인 단계를 재측정 대상으로 이름 붙일 수 있는 경로를 만들지 않는다).

#### Scenario: 이어할 단계가 없는 기록

- **WHEN** 모든 단계의 판정이 terminal인 기록으로 검증 화면을 열면
- **THEN** [이어하기]는 비활성이고 사유가 표시되며, 재측정 대상이 있으면 재측정이 기본 동작으로 제시된다

#### Scenario: 이어할 단계가 있는 기록

- **WHEN** 판정이 없는 단계가 남아 있는 기록으로 검증 화면을 열면
- **THEN** [이어하기]가 활성으로 제시된다

### Requirement: 시장별 검증 화면

검증 능력은 (계좌, 시장)의 속성이므로 콘솔의 검증 화면은 **시장을 선택**할 수 있어야 한다(SHALL — 지원 시장은 KR·US). 선택된 시장의 증거 기록만으로 진행률·남은 단계·재측정 대상·단계 목록·리포트를 렌더해야 하며(SHALL), 한 시장의 판정이 다른 시장의 단계를 완료로 만들어서는 안 된다(SHALL NOT — 측정하지 않은 시장의 능력을 측정된 것으로 표시하는 것은 이 화면이 만들 수 있는 가장 나쁜 오류다). 시장을 지정하지 않은 접근은 KR로 해석한다(SHALL — 기존 동작 보존).

검증 실행은 그 시장의 심볼에만 주문을 낸다(SHALL). run의 시장과 다른 시장의 심볼을 쓰는 주문 단계는 사유와 함께 건너뛴다(SHALL — 다른 시장의 규칙으로 시험하지 않는다).

장시간 안내는 시장별로 표시하되 **관측된 것과 관측되지 않은 것을 구분해야 한다**(SHALL): KR은 실측된 휴장 응답 코드를 인용하고, 아직 휴장 응답이 관측되지 않은 시장은 미측정임을 명시한다(SHALL NOT — 한 시장의 실측을 다른 시장의 근거로 제시하지 않는다). 어느 시장에서도 안내가 시작을 차단해서는 안 된다(SHALL NOT — 기존 advisory 규율 유지).

승인 형식은 시장과 무관하게 동일하다(SHALL — 계획 목록을 표시한 화면에서의 클릭 1회). 한 프로세스가 검증을 한 번만 수행한다는 규칙도 시장과 무관하게 유지되므로(SHALL), 두 시장을 이어서 측정하려면 프로세스 재시작이 필요함을 화면이 안내한다(SHALL).

#### Scenario: 시장 선택

- **WHEN** 검증 화면을 US로 열면
- **THEN** US 기록의 진행률·재측정 대상·단계 목록이 표시되고, KR 기록의 판정은 그 화면에 나타나지 않는다

#### Scenario: 한 시장의 판정이 다른 시장을 오염시키지 않음

- **WHEN** 한 시장에서 모든 단계가 통과한 뒤 다른 시장의 검증 화면을 열면
- **THEN** 다른 시장의 단계는 여전히 미측정로 표시되고 시작이 제공된다

#### Scenario: run의 시장과 다른 심볼

- **WHEN** run의 시장과 다른 시장의 심볼을 쓰는 주문 단계에 도달하면
- **THEN** 그 단계는 사유와 함께 건너뛰어지고 주문은 전송되지 않는다

#### Scenario: 실측 없는 시장의 장시간 안내

- **WHEN** 휴장 응답이 아직 관측되지 않은 시장에서 정규장 밖에 검증 화면을 열면
- **THEN** 안내는 미측정임을 명시하고, 다른 시장의 실측 코드를 그 시장의 근거로 제시하지 않으며, 시작을 막지 않는다

### Requirement: 운영 개요 가시성

콘솔은 운영자가 지금 내려야 하는 판단에 필요한 상태를 한 화면에 모아 read-only로 표시해야 한다(SHALL — `/dashboard`, GET, 세션 게이트 안·CSRF 게이트 밖). 기존 `/`(검증 콘솔)는 이 화면으로 대체되거나 리다이렉트되지 않는다(SHALL NOT — 검증 승인 창을 보고 있는 탭이 다른 화면으로 갈아타서는 안 된다). 화면이 답하는 것은 여섯이다(SHALL): ① 엔진 실행 여부와 기동 거부 사유, ② 계좌 평가액·평가손익·보유 종목 수(관리/관리 외 구분), ③ 오늘 실현손익·왕복 건수·승패, ④ 살아 있는 주문 건수, ⑤ 검증 상태·잔여물·Guardian 한도, ⑥ 최근 exit 이벤트. ②와 ③은 **시장별(KR/US)로 나누어** 표시하고 시장을 가로지르는 합계를 만들지 않는다(SHALL NOT — 보유 값은 시장마다 다른 통화이고 그것을 더한 숫자는 아무 뜻도 없다; 오늘의 합산 한 줄은 한 시장의 손실을 다른 시장의 이익이 가리는 것을 허용한다). "오늘"의 경계는 **시장별 현지 자정**이며 화면은 **어느 경계를 썼는지 출력한다**(SHALL — 원장의 청산 시각은 UTC이고 UTC 자정은 KR 장 시작 한 시간 뒤에 떨어진다). 이 화면은 브로커를 호출하지 않는다(SHALL NOT — 캐시를 갱신 없이 읽기만 한다). 캐시가 한 번도 채워지지 않았으면 0이 아니라 미측정으로 렌더하고 **그 값을 채우는 화면으로 가는 링크를 함께 제공한다**(SHALL — 링크가 없으면 그 숫자는 영원히 비어 있을 수 있다). 한 출처가 없거나 읽히지 않아도 나머지 패널은 계속 렌더된다(SHALL). 값을 얻지 못한 항목은 0이 아니라 미측정과 **사유**를 표시하며, 사유는 검증 중 갱신 보류·브로커 조회 실패·원장 미판독·seam 미배선·미조회·설정 판독 불가·원장 값 해석 불가·발굴 저장소 판독 불가 여덟을 구분한다(SHALL — 사유 없는 "—"는 기다릴지 고칠지 배선할지 알려주지 않는다; 열거는 운영자가 무엇을 고쳐야 하는가를 남김없이 적는 것이므로, 자유 문장으로만 존재하는 사유는 셀 수도 없음을 테스트할 수도 없다). 값이 없는 항목을 화면에서 조용히 생략하지 않는다(SHALL NOT — 라벨 없이 사라지는 것은 미측정 표시가 아니라 미측정의 은폐다). Guardian 한도는 **읽을 수 있는 값만** 표시하고, 오늘 소진분은 실현손익 대 일일 손실 한도 **한 축만** 산출하며 나머지 축은 구조적으로 미측정이라고 명시한다. 한도 통화와 시장 통화가 다르면 두 숫자를 나란히 보이되 비율을 내지 않고 그 이유를 표시한다(SHALL NOT — 환산해서 비율을 만드는 것은 시장을 가로지르는 합계 금지를 한 칸 옆에서 어기는 것이다). 게이트가 열려 있는지 여부는 이 화면이 판정하지 않는다(SHALL NOT — 콘솔은 엔진 프로세스에 묻고 그 답을 표시할 뿐이며, ⑤가 요구하는 것은 한도이지 게이트 상태가 아니다)(SHALL NOT — 소진분은 엔진이 in-process로 계산하는 값이고 원장에 기록이 없다; 지어낸 한도 소진분은 있지도 않은 보호를 믿게 만든다). 잔여물 표시는 **모든 시장의 검증 기록**을 읽는다(SHALL — 한 시장만 읽는 잔여물 패널은 다른 시장의 잔여물을 감추는데 그것이 이 패널이 존재하는 이유다). 이 화면에는 상태를 바꾸는 수단이 없다(SHALL NOT — 확인 문자열 타이핑·2단계 클릭 등 어떤 확인 마찰도 두지 않는다; 누를 것이 없으므로 확인할 것도 없다).

#### Scenario: 엔진이 꺼져 있는 상태의 개요
- **WHEN** 엔진이 실행 중이 아닌 상태에서 개요 화면을 열면
- **THEN** 엔진 패널은 미실행과 기동 거부 사유(이 콘솔 프로세스가 아는 경우)를 표시하고, 나머지 패널은 정상 렌더된다

#### Scenario: 브로커 캐시가 비어 있는 개요
- **WHEN** 브로커 캐시가 한 번도 채워지지 않은 상태에서 개요 화면을 열면
- **THEN** 브로커 호출이 한 건도 발생하지 않고, 계좌 패널은 0이 아니라 "아직 읽지 않음"과 그 값을 채우는 화면으로 가는 링크를 렌더한다

#### Scenario: 원장을 읽지 못한 상태의 개요
- **WHEN** journal을 열 수 없는 상태에서 개요 화면을 열면
- **THEN** 오늘 패널과 최근 이벤트 패널은 0이 아니라 "원장 미판독"으로 렌더되고, 엔진·검증 패널은 계속 동작한다

#### Scenario: 배선되지 않은 seam의 패널
- **WHEN** 주문 조회 seam이 배선되지 않은 빌드에서 개요 화면을 열면
- **THEN** 미체결 패널은 0건이 아니라 "seam 미배선"으로 렌더된다

#### Scenario: 두 시장의 오늘
- **WHEN** KR과 US 양쪽에 오늘 청산된 왕복이 있는 상태에서 개요 화면을 열면
- **THEN** 실현손익·건수·승패가 시장별로 각각 표시되고, 두 시장을 가로지르는 합계는 표시되지 않으며, 각 시장이 어느 자정을 경계로 썼는지가 화면에 나타난다

#### Scenario: 거래가 없는 시장
- **WHEN** 한 시장에 오늘 청산된 왕복이 하나도 없으면
- **THEN** 그 시장의 행은 실현손익 0이 아니라 "거래 없음"으로 렌더된다

#### Scenario: 읽을 수 없는 Guardian 한도
- **WHEN** config에서 한도를 읽을 수 없으면
- **THEN** 그 한도는 0도 "무제한"도 아닌 미측정으로 렌더된다

#### Scenario: 잔여물이 있는 상태의 개요
- **WHEN** 직전 검증이 어느 시장에든 잔여물을 남긴 상태에서 개요 화면을 열면
- **THEN** 잔여물의 종류·식별자·심볼이 시장과 함께 표시되고, 그것이 다음 검증의 노출 상한을 먹는다는 사실이 명시된다

#### Scenario: 개요 화면에 행위가 없다
- **WHEN** 개요 화면의 렌더 결과를 검사하면
- **THEN** POST 폼이 하나도 없고, 확인 문자열 입력란이 없다

#### Scenario: `/`는 그대로다
- **WHEN** 개요 화면이 추가된 뒤 `/`를 열면
- **THEN** 기존 검증 콘솔이 그대로 렌더되고 리다이렉트되지 않는다

### Requirement: 주문 가시성

콘솔은 계좌의 주문을 read-only로 표시해야 한다(SHALL — `/orders`, GET, 세션 게이트 안·읽기 전용 wrapper 적용·CSRF 게이트 밖). 표시 항목은 주문 시각·심볼·시장·매수/매도·상태·주문수량·체결수량·주문가·평균체결가·**주문번호**·발주 주체다(SHALL). 브로커가 주지 않은 값은 0이 아니라 "—"로 렌더한다(SHALL NOT — 0으로 채우면 아직 체결되지 않은 주문이 전량 체결된 주문과 같게 보인다; API는 시장가 주문의 가격과 미체결 주문의 체결 정보 전체를 null로 보내므로 이것은 예외가 아니라 평시다). 이를 위해 주문 값은 **브로커의 원문 문자열을 보존하는 읽기**로 가져온다(SHALL — float64를 지난 값에는 부재와 0의 차이가 남아 있지 않다). 종목명은 이 화면에 두지 않는다(SHALL NOT — 주문 응답에 그 값이 없어 전 행이 "—"가 되며, 어디에도 없는 값을 위한 열은 정보가 아니라 소음이다). 미체결 주문과 종결된 주문은 구분해 표시한다(SHALL). 미체결 조회는 **브로커가 전량 반환을 보장하는 호출 형태**를 써야 한다(SHALL — 페이지 경계에 걸려 잘릴 수 있는 조회로 미체결을 세면 101번째의 살아 있는 주문이 화면에서 사라지고, 그것이 이 화면이 존재하는 이유다). 목록을 **가져오지 못한 상태는 0건과 구분해야 한다**(SHALL — 자료형에서 구분하며 불리언 하나로 뭉치지 않는다). 페이지가 잘렸으면 건수를 숫자로 단정하지 않는다(SHALL NOT — 다음 페이지가 있으면 "N건 이상"이다). 각 주문이 엔진이 낸 것인지는 원장의 주문 시도 기록 조인으로 판정하고(SHALL), 원장을 읽지 못했을 때는 "그 밖"이 아니라 **"불명"**으로 표시한다(SHALL NOT — 원장 미판독을 수동 주문으로 읽으면 엔진이 아무 일도 안 한 것처럼 보인다). 원장 미판독 사유는 페이지 수준 안내 1회로 말한다(SHALL). 원장의 읽기 전용 핸들에 이 조인을 위한 접근자를 더할 때 그 테이블을 필수 테이블 목록에도 등록한다(SHALL — 등록하지 않으면 열기는 성공하고 질의만 실패하며, 그 실패는 0행으로 돌아와 "전부 수동 주문"으로 읽힌다). 시장·방향·상태 필터를 제공할 수 있으며 GET 쿼리 파라미터와 링크로만 구현한다(SHALL — 이 콘솔에는 JavaScript가 없다). 필터가 적용된 화면은 필터 후 건수와 전체 건수를 함께 표시하고(SHALL), 목록이 미측정이면 필터를 작동시키지 않는다(SHALL NOT — "0/—건"은 "0건이 일치"로 읽힌다). 이 화면에는 주문을 내거나 정정·취소하는 수단이 없다(SHALL NOT — 확인 문자열 타이핑 등 어떤 확인 마찰도 두지 않는다).

#### Scenario: 미체결 주문 표시
- **WHEN** 브로커에 살아 있는 주문이 있는 상태에서 주문 화면을 열면
- **THEN** 각 주문의 시각·심볼·시장·매수/매도·상태·주문수량·체결수량·주문가·평균체결가·주문번호가 표시되고, 미체결과 종결이 구분된다

#### Scenario: 아직 체결되지 않은 주문의 평균체결가
- **WHEN** 브로커가 체결 정보를 null로 보낸 미체결 주문을 렌더하면
- **THEN** 평균체결가와 체결수량은 0이 아니라 "—"로 표시된다

#### Scenario: 빈 목록과 못 가져온 목록
- **WHEN** 주문 조회가 실패해 목록을 얻지 못한 상태에서 주문 화면을 열면
- **THEN** 화면은 "주문 없음"이 아니라 미측정과 그 사유를 표시하고, 필터는 작동하지 않는다

#### Scenario: 잘린 페이지
- **WHEN** 브로커가 다음 페이지가 있음을 알리면
- **THEN** 건수는 확정된 숫자가 아니라 "N건 이상"으로 표시된다

#### Scenario: 조건주문이 살아 있는 상태의 미체결
- **WHEN** 일반 주문은 하나도 없고 조건주문 잔여물만 살아 있으면
- **THEN** 미체결이 0건으로 표시되지 않고 조건주문이 건수에 포함되어 표시된다

#### Scenario: 원장을 읽지 못한 상태의 발주 주체
- **WHEN** journal을 열 수 없는 상태에서 주문 화면을 열면
- **THEN** 모든 행의 발주 주체는 "불명"이며 "그 밖"으로 표시되지 않고, 원장 미판독 안내가 페이지에 1회 나타난다

#### Scenario: 주문 시도 테이블이 없는 원장
- **WHEN** 주문 시도 기록 테이블이 없는 원장을 읽기 전용으로 열면
- **THEN** 열기 단계에서 한 번 명확히 거부되고, 질의마다 하나씩 실패하며 0행을 돌려주지 않는다

#### Scenario: 주문 화면의 행위 부재
- **WHEN** 주문 화면의 렌더 결과와 라우트 표를 검사하면
- **THEN** 주문을 내거나 정정·취소하는 폼·링크·라우트가 없고, 확인 문자열 입력란이 없다

### Requirement: Guardian 한도 설정 화면

콘솔은 Guardian 한도(주문 수량·주문 notional·총 개방 노출·일일 손실 절대액·일일 손실 자본 비율)와 한도 통화를 표시·편집하는 표면을 제공해야 한다(SHALL — 대상은 `engine.automation_gate`의 그 여섯 값뿐이며, `enabled`·`attestation_file`은 표시만 하고 쓰지 않는다). 편집의 1차 경로는 **티어 프리셋 적용**이고 확인 대화상자 1회 외에 어떤 입력도 요구하지 않는다(SHALL — 사용자 결정 2026-07-30 "클릭 한번으로"; 타이핑 확인·추가 승인 마찰 금지 — 사용자 결정 2026-07-27). 개별 값 직접 기입은 고급 접힘 안에만 둔다(SHALL — 편입 설정 화면의 목록 기입과 같은 형태). 티어 레지스트리는 각 수치의 출처를 코드에 기록해야 하며(SHALL — 범주와 자격은 risk-management "정책 수치의 provenance"가 정한다: StockOS 파일·심볼, 또는 `measurements.md`의 관측 식별자로 인용된 TossOS 실측), 화면은 각 프리셋의 통화와 다섯 값을 적용 전에 보여준다(SHALL). 저장은 서버측에서 기동 인터록과 동일한 규칙으로 재검증한다(SHALL — 다섯 값 전부 양수·유한, 비율은 (0,1], 통화 비지 않음; 하나라도 위반이면 저장을 거부하고 사유를 표시한다 — 엔진이 기동 거부할 블록을 기록하지 않는다). 어떤 저장도 등록된 티어 상한을 넘겨서는 안 된다(SHALL NOT — 상한은 그 통화에 등록된 모든 티어의 필드별 최대이고, 미등록 통화는 fail-closed로 거부한다). 낮추는 방향은 양수인 한 제한하지 않는다(SHALL — 보수 방향). 화면은 암묵 기본값을 표시해서는 안 된다(SHALL NOT — 한도가 파일에 없으면 미설정이라고 말한다; 파일에 없는 숫자를 화면이 그리면 엔진의 인터록이 보는 것과 다른 상태를 보여주게 된다). 기본 티어는 **권장 프리셋**으로 제시하고 적용은 그 값을 파일에 명시적으로 기록한다(SHALL — 권장은 승인된 보수 기본값 집합의 통화에 둔다; 게이트 통화가 하나이므로 다른 통화를 권장하면 화면을 처음 여는 운영자에게 한쪽 시장이 닫히는 쪽을 권하게 된다). 화면은 현재 다섯 값이 등록된 티어 중 하나와 정확히 일치하는지를 표시한다(SHALL — 레지스트리와의 대조 결과이며, 운영자가 그 티어를 골랐다는 주장이 아니다). 한도 통화를 바꾸는 적용은 그 귀결을 명시해야 한다(SHALL — 게이트의 한도 통화는 하나이고 Guardian 체인이 통화 불일치 진입을 거부하므로, 통화를 바꾸면 다른 시장의 자동 진입이 닫힌다). 반영 시점은 편입 설정과 같은 규율로 명시한다(SHALL — 가동 중 엔진은 기동 시점 설정으로 동작하며 현재형 보장을 쓰지 않는다). 저장은 변경 전후 값·시각·주체를 audit에 기록한다(SHALL — engine-safety "게이트 토글·한도 변경 등 운영 설정 변경").

#### Scenario: 프리셋 클릭 한 번

- **WHEN** 한도 프리셋을 적용하면
- **THEN** 다섯 한도와 한도 통화가 그 티어의 값으로 한 번에 저장되고, `enabled`는 변경되지 않으며, 반영 시점이 안내된다

#### Scenario: 상한 초과 저장 거부

- **WHEN** 등록된 티어의 필드별 최대를 넘는 값을 개별 기입으로 저장하려 하면
- **THEN** 저장이 거부되고 어느 필드가 어느 상한을 넘었는지 표시되며 config는 변경되지 않는다

#### Scenario: 인터록이 거부할 블록의 저장 거부

- **WHEN** 다섯 한도 중 하나가 0·음수·비유한이거나 비율이 1을 넘거나 통화가 빈 채로 저장하려 하면
- **THEN** 저장이 거부되고 사유가 표시되며 config는 변경되지 않는다

#### Scenario: 사용자 미확정 상태

- **WHEN** 게이트 블록이 다섯 한도를 싣지 않은 상태로 한도 화면을 열면
- **THEN** 다섯 칸은 미설정으로 표시되고, 기본 티어가 권장 프리셋으로 제시되며, 화면은 암묵 기본값을 적용된 것처럼 그리지 않는다

#### Scenario: 부분 설정에는 나머지를 발명하지 않는다

- **WHEN** 다섯 한도 중 일부만 적힌 게이트 블록을 읽으면
- **THEN** 나머지는 채워지지 않고 미설정으로 남으며, 화면은 인터록이 그 상태를 거부한다는 것과 프리셋으로 고칠 수 있음을 안내한다

#### Scenario: 현재 값과 티어의 대조

- **WHEN** 다섯 한도와 통화가 등록된 티어와 정확히 일치하는 상태로 화면을 열면
- **THEN** 그 티어의 이름이 표시되고, 하나라도 다르면 사용자 지정값으로 표시된다

#### Scenario: 통화를 바꾸는 프리셋

- **WHEN** 현재 한도 통화와 다른 통화의 프리셋을 적용하면
- **THEN** 저장은 수행되고, 응답은 그 통화가 아닌 시장의 자동 진입이 통화 불일치로 거부된다는 사실을 명시한다

#### Scenario: 실측 출처 티어의 제시

- **WHEN** StockOS 대응 심볼이 없는 티어를 프리셋으로 제시하면
- **THEN** 그 티어도 통화와 다섯 값을 적용 전에 보여주고, 코드에는 실측 식별자가 출처로 남아 있다

### Requirement: 최적화 화면은 공통 청산 정책을 설명한다
콘솔은 `/optimization`에서 현재 승인값, 권장값, 각 등록 정책의 목표·보호선·부분익절·최종 runner 의미를 한국어로 표시해야 한다 (SHALL).

#### Scenario: 미승인 화면
- **WHEN** 공통 정책 설정이 비어 있는 상태에서 인증된 운영자가 `/optimization`을 연다
- **THEN** 현재 동작은 기존 RATCHET이고 HYBRID_50은 권장일 뿐 아직 적용되지 않았다고 표시한다

#### Scenario: 외부 구매 적용 설명
- **WHEN** 운영자가 정책 카드를 본다
- **THEN** 신규 자체 포지션과 향후 편입 외부 매수분에 적용되고 기존 활성 포지션은 바뀌지 않는다고 표시한다

### Requirement: 정책 변경은 세션과 CSRF를 통과한 사람의 POST다
최적화 정책 저장 route는 기존 `session0(mutating(...))` 체인을 사용하고 GET, 무세션, 잘못된 CSRF 요청을 거부해야 한다 (SHALL).

#### Scenario: 정상 저장
- **WHEN** 유효 세션과 CSRF를 가진 운영자가 등록 policy ID를 POST한다
- **THEN** config seam을 정확히 한 번 호출하고 `/optimization`으로 결과 안내와 함께 redirect한다

#### Scenario: CSRF 누락
- **WHEN** 세션은 있지만 CSRF가 없는 정책 저장 요청이 도착한다
- **THEN** 403을 반환하고 config, audit, broker state를 변경하지 않는다

### Requirement: 최적화 화면은 최소 권한 seam만 가진다
최적화 handler가 받는 설정 seam은 공통 exit policy의 load/save만 제공해야 하며 주문, gate, trading toggle, journal mutation capability를 제공해서는 안 된다 (MUST NOT).

#### Scenario: 정적 capability 검사
- **WHEN** console Options와 최적화 handler의 dependency closure를 검사한다
- **THEN** exit-policy 설정 이외의 mutation capability와 account mutation verb가 존재하지 않는다

### Requirement: 콘솔의 엔진 자동 시작 승인

운영 콘솔은 인증된 운영자가 `engine.autostart`를 ON/OFF로 저장할 수 있어야 한다(SHALL).
이 상태변경 라우트는 세션+CSRF 뒤에 있어야 하고(SHALL), 한 키만
기록하며 변경 전후 값·시각·주체를 audit에 남겨야 한다(SHALL). ON 저장이 성공하면
콘솔은 기존 엔진 시작 seam을 한 번 호출해야 하며(SHALL), 성공·이미 실행·startup
interlock 거부 결과를 화면에 표시해야 한다(SHALL). OFF 저장은 다음 프로세스
기동의 자동 시작만 막고 실행 중 엔진을 정지해서는 안 된다(SHALL NOT).

#### Scenario: 자동 시작 ON 저장
- **WHEN** 인증된 운영자가 유효한 CSRF와 함께 엔진 자동 시작을 ON으로 저장하면
- **THEN** `engine.autostart` 한 키가 true로 기록되고 audit가 추가되며 기존 엔진 시작 seam이 정확히 한 번 호출된다

#### Scenario: 기동 인터록 거부
- **WHEN** 자동 시작 ON 저장 뒤 기존 엔진 시작 seam이 startup interlock 오류를 반환하면
- **THEN** 설정은 ON으로 남고 엔진 자신의 거부 사유가 화면에 표시되며 인터록을 우회하는 두 번째 시작 경로는 실행되지 않는다

#### Scenario: 자동 시작 OFF 저장
- **WHEN** 실행 중 엔진이 있는 상태에서 자동 시작을 OFF로 저장하면
- **THEN** `engine.autostart`는 false가 되고 자동 정지 호출은 발생하지 않으며 화면은 [엔진 정지]를 별도 사용하라고 알린다

#### Scenario: CSRF 없는 자동 시작 변경
- **WHEN** 세션은 유효하지만 CSRF가 없거나 틀린 자동 시작 저장 요청이 도달하면
- **THEN** 요청이 거부되고 config·audit·엔진 프로세스 상태가 모두 변경되지 않는다

#### Scenario: 설정 키 격리
- **WHEN** 콘솔에서 엔진 자동 시작을 저장하면
- **THEN** `engine.autostart` 이외의 automation gate·Guardian 한도·거래 정책 바이트는 그대로다

### Requirement: 부팅 서베이는 엔진의 준비 신호를 유한하게 기다린다

자동 시작 서베이는 엔진이 재시작 복구를 완료했다는 관측 가능한 신호를 유한 시간 기다린 뒤 시작해야 한다(SHALL). 그 신호는 엔진이 복구를 완료한 뒤에만 만들어져야 하며(SHALL), 프로세스 존재나 lock 보유는 그 신호가 아니다 — 둘 다 복구보다 먼저 생긴다.

대기는 다음 세 경우에 즉시 끝나야 한다(SHALL): ① 신호가 관측됨 ② 살아 있는 엔진이 없음 — 죽은 엔진을 기다리는 것은 서베이를 이유 없이 미루는 것이다 ③ 콘솔 종료. 상한 도달 시 서베이는 그냥 시작해야 하며(SHALL — 서베이는 선택 기계장치이고 attestation 시계를 계속 세워야 한다), 어느 경우였는지를 기동 노트가 말해야 한다(SHALL). 조용한 상한 초과는 금지다(SHALL NOT).

이 대기는 운영자 콘솔 화면을 지연시켜서는 안 된다(SHALL NOT — a101: 서베이의 어떤 사정도 운영자 화면이 없는 이유가 되어서는 안 된다).

#### Scenario: 엔진이 준비를 알리면 서베이가 시작된다

- **WHEN** 자동 시작 엔진이 재시작 복구를 완료하고 준비 신호를 만들면
- **THEN** 서베이는 그 관측 후 시작되고, 노트는 준비 확인 후 시작했다고 말한다

#### Scenario: 엔진이 없으면 기다리지 않는다

- **WHEN** 서베이 자동 시작 시점에 살아 있는 엔진이 관측되지 않으면
- **THEN** 서베이는 대기 없이 시작된다 — 오늘의 동작 그대로다

#### Scenario: 상한 도달은 조용하지 않다

- **WHEN** 엔진이 살아 있으나 상한 시간 안에 준비 신호가 나타나지 않으면
- **THEN** 서베이는 시작되고, 노트는 상한 초과 후 시작했다고 명시한다

#### Scenario: 대기가 콘솔을 막지 않는다

- **WHEN** 자동 시작 서베이가 엔진 준비를 기다리는 동안
- **THEN** 운영자 콘솔 화면은 그 대기와 무관하게 뜬다
