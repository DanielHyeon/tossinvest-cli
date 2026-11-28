# a203 설계 — 돌파 KR 과거 재생

기준 HEAD `2b86ea27d`(origin/main) · go1.26.5 · 분기 근거 `analysis/ast-evidence/claims.md`(C1~C10) · a140 원장 수치
`analysis/a140-ledger-quote-stats.txt`(스크립트 `analysis/a140-ledger-quote-stats.py`, 원장 sha256 `0b802a92…` = a140 보고서 영수증과 같음).

## 0. 비례 원칙 · 적용 범위

- 위험 등급 Low. 주문 경로 밖의 오프라인 측정 도구이고 High-risk 경로(주문 · 손절 · 익절 · 사이징 · Guardian · 원장 · 대사 · 인증 · 체결)와
  게이트 판정 로직을 바꾸지 않음. `internal/` · `cmd/` 생산 Go 편집 0 을 목표로 함.
- 변이 원장 · 다중 적대 리뷰: **not-applicable** — 비례 원칙(docs/WORKFLOW.md 「비례 원칙」, 오프라인 측정 도구 · High-risk 경로 · 게이트
  판정 변경 0). 시험 + lint + 경량 리뷰 1회.
- FLM: D2 권고안 (가)는 기존 함수 `dispatch`(tools/a140-breakout-signal-shadow/main.go:55-67) 본문 1곳을 고침 → gate ⑤ 가 그 함수의
  번들 1개를 요구함(편집 전 작성, tasks 2.1). 나머지 새 코드는 새 파일.

## 1. 무엇을 재는가 (D1)

「지금 코드의 동결 v1 규칙이 과거 KR 세션에서 냈을 제안과 그 결과」 하나만 잼.

- 레인 상수는 `internal/breakoutlane/types.go:17-27` 의 v1 값 그대로(opening range 15 · 돌파 버퍼 100,000 ppm · KR timeout 8 봉 ·
  RVOL 1.5× · 윗꼬리 35%). 도구 쪽 사본 상수 `tools/a140-breakout-signal-shadow/policy.go:15-23` 도 그대로 씀.
- 범위만 정해진 설정 · 가상 사이징 · 비용은 **a140 판(lot) 값 그대로**: a140 원장 universe 행의 `params`
  (RetestTolerancePPM · MaxQuoteAgeMS · MaxSpreadPPM · MaxEntryDriftPPM · 위험 예산 · 명목 상한 · CostPPM · MinRRPPM). 현재 설정 파일을
  다시 읽지 않음 — 설정이 그 뒤 바뀌었어도 a140 과 같은 판으로 비교하려는 것. `costs.DefaultModel()` 의 현재 요율은 보고서에 함께 적음.
- **파라미터 탐색 금지.** 위 값을 바꿔 가며 돌리지 않음. 민감도는 §4 의 호가 가정 셋에만 있고 판정은 주 가정 하나로만 냄.

## 2. 도구 배치 (D2)

a140 의 평가 · 라벨 함수는 `package main` 이라 다른 패키지에서 import 할 수 없음. 재사용 방법은 둘.

| 항목 | (가) a140 패키지 안 새 파일 + `replay` 하위 명령 | (나) 새 도구 디렉터리에 옮겨 적기 + AST 동일성 시험 |
|---|---|---|
| 사본 | 0 — `evaluateSetup` · `labelProposal` · `planSizing` · `contiguousPrefix` 를 그대로 호출 | 약 300 줄(평가 · 라벨 · 사이징 · 틱 · 비용 함수 9개 + 타입 6개 + 정책 상수) |
| 기존 함수 편집 | `dispatch`(main.go:55-67) 에 case 1개(C10) | 0 |
| gate ⑤ | `dispatch` 번들 1개(분기 5) — 번들이 있으므로 `--record-landing` 가능 | required 0 → `not-applicable` 표지. 단 번들 0 이라 착지 기록 불가 → 늦게 닫으면 형제 change 의 함수가 창에 쌓임(Hindsight 교훈 a180-2) |
| a166 census `shapes.json` | `main.go||dispatch` 고정 본문 1건 갱신 필요 | 변경 0 |
| a166 census `sites.json` | 새 배선 자리 추가(클라이언트 생성 · 예산 열기 · `.RoundTrip` 참조) | 같음 + 옮겨 적은 함수의 `VALUE:plan.RoundTrip` 류 자리까지 추가(옮길 `evaluateSetup` · `rMultiplePPM` 이 이미 그런 자리를 가짐, sites.json:548 · 578) |
| 읽기 전용 가드 | a140 `static_test.go` 세 시험이 패키지의 **모든 비시험 .go** 를 glob 으로 읽으므로(static_test.go:26) 새 파일에 자동 적용 — 주문 · 계좌 · 설정 쓰기 이름 금지, import 허용 목록, `WithPriority` 는 Auxiliary 만 | 새로 써야 함(a140 시험을 옮겨 적으면 그것도 사본) |
| 계기 오염 | `run` · `report` 경로 · 정책 digest 무변경(새 파일과 case 1줄만) | 없음 |

