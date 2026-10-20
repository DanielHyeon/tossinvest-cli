# a128 Tasks

## 0. 사람 관문

- [x] 0.1 TypeSafe API 키 조달 — 사람이 발급해 `TYPESAFE_API_KEY` env 로 공급
      (저장소·기억·로그에 키 저장 금지). 키 전까지 1.x 구현은 mock 으로 진행 가능.
      — 2026-10-07 종결: 키는 `.env`(gitignore 확인)의 JEV_AI_API_KEY. 402 추적 끝에 원인 =
      콘솔 Monthly credit(프로모)은 API 크레딧이 아님 — Add funds 실결제 후 200
      (계약 문서 7·8절 영수증).
- [x] 0.2 외부 API 호출량 기본값 승인 — 세션당 샘플링 간격 10분 × 상위 10종목 × 판단 2
      (세션당 ≤ 약 780 판단 호출 × 4세션, 비용은 키 발급 시 요금표로 확인해 여기 기록).
      — 2026-10-07 종결: J1+J2 한 요청 묶음으로 실측 상한은 세션당 요청 ~340. smoke 실측
      332 input tokens/요청 → 4세션 ≈ 45만 토큰 수준(소액, $5 충전으로 충분 추정 — 1세션
      후 원장 합산으로 재확인). 승인 근거 = 사용자가 산식 2회 제시 후 결제 집행 +
      "남은 항목 진행" 위임(2026-10-07).
- [ ] 0.3 (선택) WTS 뉴스 소스용 `tossctl auth extend` — 사람 실행. 안 하면 뉴스 결손
      모드로 측정(원장에 명시됨).

## 1. 구현 (Opus 팀메이트 — 새 파일만, 생산 무접촉)

- [x] 1.1 TypeSafe API 계약 동결 — docs.typesafe.ai/api.md·primitives/noul.md 를 읽고
      요청/응답 계약을 원문 인용으로 `analysis/typesafe-api-contract.md` 에 기록.
      추측 필드 금지(계약은 영수증에서).
      — 2026-10-07 읽음(7 페이지). p = `answers.<id>.noul`. 문서 미기재 7항은 계약 문서 6 절.
- [x] 1.2 `tools/a128-jev-shadow-probe/` — judgment client(stdlib HTTP, env 키,
      타임아웃·재시도), state 조립기(공개 시장 데이터만 — 계좌 필드 차단),
      수집 루프(rankings→stocks→prices/orderbook, rate limit 백오프, 세션 경계 준수),
      원장 writer(JSONL append, 행 2종), 라벨러(h분 뒤 /prices 재조회),
      보고서 생성기(p 구간표·구간당 n<30 보류 규칙).
      — 설계와 다른 점: (a) 원장 행이 **3종**(judgment·label·**gap**) — 랭킹/시세 읽기 포기·판단
      API 포기처럼 판단 행이 생기지 못한 결손은 실을 행이 없어 침묵 생략이 되므로(spec 「결손은 행에
      명시」). (b) 뉴스는 이 로트에서 미배선 — state `news: "absent"` + `missing_inputs` 와 행
      `collection_gaps` 에 사유 명시(사람 관문 0.3 전). (c) official 은 복제가 아니라
      `internal/official` 을 import(L1c 선례) — 토큰 캐시·401 adoption 을 공유해야 엔진과 토큰
      전쟁(a082)이 안 남. 대신 브로커 표면은 읽기 GET 5개짜리 `marketSource` 인터페이스로만
      다루고 주문·계좌 이름은 AST 시험이 금지. (d) J1·J2 를 한 요청에 묶음(noul.md 권고) → 판단
      **요청** 수는 0.2 추정의 절반. (e) 판단 시각 + h 가 장 마감을 넘으면 더 판단하지 않음(라벨이
      장 밖 가격을 재지 않게) — KR 기본값에서 세션당 34 사이클. 모델 기본값 `jev-1.13.0` 버전 핀.
- [x] 1.3 시험 — mock TypeSafe 서버·mock 시세 픽스처로: 계좌 필드 차단, 뉴스 결손 명시,
      원장 행 스키마, 라벨 결손 사유, 보고서 보류 규칙, 키 부재 시 시작 거부.
      `go vet`·`make lint` 통과. FLM: not-applicable(새 파일만, High-risk 무접촉 —
      비례 원칙).
      — 영수증(2026-10-07, go1.26.5, 실 API 호출 0 — httptest 만):
      `go vet ./tools/a128-jev-shadow-probe/...` rc=0 · `go vet -tags tossos_testseams …` rc=0 ·
      `gofmt -l tools/a128-jev-shadow-probe/` 빈 출력 · `rtk proxy go test -count=1 ./tools/a128-jev-shadow-probe/...`
      rc=0 (`-v`: 최상위 29 PASS + 하위 7 PASS, FAIL 0) · 같은 명령 `-race` rc=0.
      `make lint` 전체는 돌리지 않음 — 그 gofmt 범위는 `./cmd ./internal ./tools/logic-map` 라 이 디렉터리를
      안 보고, 나머지는 `go vet ./...` 이므로 위 두 vet + gofmt 로 이 패키지 몫을 대신함.
      변이 점검(제자리 sed → 복원 sha256 대조 OK, 무변이 대조군 GREEN): 바이트 가드 삭제 2(judge·storeState)·
      기한 초과 검사 삭제·보류 기준 29·4xx 재시도·뉴스 결손 삭제·키 검사 우회·horizon 창 무시·모든 오류 재시도·
      라벨 불변식 삭제·O_TRUNC·J2 적중 반전 = 12/12 CAUGHT. (모든 오류 재시도는 처음 SURVIVED →
      503 비재시도 단언 추가 후 CAUGHT; 바이트 가드는 타입이 이미 막아 행동 시험으로 안 잡혀 AST 고정으로 못 박음.)
- [x] 1.4 팀메이트 커밋(푸시 금지) — 허용 경로: `tools/a128-jev-shadow-probe/`,
      `openspec/changes/a128-jev-shadow-judgment-probe/`. 그 밖 0.
      — 451ddba4, 14파일 전수 허용 내, Opus 트레일러 확인(Manager 배터리 2026-10-07).

