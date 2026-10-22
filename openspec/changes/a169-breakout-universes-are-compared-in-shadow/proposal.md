# a169 — breakout 그림자 신호를 종목군 셋으로 다시 잰다 (정의만)

> 상태: **열린 질문 Q1~Q12 답 받음(2026-10-23) — 전부 권장 기본값 채택**(아래 「사람 결정」 표). 지금 착수하는 것은 Q11 에 따라
> What Changes 2(`securityType` 노출)뿐이다. 나머지(종목군 선정 · 측정)는 Q6 에 따라 **a140 archive 뒤** design.md 부터 착수한다.

## Why

a140(`openspec/changes/a140-breakout-signals-are-measured-in-shadow/`)은 breakout v1 신호를 KR 5 세션(2026-10-16 · 19 · 20 · 21 · 22)
그림자로 쟀고, 사용자는 2026-10-22 「a140: 중단」을 골랐다(a140 review §4). 그 측정의 종목군은 하나였다 — 개장 30 초 뒤 공식 랭킹
`MARKET_TRADING_AMOUNT` · `realtime` 상위 100 에서 주문 상한 이하 30 종목(아래 「a140 선정 방식」). 그래서 a140 의 관측은 「breakout v1 이
**거래대금 상위 종목에서** 이렇게 나왔다」이고, 종목군을 바꾸면 달라지는지는 재지 않았다.

사용자 결정(2026-10-23): **같은 breakout 신호 · 같은 진입/청산 · 같은 비용 모델(KR 왕복 0.50%)** 로 **종목군만** 바꿔 다시 잰다. a140 은
고치지 않고 그 결과를 비교 기준으로 인용한다.

a140 의 비교 기준(인용 — 판정이 아닌 관측, a140 `analysis/report-2026-10-22-supplement.txt`, 원장 sha256 `0b802a92…740c`): 종목-세션 150 →
PROPOSED 22 · 라벨 22(close 17 · stop 3 · target 2). n=22 < 30 이라 판 합계 판정은 보류됐다. 가상 손익 절(판정 아님): 순이익 거래 7 ·
거래당 평균 순 −0.70%.

## What Changes

### 1. 종목군 셋을 같은 세션에 동시에 (사용자 결정)

| 종목군 | 정의 | 출처 |
|---|---|---|
| **U0** | 거래대금 상위 30 — a140 과 같은 선정 방식(시각 · duration · 조회 수) | 공식 `GET /api/v1/rankings` `type=MARKET_TRADING_AMOUNT` |
| **U1** | 거래량 상위 30 | 공식 `GET /api/v1/rankings` `type=MARKET_TRADING_VOLUME` |
| **U2** | 거래량 증가율 상위 30 — API 에 없어 직접 계산: 당일 누적 거래량 ÷ (20 일 일평균 거래량 × 장 경과 비율). 후보 = 거래량 상위 100 ∪ 거래대금 상위 100 (시장 전체 증가율이 아니라 **후보 안의** 순위 — 보고서에 그 한계를 적는다) | 위 두 랭킹 + 공식 `GET /api/v1/candles` `interval=1d` |

- 공통 필터: 투자유의 제외(`excludeInvestmentCaution=true`) · 1 주 가격 ≥ 1,000 원 · 1 주 가격 ≤ 주문 상한.
- **ETF 는 빼지 않고 따로 잰다**: 종목군마다 유형 갈래로 나눠 집계한다. 판별 출처는 공식 `GET /api/v1/stocks` 의 `StockInfo.securityType`
  이고, 조회 결과를 영수증으로 원장에 고정한다. **ETF 와 ETN 은 별도 갈래로 센다**(사용자 결정 2026-10-23).
- 측정 기간: KR 5 세션(a140 과 같음).
- 통과 규칙 · 최소 n 은 a140 과 같은 형식으로 **측정 전에** 적는다 — 세션별 보고(보류가 기본)와 **판 합계**(같은 정의의 세션을 합쳐
  n ≥ 30 이면 승률 · 평균 R 판정, a140 design §2.1-⑪). 판정 칸 = 종목군 × 유형 갈래. 판정 임계 수치는 열린 질문 Q8.

### 2. 공식 `securityType` 을 tossctl 출력에 드러낸다 (사용자 결정 2026-10-23 — 열린 질문 아님)