**권고 (가).** 사본 0 · 가드 자동 상속 · 착지 기록 가능이 `shapes.json` 1건 갱신 + 번들 1개보다 큼.

**결정(2026-11-26, Manager) M1 = (가).** **M2 = census 시험 데이터 갱신 허용** — 범위는 `shapes.json` 의 `main.go||dispatch` 1건 갱신과
`sites.json` 에 a203 새 파일의 배선 자리 추가뿐이고 그 밖의 항목은 바꾸지 않음. census 오류 `reviewed function shape changed` 의 재검토는
구현 뒤 Manager 의 diff 검증으로 함. (참고로 (나)였다면 동일성 시험은 원본과 사본의 `format.Node` 본문(주석 제거)을 함수 이름마다 비교해 양쪽을 함께 고정함
(기억 「옮겨 적은 코드는 양쪽을 다 못 박아야 한다」).)

기각한 길: 전송을 `tossctl ops call candles` 로 하고 재생 코드는 파일만 읽게 하면 census 자리가 0 이지만, `ops call` 은 표준 해독기
`Client.Candles`(internal/ops/read_operations.go:233-262)라 a140 이 쓴 엄격 리더 · `officialbars` 적재 경로와 다른 봉 묶음이 됨 — 재생과
a140 원장의 대조(§6)가 파이프라인 차이로 흐려짐. 깊이 실측(§9)에만 씀.

새 파일: `tools/a140-breakout-signal-shadow/a203_replay.go`(하위 명령 · 시간대 거절 · 순서), `a203_cache.go`(저장소 밖 캐시 ·
캐시 리더), `a203_universe.go`(§5), `a203_quote_proxy.go`(§4), `a203_verdict.go`(§7), 각 `_test.go`. 원장은 저장소 밖
`~/.local/state/tossos-a203/` 에 쓰고, 커밋하는 것은 `openspec/changes/a203-…/analysis/replay/` 의 요약 보고서와 원장 지문(sha256 · 줄 수 ·
바이트)뿐임(결정 M4, 2026-11-26).

## 3. 재생 절차

종목 · 세션마다 a140 과 같은 파이프라인을 과거 시각으로 돌림.

1. **봉 수집**: `official.Client.StrictMinuteCandles`(a140 과 같은 엄격 리더)로 그 세션 정규장 1분봉 페이지를 받음(페이지 200 봉,
   세션당 2~3 페이지). 페이지(`StrictMinutePage` 전체 + 본문 지문)는 저장소 밖 캐시에 쓰고, 다시 돌리면 캐시에서 읽음(GET 0).
2. **적재**: 캐시 리더(`officialbars.CandleReader` 구현)를 `officialbars.PollClosedBars` 에 넣어 도구 전용 evidence DB(a140 `checkEvidencePath`
   가드 그대로)에 적재 → `SealBarSeries`. PollAt = 그 세션 마감 + 30 분(a140 라벨 유예와 같음). 과거 PollAt 로 한 세션 전체가
   적재되는지는 미검증 — tasks 2.2 RED 첫 항목.
3. **앞부분**: `contiguousPrefix`(C7). 빈 분이 있으면 a140 처럼 앞부분만 평가하고 퍼널 끝을 `GAP` 으로 셈.
4. **단계 진행**: n = 16 … len 마다 `evaluateSetup(앞 n 봉, …, 초기 대리 호가, 평가 시각 = n 번째 봉이 닫힌 시각 + 2 분)`. 처음으로
   ARMED · INVALIDATED · TIMED_OUT 이 나오면 멈춤(a140 의 「setup 은 종목 · 세션당 하나」). ARMED 전 단계는 호가 값과 무관(C2)하므로
   초기 대리 호가는 첫 봉 시가로 만든 봉인 가능 값이면 됨. ARMED 가 처음 난 n 에서 되찾은 봉 번호 k = n − 1.