- [x] 1.5 표본 선택 규칙 교체(사용자 지시 2026-10-14): 사이클마다 랭킹 **상위 100** 을 받아
      **1주 가격이 매수 가능 예산 이하**인 종목만 남기고 그중 **무작위 10**(10 미만이면 전부).
      - 이유: 상위 10 고정은 같은 종목을 반복 측정(10-12 KR 30슬롯에 고유 12) — 표본이 독립이
        아니고, veto 가 실제로 걸릴 집합(엔진이 살 수 있는 종목)과도 다름.
      - 예산 출처(사용자 결정 1=b, 자동 추종): config `engine.automation_gate.max_order_notional`
        (+ `limit_currency`) — Guardian 이 1회 주문에 거는 금액 상한(execgw/guardian.go `MaxNotional`),
        사용자 표현 「1일 구매 예산」의 코드상 대응(이 이름의 키는 config 에 없음, 2026-10-14 실측).
        **계좌 API 는 여전히 0** — config 파일 읽기만. US 는 official `ExchangeRate`(GET)로
        limit_currency 환산(Guardian 과 같은 방향: 가격×환율 ≤ 상한). 환율·config 를 못 읽으면
        그 사이클은 gap 행(stage=selection)으로 남기고 판단하지 않음(조용한 폴백 금지).
      - 원장 judgment 행에 `selection` 필드: 규칙 id(`random10_of_top100_le_order_cap`)·시드·
        예산값·통화·후보 수(100 중 통과 수). 기존 행(필드 없음)= `top10` 로트로 보고서가 분리.
      - 보고서: selection 로트별로 표를 나눔(섞지 않음).
      - 시험: 필터 경계(=상한 통과, 초과 탈락), 10 미만 전부, 시드 고정 재현, 환율 실패 gap,
        config 부재 gap, 계좌 표면 AST 금지 유지, 보고서 로트 분리.
      — 2026-10-14 구현(새 파일 `selection.go`·`selection_test.go` + collect/ledger/main/report/시험 편집).
      상한은 `internal/config` 로더(`NewService(<config dir>/config.json).Status`, 읽기 전용)로
      **사이클마다** 읽음 — 파일 부재(로더가 기본값을 돌려주므로 `Exists` 로 거름)·상한 ≤0/비유한·
      통화가 KRW/USD 아님·출처 미배선은 gap(selection)이고 랭킹도 판단도 부르지 않음. 경로는
      `--config-dir`(없으면 기본 해석 `~/.config/tossctl/config.json`, 실측 값 500000 KRW — 두 키만 읽음).
      `marketSource` 에 `ExchangeRate`(GET `/api/v1/exchange-rate`) 추가, 시세 통화 ≠ 상한 통화일 때만
      `ExchangeRate(시세, 상한)` 의 `rate`(매수 환율 = Guardian `RateQuoteToBase` 와 같은 필드)를 곱함;
      응답 짝 불일치·0 이하도 gap. 시드 = `math/rand/v2` 전역 생성기 53비트, 추출 =
      `rand.New(rand.NewPCG(seed, 0xa128_0105)).Perm(n)[:k]` 정렬.
      — 설계와 다른 점: (a) 필터 가격 = **랭킹 응답의 lastPrice**(`price_source: rankings.last_price`) —
      100 종목 `/prices` 한 번을 더 부르지 않으려는 선택, 기준가(base_price)는 그대로 `/prices`.
      (b) 랭킹 행 통화가 비었거나 시장 통화와 다르면 「가격 모름」으로 제외·계수(추정 금지).
      (c) 규칙 id 에 K 를 넣음(`random%d_of_top100_le_order_cap`) — `--top 10` 기본값이면 지시된 id 와
      같고, 다른 K 는 다른 로트로 갈라짐. (d) selection 레코드 필드: rule·seed·cap_value·cap_currency·
      price_currency·price_source·fx_rate/fx_pair(환산 시만)·ranking_requested(100)·ranking_received·
      excluded_over_cap·excluded_unknown_price·candidates·selected. (e) 후보 목록 자체는 원장에 없음 —
      시드는 같은 후보 목록이 주어질 때의 재현만 보장함(랭킹 원문은 보관하지 않음). (f) 환산은 Guardian 의
      haircut·올림을 적용하지 않음(표본 필터이지 주문 가드가 아님).
      — 영수증(2026-10-14, go1.26.5, 실 API 호출 0 — httptest·임시 config 만, `run` 미실행, .env 미열람):
      `gofmt -l tools/a128-jev-shadow-probe/` 빈 출력 · `go vet ./tools/a128-jev-shadow-probe/...` rc=0 ·
      `go vet -tags tossos_testseams …` rc=0 · `rtk proxy go test -count=1 -v ./tools/a128-jev-shadow-probe/...`
      rc=0 (최상위 42 PASS + 하위 14 PASS, FAIL 0) · 같은 명령 `-race` rc=0.
      `report` 하위 명령을 실 원장에 읽기 전용으로 돌림 → `표본 선택 로트: top10=160`, 표 제목 `· 로트 top10`.
      변이 점검 — 행동 10건은 `go test -overlay`(디스크 무변경, 무변이 대조군 GREEN, 전후 sha256 동일):
      필터 `<=`→`<`·환율 곱 삭제·selection gap 쓰기 삭제·표본=상위 k·보고서 로트 필터 삭제·config Exists 검사
      삭제·환율 짝 검사 삭제·시드 미기록·규칙 id 의 K 무시·통화 불일치 허용 = 10/10 CAUGHT.
      정적 가드 4건은 제자리 편집 → 복원 sha256 일치: `(*tossconfig.Service).SaveEngineAutostart` 명명·
      `tossconfig.DefaultFile`(허용 목록 밖)·`marketSource` 에 `BuyingPower` 추가·`(*official.Client).BuyingPower`
      명명 = 4/4 CAUGHT. 골든(시드 7)은 pickSample 이 아니라 절차를 따로 짠 프로그램으로 계산해 넣음.
      FLM·변이 원장·다중 적대 리뷰: not-applicable(비례 원칙 — 측정 도구 코드만, High-risk 경로 무접촉,
      주문·계좌 표면 0 은 AST 시험이 유지).
      — Manager 검증(2026-10-14): 경로 11 전수 허용 내·트레일러 확인·vet/test 재실행 rc=0·경량 리뷰
      critical 0. 실측 보강(read-only GET): KR·US 랭킹 count=100 각 100행, currency(KRW/USD)·last_price 전 행
      존재 → 상한 500000 KRW 통과 KR 88·US ≈64(환율 ≈1345). official ExchangeRate 실응답 짝은 미측정 —
      첫 US 사이클의 gap(stage=selection) 유무로 확인. 푸시 b5654298.
      **운영 제약(리뷰 발견)**: `restorePending` 은 시장을 거르지 않음 — 미라벨 판단이 있는 원장에 두 번째
      프로세스를 띄우면 남의 시장 판단까지 라벨링(이중 라벨). 한 원장에 미라벨이 남은 동안 프로세스는 하나만
      (시장이 달라도). 수리는 필요해지면 후속(시장 필터 또는 원장 잠금).

