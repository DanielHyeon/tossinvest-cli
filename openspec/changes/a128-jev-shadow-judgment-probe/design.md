# a128 설계 — Jev shadow 프로브

## 비례 원칙 적용

새 코드가 새 파일(`tools/a128-jev-shadow-probe/`)에만 들어가고 High-risk 경로(주문·손절·
사이징·Guardian·원장·대사·인증·체결)를 만지지 않으므로, Function Logic Map·변이 원장·
다중 적대 리뷰는 `not-applicable`(기존 함수 편집 0, 게이트 판정 변경 0 — docs/WORKFLOW.md
「비례 원칙」). 기본 = 시험 + 경량 리뷰(gstack).

## 데이터 소스 — 2026-10-07 실측 근거 (문서가 아니라 호출로 확인)

| 소스 | 표면 | 상태 |
|---|---|---|
| 랭킹 `market rankings`(MARKET_TRADING_AMOUNT 등) | official | 작동 확인(100행 수신) |
| 시세 `/prices`·호가 `/orderbook` | official | a112 L1c 프로브 선례(leader 재사용 가능) |
| 종목 메타 `/api/v1/stocks` | official | 작동 확인(5심볼 이름 수신) |
| 뉴스 `market news` | **WTS-only** | 세션 만료 관측 — 사람 `auth extend` 후에만, 없어도 동작 (1.8: 배선, 2026-10-15 세션 유효·피드 48+5 건 수신) |
| 종목 수급 `trading-trend`·주요 지수 `overview/indicator/index` | **WTS-only** | 1.8 배선, 2026-10-15 dry run 수신(종목 5/5·지수 6 종) |

주의: official 호출에 rate limit 간헐(실측 2026-10-07, 45초 대기로 해소) — 수집 루프는
호출 간 간격과 429/limited 재시도(지수 백오프, 포기 시 그 샘플만 결손 기록)를 가진다.

## TypeSafe 연동

- HTTP API 직접 호출(stdlib `net/http`) — Go SDK 없음, SDK 의존 추가 안 함(YAGNI).
- 구현 전에 https://docs.typesafe.ai/api.md 와 https://docs.typesafe.ai/primitives/noul.md 를
  읽고 요청/응답 계약을 **그 문서 원문 인용으로** `analysis/typesafe-api-contract.md` 에
  동결한다(계약 숫자·문장은 영수증에서 — 추측 금지). 엔드포인트·필드명을 이 설계가
  미리 지정하지 않는 이유: 아직 읽지 않은 계약을 지어내지 않기 위함.
- 인증: `TYPESAFE_API_KEY` env 만. 키가 없으면 프로브는 시작을 거부하고 사유를 말한다.

## 판단 설계

- J1(방향 가설, Noul): state = {market, symbol, name, 최근 체결가, 등락률, 호가 스프레드,
  거래대금 순위, (있으면) 뉴스 헤드라인 목록}. 질문: "이 맥락에서 이 종목이 다음 h분 안에
  기준가 대비 양(+)수익으로 마감할 가능성" — **반증 대상 가설**: calibration 에서 p 구간과
  실현 적중률의 관계가 평평하면 예측력 없음으로 기록하고 J1 은 배선 후보에서 제외.
- J2(악재 veto, Noul): 같은 state. 질문: "이 맥락에 매수를 보류해야 할 명백한 악재
  신호가 있는가". 뉴스 필드가 없으면 state 에 `news: absent` 를 명시해 모델이 결손을
  알게 한다(조용한 생략 금지 — 결손을 숨기면 p 의 의미가 달라진다).
- 질문 원문·버전은 원장 행에 박는다(질문이 바뀌면 p 의 모집단이 달라진다 —
  측정은 잰 순간을 달고 다닌다).

## 원장

- JSONL, append-only, 행 2종: `judgment`(시각 RFC3339·market·symbol·question_id·
  question_rev·state_digest(sha256)·p·기준가·수집 결손 목록), `label`(judgment 참조 id·
  라벨 시각·실현가·실현 수익률·라벨 결손 사유).