5. **ARMED 시도**: 시도 i = 1 … 5 마다 §4 의 대리 호가로 `evaluateSetup(앞 k+1 봉, …)` — 앞부분을 언제나 0..k 로 고정해 「평가기는
   ARMED 에서 멈춘다」는 a140 주장(미검증 인용)에 기대지 않음. 처음 PROPOSED 면 끝, 다섯 번 모두 거절이면 `armed_expired`.
6. **라벨**: PROPOSED 가 시도 i 에서 났으면 `labelProposal(세션 전체 봉, 봉 k+2+i 의 여는 시각, plan, final=true)`(C6). a140 원장에서
   라벨 시작 봉은 ARMED 봉 + 3 분이 22건 중 19건(시도 1)이고 시도마다 1분씩 밀림 — 그 관계를 그대로 씀.
7. **원장**: a140 signal · label · gap 행과 같은 칸 + `replay` 칸(호가 가정 ID · 캐시 페이지 지문). 판 = `a203-replay/v1@<정책 digest>`,
   정책 digest 는 a140 정책 + 호가 가정 + 종목군 규칙 + 표본 목록을 담음.

## 4. 호가 관문 가정 (D3)

과거 호가는 없음. 스프레드 · 진입가 괴리 · 호가 나이 세 관문 중 대리 호가로 살아 있는 것은 괴리 하나뿐임(C4).

**a140 의 실제 모습**(`analysis/a140-ledger-quote-stats.txt`): ARMED setup 24 중 PROPOSED 22. 시도 행 36 = PROPOSED 22 ·
ENTRY_DRIFT_EXCEEDED 10 · SPREAD_TOO_WIDE 3 · QUOTE_STALE 1. 제안은 ARMED 봉이 열린 뒤 2분째 분에 19건 · 3분째 2건 · 4분째 1건(봉 k 는 k+2 분에
적재 — a140 design §2.1 리더 계약). 제안 시점 ask − 진입가는 −5 ~ +5 틱(중앙 0), 스프레드는 1 틱이 18/22, 가상 체결가 worst = max(진입가, ask)
(C5)가 진입가와 같은 경우 12/22, (worst − 진입가)/진입가 평균 0.098% · 최대 0.383%.

**주 가정 P0 — 「시도 분의 첫 거래가」**(판정은 이것으로만):

- 시도 i 의 평가 시각 = 봉 k+1+i 의 여는 시각(시도 1 = ARMED 봉 + 2분, a140 실측 최빈값).
- ask = 봉 k+1+i 의 시가, bid = ask − `krTickMinor(ask)`(1 틱), last = ask, 관측 = 수신 = 평가 시각, 통화 KRW.
- 결과: 호가 나이 · 스프레드 관문은 구성상 통과(C4 — 가격 1,000 원 이상일 때, §5 필터), 괴리 관문은 동결 평가기가 실제로 판정,
  가상 체결가는 a140 과 같은 식 worst = max(진입가, ask).
- 근거: a140 은 ARMED 감지 직후의 실시간 호가 ask 로 괴리를 재고 체결가를 정함. 과거 자료 중 그 순간에 가장 가까운 관측 가능한 가격이
  그 분의 첫 거래가(시가)임. 1 틱 스프레드는 a140 실측 최빈값.

**민감도(보고만, 판정에 안 씀)**:

- S1 낙관: ask = 진입가(시도 1 에서 언제나 통과, worst = 진입가).
- S2 비관: ask = 봉 k+1+i 의 고가(그 분에 실제로 치러야 했을 수 있는 가장 나쁜 가격) — 괴리 거절이 늘고 worst 가 오름.

보고서 **첫 줄**: 「과거 호가가 없어 호가 관문(스프레드 · 호가 나이)은 재생하지 않았고, 괴리와 체결가는 그 분 시가를 ask 로 대신함(P0).
체결 가능성(호가 잔량)도 재지 않음.」

## 5. 종목군 (D4)

과거 시점의 realtime 랭킹(a140 의 종목군, C9)은 조회할 수 없음.