- [x] 1.6 Jev 선별 후 결선(사용자 결정 2026-10-14, 권고안 전부): 상한 필터 통과 상위 100 후보에서 **무작위 풀 30**
      (기존 `pickSample`·시드, 후보 ≤ 30 이면 전부)을 뜨고 풀 **전 종목**에 J1·J2 를 물음(종목당 요청 1 — 기존과 같음).
      J2 p ≥ 0.5 제외 → 나머지 J1 p 내림차순(동률 = 랭킹 순) 상위 **K=10** 이 결선. 생존 < K 이면 결선도 그만큼(수 기록).
      판단 실패 = gap 행, 선별 대상 아님. horizon 60m·질문 불변(`questionRev` a128-r1 그대로). 프로브 전용 — `internal/`·엔진 무접촉.
      - 측정 설계: 풀 전체 판단을 메모리에 모은 뒤 선별하고 **모든** 판단 행을 쓰고 전부 라벨 대기에 올림 — 비결선 = 같은
        사이클·같은 질문의 대조군. 행마다 `screen{finalist, screen_rank(생존 중 J1 순위, 악재 제외 = 0), red_flag_excluded}`.
        `selection` 에 `pool`·`judged`·`red_flag_excluded`·`finalists_requested`·`finalists` 추가, `selected` = 실제 뽑은 풀.
        규칙 id = `jev_screen<K>_of_random<POOL>_of_top100_le_order_cap`(기본 `jev_screen10_of_random30_of_top100_le_order_cap`)
        — 자기 로트. 옛 로트(`top10`·`random10_of_top100_le_order_cap`)는 그대로 보고됨.
      - 사이클 도중 크래시: 그 사이클 판단 행은 안 남음(state 파일·판단 gap 행은 남을 수 있음 — 판단 행 없는 state 는 무해). 수용.
      - 플래그: `--pool`(기본 30, `--top` ≤ `--pool` ≤ 100, 위반은 자격증명 읽기 전 거부), `--top` = 결선 수. 시작 로그에 풀·결선·규칙 id.
      - 보고서: 선별 기록이 있는 로트 × 시장마다 J1 「결선 vs 비결선」 표(결선 / 비결선 전체 / └ J2 악재 제외 / └ 생존·순위 밖 —
        판단 n·실현 라벨 n·적중·적중률, 같은 n<30 보류). 평균 실현 수익률은 보고서가 원래 계산하지 않으므로 넣지 않음.
        수익성 결론 없음.
      — 영수증(2026-10-14, go1.26.5, 실 API 호출 0 — httptest 만, `run` 미실행, .env 미열람):
      `gofmt -l tools/a128-jev-shadow-probe/` 빈 출력 · `go vet ./tools/a128-jev-shadow-probe/...` rc=0 ·
      `go vet -tags tossos_testseams …` rc=0 · `rtk proxy go test -count=1 -v ./tools/a128-jev-shadow-probe/...` rc=0
      (최상위 49 PASS + 하위 19 PASS, FAIL 0) · 같은 명령 `-race` rc=0.
      새 시험(`screen_test.go`): `TestScreenFinalistsRules`(경계 0.5 제외·0.4999 생존, J1 내림차순·K 절단, 동률 랭킹 순,
      생존 < K, 빈 입력) · `TestScreenQuestionsAreAsked` · `TestCycleJudgesTheWholePoolAndFlagsOnlyTheFinalists`(풀 30, 판단 실패 1 →
      58 행·gap 1·대기 29·라벨 58, 결선 정확히 10 종목, 순위 핀) · `TestCycleRecordsFewerFinalistsWhenFewSurvive` ·
      `TestPoolSamplingDrawsThePoolNotTheTop` · `TestRunFlagsValidatePoolAgainstTop`; `report_test.go`:
      `TestReportComparesFinalistsWithTheControlGroup` · `TestReportLegacyLotsHaveNoScreenComparison`. 기존 시험 갱신:
      `TestLedgerRowSchema`(screen·selection 열쇠), `TestSelectionCapBoundaryAndUnknownPrices`(새 수치), `TestSelectionSeedReproduces`
      (풀 10 으로 골든 유지, 규칙 id), 1.5 의 K 추출 시험은 풀 추출 시험으로 대체.
      변이 점검: 결선 플래그 항상 true(제자리 sed → 3 시험 FAIL → 원본 복원, sha256 일치·grep 으로 변이 줄 0 확인) + `-overlay`
      11건(무변이 대조군 GREEN, 전후 소스 sha256 동일): `<`→`<=`·악재 `>=`→`>`·비안정 정렬·동률 역전·비결선 행 누락·비결선 라벨
      대기 누락·풀=K·보고서 집단 뒤바꿈·보고서 J2 혼입·`--pool < --top` 허용·규칙 id 의 풀 무시 = 12/12 CAUGHT.
      FLM·변이 원장·다중 적대 리뷰: Function Logic Map: not-applicable — 측정 도구 경로만(`tools/`), High-risk 경로 무접촉·게이트 판정 변경 0
      (비례 원칙). 비용: 사이클당 Jev 요청 10 → 30(3배), 호가 GET 도 10 → 30.