- 위치: `openspec/changes/a128-jev-shadow-judgment-probe/analysis/shadow-ledger/`
  (장중 산출물 — 2일 측정 종료 후 커밋; 계좌 데이터가 없으므로 커밋 가능).
- state 원문은 원장에 넣지 않고 digest 만 — 원문은 같은 디렉터리의 state 파일로
  별도 보존(재검증용). 어느 파일에도 계좌 식별자·토큰이 없음을 시험이 강제한다.

## 세션 경계

- KR 09:00~15:30 KST, US 22:30~05:00 KST(당일 영수증 기준, DST 는 market hours 재조회로).
  프로브는 `market hours` 를 시작 시 읽어 세션 밖이면 대기하거나 종료한다(플래그).
- 2일 = 달력 2일의 KR·US 세션 각 2회. 운영: 에이전트가 세션마다 실행(read-only 라 가능),
  장시간 실행은 백그라운드 + 로그 파일.
- 달력 우회(`--session "HH:MM-HH:MM"`): official market-calendar 가 틀릴 때만 쓰는 수동 수단
  (2026-10-12 KR 개장일을 휴장으로 오응답한 실측 대응). 지정하면 달력을 전혀 호출하지 않고
  시장 현지 시각(KR=Asia/Seoul·US=America/New_York)으로 오늘 세션창을 만든다. 명시적 우회이므로
  침묵 생략 금지 — state `session_source="manual-override"` 와 원장 gap 행(stage=session)에 남겨
  보고서가 로트의 혼입을 가려낸다. 파싱 실패·역전창은 측정 시작 전에 거부. 판단 종료 규칙
  (judgment+horizon ≤ close)·라벨 창은 수동 close 를 그대로 쓴다.

## Jev 선별과 대조군 (tasks 1.6, 2026-10-14)

- 상한 이하 상위 100 후보에서 무작위 풀 30 → 풀 전 종목 J1·J2 → J2 p ≥ 0.5 제외 → J1 p 내림차순(동률 랭킹 순) 상위 10 = 결선.
  shadow 측정이므로 결선은 「주문했을 종목」 표시일 뿐 어떤 발주 경로에도 이어지지 않는다.
- **대조군**: 결선만 적으면 「선별이 효과가 있었나」를 잴 기준이 없다. 그래서 풀 전체의 판단을 쓰고 라벨링하며, 행마다
  `screen.finalist` 를 남겨 보고서가 같은 사이클·같은 질문의 비결선과 결선의 J1 실현 적중을 나란히 비교한다(무작위 풀이라
  두 집단의 차이는 선별 규칙에서 온다). 결선 표시는 판단이 다 끝나야 정해지므로 판단 행은 사이클 끝에 한꺼번에 쓴다.
- 비용: 사이클당 Jev 요청과 호가 GET 이 10 → 30(3배). 질문·horizon 은 그대로라 p 의 모집단 정의는 바뀌지 않고, 표본 규칙만
  새 로트(`jev_screen10_of_random30_of_top100_le_order_cap`)가 된다.

## Jev 가 보는 state (s2, tasks 1.7, 2026-10-14)

- 이유: s1 state 는 한 시점의 가격·등락·랭킹·호가 맨 위 한 줄뿐이라, J1(h분 뒤 상승?)이 가격 경로 없이 판단했다. s2 는 같은
  「공개 시장 데이터만」 원칙 안에서 경로(오늘 1분봉 30 개 + 5/15/30/60분 수익률·창 고저·VWAP 거리·거래량 가속), 일 단위 맥락
  (1/5/20일 수익률·20일 고저), 시장 맥락(KR 지수·시장 투자자 순매수), J2 의 재료(매수 유의·상하한가 거리)를 더한다.
- 정직성 규칙: 계산 못 하는 값은 0 이 아니라 생략(0 은 「변화 없음」으로 읽힘), 단위가 미확인인 값은 쓰지 않고 가격에서 계산,
  공식 표면에 없는 것(US 지수, 종목별 투자자 매매)은 지어내지 않고 `missing_inputs` 에 이름으로 남긴다.