- 사실: 공식 응답 구조체는 이 값을 받는다(`internal/official/stock_reads.go:37` `SecurityType string \`json:"securityType"\``). 그러나
  `adaptStocks`(`internal/official/stock_reads.go:81~97`)가 만드는 `domain.Quote`(`internal/domain/models.go:395~426`)에는 칸이 없어 버려진다.
  `adaptStocks` 의 분기는 `range` 하나(B1 @83)와 return 하나(@96)뿐이다(AST: `analysis/ast-evidence/internal-official-stock_reads.go--adaptStocks.json`).
- enum(`docs/migration/openapi.latest.json` `components.schemas.StockInfo.properties.securityType`): `STOCK, FOREIGN_STOCK, DEPOSITARY_RECEIPT,
  INFRASTRUCTURE_FUND, REIT, ETF, FOREIGN_ETF, ETN, STOCK_WARRANTS`.
- 변경안: `domain.Quote` 에 `SecurityType string \`json:"security_type,omitempty"\`` 칸을 더하고 `adaptStocks` 에서 채운다. 다른 생산자(WTS ·
  `/prices`)는 채우지 않으므로 그 JSON 바이트는 그대로다 — a163 `SourceTimestamp` 가 같은 이유로 `string+omitempty` 를 쓴 선례
  (`internal/domain/models.go:422~425` 주석).
- 노출 경로: `tossctl ops call stocks` → `internal/ops/read_operations.go:190~198` 의 `d.Client.Stocks` → `adaptStocks`.
- 소비자:
  1. 이 change 의 측정 — 종목군별 유형 갈래(위 1).
  2. **a128 의 D2 분석** — `openspec/changes/a128-jev-shadow-judgment-probe/analysis/prereg-2026-10-23-d2-twice-daily-index-correlation.md` §7 은
     같은 enum 을 STOCK(`STOCK`·`FOREIGN_STOCK`) · ETF(`ETF`·`FOREIGN_ETF`) · ETN · 기타 · 미분류로 사전 고정했고, 입력을 영수증 파일
     `analysis/d2-receipts/security-types.json` 으로 받는다. 같은 문서가 「현재 tossctl 은 이 값을 출력하지 않음 … 드러내는 변경은 후속 change(a169)
     범위」라고 적었다. 이 change 는 tossctl 출력의 `security_type` 칸만 만든다 — a128 의 영수증 파일(조회 시각 · 호출 수 포함)은 a128 이 만든다.
- 위험: 읽기 전용 경로(High-risk 목록 밖). `adaptStocks` 는 기존 함수라 구현 때 FLM/BTM 또는 `not-applicable` 사유를 남긴다(tasks 2.x).

## a140 선정 방식 (코드 사실 — 이 change 가 U0 로 옮길 것)

근거 AST: `analysis/ast-evidence/tools-a140-breakout-signal-shadow-shadow.go--probe.run.json` · `…--probe.selectUniverse.json` ·
`…--probe.startSession.json`(HEAD `10377cb4b`, `source_sha256` 은 각 JSON 머리).

- 시각: `probe.run` 이 개장 + 30 초까지 기다린 뒤(`tools/a140-breakout-signal-shadow/shadow.go:547`, B3) `startSession` 을 부른다(:550, B4).
  원장의 universe 행 5 개가 모두 `…T00:00:30Z`(09:00:30 KST)에 찍혔다(a140 `analysis/shadow-ledger/ledger.jsonl`).
- 한 세션에 한 번: 원장에 그 세션의 universe 행이 있으면 그 종목을 쓰고(`shadow.go:196`), 없을 때만 `selectUniverse` 를 부른다
  (`startSession` B9 @200 · :201, AST `…--probe.startSession.json`) — 재시작 경로.
- 호출: `Rankings(ctx, "MARKET_TRADING_AMOUNT", "KR", "realtime", false, 100)`(`shadow.go:23~25` 상수 · :265). **`excludeCaution` 은 `false`** 였다.
- 거름(`selectUniverse` B3 @271 · B4 @275): 순위대로 보며 30 개(`defaultProbeConfig` `Top: 30`, :52)가 차면 멈추고, 6 자리 KR 심볼 · `KRW` ·
  NaN 아님 · 가격 > 0 · 가격 ≤ `NotionalCapKRW` 가 아니면 건너뛴다. **1,000 원 하한은 없었다.** 주문 상한은 설정에서 읽는다(`main.go:99` `loadLimits`).
- 비용: `costRatePPM(costs.DefaultModel())`(`main.go:103`) — a140 은 이 값으로 왕복 비용을 매겼다(사용자 결정의 「KR 왕복 0.50%」).