- [x] 1.7 Jev 에 충분한 정보(사용자 지시 2026-10-14 「Jev에게 충분한 정보를 줘서 올바른 판단을 할 수 있도록 해」 +
      Manager 수정 「API 호출 최소화·묶음」): state 를 공개 시장 데이터로 넓힘 — `state_rev = a128-s2`(원장 judgment 행
      `state_rev`, 보고서 로트 = `규칙@a128-s2` 로 갈라짐; 질문 불변 → `questionRev` a128-r1 그대로). 기존 `official.Client`
      공개 읽기만, 새 전송 경로 0, `internal/` 무접촉. 주문·계좌·보유·매수가능 이름은 AST 시험이 계속 금지.
      - 새 블록(features.go): `intraday`(오늘 세션 1분봉 최근 30 + 5/15/30/60분 수익률(N분 전 봉 허용폭 5분)·창 고저와 그 안
        위치·창 근사 VWAP 거리·거래량 가속 5/30 봉) · `daily`(수정주가 직전 거래일 봉: 1/5/20일 수익률·20일 고저 거리·20일 평균
        거래량) · `market_context`(KR: KOSPI·KOSDAQ 오늘·30분 등락 + 시장 전체 투자자별 순매수 원) · `warning_types`(활성 매수 유의,
        0 건 = 빈 목록) · `price_limits`(KR 상하한가 거리). 못 읽은 블록은 생략 + `missing_inputs` 에 이름(판단은 계속, stocks_meta 정책).
      - 계약 근거(영수증): 지표 카탈로그 = KOSPI·KOSDAQ·KR_BOND_* 8종뿐(openapi Market Indicators) → **US 지수 없음** →
        US 는 `market_context:` 결손 이름. investor-trading = KOSPI/KOSDAQ 시장 단위만, 금액 원화 정수, 당일 = 잠정치(openapi) →
        **종목별 투자자 매매는 공식에 없음** → KR state 에 `investor_trading_per_stock:` 상시 이름. 1분봉 timestamp = 닫는 시각
        (a112 결정 30) → `before = 판단 시각` 으로 자라는 봉 제외. 일봉 timestamp 의미는 미실측 — 「시각 < 오늘 개장 = 직전
        거래일」 경계는 닫는 시각·개장 시각 해석 모두에서 같고, 「자기 날짜 자정」 표기만 틀림 → `last_prior_bar_at` 을 state 에 남김.
      - 넣지 않은 것: 호가 깊이 — `StrictOrderbookTop` 은 맨 위 한 줄만 내주고 깊이를 얻으려면 `/orderbook` 을 한 번 더 부르거나
        `internal/official` 을 고쳐야 함(엄격 경로 약화 없이 불가) → 제외. 체결(tape) — 공식 Trade 에 매수/매도 구분 없음 + 1분봉이
        거래량을 실음 → 제외. 「오늘 거래량/20일 평균」 — 일봉 캐시와 맞지 않아 제외(오늘은 intraday 가 실음).
        KRX 거래정지(`koreanMarketDetail.krxTradingSuspended`)는 공식 스키마에 있으나 `adaptStocks` 가 버림 → `/stocks` `status`
        (ACTIVE 외 = SCHEDULED·DELISTED)로만 종목별 읽기 생략(`per_symbol_reads: skipped …`). 정지 플래그는 열린 항목.
      - 호출 예산(풀 30, 시험 실측 — mock 의 경로·rate-limit 그룹별 계수): **KR 정상 사이클 = 96 GET**(RANKING 1·STOCK 31·
        MARKET_DATA 31·MARKET_DATA_CHART 30·MARKET_INDICATOR 1·MARKET_INDICATOR_CHART 2 = Manager 추정치와 같음), **그날 처음 보는 종목**
        은 일봉 1·가격제한폭 1 추가(첫 사이클 30 종목 전부 처음 = 160 GET, `TestCycleJudgesTheWholePoolAndFlagsOnlyTheFinalists`).
        2 종목 시험: 첫 사이클 20 GET·둘째 12 GET·캐시 적중 8(`TestCycleReadsTheRicherStateWithinItsCallBudget`). 하루 캐시 =
        (시장|현지 날짜|종류|심볼): 종목 일봉·가격제한폭·지수 일봉·시장 투자자 매매. 묶음: `/stocks`·`/prices`·지수 시세는 풀/지수
        전체 한 번. 사이클 원장 `selection` 에 `broker_reads`·`broker_reads_by_group`·`day_cache_hits`·`collect_seconds`·
        `interval_seconds`, 로그에 사이클마다 같은 수 + 수집 시간이 간격 절반을 넘으면 WARNING, 보고서에 로트별 「사이클 예산」 줄
        (GET 중앙·최대·수집 중앙·최대·간격 절반 초과 사이클 수).
      - 속도: 모든 브로커 GET 앞에 `ReadGap` 250ms(기존) → 한 시장 프로세스의 전체 GET ≤ 4/s, 종목 루프가 MARKET_DATA·CHART·STOCK
        를 번갈아 부르므로 MARKET_DATA 는 대략 ≤ 1.5/s(엔진과 공유하는 10/s 의 절반 아래). KR·US 세션은 겹치지 않음.
        추정 수집 시간(실측 아님): 96 GET × (0.25s + 지연) + Jev 30 요청 ≈ 1~2.5 분 < 간격 절반 5 분. 첫 장중 사이클의 로그·
        `collect_seconds` 로 확인할 것.
      - **입력 토큰(추정, 실측 아님)**: s1 실측 — 원장 j1 요청 90 건 median `input_tokens` **734.5**(726~751), state 파일 median
        659.5 B, 선형 적합 tokens ≈ 357 + 0.576 × state bytes. s2 state(시험 mock, KR 1분봉 30 개) = **3,869 B** → 추정 **≈ 2,590
        tokens/요청(약 3.5 배)**, 사이클(30 요청) ≈ 7.8 만, KR 세션(34 사이클) ≈ 260 만 토큰. 적합이 좁은 크기 범위(약 650~700 B)에서
        나와 외삽 오차가 큼 — 실제 median 은 첫 s2 사이클 원장의 `input_tokens` 로 다시 적을 것. 크기의 약 절반이 1분봉 30 개.
      — 영수증(2026-10-14, go1.26.5, 실 API 호출 0 — httptest 만, `run` 미실행, .env 미열람; 원장·states 는 **읽기만**):
      `gofmt -l` 빈 출력 · `go vet ./tools/a128-jev-shadow-probe/...` rc=0 · `go vet -tags tossos_testseams …` rc=0 ·
      `rtk proxy go test -count=1 -v ./tools/a128-jev-shadow-probe/...` rc=0(최상위 60 PASS + 하위 23 PASS, FAIL 0) · `-race` rc=0.
      새 시험(`features_test.go`, RED 먼저 — 컴파일 실패 확인 후 구현): `TestIntradayFeaturesOnAHandComputedSession`(60 봉 손계산 —
      개장 전·판단 뒤 봉 거름) · `TestIntradayFeaturesBoundaries`(빈 봉·봉 1 개·거래량 0·허용폭 5분 경계) ·
      `TestDailyFeaturesOnHandComputedBars`(21 직전 봉 + 오늘 봉 + 시각 없는 봉, 6 봉·0 봉) · `TestLimitsAndIndexFeatures` ·
      `TestAssembleStateCarriesEachNewInputOrNamesItsAbsence` · `TestCycleReadsTheRicherStateWithinItsCallBudget`(경로·그룹별 수, 둘째 사이클
      캐시, US 무호출) · `TestInactiveListingSkipsPerSymbolReads` · `TestReadFailuresBecomeMissingInputsNotGaps` ·
      `TestReportSplitsLotsByStateRev` · `TestReportShowsCycleBudgetPerLot` · `TestMarketContextGapDoesNotMaskAMissingStockMeta`
      (자기 리뷰에서 찾은 결손 가림 — RED 확인 후 수리). 갱신: `TestPublicStateKeySetIsPinned`(슬라이스 안 구조체까지 열쇠 고정,
      새 열쇠 전부 금지 정규식 비통과 확인 — 처음 쓴 `imbalance_pct` 가 `balance` 에 걸려 가드가 막음 → 깊이 블록과 함께 제거),
      `TestMarketSourceIsTheReadOnlyGets`(13 메서드), 계좌 표식 mock(새 읽기 응답 봉투·행에 `accountNo`·`holdingQuantity`·
      `accountBalance` 표식) 으로 `TestAccountFieldsInBrokerResponsesNeverReachTypeSafe` 통과.
      변이(`-overlay`, 무변이 대조군 GREEN, 전후 소스 sha256 동일): 캐시 무력화·상장 상태 관문 삭제·허용폭 삭제·세션 거름 삭제·
      오늘 봉을 직전으로·state_rev 미기록·US 가격제한폭 호출·종목별 투자자 결손 이름 삭제·GET 계수 삭제·보고서 절반 기준 → 간격·
      경고 0 건을 nil 로·US 0 띠 허용·로트가 state_rev 무시·거래량 0 나눗셈 = 14/14 CAUGHT.
      Function Logic Map: not-applicable — 측정 도구 경로만(`tools/`), High-risk 경로 무접촉·게이트 판정 변경 0(비례 원칙).