- 비용(사이클당 GET): 묶음(`/stocks`·`/prices`·지수 시세 한 번) + 하루 캐시(일봉·가격제한폭·지수 일봉·시장 투자자 매매) +
  `/orderbook` 은 엄격 맨 위 한 줄 한 번(깊이 제외) + 체결 제외 → KR 정상 96 GET/사이클, US 94(지수 맥락 0 · 환율 1 · 가격제한폭 0). 엔진과 공유하는 Open API
  키이므로 사이클의 그룹별 GET 수와 수집 시간을 원장·로그·보고서에 남긴다. Jev 입력 토큰은 요청당 약 3.5 배로 추정(tasks 1.7) —
  첫 s2 사이클에서 실측해 다시 적는다.

## Jev 가 보는 state (s3, tasks 1.8, 2026-10-15)

- 이유: s2 판단 전부가 같은 결손 셋을 달고 있었음(원장 실측) — 뉴스(KR·US 100%), KR 종목별 투자자 매매(공식은 시장 단위뿐),
  US 시장 맥락(공식 카탈로그에 US 지수 없음). 셋 다 **기존 WTS 조회 함수**로 채움(사용자 승인 2026-10-15). WTS 는 조회 전용
  (주문 경로 0) — `wtsSource` 인터페이스 = `GetMarketNews`·`GetMarketIndices`·`GetTradingFlows` 셋뿐이고, 나머지 `*client.Client`
  공개 메서드 전부(주문·계좌·보유·관심종목 쓰기)는 AST 시험이 이름조차 금지(`internal/client` 소스에서 자동 열거). 뉴스 피드는
  POST 이지만 읽기 엔드포인트(대시보드 피드)임.
- 새 필드(`wts.go`):
  - `news` = `present`|`absent` + `news_headlines{symbol[], market[]}` — 피드 범위 `ALL_HIGHLIGHT`·`SOARING_STOCK`(계좌 맥락
    없는 범위만; 보유·관심·추천 개인화 범위는 계좌의 종목을 드러내므로 소스 문자열·요청 본문 양쪽 시험으로 금지). 사이클마다
    읽어 프로세스 안 보관소(24 시간)에 쌓고 종목 기사(관련 종목 코드 = `A`+심볼 또는 이름 = 공식 `/stocks` 이름; ≤3, 제목 ≤120자·
    요약 ≤120자)와 시장 기사(국적 = 시장, ≤5, 제목만)를 고름. **판단 시각보다 늦은 기사는 넣지 않음**(라벨이 재는 미래 차단).
    시각 없는 createdAt 은 Asia/Seoul 로 읽음(클라이언트 fixture 모양 `2006-01-02 15:04:05` — 가정, 첫 s3 사이클 로그의
    `newest raw` 를 벽시계와 대조할 것). 받은 피드의 기사가 전부 시각 해석 실패면 그 범위는 못 읽은 것으로 침(`news=present`·빈
    목록을 내보내지 않음 — dry run 에서 실제로 난 결함).
  - `investor_flows_per_stock`(KR) — 직전 거래일까지 최신 5 일의 개인·외국인·기관 순매수 **주 수**(WTS `*BuyVolume`). 오늘 행은
    장중 잠정치이고 하루 캐시가 그 값을 얼리므로 버림 → 캐시가 정확함. 날짜 하나라도 못 읽으면 전체 결손.
  - `market_context.major_indices`(양 시장) — 코스피·코스닥·나스닥·S&P 500·필라델피아 반도체·VIX(실측 6 종), `last` +
    `change_pct_vs_base`. base = 직전 종가(KR 두 지수에서 공식 지수 직전 종가 기준 등락과 소수 넷째 자리까지 일치 실측). US 는
    이것이 유일한 지수 맥락. `indices`·`investor_flows`(공식, KR)는 그대로이고 빈 목록이면 생략으로 바뀜(s3).