**공식 종목 목록의 출처**: `docs/migration/openapi.latest.json` 에 `GET /api/v1/stocks/all`(market = KOSPI · KOSDAQ …, status =
ACTIVE · SCHEDULED · DELISTED, securityType, commonShare; Rate Limits Group `STOCK_ALL`)이 있음 — 상장폐지 종목도 조회 가능한 목록임.
그러나 `internal/official` 클라이언트와 `internal/ops` 카탈로그 어디에도 이 경로의 읽기가 없음(`grep -rn 'stocks/all' --include=*.go internal cmd` → 0건, 2026-11-26). 쓰려면
`internal/official` 에 새 읽기 함수가 필요함 → D7 목표(생산 코드 0) 밖이라 열린 질문 H4 로 올림.

**주 규칙 U-A(권고, 코드 범위 안)**:

- 후보 모집단 P = 공식 랭킹 `MARKET_TRADING_AMOUNT` · KR · `duration=6mo` · `excludeInvestmentCaution=false` · count 100 을 수집 시작 때 1회
  읽은 목록(캐시 · 지문 기록). 6mo 를 고른 이유: 표본 창(§6, 약 2026-06 ~ 2026-11)과 기간이 가장 가깝고, 투자유의 미제외는 a140 과 같음.
- 세션 d 의 종목군 = P 중 **직전 세션 d−1 의 일봉**이 있고 d−1 종가가 1,000 원 이상 · a140 판 명목 상한 이하인 종목을, d−1 일봉의
  종가 × 거래량(거래대금 근사) 내림차순으로 30개(같으면 종목 코드 오름차순). d 당일 이후 정보는 쓰지 않음.
- 1,000 원 하한: a169 공통 필터와 같고, 1 틱 대리 스프레드가 MaxSpreadPPM 안에 들게 하는 조건(C4).
- 일봉은 `Client.Candles(symbol, "1d", 200, "", false)`(표준 해독기 — 엄격 일봉 리더는 없음, a182 와 같은 한계) 종목당 1회.

**한계(보고서 첫 줄 다음)**: P 를 지금 랭킹으로 고르므로 (1) 생존 편향 — 그 사이 상장폐지 · 거래 급감 종목이 빠짐, (2) 모집단 미래 정보 —
「앞으로 6개월 동안 거래가 많았던 종목」에서 고름. 세션별 30개 선정만 미래 정보가 없음. 하루짜리 급등 소형주처럼 6개월 상위 100 에 없는
종목은 재생 종목군에 들 수 없어 a140 종목군과 성격이 다름 — §6 대조 2 가 그 차이를 잼. U-B(`/stocks/all` ACTIVE + DELISTED 전 종목에서 같은
세션별 규칙)는 (1)(2)를 크게 줄이지만 생산 읽기 함수 1개 · 일봉 약 2,500 GET 이 더 듦(H4).

## 6. 표본 · 기간 · 대조 (D5)

- **주 표본(결과 보기 전 고정)**: 기준일 2026-11-26(이 제안 날짜) 이전 KR 정규장 세션 중, 대조 세션 5개(2026-10-16 · 10-19 · 10-20 · 10-21 ·
  10-22)를 뺀 **가장 최근 100 세션**. 세션 목록은 공식 달력(또는 §9 G7 결과에 따라 일봉 날짜)으로 tasks 1.1 에서 확정하고, 그 목록의
  지문을 판 digest 에 넣음. 1분봉 수집 전에 커밋함.
- **깊이 분기(§9 결과로 기계적으로 정함)**: 1분봉 소급 깊이가 주 표본의 가장 이른 세션을 덮지 못하면, 주 표본 = 깊이 안에서 대조 5세션을
  뺀 모든 세션. 그 수가 40 미만이면 수집하지 않고 사람에게 보고(40 = a140 세션당 제안 4.4건 기준 약 175건, §7 최소 표본의 여유). 결정(2026-11-26, Manager) M5: 40 확정.