- [x] 1.8 Jev 가 모르는 정보 채우기(사용자 승인 2026-10-15 「권장대로」): s2 판단 전부의 결손 셋(뉴스 KR·US 100% ·
      KR 종목별 투자자 매매 · US 시장 맥락)을 **기존 WTS 조회 함수**로 채움 — `state_rev = a128-s3`(질문 불변, `questionRev`
      a128-r1). 새 파일 `wts.go`·`wts_test.go`, 기존 파일은 배선만. `internal/` 무접촉(client·session 은 import 만).
      - 새 필드(design.md 「s3」): `news`(present|absent) + `news_headlines{symbol ≤3(제목·요약·시각·출처), market ≤5(제목)}` —
        범위 `ALL_HIGHLIGHT`·`SOARING_STOCK`(계좌 맥락 없는 범위만), 판단 시각 뒤 기사 제외, 24 시간 보관소 · KR
        `investor_flows_per_stock`(직전 거래일까지 5 일, 순매수 주 수, 하루 캐시) · `market_context.major_indices`(양 시장, WTS
        주요 지수 6 종 + 직전 종가 대비 %).
      - 결손 정책: 소스 실패 → 블록 생략 + `missing_inputs` 이름, 사이클 계속. 401/403 이면 그 사이클 나머지 WTS 읽기 생략
        (`skipped, the WTS session was rejected …`). 세션 없음 → 뉴스만 결손(요청 0). 피드를 받았어도 시각 해석 0 건이면 결손.
      - 호출(사이클당): **Open API GET 은 s2 와 같음**(시험 단언 — KR 2 종목 20/12). s2 원장 실측(2026-10-15, KR 42 사이클·US 33):
        KR GET 중앙 105(97~164) · US 99(94~124). **WTS 요청(별도 서버, `selection.wts_requests`)**: KR = 지수 1 + 뉴스 2 + 2 × (그날
        처음 보는 종목), US = 3. 풀 30 추정(s2 의 새 종목 수 중앙 4 로 계산 — 실측 아님): KR 첫 사이클 63 · 정상 중앙 ≈ 11, US 3.
      - **입력 토큰**: s2 원장 median KR **3,517** · US 3,184(전체 3,260). s3 dry run(아래, KR 5 요청) median **4,204**(4,134~4,213,
        state 5.0~5.1 KB, 약 +20%) — 단 이 dry run 의 뉴스 블록은 비었음(시각 해석 결함, 아래에서 수리) → **뉴스 포함 토큰은 미실측**.
        상한: 종목 기사 3 × (제목 ≤120 + 요약 ≤120 자) + 시장 기사 5 × 제목 ≤120 자. 첫 s3 사이클 원장 `input_tokens` 로 다시 적을 것.
      - dry run 1 회(2026-10-15 15:58 KST, 장 마감 뒤 — `run-probe.sh KR … --out <scratch> --pool 5 --top 2 --session 00:00-23:59`,
        공유 원장 밖, 사이클 1 회 후 SIGINT): Open API GET 35 · WTS 13(= 1 + 2 + 5×2 일치) · 수집 18s · 세션 `loaded`. 수급 5/5 종목
        수신, 오늘 행 제외 확인(최신 2026-10-14). 주요 지수 6 종, KR 두 지수의 base 등락이 공식 직전 종가 등락과 소수 넷째 자리까지
        일치(-7.5654 / -7.654) → base = 직전 종가. **뉴스: 피드 48+5 건 수신, createdAt 53/53 이 RFC3339 해석 실패 → `news=present`·
        빈 목록**(결함) → 수리: 시간대 없는 모양(`2006-01-02 15:04:05` 등 — 클라이언트 fixture 모양)을 Asia/Seoul 로 읽고, 받은
        기사가 전부 해석 실패면 그 범위를 결손으로(`TestFeedWithNoParseableTimeIsAGapNotEmptyNews`). 시간대는 가정 — 첫 s3 사이클
        로그 `news feed …: newest raw "…"` 를 벽시계와 대조할 것. 두 번째 dry run 은 지시(1 회)대로 하지 않음.
      - 남은 결손: US 시장 단위 투자자 매매(어느 표면에도 없음 — 이름으로 남김) · 영숫자 KR 코드(0167A0 등: WTS 클라이언트가 국내
        종목으로 보지 않고 공식 심볼 검증도 거부 — 후속 change, `internal/official` 검증은 이 task 에서 안 고침) · US 종목 기사는
        관련 종목 이름 = 공식 `/stocks` 이름 일치로만 맞음(US 상품 코드는 티커 아님, 일치율 미실측).
      — 영수증(2026-10-15, go1.26.5, 시험은 httptest·in-process fake 만): `gofmt -l` 빈 출력 · `go vet ./tools/a128-jev-shadow-probe/...`
      rc=0 · `go vet -tags tossos_testseams …` rc=0 · `rtk proxy go test -count=1 -v …` rc=0(최상위 75 PASS + 하위 23 PASS, FAIL 0) ·
      `-race` rc=0(리뷰 수리 뒤 재실행). 새 시험(`wts_test.go`): `TestWTSInputsReachTheStateWithinTheirBudget`(선행·24h 밖·시각 없는 기사 제외, 오늘 수급
      행 제외, 둘째 사이클 수급 캐시, Open API 20 불변) · `TestUSStateCarriesMajorIndicesAndNoStockFlow` ·
      `TestWTSFailuresBecomeNamedGapsNotFabrications`(5xx → 생략 + 이름, 다음 사이클 회복) · `TestWTSSessionRejectionStopsTheCycleWTSReads`
      · `TestNoSessionStillReadsWhatNeedsNoSession` · `TestWTSAccountFieldsAndSessionNeverLeave`(쿠키·계좌 표식) ·
      `TestPriorDayFlowsAndClip` · `TestFeedWithNoParseableTimeIsAGapNotEmptyNews` · `TestReportShowsWTSRequestsInTheCycleBudget` ·
      (리뷰 수리) `TestEmptyFeedAndEmptyFlowsAreGaps` · `TestConsecutiveServerFailuresStopTheCycleFlowReads` ·
      `TestFlowFailureCountIsConsecutive` · `TestHeadlinesApplyTheWindowAtTheirOwnJudgment`.
      정적: `TestWTSSourceIsTheThreeReads`(인터페이스 = 셋, `internal/client` 의 `*Client` 나머지 공개 메서드 전부 이름 금지 —
      AST 자동 열거) · `TestNoPersonalizedNewsScopeInSource` · import 허용 목록에 client·session. 갱신: 열쇠 고정 목록, 뉴스·수급
      결손 이름. 빠른 변이(사본 복원, 전후 sha 동일): 선행 기사 필터·오늘 수급 행·거절 차단기·수급 캐시·개인화 범위·이름 일치·
      시장 목록 중복 제거·0 지수·보관소 24h 정리 = 9/9 CAUGHT(처음 「24h 창」이 살아남음 → 한 곳으로 합쳤다가, 리뷰 7 로 고르는 쪽
      창을 되살리며 시험을 따로 세움). 리뷰 수리 뒤 2 판: headlinesFor 선행 필터·창·보관소 정리·받은 순간 뒤 기사·쓸 수 있는 기사 0
      결손·빈 수급 결손·연속 실패 멈춤·연속 수 되돌림·0 지수 이름·거절 차단기 = 10/10 CAUGHT(처음 「연속 수 되돌림」 생존 → 시험 추가).
      Function Logic Map: not-applicable — 측정 도구 경로만(`tools/`), High-risk 경로 무접촉·게이트 판정 변경 0(비례 원칙).
      리뷰: review.md 「1.8」(경량).