- 결손 정책: 각 소스 실패 → 그 블록 생략 + `missing_inputs` 이름(`news: <범위>: …`, `investor_flows_per_stock: …`,
  `market_context: major_indices: …`), 사이클·판단은 계속. 지어낸 0·빈 목록 없음(빈 목록 = 읽었고 0 건). 한 사이클에서 WTS 가
  세션을 거절(401/403)하면 그 사이클의 나머지 WTS 읽기는 부르지 않고 `skipped, the WTS session was rejected …` 로 이름을 남김
  (만료 세션에 종목마다 실패 요청 금지) — 다음 사이클은 다시 시도. 세션 파일이 없으면 클라이언트는 세션 없이 만들어지고
  세션을 요구하는 뉴스만 결손(`no active session`, 요청 0). 리뷰 수리(2026-10-15): 피드에 쓸 수 있는 기사(시각 해석 성공 +
  받은 순간 + 5분 이전)가 0 이면(0 건 응답 포함) 그 범위는 결손 — 받은 순간보다 늦은 기사는 시간대 오류 신호로 보관하지 않고
  셈 · 직전 거래일 수급 0 행은 결손(하루 캐시하지 않음) · 서버 쪽 수급 실패(HTTP 상태·전송)가 연속 3 번이면 그 사이클의 나머지
  수급 읽기를 멈춤(종목별 클라이언트 오류는 세지 않음) · 값 0 인 주요 지수는 이름으로 남김 · 24 시간 창은 보관소 정리(메모리)와
  고르는 쪽(이 판단의 at) 양쪽 — 두 시장이 보관소를 공유하므로. 남는 구조적 결손: US `market_context: no US market-wide investor
  flow …`, 그리고 영숫자 KR 코드(예 0167A0 — WTS 클라이언트가 국내 종목으로 보지 않음, 공식 심볼 검증도 거부; 후속 change).
- 세션·비밀: 세션은 기존 세션 저장소(`session.NewFileStore(paths.SessionFile)`)를 기존 클라이언트가 읽음. 로그에는 `loaded`/
  `not found`/`unreadable` 낱말만. 쿠키 표식·WTS 응답의 계좌 모양 표식이 TypeSafe 요청·원장 어느 파일에도 없음을 시험이 강제.
- 비용: Open API GET 수는 s2 와 같음(시험 단언). WTS 요청(HTTP 단위 — `GetTradingFlows` 는 종목 정보 + 수급 2 요청)은
  `selection.wts_requests` 로 따로 셈: KR = 지수 1 + 뉴스 2 + 2 × (그날 처음 보는 종목), US = 3. 보고서 「사이클 예산」 줄에
  `WTS 요청 중앙·최대` 가 붙음. WTS 는 엔진과 공유하는 Open API 키가 아님(다른 서버).
- state_rev = `a128-s3`; 질문 원문 불변(`questionRev` a128-r1). 원장은 s1(빈 값)·s2·s3 로트로 갈라짐.

## 질문 추가 (s4, tasks 5.x, 2026-10-20)

- 사용자 결정(2026-10-20): 4.2 를 「측정 계속」(선택지 2)으로 정하고, 같은 요청에 질문 둘(J3·J4)을 **더한다**. J1·J2 는
  원문·id·rev 그대로 계속 묻는다(옛 행과 새 행의 J1·J2 가 같은 모집단으로 비교되도록).
- 이유(`analysis/calibration-report-2026-10-20.md` 의 관측에서 — J1 구간 적중 0.581·0.547·0.549 로 평평, 사이클 안 ρ −0.000;
  J2 낮은 구간은 60분 하락 라벨에 대해 역방향):
  1. J1 의 「> 0」 은 그날 시장 흐름에 묻힌다 — 적중률이 0.5 를 넘은 것은 상승일 탓이고 구간 사이 차이가 없었다. 0 근처의
     작은 움직임이 적중을 좌우하므로 방향 질문이 국면 질문이 된다.
  2. +1% 는 「비용(수수료·세금·스프레드)을 넘는 움직임」의 근사다. 이 근사는 비용 모형이 아니고, 비용 차감 결론을 내지 않는다.
  3. J2 는 「명백한 악재 신호」를 물었는데 라벨은 60분 하락이었다 — 질문과 라벨이 다른 것을 쟀다. J4 는 하방 위험을 라벨과
     같은 말로 직접 묻는다.
  4. J4 가 J3 의 거울이라 두 질문의 p 를 같은 state 에서 대조할 수 있다(둘 다 높으면 「변동성이 크다」로 읽힘 — 관측 항목).