- **대조(검증 단계, 판정에 안 씀)**: 대조 5세션을 두 번 재생함.
  - 대조 1 — **a140 원장의 종목군 그대로**(universe 행): 파이프라인 차이만 남김. ARMED 전 종결 단계(RANGE_LOCKED · RETEST_WAIT · INVALIDATED ·
    TIMED_OUT · ARMED 도달)는 같은 봉이면 결정적이므로 **종목-세션 150 개 전부 일치**해야 함. 불일치가 하나라도 있으면 그 봉 차이(캐시 페이지 대
    a140 evidence)를 설명하기 전에는 주 표본을 수집하지 않음(fail-stop). PROPOSED 는 호가 차이로 달라질 수 있음 — a140 22건 대비 재생 제안의
    일치 수, 대리 ask − a140 ask(틱) 분포, 건당 순 수익률 차이를 보고. 이것이 P0 가정의 실측 오차임.
  - 대조 2 — **§5 규칙 U-A 의 종목군**: 세션마다 a140 종목군과의 겹침(교집합 / 30)과, 제안 수 · 건당 순 수익률을 a140 과 나란히 보고.
- a140 원장(5세션 22건)은 주 표본 판정에 넣지 않음(가설을 만든 표본).

## 7. 판정 사전등록 (D6)

결과를 보기 전에 이 절을 커밋하고, 커밋 뒤 숫자를 바꾸면 새 사전등록 · 새 표본임.

- **단위**: 주 가정 P0 의 PROPOSED 1건의 건당 순 수익률 r = (exit − worst − round_trip) / worst. a140 보고서 「가상 손익」과 같은 식
  (a140 archive `analysis/report_supplement.py:76-79`). 수량과 무관함(ZERO_QUANTITY 거절만 위험 예산 · 명목 상한의 영향을 받음).
- **일 평균**: 그 세션 제안의 r 단순 평균. 제안 0건 세션은 일 평균 배열에 넣지 않고 수만 보고.
- **INSUFFICIENT**: 제안(라벨 완료) < 30 건(a140 `minVerdictSamples = 30`, report.go:12), 또는 제안이 1건 이상인 세션 < 20, 또는 수집 결손(봉
  수집 실패 · GAP 로 끝난 종목-세션)이 전체 종목-세션의 20% 초과(a184 의 `unobserved` 20% 와 같은 상한), 또는 라벨 못 단 제안이 있음.
  - 세션 20 의 근거: 판정을 일 단위로 내므로 일 수가 표본 크기임. a184 는 20일 창에서 8일을 요구함(40%). 100 세션 창에서 같은 비율은 40 이지만,
    a140 실측에서 5세션 모두 제안이 있었으므로 40 은 거의 언제나 충족되어 거르는 힘이 약함 — 「부트스트랩이 의미를 갖는 최소 일 수」로 20 을 둠. 결정(2026-11-26, Manager) M5: 20 확정.
- **PASS**(INSUFFICIENT 가 아니고 셋 모두): (1) 일 평균의 평균 > 0, (2) 일 평균의 부트스트랩 0.025 하한 > 0, (3) 가장 좋은 하루를 뺀 일 평균의
  평균 > 0. 그 밖은 **FAIL**.
  - 근거: a184 design 「판정」(브랜치 `feat/a184-kr-jev-warning-live-lane`)과 같은 세 조건. (1) 평균이 비용 뒤 양수, (2) 하루 단위 상관을 그대로
    둔 채 우연을 거름, (3) 한 날의 대박이 결론을 만들지 않게 함. 비용은 이미 r 에 들어 있으므로 따로 빼지 않음.
- **부트스트랩**: a182 `report.go` 의 `bootstrapLower` · `percentile` 정의(a184 가 인용)와 같은 식 — 날짜 오름차순 일 평균 배열, `math/rand`
  `NewSource(203)`, 같은 크기 복원추출 2,000 회, 0.025 분위, 선형 보간. a182 코드는 main 에 없으므로(미병합 브랜치) 같은 정의를 새 파일에
  구현하고, a182 브랜치 함수와 같은 입력에서 같은 값을 내는 시험을 둠(tasks 2.4).
- 민감도 S1 · S2 의 같은 세 지표는 「판정 아님」으로 표시해 나란히 보고.
- **PASS 는 「실거래 전환 Story 를 제안할 근거」일 뿐임.** 엔진 배선 · 서명 · 토글은 별도 Story 와 사람 승인이 필요함(a140 design §8 S4~S6).
  FAIL 이면 「이 정의 그대로는 근거 없음」으로 닫고, 정의를 바꾼 재측정은 새 change · 새 표본임.

## 8. 호출 예산과 안전 (D7)

- **공식 GET 만**: `StrictMinuteCandles` · `Candles`(1d) · `Rankings`(1회) · `TypedMarketCalendar`. 주문 · 서명 · 엔진 전달 · 운영 설정 변경 0 —
  (가)안이면 a140 `static_test.go` 세 시험이 새 파일에도 그대로 적용됨(§2).