## 2. 리뷰·착지

- [x] 2.1 gstack 리뷰(경량 — 비례 원칙) + Manager 검증(경로 전수·트레일러·시험 재실행).
      — 2026-10-07 Manager /review: 구현 6파일 정독, critical 0 · informational 2
      (① input_tokens 요청 단위 의미 — 주석 명기로 수리 d80ae85a ② 원장 꼬리 torn line 시
      readLedger 가 전체 거부 = fail-closed, 수리 불요·운영 시 꼬리 절단으로 복구).
      Review Army·적대 다중 리뷰: not-applicable(비례 원칙 — High-risk 무접촉·새 파일만).
      Manager 재실행 영수증: go vet rc=0 · go test rc=0 · gofmt 빈 출력.
- [x] 2.2 Manager 푸시 — 451ddba4 + d80ae85a (e67c514a..d80ae85a).

## 3. 측정 (2일 — 사용자 지시 2026-10-07)

> **2026-10-10 재편성.** 10-07 US 1일차 시도는 무산 — 하네스 background 프로브가 Claude
> 세션 종료와 함께 죽음("context canceled", 개장 대기 중이라 원장 0행·API 비용 0).
> 10-08(목)·10-09(금, 한글날 KR 휴장)은 세션이 없어 미기동. 10-10(토)·10-11(일) 휴장.
> → **1일차 = 10-12(월), 2일차 = 10-13(화)**, 각 KR(09:00~15:30 KST)+US(22:30~05:00 KST).
> 기동 래퍼는 `run-probe.sh`(이 디렉터리, 2026-10-10 실동작 검증 — 토요일 KR 달력 거부 rc=1
> 확인). 세션 밖 자가 지속 예약은 금지 유지(2026-10-07 분류기 차단 outcome) — 측정일
> 당일 in-harness 기동, 세션을 장 마감+horizon까지 유지하거나 사람이 setsid 로 띄운다.