즉 U0 에 공통 필터(투자유의 제외 · 1,000 원 하한)를 걸면 **a140 의 선정과 바이트가 같지 않다** — Q7.

## search-first 결과 (U2 계산을 재사용할 수 있는가)

- a140 의 「RVOL 반사실」은 **봉 단위** 지표다 — `RVOLPPM` = 그 봉 거래량 × 15 / 직전 15 봉 거래량 합(`internal/breakoutsignal/signal.go:151~153`,
  `rvolWindow = 15` :34). 일평균 기준선이 없어 U2(당일 누적 ÷ 20 일 평균 × 경과 비율)로 쓸 수 없다. 재사용 불가.
- `internal/candidate` 의 `Accelerate`(`internal/candidate/metrics.go:666`)는 최근 창 율 ÷ 직전 창 율(`Acceleration` :279~280), 창은 30 초
  (`DefaultAccelerationWindow` :193)다. 역시 일 단위 기준선이 없다. 재사용 불가.
- 당일 누적 거래량 후보: 공식 랭킹 응답 항목의 `tradingVolume`(`domain.RankingItem.TradingVolume`, `internal/domain/models.go:705`). OpenAPI 는
  「`MARKET_*` 는 시장 전체 기준」이라고만 적고 `realtime` 의 누적 기준은 적지 않는다 — Q3.
- 일봉: 공식 `GET /api/v1/candles` `interval` enum `["1m","1d"]` · `count` 최대 200 · Rate Limits Group `MARKET_DATA_CHART`
  (`docs/migration/openapi.latest.json` `/api/v1/candles`), 리더는 `internal/official/candle_reads.go:53` `Client.Candles`.
- 유형: `GET /api/v1/stocks` 는 `symbols` 콤마 최대 200 건 · Rate Limits Group `STOCK` — 후보 ≤ 200 이면 세션당 1 GET.
  대안 `GET /api/v1/stocks/all`(`securityType` · `market` 필터, 페이지 없이 한 번에, 그룹 `STOCK_ALL`, 일 배치 갱신)은 보았다 — 유형별 목록을
  세션 전에 받을 수 있지만 KR 은 시장(KOSPI · KOSDAQ · KR_ETC)마다 호출이 필요하고, 이 change 는 선정된 종목의 유형만 있으면 되므로
  `/stocks` 를 권장한다(design 에서 확정).
- 외부(레지스트리 · GitHub · MCP)는 이 질문에 쓸 이유가 없어 **보지 않았다**(공식 API 계산식 정의 문제).

## 사람 결정 (2026-10-23) — Q1~Q12 전부 권장 기본값 채택

사용자 답(2026-10-23, Manager 경유): 「Q1~Q12 전부 권장」. 아래 「열린 질문」 본문은 질문 당시 기록으로 남기고, 결정은 이 표가 정본이다.
같은 표가 `issues.md` 에 있다.