- 질문 원문(영어 — models.md 「English is the primary training language」, `state.go` `probeQuestions` 와 바이트 단위로 같음):
  - J3 `j3_up_1pct_within_horizon`
    - instructions: "Will this stock's last traded price, exactly `horizon_minutes` minutes after `observed_at`, be at least
      1% above `base_price`?"
    - criteria.true: "The last traded price at `observed_at` plus `horizon_minutes` is at least 1.01 times `base_price`."
    - criteria.false: "That price is below 1.01 times `base_price`."
  - J4 `j4_down_1pct_within_horizon`
    - instructions: "Will this stock's last traded price, exactly `horizon_minutes` minutes after `observed_at`, be at least
      1% below `base_price`?"
    - criteria.true: "The last traded price at `observed_at` plus `horizon_minutes` is at most 0.99 times `base_price`."
    - criteria.false: "That price is above 0.99 times `base_price`."
- 참/거짓 판정(라벨러 불변 — 기존 endpoint `realized_return` 하나로 셈): J3 참 ⇔ `realized_return ≥ +0.01`,
  J4 참 ⇔ `realized_return ≤ −0.01`. 보고서 적중 규칙 = 이 참 조건(`report.go` `hitRules`). 경계 ±0.01 은 참에 포함.
  `realized_return` 은 라벨러가 소수 8 자리로 반올림해 적으므로(`writeLabels`) 정확히 1% 움직임은 0.01 로 저장되어 부동소수
  오차로 경계가 뒤집히지 않는다. 라벨 결손·미라벨은 기존과 같이 분모에서 빠진다.
- rev 는 **질문마다** 붙는다(`question.Rev` 필드 → 원장 `question_rev` 열, 행 하나 = 질문 하나). J1·J2 = `a128-r1` 유지,
  J3·J4 = `a128-r2`. 원문이 바뀌면 `question_digest`(전송 본문 sha256)가 자동으로 바뀌므로, 시험이 J1·J2 의 digest 를 현재 값
  (원장 7,710 행씩과 같은 값, 2026-10-20 실측)으로 못 박는다.
- 그대로인 것: J1·J2 원문·id·rev·digest, 선별(`screen.go` — J2 p ≥ 0.5 제외, J1 내림차순 상위 K; J3·J4 는 읽지 않음), 결선 비교
  표(J1 만), state 모양(`state_rev` a128-s3 — state 를 바꾸지 않으므로 올리지 않음), 표본 선택 규칙 id, 라벨러, 원장 스키마
  (`a128-shadow-ledger/v1` — 새 열 없음, 새 행은 `question_id` 로 가려짐), spec 델타(질문 일반 요구라 바뀔 문장 없음).
- 로트 분리: J3·J4 는 자기 `question_id` 로 따로 표가 됨(보고서는 질문 id × 로트 × 시장마다 표를 만듦 — 섞일 자리가 없음).
  J1·J2 는 s4 전후 행이 **의도적으로 한 로트**다(원문·digest·state_rev 가 같으므로). s4 이후만 보려면 첫 J3 행의 `at` 이 경계다.
  J3·J4 를 더해 J1·J2 의 답이 달라지지 않는다는 근거는 계약 문서 인용뿐이다(concepts/state.md 「All questions see the same state
  and are evaluated independently」, api.md 「The key is not sent to the underlying model」) — 실측 아님. s4 측정 뒤 J1·J2 의 p 분포를
  s4 전후로 대조해 적을 것(tasks 5.7).