> **2026-10-12 정정(실측).** 이날 KR 은 **휴장**이다 — official market-calendar 실측:
> today(10-12) regularMarket=null, 영업일은 previousBusinessDay=10-08·nextBusinessDay=10-13.
> (한글날 10-09 금 전후 연휴·대체.) 프로브의 "no regular session today" 거부는 **정상 동작**이다.
> 이날 Manager 가 프로브를 버그로 오판해 "빈 date→null" 가정의 수리를 커밋했다가(6549432c)
> 되돌림(720e89a9) — 그 가정은 **개장일에서 측정된 적이 없는 거짓**이었다(교훈:
> [[generated-evidence-must-be-measured]]·[[a-measurement-carries-its-moment]]). 빈 date 는
> 서버 실제 today 를 앵커해 올바른 데이터를 준다(US today 10-12 는 정상 세션 반환으로 재확인).
> **US 는 10-12 개장**(Columbus Day — 미 증시 정상, 22:30~05:00 KST). → 1일차 KR 은
> **10-13(화)로**, US 1일차는 **오늘 밤(10-12)** 가능. 날짜 재배치: KR 1일차 10-13·2일차 10-14,
> US 1일차 10-12·2일차 10-13. (프로브는 세션 종료와 함께 죽으므로 US 밤 세션은 사람이
> `setsid nohup ./run-probe.sh US <log> --wait-open &` 로 띄워 장 마감+horizon 까지 유지.)