| Q | 결정(원문 「권장」) |
|---|---|
| Q1 | U0 · U1 = 개장 + 30 초. U2 = 09:30:30 KST 1 회 선정. 선행 정보 차단은 평가기 밖 종목군별 귀속 필터(선정 시각 이하에 닫힌 BREAKOUT_CLOSED 봉은 그 종목군에 귀속 안 함) — 평가기 출력은 a140 과 바이트 동일 |
| Q2 | 세션당 1 회 고정(재시작은 원장 universe 행으로 되살림) |
| Q3 | 분모 = 오늘 뺀 직전 20 거래일 일봉 거래량 평균(`interval=1d`, `adjusted=true`), 20 일 미만 이력은 U2 제외 · 사유 원장 기록. 개장 · 마감은 공식 달력. 분자 `tradingVolume` 이 당일 누적인지 첫 세션에 같은 종목 당일 1 분봉 합과 대조(다르면 1 분봉 합으로). 동률은 거래량 순위 높은 쪽 먼저 |
| Q4 | 일봉 = 선정 때 후보 × 1 GET(분산 일정은 design), 유형 = `/stocks` 1 회, 분봉 = 세 종목군 합집합 1 회 폴. 측정 5 세션 전 리허설 1 세션(측정에 안 셈). 10 분 창 수집 시간 300 초 초과(WARNING)면 사람에게 다시 묻는다. `ReadGap` 250 ms 불변 |
| Q5 | 종목당 평가 1 회, Q1 귀속 필터를 통과한 종목군 전부에 귀속. 보고서는 종목군별 합계(중복 포함) · 겹침 표 · 고유 종목 부분 집계 |
| Q6 | (a) a140 도구에 종목군 입력 추가, 기본값 = a140 동작 그대로(판 digest 불변 시험). **a140 archive 뒤 착수** |
| Q7 | U0 랭킹은 `excludeInvestmentCaution=true`, universe 행에 필터 전 상위 100 과 종목별 제외 사유 기록. a140 식 재구성용 `excludeInvestmentCaution=false` 랭킹 1 GET/세션 추가 |
| Q8 | 칸(종목군 × 유형 갈래) 판 합계 n ≥ 30 이고 평균 순 R > 0(왕복 비용 포함) → 「후속 검토 후보」, 아니면 「중단 근거 유지」. n < 30 은 보류. 통계 판정이 아니라 사람 판단 입력 |
| Q9 | a128 D2 §7 대응 그대로 — STOCK(`STOCK` · `FOREIGN_STOCK`) · ETF(`ETF` · `FOREIGN_ETF`) · ETN · 기타 · 미분류. 판정 칸은 STOCK · ETF · ETN, 기타 · 미분류는 개수만 |
| Q10 | 30 미만 종목군은 받고 실제 수를 기록, 「30 미만 세션」 표시로 판 합계에서 따로 셀 수 있게. 세션 무효 아님 |
| Q11 | What Changes 2(`securityType` 노출)를 먼저 진행 — tasks 2.1. 나머지는 Q6 전제대로 a140 archive 뒤 |
| Q12 | 첫 세션(또는 리허설) 영수증으로 랭킹에 든 ETN 심볼 형식 확인, `^[0-9]{6}$` 밖이면 거름을 넓힐지 사람에게 묻는다 |

## 열린 질문 + 권장 기본값 (질문 당시 기록 — 결정은 위 표)

1. **Q1 선정 시각.** U0 은 a140 그대로 개장 + 30 초(결정). U2 를 같은 순간에 고르면 30 초 누적 거래량이 시가 단일가 체결에 지배되어
   증가율이 잡음이 된다. (선정 시각이 하나면 「장 경과 비율」은 모든 후보에 같은 상수라 **U2 순위를 바꾸지 않는다** — 보고용 절대값에만 쓰인다.)
   - 권장: U0 · U1 = 개장 + 30 초(두 API 랭킹은 같은 순간에 잰다 — 차이가 지표 하나뿐이 되게). U2 = **09:30:30 KST** 한 번.
   - 선행 정보 차단은 **평가기가 아니라 귀속 단계**에 둔다: 평가기 출력(setup · 전이 · 라벨)은 a140 과 바이트가 같고, 종목군마다 「BREAKOUT_CLOSED
     봉이 그 종목군의 선정 시각 이하이면 그 종목군에는 귀속하지 않는다」(U2 를 09:30 에 고르면 09:15~09:30 의 돌파가 선정 이유 자체가 되므로).
     U0 · U1 은 09:00:30 에 닫힌 봉이 없어 a140 과 같은 결과다.
   - 대가: U2 는 첫 30 분의 돌파를 못 센다 — U2 와 U0 의 차이에 시각 효과가 섞인다. 보고서에 그 한계를 적는다.
2. **Q2 고정 vs 주기 갱신.** 권장: 세션당 **한 번 고정**(a140 과 같음 — 재시작은 원장의 universe 행으로 되살림). 주기 갱신은 세션 중 모집단이
   바뀌어 a140 과 비교가 안 된다.
3. **Q3 U2 계산 세부.** 권장:
   - 분모 = 오늘을 뺀 직전 20 거래일 일봉 거래량의 평균(`interval=1d`, 수정주가 기본값 `adjusted=true`). 20 일 미만 이력(신규 상장 등)은 U2 에서
     빼고 원장에 사유를 남긴다.
   - 장 경과 비율 = (선정 시각 − 정규장 개장) ÷ (정규장 마감 − 개장), 개장 · 마감은 a140 과 같은 공식 달력(`cal.Today.Regular`, `shadow.go:543`).
   - 분자 = 랭킹 응답의 `tradingVolume`. 이것이 「당일 누적」인지는 **첫 세션에서 영수증으로 확인**한다(같은 종목의 당일 1 분봉 거래량 합과 대조).
     다르면 분자를 당일 1 분봉 합으로 바꾼다.
   - 동률은 거래량 순위가 높은 쪽 먼저.