- **시간대 거절 — 결정(2026-11-26, Manager) M3**: 매 GET 직전, **평일 08:00~20:00 KST**(KR 정규장 + NXT 프리 · 애프터마켓) 또는 **미국 정규장**
  (ET 09:30~16:00, 서머타임 반영)이면 체크포인트를 남기고 전송 없이 멈춤. 허용 창 = 평일 20:00 KST ~ 미국 정규장 개장 전, 또는 주말. 본 수집의 기본 권고 창은 주말.
  사유: NXT 애프터마켓 동안에도 엔진이 KR 포지션의 손절 가격을 조회하고, 계정 예산(a166)은 운영에서 OFF 라 같은 키 한도를 보호하는 장치가 없음.
  미국장은 엔진이 같은 키로 US 레인을 돌리는 시간임. 「NXT 마감 20:00」은 Manager 판단이며 달력으로 확인하지 않았음 — 실행 전 확인 항목.
  달력 호출 없이 시각만으로 판정.
- **429**: 첫 429(또는 예산 거절)에서 그 실행 전체를 멈추고 체크포인트를 남김. 재시도 없음. 다시 돌리면 캐시에 있는 것은 GET 0.
- **간격**: GET 사이 1초 이상(벽시계 기준, 응답 시간과 무관하게 직전 전송 시작부터).
- **캐시**: `~/.local/state/tossos-a203/cache/`(저장소 밖) — 페이지마다 요청 인자 · `StrictMinutePage` · 본문 지문. 원장 · 보고서는 캐시 디렉터리
  지문(파일별 sha256 목록의 sha256)을 적음. 캐시는 커밋하지 않음.
- **예산 · 키**: a140 처럼 `accountbudgetruntime.Open` + 팩토리 클라이언트 + `accountbudget.Auxiliary` 우선순위. 어느 프로필(`--config-dir`)로
  돌릴지, 엔진과 같은 키라면 계정 예산(a166) 몫이 엔진 보조 읽기와 어떻게 나뉘는지는 사람 결정(H1).
- **도구 상한**: 실행당 GET 상한 플래그(기본 = §9 추정 상한)를 넘으면 멈춤.

**추정 GET**(주 표본 100 세션 기준):

| 구분 | GET | 근거 |
|---|---|---|
| 1분봉(주 표본) | 6,000 ~ 9,000 | 100 세션 × 30 종목 × 2~3 페이지 |
| 1분봉(대조 1 · 2) | 600 ~ 900 | 5 세션 × (30 + 겹치지 않는 종목 ≤ 30) × 2~3 |
| 일봉 | ≤ 100 | 모집단 P 종목당 1회(200 봉 ≈ 9.5 개월) |
| 랭킹 · 달력 | ≤ 110 | 랭킹 1 + 세션 달력(§9 G7 결과에 따라 0 가능) |
| 합계 | 약 6,800 ~ 10,100 | 1초 간격이면 약 2 ~ 3 시간 — 평일 20:00 KST ~ 미국 정규장 개장 전, 또는 주말(기본 권고: 주말) |

## 9. 소급 깊이 실측 계획 (D8 — tasks 0, 사람 승인 뒤 실행)

출처 충돌: a182 브랜치 design 「종목 1분봉·일봉 깊이 | 1분봉 730일 … | 앞선 깊이 측정」(영수증 미발견) 대 a128 tasks 6.4 「공식 API 는 최소 9일
전(KR 10-14) 1분봉을 줌(상한은 미측정)」. 새 코드 없이 기존 CLI 로 잼. 종목은 공개 예시 1종목(삼성전자 005930, a182 와 같음).