- [x] 3.1 1일차: KR 세션 + US 세션 각 1회 프로브 실행(에이전트 가능 — read-only).
      실행 로그·원장 보존.
      — 달력 오판 대응 수단: official market-calendar 가 개장일을 휴장으로 오응답하면(2026-10-12 KR 실측)
      `--session "HH:MM-HH:MM"`(시장 현지 시각, KR=Asia/Seoul·US=America/New_York)로 달력을 우회해 측정을 돌린다 —
      달력 호출 0, 우회는 state `session_source=manual-override` + gap 행(stage=session, "calendar bypassed by
      --session, …, reason=operator")에 남아 보고서가 혼입을 가려낸다. 파싱 실패·역전창(start≥end)은 시작 전 거부.
      (run-probe.sh 가 그대로 전달: `./run-probe.sh KR <log> --session "09:00-15:30"`.)
      — 2026-10-20 read-only reconciliation: current `a128-s3` core KR 10-16·10-19와 US 10-15·10-19의 실행 기록·원장·state·라벨 시각을 대조했다. KR 두 세션은 기록형 `manual-override`, US 두 세션은 calendar provenance다. 자세한 집계·제한은 `analysis/reconciliation-2026-10-20.md`와 동명 JSON에 보관한다.
- [x] 3.2 2일차: 동일. 라벨러 완주 확인(마지막 판단 + h분까지).
      — KR 10-19는 첫 6 cycles가 context canceled 된 뒤 28 closed cycles를 재개해 합계 34 cycles가 됐다. 이를 uninterrupted run으로 표현하지 않는다. 네 core 세션의 미라벨은 0이고 label_at<due_at는 0이다. KR 10-16의 명시 라벨 결손 60건은 실현으로 바꾸지 않는다.
- [x] 3.3 원장·state 파일 커밋(계좌 데이터 0 확인 후).
      — 2026-10-17 offline audit: 9,000 judgment·9,000 label·6 explicit gap,
      4,500 referenced state files with link/hash/JSON errors 0 and no
      credential-shaped scan hit. Raw snapshot hashes and aggregate are
      `analysis/audit-2026-10-17.md`; independent review와 staged-index scan
      통과 후 evidence-only commit. 이 기록은 3.1/3.2 또는 4.x 완료를 뜻하지 않는다.

## 4. 판정·종결

- [x] 4.1 calibration 보고서 생성 — J1 예측력 유무(평평하면 J1 폐기 기록),
      J2 p≥0.9 구간의 적중률·표본 수, 2일 한계 명기.
      — 2026-10-20 `analysis/calibration-report-2026-10-20.md`(동결 사본 sha256 `a4b1b99a…`, s3 core 4 세션 3,990 판단 쌍):
      J1 예측력 없음(구간 적중 0.581·0.547·0.549, 사이클 안 ρ −0.000) → 배선 후보 제외 기록; J2 p≥0.9 판단 0(최대 0.85) →
      보류, 낮은 구간은 역방향. 비용 미반영 적중 빈도 관측 — 수익성 결론 아님. 결정은 4.2(사람).
- [x] 4.2 사람 결정 기록 — 90% 임계 채택/기각/보류, 후속 change(수동 보조·엔진 veto
      gate) 착수 여부. 후속 change 범위 초안을 review.md 에 남긴다.
      — 2026-10-20 사용자 결정: **측정 계속 + 질문 추가(s4)** (4.2 선택지 2, Manager 전달). 같은 요청에 J3·J4 를 더하고
      J1·J2 는 그대로 묻는다(design.md 「질문 추가 (s4)」). 측정을 이어 가므로 후속 change(수동 보조·엔진 veto) 범위 초안은
      이번에 쓰지 않음 — s4 측정 뒤 사람 결정(5.8)으로 넘김. 선택지 목록 원문은 이 기록에 없음(Manager 전달 요지만).
- [ ] 4.3 gate → archive(Opus) → Story 경로 갱신.

## 5. 질문 추가 s4 (사용자 결정 2026-10-20)

- [x] 5.1 설계 — design.md 「질문 추가 (s4)」: J3 `j3_up_1pct_within_horizon`(60분 뒤 +1% 이상)·J4 `j4_down_1pct_within_horizon`
      (60분 뒤 −1% 이하) 원문, 참/거짓 기준(J3 ⇔ `realized_return ≥ +0.01`, J4 ⇔ `≤ −0.01`), 질문별 rev(J3·J4 = `a128-r2`,
      J1·J2 = `a128-r1` 유지), 그대로인 것, 로트 분리, 시간대 표, 비용·한계. spec 델타 변경 없음(질문 일반 요구).
- [x] 5.2 코드 — `state.go` `probeQuestions` 에 J3·J4 추가(J1·J2 바이트 불변), `report.go` 적중 규칙 둘 + 개장 후 경과 시간대
      표(state 파일 읽기 전용), `main.go` report 가 states 를 읽음. 선별·라벨러·원장 스키마·state_rev 무변경. `internal/` 무접촉.
- [x] 5.3 시험 — J3·J4 가 요청에 원문대로 실림 · J1·J2 digest 고정(원장 실측 값) · 적중 경계(+0.01 참·+0.0099 거짓·−0.01 참) ·
      선별이 J3·J4 를 무시 · 옛 원장 픽스처가 base 코드 골든을 접두로 그대로 보고 · 시간대 구간 경계 · J3·J4 행이 question_id 로
      가려짐. 변이 1 건 이상 확인.
      — 영수증(2026-10-20, go1.26.5, 실 API·브로커 호출 0 — httptest 만, `run` 미실행, 공유 트리 무편집):
      `gofmt -l tools/a128-jev-shadow-probe/` 빈 출력 · `go vet ./tools/a128-jev-shadow-probe/...` rc=0 ·
      `go vet -tags tossos_testseams …` rc=0 · `rtk proxy go test -count=1 -v ./tools/a128-jev-shadow-probe/...` rc=0(최상위 84 PASS +
      하위 23 PASS, FAIL 0) · 같은 명령 `-race` rc=0. 새 시험 `s4_test.go` 9 개 + 픽스처 `testdata/pre-s4/`(옛 모양 원장 23 줄 · state 5 —
      파일 없음 1 · 해시 불일치 1 포함 · base 코드 bbe4ceaa6 로 만든 `report-base.golden.md`). 기존 시험은 행 수만 갱신(2 → 4 질문).
      변이(`-overlay`, 무변이 대조군 GREEN, 전후 소스 sha256 동일): J3 `>=`→`>` · J4 `<=`→`<` · 구간 경계 29→30 · state 해시 검사 삭제 ·
      선별이 J1 대신 J3 를 읽음 · 선별이 J2 대신 J4 를 읽음 · J1 원문 한 글자 · J3 rev r1 = 8/8 CAUGHT(처음 「선별이 J3 를 읽음」이
      살아남음 — 두 사이클의 J3 순서가 우연히 같은 결선을 내서; J3 순서를 반대로 두고 결선을 손계산으로 고정해 CAUGHT). 제자리 변이 1 건
      (J3 `>=`→`>` → `TestS4HitRuleBoundaries` FAIL → 복원, `diff` 빈 출력). 라이브 원장(sha256 접두 a4b1b99acef0e505)을 읽기 전용으로
      base 코드(overlay)·새 코드 둘로 report → 새 출력이 base 출력 23,235 바이트를 접두로 바이트 그대로 포함, 뒤는 시간대 표 18 개뿐,
      「알 수 없음」 0(참조 state 전부 해시 일치).
      Function Logic Map: 기존 함수 내부 분기 변경이 `renderCalibrationReport` 끝의 덧붙임 루프 1 개(AST B19~B22)라 AST 산출물을 만듦 —
      `analysis/function-logic/tools-a128-jev-shadow-probe--rendercalibrationreport/`(ast.json 분기 22 · FLM · BTM · risk). B1~B18 은 base 와
      같고 골든 접두 시험이 고정. `check_analysis.py --change a128-…` 는 「missing base-commit.txt」로 판정 불가(이 change 에 base 가 동결된
      적 없음 — 동결은 Manager 결정). 그 밖의 편집(`probeQuestions` 데이터·`reportCommand` 한 줄·새 함수 `renderSessionTimeSection`·
      `readStateMinutes`·`sessionBucket`): not-applicable(분기 변경 없음 / 새 함수). 변이 원장·다중 적대 리뷰: not-applicable(비례 원칙 —
      측정 도구 경로만, High-risk 무접촉·게이트 판정 변경 0).
- [x] 5.4 리뷰 — 경량(비례 원칙): `analysis/gstack/terra-s4-review-2026-10-20.md`.
- [ ] 5.5 착지 — 공유 트리 반영(systemd 타이머가 공유 트리에서 `go run` 하므로 **반영 순간부터 s4 가 측정됨**). 반영 시각을 여기
      적어 s4 경계로 씀. Manager. **선행 조건(리뷰 발견)**: 원장을 읽는 다른 도구 둘이 「한 state = J1·J2 두 행」을 가정함 —
      `tools/a145-jev-offline-analysis/analysis.py` 는 `set(pair) != {J1,J2}` 면 `ValueError('unpaired questions')` 로 **전체가 멈추고**,
      `tools/a147-jev-mfi-offline/analysis.py` 는 `len(rows) == 2` 가 거짓이라 s4 행을 전부 `invalid_pair` 로 셈(조용한 강등). s4 행이
      원장에 들어가기 전에 두 도구가 question_id 를 J1·J2 로 거르게 하거나, s4 이전 스냅숏에만 돌린다는 제한을 적을 것(a145·a147 change 범위).
- [ ] 5.6 측정 — 시장당 s4 세션 5 회 이상(KR ≥5 · US ≥5). 첫 사이클의 `input_tokens` 중앙값을 s3 와 대조해 기록.
- [ ] 5.7 보고 — J3·J4 구간표·시간대 표, J1·J2 의 p 분포 s4 전후 대조(질문 추가가 J1·J2 를 바꾸지 않았는지), 수익성 결론 없음.
- [ ] 5.8 사람 결정 — J3·J4 결과로 후속(수동 보조·엔진 veto) 착수 여부와 임계. 후속 change 범위 초안을 review.md 에.