4. **Q4 호출 예산(a140 D11 의 영향).** 권장:
   - 일봉: 선정 때 후보(≤ 200) × 1 GET(`MARKET_DATA_CHART`) — 세션당 한 번.
   - 유형: 후보 ≤ 200 → `GET /api/v1/stocks` 1 회(`STOCK` 그룹).
   - 분봉: 세 종목군의 **합집합**을 한 번만 폴(중복 제거). 최대 90 종목. a140 실측(30 종목)은 10 분 창 수집 시간 중앙 95.3~196.0 초 · 최대
     359.0 초였다(a140 supplement 「호출 예산 실측」). 비례하면 90 종목은 1 분 안에 못 끝날 수 있다.
   - 그룹 점유율: 분봉 폴은 종목당 분당 1 GET 이라 90 종목이면 1.5 GET/s — a140 design §6 이 적은 `MARKET_DATA_CHART` 초당 5 의 30%
     (a140 의 N = 30 은 10%). **2026-10-22 사용자의 「D11 지금대로」는 N = 30 기준이라 이 change 로 승계되지 않는다** — 다시 확인받는다.
   - 버스트 충돌: 09:30:30 U2 선정 = 랭킹 2 GET + 일봉 ≤ 200 GET 이 같은 paced source(`ReadGap` 250 ms → 최소 50 초)와 같은 그룹에서 이미 도는
     U0 ∪ U1(≤ 60 종목) 분봉 폴과 겹친다. 그 분의 폴이 밀리면 결손 행이 생겨 a140 식 세션 유효성(`no_minute_gap`)이 깨진다.
     권장: 일봉 200 GET 을 선정 전 몇 분에 나눠 미리 읽거나(후보는 그때의 랭킹으로) 분봉 폴 사이 빈 시간에만 넣고, 일정은 design 에서 정한다.
   - 그래서 **측정 5 세션 전에 리허설 1 세션**(측정에 안 셈)으로 실측한다. 판단 기준은 a140 과 같게 「10 분 창 수집 시간이 창의 절반(300 초)
     초과 → WARNING」(a140 design §6 · supplement 「호출 예산 실측」)이고, 그것이 나오면 사람에게 다시 묻는다(종목군당 수 줄이기 등).
     a140 의 폴 간격(`ReadGap` 250 ms)은 바꾸지 않는다.
5. **Q5 종목군 사이 중복 종목.** 권장: 종목당 평가는 한 번, 그 setup · 라벨을 그 종목이 든 종목군 중 **Q1 귀속 필터를 통과한 모든** 종목군에 귀속한다. 보고서는 종목군별 합계
   (중복 포함)와 겹침 표(U0∩U1 · U0∩U2 · U1∩U2 · 셋 다), 그리고 종목군 고유 종목만의 부분 집계를 함께 낸다.
6. **Q6 실행 형태.** a140 도구는 `package main`(AST `package: main`)이라 import 할 수 없다. 선택지: (a) a140 도구에 종목군 선택 입력을 더하고
   기본값은 a140 동작 그대로 — 판(lot) digest 불변을 시험으로 고정, (b) 새 도구 디렉터리에 복제(두 정본이 갈릴 위험), (c) 공용 부분을
   `internal/` 패키지로 옮김(a140 도구 리팩터). 권장: **(a)** — 신호 · 진입/청산 · 비용 코드가 물리적으로 하나라 「같은 신호」가 구성상 참이다.
   a140 change 문서와 측정 결과는 고치지 않는다. **전제: (a) 는 a140 archive 뒤에 착수한다** — a140 은 지금 완료 gate 가 7 단계에서 막혀
   (a140 review §6.2) 아직 활성이고, 그 사이 `tools/a140-breakout-signal-shadow/` 를 고치면 a140 의 gate 창에 들어간다.
7. **Q7 a140 식 선정의 재구성.** U0 에도 공통 필터를 거는 것은 결정이다(사용자 결정). 그래서 U0 은 a140 의 선정(`excludeCaution=false` ·
   1,000 원 하한 없음)과 같지 않다. 열린 것은 그 차이를 원장에 얼마나 남길지다. 권장: U0 의 랭킹은 `excludeInvestmentCaution=true` 로 읽고,
   universe 행에 필터 전 상위 100 과 종목마다 제외 사유(1,000 원 미만 · 상한 초과 등)를 적는다. 투자유의 종목은 API 가 빼므로 a140 식 완전
   재구성에는 `false` 랭킹 1 GET 이 더 필요하다 — 그 1 GET 을 더할지가 질문이다(권장: 더한다, 세션당 1 회). 보고서는 「a140 대비 바뀐 것 =
   종목군 + 공통 필터」를 밝힌다.