```bash
CFG=<H1 에서 정한 설정 디렉터리>
OUT=~/.local/state/tossos-a203/depth; mkdir -p "$OUT"
call() {  # $1 = 이름, $2 = 작업, $3 = params — 실패(429 포함)면 즉시 멈춤
  tossctl --config-dir "$CFG" ops call "$2" --params "$3" > "$OUT/$1.json" || { echo "STOP at $1"; exit 1; }
  sleep 1
}
call G1 candles '{"symbol":"005930","interval":"1m","count":3,"before":"2024-11-27T15:30:00+09:00","adjusted":false}'  # 약 730일
call G2 candles '{"symbol":"005930","interval":"1m","count":3,"before":"2025-11-26T15:30:00+09:00","adjusted":false}'  # 약 365일
call G3 candles '{"symbol":"005930","interval":"1m","count":3,"before":"2026-06-10T15:30:00+09:00","adjusted":false}'  # 주 표본 시작 이전(여유)
call G4 candles '{"symbol":"005930","interval":"1m","count":3,"before":"2026-08-26T15:30:00+09:00","adjusted":false}'  # 약 92일
call G5 candles '{"symbol":"005930","interval":"1m","count":3,"before":"2026-10-15T15:30:00+09:00","adjusted":false}'  # 대조 직전
call G6 candles '{"symbol":"005930","interval":"1d","count":200,"adjusted":false}'                                       # 일봉 200개가 닿는 날짜
call G7 market_calendar '{"country":"KR","date":"2026-06-10"}'                                                          # 과거 날짜 달력 응답 여부
```

- 기본 7 GET. G3 가 실패하고 G4 가 성공하면 그 사이를 이분해 **최대 2 GET**(G8 · G9) 더 → 합계 ≤ 9.
- 판정: 각 Gn 이 「요청한 날짜의 봉」을 1개 이상 돌려주면 그 날짜까지 깊이 있음. 빈 배열 · 다른 날짜 봉 · 오류면 없음. 원문 JSON 은 저장소 밖에 두고,
  커밋 영수증(`analysis/depth/receipt.md`)에는 요청 인자 · 응답 봉 timestamp · 개수 · 본문 sha256 만 적음.
- 분기: G3 성공 → §6 주 표본 그대로. G3 실패 → §6 깊이 분기. G7 이 과거 날짜를 주지 않으면 세션 개장 · 마감 시각은 일봉 날짜 + KRX 정규장 기본
  09:00~15:30 으로 만들고, 개장이 늦은 특수일(수능일 · 새해 첫날)은 그 날짜 1분봉의 첫 봉 시각으로 확인해 표본 목록에 표시.
- 한계: 한 종목만 잼 — 종목마다 깊이가 다를 수 있음. 주 수집에서 깊이 밖 종목-세션은 결손으로 셈(§7 결손 20% 상한에 들어감).

## 10. spec delta (D9)

생산 동작 변경 0 이므로 spec delta 를 만들지 않음. `.openspec.yaml` 에 `skip_specs: true`(선례 a202). 재생 도구는 tools/ 아래 오프라인 측정
계기이고 `breakout-retest-strategy-lane` 의 요구사항(평가 · 증거 권위)을 바꾸지 않음.

## 11. 한계 (보고서에 그대로)

1. 호가 관문 미재생(§4 첫 줄), 체결 가능성 미측정.
2. 종목군 생존 편향 · 모집단 미래 정보(§5).
3. 일봉은 표준 해독기(엄격 리더 없음).
4. 비용은 a140 판 CostPPM 하나(세율 변경 시점 미반영 — 현재 `costs.DefaultModel()` 값을 함께 적음).
5. VI · 단일가 구간이 만든 빈 분은 GAP 로 셈(a140 과 같음).
6. 표본은 한 시장 국면(2026 하반기)의 결과일 수 있음.
7. 원장은 커밋하지 않으므로(결정 M4) 판정 재계산은 그 로컬 원장(`~/.local/state/tossos-a203/`)에 의존함 — 커밋된 것은 지문뿐.

## 12. 기억 회고 (2026-11-26)

- ai-memory `memory_query "breakout replay 1분봉 candles depth a140"`: 직접 관련 결정 0. 관련 함정 1 — `gotchas/a166-census-freezes-wiring-function-bodies.md`
  (a166 census 가 `tools/*` 배선 함수 본문을 고정, 선택 시험에 안 잡힘) → §2 비교표 반영, `shapes.json` 실측으로 재확인(`main.go||dispatch` 외 3건).
- Hindsight `recall`(bank `tossos-dev-episodes`): breakout · replay · 깊이 관련 기억 0. 관련 교훈 1 — a180 종결 교훈 2(고정 FLM 번들 0 인 도구 change 는
  착지 기록 불가 → 늦게 닫으면 형제 함수가 쌓임) → §2 비교표 반영.
- 둘 다 advisory 이며 위 반영 내용은 현재 HEAD 로 다시 확인함.