- 원장을 읽는 다른 도구(호환, tasks 5.5): `tools/a145-jev-offline-analysis`·`tools/a147-jev-mfi-offline` 은 한 state 의 행을
  J1·J2 두 개로 가정했음(a145 는 모르는 질문 id 를 `invalid judgment` 로 거부, a147 은 행 수 2 가 아니면 `invalid_pair`). 두 도구가
  `S4_QUESTIONS`(J3·J4 두 id — a145 에 한 번 정의, a147 은 그것을 읽음) 행과 그 라벨을 짝 판정 **전에** 빼고, 뺀 수를 출력에
  `ignored_s4_question_rows` 로 남기게 함(s4 이전 원장에서는 이 열쇠가 없어 출력 불변). 목록 밖 질문 id 는 예전처럼 거부·`invalid_pair`.
- 보고서 추가: (1) J3·J4 의 p 구간표·p ≥ 0.90 예비 판정(기존 양식 그대로). (2) **개장 후 경과 시간대별 표** — 질문 × 로트 × 시장마다
  `minutes_since_open` 구간 0–29 · 30–60 · 61–120 · 121+ 의 판단 n · 실현 라벨 n · 적중 · 적중률 · 평균 p(실현 라벨 행) · 판정(같은
  n<30 보류). 개장 30분(가장 변동이 큰 구간)을 따로 보이기 위함. `minutes_since_open` 은 원장 행에 없고 state 파일에 있으므로
  보고서가 `<원장 디렉터리>/states/<digest>.json` 을 **읽기만** 한다 — 파일 내용 sha256 이 행의 `state_digest` 와 같을 때만 쓰고,
  파일 없음·해시 불일치·값 없음·음수는 「알 수 없음」 줄로 센다(조용히 빼지 않음). 기존 J1·J2 출력은 바뀌지 않고 뒤에 덧붙기만 한다
  (base 코드로 만든 골든을 접두로 고정하는 시험).
- 비용: 질문이 요청당 2 → 4. 요청 수·GET 수는 같다. 과금은 입력 토큰 단위이고(models.md) 질문 원문만큼 입력이 늘지만, state 가
  질문마다 다시 세어지는지는 계약 문서 범위에서 미기재 — 첫 s4 사이클 원장 `input_tokens` 중앙값을 s3 중앙값과 대조해 적을 것.
- 한계: +1% 는 비용 근사일 뿐이며 이 측정은 여전히 수익성 결론을 내지 않는다. 60분 창 겹침·같은 종목 반복으로 표본이 독립이
  아님(calibration-report 「한계」)은 J3·J4 에도 그대로다. J3·J4 의 기저율(±1% 를 넘는 60분 움직임 비율)은 종목·국면에 따라 작을 수
  있어 구간 n≥30 이 늦게 찰 수 있다. 측정 창은 시장당 5 세션 이상(5.6)이며 그 뒤에도 판정은 예비다.

## 임계 판정 보고서

- p 구간(0.50~1.00, 0.05 폭)별: 표본 수, J1 실현 적중률, J2 는 라벨 정의(악재 판단 뒤
  h분 수익률이 음수였는가)로 적중률. **표본 수가 구간당 30 미만이면 그 구간 판정 보류**를
  보고서 양식에 박는다 — 2일 측정의 한계를 숫자가 직접 말하게.
- 90% 임계의 채택/기각/보류 권고와 근거를 보고서가 내고, 결정은 사람이 한다.

## 후속 (이 change 밖)

1. 수동 보조: 콘솔/CLI 에 J2 판단 표시(읽기 전용 UI) — calibration 보고 후 별도 change.
2. 엔진 veto gate: J2 를 기존 매수 신호의 AND 거부 조건으로 — High-risk change,
   사람 토글·audit·FLM 전 절차. J1 배선은 calibration 이 예측력을 증명한 경우에만 후보.
3. 손절·비상 청산 경로에는 어떤 Jev 판단도 배선하지 않는다(안전 불변식 4) — 이 금지는
   spec 델타에 요구로 박는다.