8. **Q8 판정 임계.** 최소 n = 30(판 합계, a140 과 같은 형식). 권장: 칸(종목군 × 유형 갈래)의 판 합계 n ≥ 30 이고 **평균 순 R > 0**(왕복 비용
   포함)이면 그 칸은 「후속 검토 후보」, 아니면 「중단 근거 유지」. n < 30 이면 보류. 세션끼리 상관이라 이것은 통계 판정이 아니라 사람 판단의
   입력이다(a140 review §4 과 같은 성격).
9. **Q9 유형 갈래.** 권장: a128 D2 §7 의 대응을 그대로 쓴다 — 비ETF = STOCK 갈래(`STOCK` · `FOREIGN_STOCK`), ETF(`ETF` · `FOREIGN_ETF`),
   ETN, 기타(`DEPOSITARY_RECEIPT` · `INFRASTRUCTURE_FUND` · `REIT` · `STOCK_WARRANTS`), 미분류(enum 밖 · 조회 실패 · 응답 없음). 판정 칸은
   비ETF · ETF · ETN 셋, 기타 · 미분류는 개수만 보고한다.
10. **Q10 종목군이 30 에 못 미칠 때.** a140 `selectUniverse` 는 0 개일 때만 거부하고(B5 @281) 30 미만은 받았으며, a140 은 세션 유효성
    `universe30` 으로 그것을 걸렀다. 공통 필터 · U2 의 20 일 이력 제외 · 랭킹이 `count` 보다 적게 오는 경우(OpenAPI: 시세 실패 종목 제외)로
    30 미만이 될 수 있다. 권장: 받고 기록하되(실제 수를 universe 행에), 종목군별로 「30 미만 세션」을 표시해 판 합계에서 따로 셀 수 있게 한다.
    세션을 무효로 하지는 않는다(다른 종목군까지 버리게 되므로).
11. **Q11 `securityType` 노출(What Changes 2)을 먼저 진행할지.** 노출 자체는 결정됐고 모호함이 없다. 그런데 이 change 전체가 Q1~Q10 답을
    기다리는 동안 a128 D2 의 유형 칸은 「미분류」로 남는다(a128 prereg §7). 권장: **2 만 먼저 진행**한다(tasks 2.1 — Q 답과 무관). 나머지는 답 뒤에.
12. **Q12 ETN 심볼 형식.** a140 의 심볼 거름은 `^[0-9]{6}$`(`shadow.go:33`)이다. KR ETN 심볼이 이 형식이 아니면 ETN 갈래가 조용히 빈다
    (미검증 가정). 권장: 첫 세션(또는 리허설) 영수증에서 랭킹에 든 ETN 의 심볼 형식을 확인하고, 다르면 거름을 넓힐지 사람에게 묻는다.

## Non-goals

- a140 의 문서 · 측정 결과 · 판정을 고치지 않는다. breakout 지표 판본 `breakoutsignal-v1` 과 진입/청산 · 비용 규칙을 바꾸지 않는다.
- 엔진 배선(S4) · 서명 매니페스트 · Jev 호출 · 주문 · 계좌 조회 0 — a140 과 같은 도구 전용 evidence 경계.
- US 시장, 다른 전략 가족.
- 수익성 결론. 결과는 사람 판단의 관측 근거다.

## Impact

- 코드(착수 뒤, Q6 에 따라): `internal/domain/models.go`(`Quote` 칸 하나) · `internal/official/stock_reads.go`(`adaptStocks`) ·
  `tools/a140-breakout-signal-shadow/`(종목군 선택) 또는 새 도구.
- 위험도: Normal(읽기 전용 · 그림자). High-risk 경로(주문 · 손절 · 원장 · Guardian · 인증) 무접촉. 공식 API 호출 추가(일봉 · stocks)는
  호출 예산에 계상(안전 불변식 §0-4, Q4).
- PM: `STORY-TOS-a169` → FEAT-TOS-013(a140 과 같은 Feature).
- 정의 단계의 증거: Function Logic Map 은 정의 단계라 not-applicable(기존 함수 편집 0). 분기 근거는 위 AST 산출물 넷.
