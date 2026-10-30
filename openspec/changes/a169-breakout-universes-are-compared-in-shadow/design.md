# a169 설계 — breakout 그림자 신호를 종목군 셋으로 다시 잰다

> 2026-10-31 정합화 진행: 사용자 승인 정책은 공통 그룹 합계 3건/초 우선·보호 대기 허용이다. 아래 과거 판의 4 GET/s, 별도 키의 독립 예산, engine.log 감시만으로 취소 후 새 wire0을 보장한다는 설명을 현재 안전 근거로 사용하지 않는다. 별도 키는 토큰 격리를 위해 유지하고 U2 09:30:55 마감·미완 전체 late도 유지한다. a166 부분 착지는 재개 승인이 아니다. `analysis/a166-bridge-draft.md`, `a166-bridge-research.md`, `a166-bridge-independent-review.md`에 현재 source와 미해결 event fence·활성 엔진 신원·영속 abort 계약을 기록했다. 초안은 freeze 전이며 1.3 FAIL·리허설/측정 보류는 그대로다.

> 2026-10-31 추가 사람 결정: 안정된 현재 cursor를 기준점으로 시작하고 이전 gap은 제외한다(issues §7). 현재 pending/dirty/오류를 허용하거나 기록을 지우지 않는다. `analysis/prewire-fence-plan.md`의 선택적 내부 전송 fence 로트는 **High-risk**로 FLM/BTM·독립 적대 검토를 적용한다. 아래 과거 Normal 분류는 이 새 전송 계층 편집에 적용하지 않는다. 전체 측정 모드 freeze와 운영 재개는 여전히 미완이다.

기준 HEAD `0c9bfd272` · codegraph 1.6.0 · 분기 근거 `analysis/evidence/claims.md`(AST 26 개 — 13 개는 초안 전, 2 개는 review §4 반영 전, 5 개는 review §6 반영 전, 6 개는 review §8 반영 전에 생성) ·
호출 근거 `analysis/code-context/`(0.3 절). 사람 결정은 proposal 「사람 결정 (2026-10-23)」 표(Q1~Q12 전부 권장)가 정본이고, 이 문서는
그 답을 코드 위에 옮긴다. 줄 번호는 모두 HEAD `0c9bfd272` 기준이다.

## 0. 전제와 비례 원칙

- **선행 조건 충족**: Q6 의 전제 「a140 archive 뒤 착수」는 성립했다 — a140 은 `openspec/changes/archive/2026-10-22-a140-breakout-signals-are-measured-in-shadow/`
  (커밋 `a25db35f4`)로 닫혔다. a140 도구 `shadow.go` 의 sha256 은 정의 단계 AST(`analysis/ast-evidence/`, HEAD `10377cb4b`)와 같은
  `62c2fe84…` 다 — 그 뒤 바뀌지 않았다.
- **위험 등급: Normal(도구 전용 그림자)**. 주문 · 손절 · 익절 · 사이징 · Guardian · 원장(journal) · 대사 · 인증 · 체결 경로의 함수를 편집하지 않는다.
  운영 판정 · 게이트 판정을 바꾸지 않는다(결과는 사람 판단의 관측 입력). 그래서 변이 원장 · 다중 적대 리뷰는 not-applicable 이고 기본은
  시험 + lint + 경량 리뷰다(docs/WORKFLOW.md 「비례 원칙」).
- **단 gate ⑤ 는 등급과 무관**하다 — 이 change 가 본문을 고치는 기존 Go 함수(시험 파일 포함)마다 FLM 번들이 필요하다. 그래서 새 코드는 새 파일에 두고,
  기존 함수 편집은 §9 의 다섯으로 줄인다.
- 안전 경계 인접 항목 하나: a140 은 엔진과 같은 Open API 키를 썼다(a140 design §6). **a169 모드는 측정 전용 별도 앱 키만 쓴다**(사람 결정 D5,
  2026-10-28 — §7.5 · §14). 같은 키면 MARKET_DATA 예산과 토큰(client 당 유효 토큰 1 개)을 엔진과 나눠 엔진 보호 호가가 굶을 수 있다(review §6 P1-3).
  §7 에서 호출 수를 세고, 개장 직후 MARKET_DATA 시작 호가는 a140 과 같이 읽는다(§7.3 — 생략해도 GET 수가 줄지 않음, D2 · review §4 P1-2).
- **review §4 반영(2026-10-28)**: 독립 리뷰 FAIL(P1 2 · P2 6 · P3 7)의 발견을 본문에 옮겼다. 발견별 위치는 review §5 표.
- **review §6 반영(2026-10-29)**: 재리뷰 FAIL(P1 1 · P2 2)의 발견을 본문에 옮겼다 — P1-3 별도 키(§7.5 · §14), P2-1 `--top` 고정(§1),
  P2-2 U2 마감(§2 · §4.5). 발견별 위치는 review §7 표.
- **review §8 반영(2026-10-29)**: 재리뷰 FAIL(P1 4 · P2 2) — 토큰 동등 판정 · `--profile-preflight` · 엔진 rate limit 실시간 감시기 · 엔진 경로 단일 변수 ·
  경로 정규화 · 좁은 reader(§7.5 · §14). 발견별 위치는 review §9 표.

## 1. 실행 형태 (Q6 = (a))

- a140 도구 `tools/a140-breakout-signal-shadow/`(`package main`)에 **종목군 모드** 하나를 더한다: `run --universes a169`. 기본값 `a140` 은 a140 의
  동작 그대로다. 신호(`evaluateSetup`, evaluate.go:66~162) · 진입/청산(`armedAttempt` · `labelProposal`) · 비용(`costRatePPM(costs.DefaultModel())`,
  main.go:103) 코드는 물리적으로 하나라 「같은 신호」가 구성상 참이다.
- **판(lot) 분리**: 판 = `effectivePolicy` JSON 의 sha256(policy.go:71~80). 종목군 설정을 `Universes *universePolicy \`json:"universes,omitempty"\``
  칸으로 더하면, 기본 모드(nil)의 JSON 바이트는 지금과 같아 a140 판이 그대로이고, a169 모드는 다른 판이 된다. 같은 세션 · 같은 원장 파일에
  다른 판이 섞이면 시작을 거부하는 a140 규칙(startSession B5@186)이 그대로 지킨다.
- **`universePolicy` 칸(review §4 P2-1)** — 판을 가르는 값은 전부 여기 들어간다. 그래서 D3 (a) 로 N 을 20 으로 바꾸면 **다른 판**이 되어
  30 판과 20 판의 세션이 한 판 합계로 섞이지 않는다. 칸은 정수(시간은 ns 정수) · 불리언 · 문자열 · 문자열 슬라이스만 쓴다 — float · `time.Time` 금지(원장에서 읽어
  다시 digest 해도 같은 바이트여야 함, startSession B6@189 → 190 `row.Policy.digest()` · review §4 P3-5).

  | 칸(JSON) | 값(정상 판) | 뜻 |
  |---|---|---|
  | `version` | `a169-universe-policy/v1` | 이 표의 판본. 귀속 규칙(§5.2) · U2 식(§4.2) · 일봉 경계(§4.2)를 바꾸면 판본을 올린다 |
  | `per_universe` | 30(D3 (a) 적용 판은 20) | 종목군당 N. **`cfg.Top` 과 다른 칸**이다 — A140 재구성(`selectUniverse`, shadow.go:271 `p.cfg.Top`)과 `effectivePolicy.top` 은 30 그대로 |
  | `cohorts` | `["U0","U1","U2"]` | 측정 종목군 |
  | `u0u1_offset_ns` · `u2_offset_ns` | 30 초 · 30 분 30 초(D1) | 개장으로부터의 선정 시각 |
  | `ranking_fetch` | 100 | 랭킹 항목 수(Q7) |
  | `exclude_caution` | `true` | U0 · U1 · U2 랭킹 요청 값 |
  | `price_floor_krw` | 1000 | 1 주 가격 하한(상한은 `params.NotionalCapKRW` — 이미 판에 있음, evaluate.go:17 태그 없음) |
  | `symbol_pattern` | `^[0-9]{6}$` | 심볼 거름(D4) |
  | `numerator` | `ranking-trading-volume` | U2 분자. Q3 대체식(1 분봉 합)을 쓰면 `minute-bar-sum` — 판이 바뀜(§4.4) |
  | `baseline_days` · `daily_fetch` | 20 · 25 | 분모 일수 · 일봉 요청 수 |
  | `daily_boundary` | `before-session-date-midnight-kst` | 직전 일봉 경계(§4.2, review §4 P1-1) |
  | `prefetch_per_minute` · `prefetch_stop_before_ns` | 8 · 10 초 | 일봉 미리 읽기(§7.2) |
  | `late_selection_ns` | 60 초 | 늦은 선정 기준(§2) |

  시작 호가 생략 여부는 칸이 없다 — D2 로 「생략 안 함」이 고정이라(§7.3) 판을 가를 값이 아니다.
- **A140 재구성 30 고정(review §6 P2-1)**: `--top` 은 지금 1..100 을 받아 그대로 `cfg.Top` 에 넣고(main.go:75 · 89~90 · 107) `selectUniverse` 가
  그 값을 쓴다(shadow.go:271). 그래서 `run --universes a169` 에서 `--top` 값이 30 이 아니면 **브로커 호출 전에 거부**한다(오류 문구: 「a169 는
  A140 재구성을 30 으로 고정 — 종목군 크기는 `per_universe`」). 무시하지 않고 거부하는 이유: 무시하면 사람이 넣은 값과 원장 값이 조용히 갈린다.
  기본값(30)은 통과한다. 그래서 a169 모드에서는 `effectivePolicy.top` = 30 이 늘 참이고, A140 cohort 는 30 · U0/U1 은 `per_universe`(30 또는 20)다.
  기본 모드(a140)의 `--top` 동작은 그대로다.
- **판 digest 불변 시험(Q6 의 조건)**: 아카이브된 a140 원장
  (`openspec/changes/archive/2026-10-22-a140-…/analysis/shadow-ledger/ledger.jsonl`, sha256 `0b802a92…740c`)의 universe 행 5 개는 모두
  `policy_digest` `sha256:3ff46b8b0f7648c6008ed7a3ae463a9262af26fe1c1cc74267d97b26417101ff` 다. 시험은 그 행의 `effective_policy` 에서 `Top` ·
  `Params` · `session_source` 만 꺼내 `defaultProbeConfig()` 에 넣고 `effectivePolicyFor(...).digest()` 가 그 값과 **같음**을, a169 모드에서는
  **다름**을 단언한다. 기대값은 실행 중인 시스템이 아니라 커밋된 측정 산출물에서 온다.
- **원장 위치**: a169 모드는 `--out` 을 명시해야 한다(기본값 `defaultLedgerDir` 는 아카이브로 사라진 a140 경로, main.go:42). 측정 원장 =
  `openspec/changes/a169-…/analysis/shadow-ledger/`, 리허설 원장 = `…/analysis/rehearsal-ledger/`(측정에 안 섞이게 파일을 가른다).
  evidence DB 는 도구 전용 새 파일(`--evidence-db`, 생산 DB 가드 main.go:96 그대로).
- 원장 schema 문자열은 `a140-signal-shadow/v1`(ledger.go:15) 그대로 두고 행 종류만 더한다(§8) — `readLedger` 가 schema 를 등식으로 검사하므로
  (ledger.go:153) 바꾸면 a140 `report` 가 a169 원장을 못 읽는다. a140 `report` 는 모르는 행 종류를 건너뛰므로(report.go:59 switch) a169
  원장에 대해 「측정 종목 합집합 기준」 퍼널을 그대로 낸다 — 교차 확인용.
- **a140 `report` 의 a169 원장 출력은 측정치로 인용하지 않는다(review §4 P3-6)**. 그 합계에는 어느 종목군에도 귀속되지 않는 라벨이 섞인다 —
  U2 에만 든 종목이 선정 전에 돌파했으면 09:31:05 첫 평가에서 바로 ARMED 로 가 `armedAttempt` 가 MARKET_DATA 를 읽고 제안 · 라벨까지 낸다
  (`stepSymbol` B17@414). 그 읽기를 막으면 a140 과의 신호 동일성이 깨지므로 막지 않고, 비용으로만 센다(§7.1). 측정치는 `report-universes`(§9) 뿐이다.

## 2. 세션 타임라인 (Q1 · Q2)

| 시각(정규장 개장 09:00 기준) | 일 | 근거 |
|---|---|---|
| 개장 + 30 초 = 09:00:30 | U0 · U1 선정 + a140 재구성 랭킹 1 GET(Q7). 측정 종목 = U0 ∪ U1 을 universe 행으로 | `run` B3@547 이 이미 개장 + 30 초까지 기다린 뒤 B4@550 `startSession` |
| 09:01:05 ~ 09:30:05 | 분 폴(a140 그대로) + 남는 시간에 U2 후보 일봉 미리 읽기(§7.2) | `run` 루프 B5@553, 분마다 `:05`(554) |
| 개장 + 30 분 30 초 = 09:30:30 | U2 랭킹 2 GET → 새 후보 일봉 보충 → U2 선정 → 새 종목을 측정 종목에 더함 → universe 행(합집합) 다시 씀 → 유형 영수증 `/stocks` 1 GET. **마감 = 다음 분 폴 10 초 전(정상일 09:30:55)** — 그때까지 선정을 못 끝내면 그 세션 U2 는 측정하지 않는 「늦은 선정」(§4.5) | 새 함수, `run` 루프에서 호출 |
| 09:31:05 ~ 마감 + 30 분 | 분 폴(U0 ∪ U1 ∪ U2, 최대 90) · 라벨 | `step` B1@317 이 `p.symbols` 를 순서대로 돎 |
| 그 뒤 | `finish`(a140 그대로) → 돌파 봉 행 · 분자 대조 행(§4.4 · §5) | `run` 565 |

- **U2 의 시각은 「개장 + 30 분 30 초」로 정의**한다(정상일 09:30:30 KST = Q1 의 값). 공식 달력이 개장을 늦춘 날에도 같은 경과 시간에 고르기 위해서다
  — Q1 원문은 시계 시각이라, 이 해석은 review §3 에 사람 확인 항목으로 올린다.
- **세션당 한 번 고정(Q2)**: 선정 결과는 `cohort` 행(§8)에 남고, 재시작은 원장에서 되살린다. 측정 종목 집합은 a140 과 같이 **마지막 universe 행**
  (startSession B3@182 · B6@189 → 196)에서 되살아나므로, U2 선정 뒤에 쓰는 universe 행은 합집합 전체를 담는다.
- **늦은 선정**: 선정 행은 `target_at` 과 실제 `selected_at` 을 같이 적는다. 차이가 60 초를 넘으면(재시작 · 지연) 그 종목군-세션은 「늦은 선정」으로
  표시하고 판 합계에서 빼서 따로 보인다(§9). 선정을 거부하지는 않는다 — 다른 종목군까지 버리지 않게(Q10 과 같은 이유).
  U2 는 여기에 마감 하나가 더 붙는다(§4.5): 일봉 보충이 마감을 넘기면 선정 자체를 하지 않고 「늦은 선정」으로 남긴다.
- 마감 뒤에는 새 setup 을 평가하지 않는 a140 규칙(`step` B4@324 · `stepSymbol` B12@396)이 그대로 적용된다.

## 3. 종목군 정의와 공통 필터

### 3.1 공통 필터(U0 · U1 · U2 모두)

| 조건 | 값 · 출처 |
|---|---|
| 투자유의 제외 | 랭킹 요청 `excludeInvestmentCaution=true` — `Client.Rankings` 는 true 일 때만 쿼리에 붙인다(market_reads.go:367~369) |
| 심볼 형식 | `^[0-9]{6}$`(shadow.go:33) — a140 과 같음. **Q12 미결**: OpenAPI `/api/v1/candles` `symbol` 설명이 「KRX: 6자리 종목코드 (숫자 또는 영문·숫자 조합; 예: 005930, 0101N0)」라 영문 섞인 코드가 실제로 있다. 이 거름에 걸린 후보는 사유 `symbol_format` 으로 선정 행에 남겨 리허설 영수증에서 센다(§7.4) |
| 통화 · 가격 유효 | `Currency == "KRW"`, `LastPrice` 유한 · > 0(shadow.go:275~276 과 같은 식) |
| 1 주 가격 하한 | `LastPrice ≥ 1,000` — a140 에는 없던 조건(proposal 「a140 선정 방식」) |
| 1 주 가격 상한 | `LastPrice ≤ NotionalCapKRW` = 설정 `max_order_notional`(main.go:194~199, `--notional-cap-krw` 로 덮을 수 있음) |
| 가격 출처 | 그 선정 순간 랭킹 항목의 `LastPrice`(`domain.RankingItem`, models.go:706) — a140 의 상한 검사와 같은 값(shadow.go:274) |

거른 후보마다 사유(`caution_absent` 는 쓰지 않음 — API 가 빼므로 알 수 없음, `symbol_format` · `currency` · `price_invalid` · `price_below_floor` ·
`price_above_cap` · U2 전용 사유 §4.3)를 선정 행에 적는다(Q7).

### 3.2 U0 · U1 · a140 재구성

- **U0** = `Rankings(ctx, "MARKET_TRADING_AMOUNT", "KR", "realtime", true, 100)` 순위대로 공통 필터를 통과한 앞 N(`universePolicy.per_universe`, §1).
- **U1** = `Rankings(ctx, "MARKET_TRADING_VOLUME", "KR", "realtime", true, 100)` 순위대로 공통 필터를 통과한 앞 N. enum 은 OpenAPI
  `/api/v1/rankings` `type`(「`*_TRADING_VOLUME` → `tradingVolume`」 기준) 그대로.
- 두 랭킹은 같은 순간을 노리지만 GET 은 직렬이라(모든 GET 앞 250 ms, `retrying` 88~94 — B2@89 · B3@92) 수백 ms 떨어진다 — 각 랭킹의 `rankedAt` · 받은 시각을
  선정 행에 적는다.
- **`selected_at` 의 정의(review §4 P3-2)**: 그 종목군을 고르는 데 쓴 랭킹 둘 중 **늦게 받은 쪽의 받은 시각**(보충 일봉 · 유형 GET 의 끝이 아님).
  U0 · U1 · U2 모두 같고, 귀속의 T_C(§5.2)와 늦은 선정 판단(§2)도 이 값이다. 늦은 선정의 기준은 `target_at` 과 이 값의 차이다.
- **a140 재구성(Q7)**: 기존 `probe.selectUniverse`(shadow.go:264~285)를 **그대로** 한 번 부른다 — 그 함수가 바로 `excludeCaution=false` 랭킹 1 GET +
  a140 거름(6 자리 · KRW · > 0 · ≤ 상한 · 앞 30)이다. 결과는 `cohort` 행 `A140`(`measured: false`)으로만 적고 측정 종목에 넣지 않는다.
  U0 와의 차집합 사유는 U0 랭킹 영수증에서 유도한다: U0 랭킹에 있는데 빠졌으면 그 거름 사유(대개 1,000 원 미만), U0 랭킹에 없으면
  「투자유의(추정 — `true` 응답에 없음)」. 이 GET 이 실패하면 결손 행만 남기고 측정은 계속한다(재구성은 기록용).
- 시작 때 U0 · U1 랭킹 둘 중 하나라도 포기되면 a140 과 같이 결손 행을 남기고 시작을 거부한다(startSession B10@202 · B11@204 와 같은 처리).

### 3.3 N 에 못 미치는 종목군(Q10)

a140 `selectUniverse` 는 0 개일 때만 거부했다(B5@281). a169 는 각 종목군의 실제 수를 선정 행에 적고 N(`per_universe`) 미만이면 `short: true` 로 표시한다.
세션은 무효로 하지 않는다. 보고서는 판 합계를 「전부」와 「N 미만 종목군-세션 뺌」 둘로 낸다(§9).

## 4. U2 — 거래량 증가율 상위 N (Q3)

### 4.1 후보

09:30:30 에 `MARKET_TRADING_AMOUNT` · `MARKET_TRADING_VOLUME` 랭킹(둘 다 `true`, 각 100)을 다시 읽고, 공통 필터를 통과한 종목의 합집합(≤ 200,
심볼 중복 제거). 시장 전체가 아니라 **이 후보 안의** 순위다(보고서 한계 문구).

### 4.2 식 — 정수로

- 분자 `V` = 그 순간 랭킹 항목의 `tradingVolume`(`RankingItem.TradingVolume float64`, models.go:709). 거래량 랭킹에 있으면 그 값, 없으면 거래대금 랭킹의 값.
  유한 · ≥ 0 · 정수가 아니면 제외(`numerator_invalid`).
- 분모: 공식 일봉 `Client.Candles(ctx, symbol, "1d", 25, "", true)`(candle_reads.go:53) — a128 선례와 같은 요청(collect.go:568,
  `dailyFetchCount = 25` features.go:37). **직전 거래일 봉 = `Time` 이 0 이 아니고 KST 세션 날짜(공식 달력의 오늘 세션 날짜)의 00:00 보다
  앞선 봉**(`universePolicy.daily_boundary`). 시각 순 정렬 뒤 마지막 20 봉의 `Volume`(models.go:382) 합을 `S20` 으로 한다.
  - **왜 개장 시각이 아닌가(review §4 P1-1)**: 처음 초안은 a128 `dailyFeatures`(features.go:219~222, `Time.Before(open)`)의 경계를 빌렸는데,
    그 함수 주석(214~216)이 스스로 「자기 날짜의 자정 표기라면 오늘 봉이 직전으로 섞인다」고 적고 있고, 커밋된 a128 영수증이 바로 그 표기다.
    a128 `analysis/shadow-ledger/states/` 의 git 추적 KR state 7,406 개(관측일 8 일: 10-14 · 10-15 · 10-16 · 10-19 · 10-20 · 10-21 · 10-22 · 10-23)
    전부에서 `daily.last_prior_bar_at` 이 **관측일 당일 00:00 KST** 다. 예: `00b48ddf69da91b2…json`(003670) `observed_at`
    2026-10-19T00:20Z(월 09:20 KST) · `last_prior_bar_at` 2026-10-18T15:00Z = 2026-10-19 00:00 KST · `prior_bars` 25 — 전날은 일요일이라 이 봉은
    그날(월) 봉이고, 요청한 25 봉이 전부 「직전」으로 들어갔다. 개장 경계로는 오늘 부분 일봉이 S20 에 들어가고, 그 부분 거래량은 종목마다
    미리 읽은 시각(09:01~09:30, §7.2)에 따라 달라 U2 순위에 종목별 잡음을 넣는다.
  - 자정 경계는 실측 표기(그날 00:00 KST)에서 오늘 봉을 빼고, 어제 봉은 여는 시각 · 닫는 시각 · 자기 날짜 자정 어느 표기로 찍혀도 넣는다.
    유일하게 틀리는 표기는 「어제 봉을 오늘 00:00 으로 찍음(닫는 자정)」인데 위 영수증의 월요일 표본이 그 표기를 반증한다(그 표기라면 금요일 봉은
    토요일 00:00 이어야 함).
  - 리허설 영수증으로 다시 본다: 후보 행에 마지막 직전 봉 시각 `last_prior_bar_at` 과 그 다음 봉(있으면 오늘 봉) 시각 `next_bar_at` 을 남기고,
    `last_prior_bar_at` 이 직전 거래일 00:00 KST 가 아니거나 `next_bar_at` 이 오늘 00:00 KST 가 아닌 후보가 있으면 측정 전에 멈추고 사람에게 묻는다.
  - 직전 봉 < 20 → 제외(`history_lt_20`), 거래량이 유한 · ≥ 0 · 정수가 아님 → 제외(`baseline_invalid`), `S20 = 0` → 제외(`baseline_zero`),
    일봉 읽기 포기 → 제외(`baseline_unavailable`) + 결손 행. 사유는 선정 행에 남는다.
  - **한계 — 결측과 0 이 구분되지 않는다(review §4 P3-4)**: 랭킹 `tradingVolume` 은 파싱 실패가 0 이 되고(market_reads.go:394 `parseDecimal`),
    일봉 `Candle.Volume` 은 `float64 omitempty`(models.go:382)라 빠진 값도 0 이다. `numerator_invalid` · `baseline_invalid` 는 이것을 못 잡으므로,
    후보 행에 `V == 0` 여부와 직전 20 봉 중 거래량 0 인 봉 수 `zero_volume_days` 를 남기고 보고서 한계에 적는다(거르지는 않음 — 거래정지 일은
    실제로 0 일 수 있다).
- 순위 키: `ratio_ppm = floor(V × 20 × 10⁶ / S20)`(math/big — a140 sizing 의 big 정수 선례 sizing.go:96~106). 이것은
  「당일 누적 ÷ 20 일 일평균」이고, 선정 시각이 하나라 장 경과 비율은 모든 후보에 같은 상수여서 **순위를 바꾸지 않는다**(proposal Q1).
- 보고용 절대값: `increase_ppm = floor(ratio_ppm × (마감 − 개장) / (선정 시각 − 개장))`, 시각은 공식 달력의 정규장 창(`cal.Today.Regular`,
  shadow.go:543)을 ns 정수로. 선정 행에만 적고 순위에는 안 쓴다.
- 동률: 거래량 랭킹 순위가 높은 쪽 먼저(거래량 랭킹에 없는 후보는 있는 후보 모두 뒤), 그다음 거래대금 랭킹 순위, 그다음 심볼 오름차순 —
  전순서라 같은 입력은 같은 N 개를 낸다.
- 앞 N 을 U2 로 한다. U2 에 새로 든 종목(U0 ∪ U1 밖)은 다음 분부터 폴한다. 생산자는 페이지당 200 봉을 읽으므로(officialbars producer.go:131)
  09:31 의 첫 폴 한 페이지로 개장부터의 봉 31 개가 소급 적재된다.

### 4.3 선정 행에 남기는 것

후보마다 `{symbol, vol_rank, amt_rank, last_price, trading_volume, s20, prior_days, zero_volume_days, last_prior_bar_at, next_bar_at, ratio_ppm,
increase_ppm, excluded}`, 그리고 두 랭킹의 `rankedAt` · 받은 시각 · 항목 수. 일봉 원문은 원장에 넣지 않는다(크기) — `s20` · `prior_days` ·
`zero_volume_days` 와 경계 양쪽 봉 시각(§4.2)만.

### 4.4 분자가 「당일 누적」인지 대조(Q3)

OpenAPI 는 `MARKET_*` 가 시장 전체 기준이라고만 하고 `realtime` 누적 기준은 적지 않는다(proposal). 세션 끝에 도구가 `numerator_check` 행을 쓴다:
U2 후보 중 **개장부터 폴한 종목(U0 ∪ U1)** 마다, 도구 evidence 에 저장된 1 분봉 거래량(`strategyevidence` 봉 `Volume uint64`, breakout_bar.go:69)으로

- 기준 시각 `R` = 그 랭킹 항목이 온 랭킹의 `rankedAt`(파싱되면), 아니면 `selected_at`(review §4 P3-3 — 랭킹 집계가 수십 초 늦으면 선정 시각
  기준 경계는 거짓 불합격을 낸다),
- `lower` = `R` **전에 닫힌** 봉의 거래량 합, `upper` = `lower` + `R` 이 든 봉의 거래량,
- 판정 `lower ≤ V ≤ upper` 와 `R` · 그 출처(`ranked_at` | `selected_at`) · 불합격이면 크기 `V − upper` 또는 `lower − V` 를 적는다. 크기가 체계적으로
  한쪽이면(예: 늘 `V > upper`) 정규장 1 분봉 밖 거래(장전 시간외 · NXT 등)가 랭킹 누적에 든 것일 수 있다 — 미확인이고 리허설이 답한다(issues §3).

리허설(§7.4)에서 하나라도 어긋나면 Q3 의 대체식(분자 = 당일 1 분봉 합)으로 바꾼다. 그 대체식은 U0 ∪ U1 밖 후보마다 선정 순간 1 분봉 1 페이지를
더 읽어야 하므로(≤ 200 GET, MARKET_DATA_CHART) 판(lot)이 바뀌고 예산을 다시 잰다 — 측정 전에 사람에게 그 결과와 함께 묻는다.

### 4.5 U2 선정 마감 — 보충이 첫 폴을 밀지 않는다(review §6 P2-2)

문제: 모든 GET 은 250 ms 직렬이라(§7.1) 미리 읽기가 비어 있으면 보충 200 GET 만으로 약 50 초다. `run` 은 동기 작업이 끝난 뒤에야 다음 `:05` 를
계산하므로(shadow.go:553~563) 보충이 09:31:05 를 넘기면 U0 ∪ U1 의 09:31 폴까지 한 분 밀린다(결손 = a140 식 세션 유효성 손상). 그리고 `selected_at`
(랭킹 수신 시각, §3.2)과 실제 편입 시각이 갈린다.

- **마감 `D` = 개장 + 31 분 5 초 − `prefetch_stop_before_ns`(10 초) = 개장 + 30 분 55 초**(정상일 09:30:55). 미리 읽기의 「다음 분 폴 10 초 전에
  멈춤」(§7.2)과 같은 규칙이라 새 판 칸이 없다(판본 `a169-universe-policy/v1` 에 처음부터 든 규칙).
- 순서: U2 랭킹 2 GET(→ `selected_at`) → 캐시에 `S20` 이 없는 후보만 보충 → 선정 → universe 행 → `ready_at`. 보충 GET 하나를 시작하기 전에
  `now + ReadGap ≥ D` 이면 멈춘다. rate limit 재시도의 대기(`retrying` 5 · 10 · 20 초)도 `D` 를 넘기면 기다리지 않고 그 후보를 미완으로 둔다.
- **완료**(후보마다 `S20` 또는 확정 제외 사유가 있음) → 정상 선정. `ready_at` = 합집합 universe 행을 쓴 시각(≤ `D`)이라 U2 새 종목은 늘 09:31:05 첫 폴에
  든다. 귀속 T_C 는 그대로 `selected_at`(§5.2).
- **미완**(`D` 에 보충이 덜 끝남) → **그 세션 U2 는 고르지 않는다**(fail-closed). `cohort` 행 U2 = `measured: false` · `status: "late"` ·
  `late_reason: "daily_topup_deadline"` · `topup_needed` · `topup_done` · 빈 `symbols`. universe 행은 다시 쓰지 않고(측정 종목 = U0 ∪ U1 그대로),
  그 세션의 U2 는 판 합계 둘 다에서 빠져 「늦은 선정 — U2 마감」 수로만 보인다(§9). 나중 분에 다시 고르지 않는다 — 60 초를 넘긴 선정은 어차피
  판 합계에서 빠지므로(§2) GET 만 쓴다.
- 일부 후보만으로 순위를 매기지 않는 이유: `S20` 이 빠진 후보를 버리면 U2 정의(후보 전체 안의 순위, §4.1)가 세션마다 달라진다.
- 유형 영수증 `/stocks` 1 GET 은 `now + ReadGap < D` 일 때만 마감 안에서 읽고, 아니면 09:31:05 폴 **뒤** 남는 시간에 읽는다(미리 읽기와 같은 자리).
  유형은 보고서 갈래에만 쓰여 폴에 필요 없다. U2 가 늦은 선정이어도 U0 ∪ U1 의 유형은 읽는다.
- 재시작: `D` 전 재시작이고 U2 cohort 행이 없으면 다시 시도한다(메모리 일봉 캐시가 없으므로 대개 미완 → 늦은 선정). `D` 뒤 재시작이고 cohort 행이
  없으면 늦은 선정 행만 쓴다.
- 용량(측정 전 추정): `selected_at` 에서 `D` 까지 약 24 초 = 응답 시간 0 이어도 GET ≤ 96. 보충 대상은 09:00:30 뒤에 두 top-100 에 새로 든 후보뿐이라
  (미리 읽기가 09:00:30 후보를 덮음, §7.2) 보통 그보다 적을 것으로 보지만 **[미측정]** — 리허설 `budget` 행 `daily_topup` 과 U2 `status` 가 답한다.
  리허설에서 U2 가 늦은 선정이면 측정 전에 사람에게 묻는다(§12-8, issues §3 P3-2 의 09:2x 후보 갱신 선택지와 함께).

## 5. 귀속 필터와 중복 종목 (Q1 · Q5)

### 5.1 왜 필요한가 — 코드 사실

평가기는 opening range 뒤 봉을 순서대로 보다가 **처음** 돌파 조건을 만족한 봉에서 멈춘다(`evaluateFresh` B4@52 · B6@57 → `break` 66).
setup ID 는 첫 봉 · 15 번째 봉 ID 로 정해진다(machine.go:205~206). 그래서 **한 종목 · 한 세션에 돌파 봉은 하나**다. U2 를 09:30:30 에 고르면
09:16~09:30 에 이미 돌파한 종목이 높은 증가율 때문에 뽑히고, 그 setup 은 선정 이유와 같은 정보를 담는다(선행 정보).

### 5.2 규칙

- 평가는 종목당 한 번(측정 종목 합집합 기준) — 평가기 · 신호 행은 a140 과 같다.
- 종목군 C(선정 시각 T_C)에 setup 을 귀속하는 조건: 종목이 C 에 있고, **돌파 봉이 없거나**, 돌파 봉 i 의 닫는 시각 `개장 + (i+1) 분` 이 **T_C 보다 뒤**.
  T_C 이하에 닫힌 돌파의 setup 은 C 의 퍼널에서 「선정 전 돌파 — 귀속 제외」로만 센다. 그 종목은 그 세션에 다른 setup 이 없으므로(§5.1)
  C 에서 라벨 0 이다 — U2 가 첫 30 분 돌파를 못 세는 대가이며 보고서 한계에 적는다(proposal Q1).
- U0 · U1(T = 09:00:30)은 첫 가능 돌파 봉이 15 번(09:15 봉, 09:16 에 닫힘)이라 언제나 귀속된다 — a140 과 같은 결과.
- 늦은 선정 종목군은 실제 `selected_at` 을 T_C 로 쓴다(`selected_at` 의 정의 §3.2).
- **선정 시각 대조 열(review §4 P2-5)**: U2 대 U0 · U1 의 차이에는 「09:30:30 뒤 돌파만 센다」는 선정 시각 효과가 섞인다. 그래서 보고서는 U0 · U1 에
  같은 귀속 필터를 T_U2(그 세션 U2 의 `selected_at`)로 한 번 더 적용한 보조 열 「U0@T_U2 · U1@T_U2 — 돌파 봉이 T_U2 뒤에 닫힌 setup 만」을
  같은 9 칸 모양으로 낸다. 원장에 이미 있는 값만 쓰므로 GET 0 이다. 판정 칸이 아니라 대조용이다(판정 셋은 붙이지 않고 n · 평균 · 표준편차만).

### 5.3 돌파 봉 번호를 어떻게 얻나

`Decision` 은 돌파 봉 번호를 내보내지 않는다(Provenance 는 전이 이름만, types.go:197~201). a140 이 ARMED 봉을 찾은 방식(앞부분 길이를 늘려 가며
평가, `evaluateSetup` B9@124 · B10@126)을 따른다: 세션 끝에 새 함수가 종목마다 저장된 봉 묶음(`loadSeries`, shadow.go:351~354 — 브로커 GET 0)의
이어진 앞부분에서 n = 16, 17, … 로 `evaluateSetup` 을 불러 전이에 `BREAKOUT_CLOSED` 가 처음 나오는 n 을 찾고, i = n − 1 을 `breakout` 행에 적는다.

- 앞부분 평가가 전체 평가와 같은 근거: ATR 은 첫 15 봉만 읽고(`OpeningRangeATR` B2@125, 주석 117~118), RVOL 은 직전 15 봉만 읽는다
  (`Derive` B13@245 → `RVOLPPM` 246). 돌파 판정은 호가를 읽지 않는다(호가 검사는 ARMED 뒤 — `evaluateFresh` B19@114 `validateQuote`). 그래도 가정에 기대지 않게
  행에 `setup_id` 를 같이 적고, 보고서는 신호 행의 `setup_id` 와 다르면 「귀속 불일치」로 따로 센다.
- **장중에 본 돌파와 대조(review §4 P2-6)** — `setup_id` 비교만으로는 모자란다. setup ID 는 첫 봉 · 15 번째 봉의 `LaneBarID`(세션:여는 시각,
  signal.go:281~283)와 config digest 만 쓰므로(machine.go:205~206) 15 번째 뒤 봉이 장중 이후 정정돼 돌파 번호가 바뀌어도 같다. 그래서 장중 원장 행의
  앞부분 길이로 대조한다:
  - 장중 행 = 그 종목의 원장 행 중 `transitions` 에 `BREAKOUT_CLOSED` 가 **처음** 든 행. 신호 행은 `stepSymbol` 의 단계 변화(B18@417 — 같으면 행 없음)와
    `armedAttempt` 에서만 나오고 모두 `signalRow`(shadow.go:569~572)라 `bars` = 그 평가의 앞부분 봉 수(ledger.go:60, evaluate.go:122)와 `at` 을 싣는다.
    그 행의 `bars` 를 `N` 이라 하면 장중에 본 돌파 봉 번호는 `N − 1` 이하다.
  - `N = i + 1` → 일치. `N < i + 1` → 「귀속 불일치」(장중이 세션 끝 재계산보다 이른 돌파를 봄 = 봉 정정).
  - `N > i + 1` → 장중 평가가 늦었다(U2 새 종목의 소급 적재 · 재시작 · 결손 — 예: `armedAttempt` B3@446 호가 실패는 결손 행만 남기고 다음 분에 다시
    평가함). 이때 장중 번호를 정확히 알 수 없으므로 봉 `i` 와 봉 `N − 1` 의 닫는 시각이 T_C 의 **같은 쪽**이면 귀속이 같아 「일치(지연 N − i − 1 분)」로,
    **다른 쪽**이면 「귀속 불일치」로 센다(보수 쪽). 지연 분 수는 `breakout` 행에 남는다.
  - 세션 끝 재계산은 측정 종목 **전부**에 대해 돌린다(GET 0). 장중 행에 `BREAKOUT_CLOSED` 가 없는데 재계산이 돌파를 찾으면 그것도 「귀속 불일치」.
  - 「귀속 불일치」 종목-세션은 판 합계에서 빼고 수를 따로 보인다(늦은 선정과 같은 처리).
- 호가 입력: `evaluateSetup` 은 호가 봉인을 요구하지만(evaluate.go:90~96) 이 계산은 전이만 본다. 재시작 뒤 끝난 종목은 시작 호가를 읽지 않으므로
  (startSession B22@249 는 끝나지 않은 종목만) 메모리 호가가 없을 수 있다 — 그 종목 마지막 저장 봉의 종가로 만든 고정 호가(bid = ask = last, 관측 · 수신 =
  평가 시각)를 쓴다. 시험 5 가 서로 다른 호가 둘로 같은 번호가 나옴을 단언한다 — 호가가 결과에 들어가지 않음을 가정이 아니라 시험으로 못 박는다.
- 귀속 대상(review §4 P3-1): **어느 원장 행이든 `transitions` 에 `BREAKOUT_CLOSED` 가 있는 종목**. 처음 초안의 「가장 멀리 간 단계가 RETEST_WAIT 이상」은
  `phaseRank`(report.go:15)에 INVALIDATED · TIMED_OUT 이 없어, 소급 적재 · 재시작 뒤 RANGE_LOCKED → INVALIDATED 로 바로 간 종목을 「돌파 없음」으로
  잘못 셌다. `BREAKOUT_CLOSED` 가 어느 행에도 없는 종목은 「돌파 없음 — 귀속」. 그 종목의 신호 행이 하나도 없으면(세션 끝 전에 도구가 죽음)
  U2 귀속은 「미정」으로 센다.
- 이 계산을 분 루프(`stepSymbol`)가 아니라 세션 끝에 두는 이유: `stepSymbol` · `finish` 를 편집하지 않고(gate ⑤ 번들 둘 감소) 재시작에도
  같은 답을 낸다. 대가는 장중에 끊기면 U2 귀속이 미정으로 남는다는 것 하나다.

### 5.4 중복(Q5)

종목군별 합계는 중복을 포함한다. 보고서는 겹침 표(U0∩U1 · U0∩U2 · U1∩U2 · 셋 다)와 「그 종목군에만 든 종목」의 부분 집계를 함께 낸다.

## 6. 유형 영수증과 갈래 (Q9)

- 출처: `Client.Stocks`(stock_reads.go:44~52) → `domain.Quote.SecurityType`(models.go:429, 2.1 에서 노출). 세션당 **1 GET**: U2 선정 직후
  측정 종목 합집합(≤ 90 ≤ 200 — `symbols` 상한, OpenAPI `/api/v1/stocks`)을 한 번에 묻는다. 그룹 `STOCK`(초당 5, docs/migration/open-api.md:73).
- 실패: rate limit 은 a140 재시도(`retrying`), 그 밖은 결손 행 뒤 다음 세 분 동안 분마다 한 번 다시. 끝까지 못 받은 종목 · 응답에서 빠진 종목 ·
  빈 값은 「미분류」. 원문은 `types` 행에 `{symbol: 원문 | null}` 로 남긴다(a128 `tools/a128-d2-receipts/types.go:96~113` 과 같은 규칙 —
  null = 미분류).
- 갈래 대응 = a128 D2 §7 그대로(`prereg_eval.py:372~373` `D2_TYPE_GROUP`):

| 갈래 | securityType 원문 | 판정 칸 |
|---|---|---|
| STOCK | `STOCK` · `FOREIGN_STOCK` | 예 |
| ETF | `ETF` · `FOREIGN_ETF` | 예 |
| ETN | `ETN` | 예 |
| 기타 | `DEPOSITARY_RECEIPT` · `INFRASTRUCTURE_FUND` · `REIT` · `STOCK_WARRANTS` | 개수만 |
| 미분류 | enum 밖 · 빈 값 · 응답 없음 · 조회 실패 | 개수만 |

- 시험: 9 값 표 골든 + `docs/migration/openapi.latest.json` `components.schemas.StockInfo.properties.securityType.enum` 의 모든 값이 미분류가 아닌
  갈래로 가는지(enum 이 늘면 시험이 깨져 사람이 갈래를 정하게).

## 7. 호출 예산 (Q4)

### 7.1 세션당 GET

| 그룹(한도, open-api.md:67~78) | 읽기 | 수 |
|---|---|---|
| MARKET_INFO(초당 3) | 달력 | 1 |
| RANKING(표에 한도 없음) | 09:00:30 거래대금 `true` · 거래량 `true` · 거래대금 `false`(Q7), 09:30:30 거래대금 · 거래량 `true` | 5 |
| MARKET_DATA_CHART(초당 5) | 일봉: 미리 읽기 ≤ 200 + 선정 순간 보충(그때 새로 든 후보만) | ≤ 200 + 보충 |
| MARKET_DATA_CHART | 1 분봉 폴: 측정 종목 × 분당 1(폴 하한 규칙 a140 §2.1-4) | 09:31 부터 ≤ 90/분 |
| STOCK(초당 5) | 유형 | 1(실패 시 ≤ 3 재시도) |
| MARKET_DATA(초당 10. a169 는 별도 키라 엔진 `/prices` · `/orderbook` 과 공유하지 않음 — **한도가 키별일 때만**, §7.5 [미측정]) **〔3 req/s 계좌 공유로 재계산 필요 — 보류, issues §6〕** | 구조 호가 2: U0 ∪ U1 은 시작 때(개장 + 30 초, a140 그대로 — §7.3), U2 새 종목은 첫 평가 때(09:31:05 무렵) · ARMED 시도마다 2(선정 전 돌파한 U2 새 종목의 ARMED 포함 — §1) | 종목 수 × 2 + 시도 |

- 도구의 모든 GET 은 한 줄로 서서 250 ms 간격을 둔다(`retrying` 88~94 — B2@89 · B3@92 · `quoteRead`, source.go:107~126) — **종목 수와 무관하게 도구 전체 최대
  초당 4 GET** 이다. 종목 수가 바꾸는 것은 순간 속도가 아니라 그 속도가 지속되는 시간이다.
- MARKET_DATA_CHART 평균 점유: 90 종목이면 1.5 GET/s = 한도 5 의 30%(a140 의 30 종목은 10%). **〔3 req/s 계좌 공유로 재계산 필요 — 보류, issues §6〕** 2026-10-22 사용자의 「D11 지금대로」는 N = 30
  기준이라 승계되지 않는다 — 리허설 결과와 함께 다시 확인받는다(proposal Q4).

### 7.2 일봉 분산 일정(Q4 가 design 에 맡긴 것)

- 09:00:30 의 U0 · U1 랭킹(둘 다 `true`)으로 같은 정의의 후보(공통 필터 통과 합집합 ≤ 200)를 만들고, **09:01:05 ~ 09:30:05 의 분 폴이 끝난 뒤
  남는 시간에** 분당 최대 8 개씩 일봉을 미리 읽는다. 다음 분 폴 10 초 전이 되면 그 분의 미리 읽기를 멈춘다 — 분 폴을 밀지 않는다(결손 행이
  생기면 a140 식 세션 유효성이 깨지므로, proposal Q4). 29 분 × 8 = 232 ≥ 200.
- 직전 거래일 봉은 장중에 바뀌지 않는다(a128 collect.go:565 주석 「과거 봉은 하루 동안 변하지 않으므로 하루 캐시」, 그리고 §4.2 의 경계가
  오늘 봉을 버림). 그래서 09:0x 에 읽은 일봉으로 09:30:30 의 분모를 만들어도 값이 같다. 메모리 캐시라 재시작하면 다시 읽는다.
- 09:30:30 에는 그때 새로 든 후보만 보충해서 읽는다(보충 수를 `budget` 행 계수 `daily_topup` 에 남김). 미리 읽기 수는 `daily_prefetch`.
- 미리 읽기 시간은 `collect_seconds`(분 폴 시간, shadow.go:331)에 넣지 않고 따로 센다 — WARNING 기준을 a140 과 같은 뜻으로 유지하려는 것.

### 7.3 시작 호가 — a140 그대로 읽는다(D2, 2026-10-28 변경)

a140 은 시작 때 끝나지 않은 종목마다 구조 호가 1 줄(GET 2)을 읽는다(startSession B22@249 · B23@250 → 251). a169 모드도 **그대로 읽는다**.

- 처음 초안은 이것을 건너뛰어 「개장 직후 MARKET_DATA 묶음을 없앤다 · GET 120 절감」이라 적었고 사람 결정 D2 도 그 전제로 났다. review §4 P1-2 가
  그 전제가 틀렸음을 보였다: 시작 호가가 없으면 `stepSymbol` B13@399 → 400 이 첫 평가 직전에 **같은 2 GET** 을 읽고, `state.quote` 를 nil 로 되돌리는
  경로가 없어(대입은 shadow.go:256 · 404 · 449 뿐) 종목마다 세션에 정확히 한 번은 어느 쪽에서든 읽힌다. 생략은 GET 을 줄이지 않고 **시각만 옮긴다**.
  게다가 옮긴 곳은 흩어지지 않는다 — U0 ∪ U1 의 모든 종목이 첫 15 봉이 모이는 같은 분(B12@396 — 대개 09:15:05 또는 09:16:05)에 첫 평가에 들어가
  그 한 분에 몰린다(처음 초안의 「대개 09:17 이후로 흩어짐」도 틀림).
- 그래서 2026-10-28 사람 결정(「권장대로」)으로 D2 를 「생략 안 함」으로 바꿨다(§13). 얻는 것: ARMED 전 신호 행의 `snapshot_digest` · `quote_digest` 까지
  a140 과 같은 경로로 만들어져, spec 의 「a140 과 바이트 동일」을 setup · 단계 전이 · 라벨 전체로 되돌린다. 판(lot)을 가를 값도 없어진다(§1).
  U2 에 새로 든 종목은 시작 뒤에 들어오므로 첫 평가 직전(B13@399)에 호가를 읽는다 — a140 도 시작 호가가 없을 때(B23@250 실패) 타는 같은 경로이고
  평가기 입력(봉 · 그 호가)에 대해 같은 출력이다. 그 종목의 호가 시각이 U0 ∪ U1 과 다른 것은 늦게 들어온 탓이지 a169 의 다른 동작이 아니다.
- 비용(측정 전 추정): U0 ∪ U1 ≤ 60 종목이면 09:00:30 부터 GET ≤ 120 — 도구는 한 줄로 서서 초당 ≤ 4 GET(§7.1)이라 약 30 초 동안 MARKET_DATA 를
  쓴다(한도 10/s). **〔3 req/s 계좌 공유로 재계산 필요 — 보류, issues §6〕** 처음 초안의 「엔진 몫 ≥ 6/s」는 도구 자신의 직렬화만 본 계산이라 같은 키에서는 보장이 아니었다(review §6 P1-3) — 엔진 몫은
  별도 키(§7.5)로 지키고, 한도가 계좌별이면 그 보장도 없다. U2 새 종목 ≤ N 개는 09:31 첫 폴의 소급 적재로 앞부분이 바로 15 봉을 넘어 09:31:05 첫 평가 직전에 GET ≤ 2N 을
  봉 GET 과 섞어 읽는다. a140 실측 MARKET_DATA 는 세션당 70~135 GET(supplement 「호출 예산 실측」) 중 시작 호가 몫이 최대 60 이었다. 실측은 리허설
  `budget` 행이 하고, D11 재승인(§12-2)의 입력이다.

### 7.4 리허설 1 세션과 판단 기준

- 측정 5 세션 전에 KR 1 세션을 같은 모드 · 같은 판으로 리허설 원장(`analysis/rehearsal-ledger/`)에 돌린다. 측정에 안 센다.
- 판단 기준(a140 과 같음): `budget` 행의 10 분 창 수집 시간이 창의 절반(300 초)을 넘는 창이 하나라도 있으면 WARNING → 사람에게 다시 묻는다.
  `ReadGap` 250 ms 는 바꾸지 않는다.
- **예상(측정 아님)**: a140 의 30 종목 실측은 세션별 10 분 창 수집 시간 중앙 95.3~196.0 초 · 최대 254.5~359.0 초, 300 초 초과 창 0~2
  (a140 `analysis/report-2026-10-22-supplement.txt` 13~17 줄). 90 종목으로 비례하면 중앙이 약 290~590 초라 **WARNING 이 날 가능성이 높다.**
  그때 선택지는 (a) 종목군당 30 → 20(합집합 ≤ 60), (b) 그대로 받아들임(D11 재승인), (c) 종목군을 세션마다 번갈아 잼(사용자 결정 「같은 세션」을
  바꿈)이고, 사람 결정 D3(2026-10-28)이 **WARNING 이면 (a)** 로 정했다.
- **(a) 적용 뒤의 순서(review §4 P2-2)** — 90 종목 리허설은 20 판(합집합 ≤ 60)의 예산을 재지 않는다. 그리고 비례 추정으로 60 종목도 중앙 약
  190~390 초라 20 으로 줄여도 WARNING 이 남을 수 있다. 그래서:
  1. 리허설 1(N = 30 판)이 WARNING → `per_universe` = 20 인 **새 판**(§1)으로 바꾼다.
  2. 그 판으로 리허설 2 를 1 세션 더 돌린다 — 원장은 `analysis/rehearsal-ledger/` 의 다른 파일, 측정에 안 센다. 판단 기준은 위와 같다.
  3. 리허설 2 가 WARNING 아님 → 20 판으로 측정 5 세션. D11 재승인(§12-2)은 N ≤ 60 기준으로 받는다.
  4. 리허설 2 도 WARNING → 측정하지 않고 사람에게 다시 묻는다(남은 선택지 (b) · (c), 또는 그 밖). 이 design 은 그 다음을 고르지 않는다.
  - 리허설 1 이 WARNING 아님 → 30 판으로 측정, D11 재승인은 N ≤ 90 기준.
- 리허설이 함께 답하는 것: Q12(선정 행의 `symbol_format` 제외 수와 그 후보의 랭킹 정보, 유형 영수증의 ETN 행), Q3(`numerator_check`), 일봉 미리 읽기가
  분 폴을 밀었는지(결손 행 · `collect_seconds`).
- 리허설은 §7.5 의 엔진 429 확인과 한도 범위 판별도 함께 한다. 그 둘 중 하나라도 「공유 증거 있음」이면 측정하지 않는다(§7.5).

### 7.5 측정 전용 별도 앱 키(사람 결정 D5 — review §6 P1-3)

**왜**: 도구의 250 ms 직렬화와 429 정지는 도구 프로세스 안에만 있고(`pacedSource`, source.go:43~125), `official.Client` 의 rate budget 은 마지막
응답을 기록만 할 뿐 판단에 안 쓴다(client.go:46~49 · 191~206). 엔진과 같은 키면 도구 4/s 와 엔진 `/prices` · `/orderbook` 이 한 예산에서 합산되고,
엔진 보호 호가가 먼저 429 를 받을 수 있다. 엔진 쪽 예약(admission)을 만드는 것은 엔진 코드 변경이라 이 change 범위 밖이다. 그래서 **같은 예산을
나누지 않는 쪽**을 고른다: 2026-10-28 사용자가 측정 도구 전용 앱 키를 발급했고(「4는 가능」), a169 모드는 그 키만 쓴다.

**자격증명을 지금 어디서 읽나(HEAD `0c9bfd272`)**

| 자리 | file:line | 사실 |
|---|---|---|
| 도구 `--config-dir` 플래그 | tools/a140-breakout-signal-shadow/main.go:74 | 「tossctl config directory (Open API credentials + config.json)」 — 한 플래그가 자격증명 · 토큰 · config.json(한도) · evidence 가드 후보를 모두 정함 |
| config.json · evidence 가드 | main.go:92 `apppaths.Resolve` → 96 `productionEvidenceCandidates(paths.ConfigDir)` → 99 `loadLimits(paths.ConfigFile)` | 가드 후보는 guard.go:21~30(journal 기본 경로 옆 + `configDir` 안) |
| 자격증명 · 토큰 경로 | main.go:121 `apppaths.OpenAPI(*configDir)` → internal/app/paths/paths.go:48~61 | `configDir` 가 있으면 둘 다 그 안(B1@51 → 52), 없으면 기본값(B2@57 실패 아니면 60) |
| 기본 경로 | internal/config/paths.go:20~41 | 자격증명 `~/.config/tossctl/openapi-credentials.json`(:39), **토큰 `~/.cache/tossctl/openapi-token.json`**(:40) |
| 자격증명 읽기 | main.go:125 → internal/official/credentials.go:29~49 | **환경 변수 `TOSSCTL_OPENAPI_KEY` · `TOSSCTL_OPENAPI_SECRET` 이 둘 다 있으면 파일보다 우선**(B1@32 → 33), 파일 없음은 (nil, nil)(B3@38 → 39) |
| client 생성 | main.go:133 `official.New(*credentials, tokenFile)` | 토큰 캐시는 키와 묶이지 않음 — 파일에 아직 유효한 토큰이 있으면 그대로 씀(`tokenManager.token` B1@67 메모리 · B2@72 디스크 → 74, 캐시 형식 token.go:24~28 에 키 칸 없음) |
| 엔진 | internal/app/engine/engine.go:365~392 | 같은 `apppaths.OpenAPI(opts.ConfigDir)` · `LoadCredentials(getenv, …)`. 컨테이너는 `--config-dir /var/lib/tossos/config`(compose.yaml:18~19) = 호스트 `${TOSSOS_CONFIG_DIR}` bind(compose.yaml:44 · 125) |
| 키 등록 | cmd/tossctl/openapi.go:345~365 `openapi login --key --secret` · cmd/tossctl/init.go:60~96 `init`(숨김 입력) | 둘 다 루트 `--config-dir` 를 따름(openapi.go:97 → app.go:178). `login` 은 값이 명령 인자라 셸 기록 · 프로세스 목록에 남음 |
| 키 · 토큰 관계 | docs/migration/open-api.md:22 · OpenAPI `issueOAuth2Token` | 「client 당 유효 토큰 1 개 — 재발급 시 이전 토큰 즉시 무효」. 허용 IP 는 client 마다 등록(OpenAPI 403 `IP address not allowed`) |

토큰 전쟁 선례(2026-10-19 실측): `--config-dir` 없이 돈 도구가 기본 토큰 캐시 `~/.cache/tossctl/` 로 새 토큰을 사서 엔진(`~/.config/tossctl/`
의 토큰)을 무효화했다. 키가 다르면 토큰도 client 마다 따로라 이 전쟁은 구조상 없다 — 단 **키가 정말 다를 때만**이다. 아래 가드가 그것을 시작 전에 증명한다.

**배선 — `--openapi-dir <dir>`(a169 모드 필수) · `--config-dir` = 사람이 확인한 엔진 프로필(a169 모드 필수)**

- `--openapi-dir` 은 자격증명 · 토큰 **만** 정한다: `apppaths.OpenAPI(openapiDir)` → `<dir>/openapi-credentials.json` · `<dir>/openapi-token.json`.
- `--config-dir` 은 지금 뜻 그대로 **엔진 프로필** `E`(읽기 전용 config.json 한도 · evidence 가드 후보)이고, a169 모드에서는 **비우면 거부**한다(기본값
  `~/.config/tossctl` 로 떨어지지 않게 — review §8 P1-7). runbook 은 사람이 확인한 값 하나 `ENGINE_CONFIG_DIR` 를 `--config-dir` · G2/G5/G6 · 엔진 로그
  감시기 · 사후 mtime 확인에 똑같이 넘긴다(§14).
- 기본 모드(a140)는 바뀌지 않는다 — `--openapi-dir` · `--profile-preflight` 를 주면 거부한다(기본 모드는 아카이브된 판 재현용이고 측정에 쓰지 않음).
- 새 함수 `checkSeparateProfile`(새 파일 `profile.go`, §10.1)은 `runCommand` 에서 **일반 자격증명 loader(`LoadCredentials`, main.go:125) · client 생성(133) ·
  evidence 저장소 · 원장 · 브로커보다 먼저** 돈다(review §8 「확인한 것」 — G4~G6 자체는 파일을 읽으므로 「자격증명 읽기 전」이 아니라 이 순서가 계약).
  하나라도 걸리면 HTTP 요청 0 · 원장 파일 0 으로 거부하고, 오류에는 가드 번호와 경로만 싣는다.
- **`--profile-preflight`(review §8 P1-5)**: `run --universes a169 --openapi-dir … --config-dir … --profile-preflight` 는 플래그 검사 · G1~G6 만 하고
  `preflight ok: openapi-dir=<경로> engine=<경로>` 한 줄로 끝난다(rc 0). config.json · evidence · 원장 · 달력 · 네트워크 0 — `--evidence-db` · `--out` 도
  요구하지 않는다. runbook 은 이것을 `tossctl … openapi test`(네트워크로 토큰을 받음)보다 **먼저** 돌린다(§14-4).

**경로 정규화(review §8 P2-3)** — 모든 비교에 같은 알고리즘을 쓴다.

- `canon(p)` = `filepath.Abs(p)` → 그 경로가 있으면 `filepath.EvalSymlinks` → `filepath.Clean`. 없는 경로는 `Abs` · `Clean` 까지만(있어야 하는 경로는 따로
  존재 검사로 거부).
- `inside(a, b)` = `rel, err := filepath.Rel(canon(b), canon(a))` 가 오류 없이 `.` 이거나, `..` 도 아니고 `../` 로 시작하지도 않음. 문자열 접두사 비교는
  쓰지 않는다(`/x/engine2` 는 `/x/engine` 의 안이 아님).
- 비교 대상: `E`(= `canon(--config-dir)`), 기본 설정 `canon(~/.config/tossctl)`, 기본 캐시 `canon(~/.cache/tossctl)`(둘 다 `apppaths.Resolve("", "")` 의
  `ConfigDir` · `CacheDir` — 그 함수는 받은 경로를 그대로 `Join` 하므로(paths.go:31~36 · 48~54) 정규화는 이 함수가 한다).

**fail-closed 가드(G1~G7)**

| # | 거부 조건 | 막는 것 |
|---|---|---|
| G1 | a169 모드인데 `--openapi-dir` 또는 `--config-dir` 이 비었음 | 기본 경로(엔진 자격증명 + `~/.cache/tossctl` 토큰) = 10-19 토큰 전쟁 · 사람이 확인 안 한 E |
| G2 | (a) `--openapi-dir` 이 없거나 디렉터리가 아님, 또는 `inside(openapiDir, x)` 가 참인 `x` ∈ {E, 기본 설정, 기본 캐시}. (b) **E 가 엔진 프로필이 아님**: `E/engine.lock`(enginelock.go:82 `LockFileName`) · `E/engine.log`(engineproc.go:67 `engineLogName`) · `E/openapi-credentials.json` 중 하나라도 일반 파일로 없음 | 엔진 프로필 · 기본 토큰 캐시 재사용, **틀린 E**(엔진이 안 쓰는 디렉터리를 비교 대상으로 삼는 것 — review §8 P1-7) |
| G3 | 환경 변수 `TOSSCTL_OPENAPI_KEY` 또는 `TOSSCTL_OPENAPI_SECRET` 이 비어 있지 않음 | `LoadCredentials` 의 환경 우선(B1@32 → 33)이 `--openapi-dir` 을 조용히 건너뜀 |
| G4 | `<dir>/openapi-credentials.json` 이 일반 파일이 아님(`Lstat` — 링크 거부) 또는 권한이 `0600` 이 아님 | 링크로 엔진 파일을 가리킴 · 권한이 무의미한 파일 시스템(fuseblk 등) 위의 비밀. 키 · 비밀이 빈 경우는 G1~G6 뒤 일반 loader 결과로 `runCommand` 가 거부(엔진 engine.go:386~388 과 같은 검사) |
| G5 | 엔진 `E/openapi-credentials.json` 의 `apiKey` 를 못 읽음, 또는 도구 파일의 `apiKey` 와 같음(`crypto/subtle.ConstantTimeCompare`) | 엔진 키 재사용. 못 읽으면 「다름」을 증명 못 하므로 거부 |
| G6 | 도구 토큰 `<dir>/openapi-token.json` 이 있고 **파싱한 `access_token` 이** 엔진 토큰 캐시 `E/openapi-token.json` · `canon(~/.cache/tossctl)/openapi-token.json` 중 하나의 파싱한 `access_token` 과 같음(상수 시간 비교, 빈 값끼리는 같다고 보지 않음). 보조: 도구 토큰 경로가 링크이거나(`Lstat`) 엔진 토큰과 `os.SameFile` 이면 거부. 엔진 토큰이 있는데 못 읽거나 파싱 못 하면 거부. 도구 토큰을 파싱 못 하면 통과(그 경우 `tokenManager` 는 캐시를 버리고 새로 받음 — `loadCache` B3@190 → 191, `token` B2@72) | 토큰 복사 · 재직렬화 · 하드링크 · 링크 — 토큰 캐시에 키 칸이 없어(token.go:24~28) 유효한 엔진 토큰을 그대로 씀(`token` B2@72 → 74). **바이트 비교로는 공백 · 필드 순서만 바꾼 사본이 통과한다**(review §8 P1-4) |
| G7 | 정적 시험(§10.3-13) — `profile.go` · `enginewatch.go` 가 키 · 비밀 · 토큰 · 엔진 로그 줄을 `fmt` · `log` · `os.Stderr` · 원장 쓰기 호출의 인자로 넘기지 않고, `profile.go` 가 `official.Credentials` · `LoadCredentials` · 문자열 `secretKey` · 필드 `SecretKey` 를 이름으로 쓰지 않음 | 값 출력 · 엔진 비밀의 Go 값 생성 |

- **좁은 reader(review §8 P2-4)**: G5 는 엔진 파일을 `struct{ APIKey string `json:"apiKey"` }` 로만 디코드한다 — `SecretKey` 를 담는 Go 값(`official.Credentials`)을
  만들지 않는다. 도구 쪽 비교 키도 같은 좁은 reader 로 읽는다. G6 도 `struct{ AccessToken string `json:"access_token"` }` 만 쓴다. 한계: 디코드할 때
  파일 바이트(엔진 비밀 포함)는 버퍼에 잠시 있고 Go 는 그 버퍼를 지우는 것을 보장하지 않는다 — 버퍼는 그 함수 밖으로 나가지 않는다. 신뢰된 등록
  단계가 남긴 지문만 읽는 더 강한 격리는 등록 도구가 하나 더 필요해 넣지 않았다(issues §5).
- 원장에는 `cohort` U0 행에 `separate_key_checked: true` 한 칸만 남긴다(경로 · 키 정보 0). 시작 stderr 한 줄에는 두 디렉터리의 **경로**만 더한다.

**엔진 rate limit 실시간 감시기(review §8 P1-6) — 엔진 편집 0**

한도 단위가 [미측정]인 동안의 interlock 이다. 사후 계수와 「보유 0 인 날」 권장은 **대체물이 아니다**(아래 권장은 유지하되 보조).

- 엔진 로그 위치(코드 사실): 콘솔이 띄운 엔진은 `--config-dir` 이 있으면 그 디렉터리를 엔진 디렉터리로 쓰고(`engineJournalDir` B1@186 → 187),
  `startEngine` 이 `engineLogPath(dir)`(engineproc.go:152, 208)로 `engine.log` 를 열어 엔진 프로세스의 stdout · stderr 를 거기에 붙인다
  (`spawnDetachedEngine` 440 `O_APPEND|O_CREATE|O_WRONLY` · 446~447). 운영 컨테이너는 `--config-dir /var/lib/tossos/config`(compose.yaml:18~19)로 뜨므로
  호스트에서는 `ENGINE_CONFIG_DIR/engine.log` 다. 회전 · 잘림 코드는 `cmd/tossctl/engine*.go` · `internal/app/engine` 에서 찾지 못했다(`rotat|truncat` 검색 0).
- 엔진 429 문구(코드 사실): 브로커 429 는 `classifyStatus` B4@54 → 55 에서 `ErrRateLimited` = `official: rate limited`(errors.go:15) 로 바뀌고 HTTP 코드를
  싣지 않는다(401 · 403 만 `(HTTP %d)`, B3@50 → 51 · 53). 엔진이 **모든** 429 를 로그 한 줄로 남기는지는 **[미측정]** — `internal/app/engine` 에 rate limit
  전용 로그 줄은 없고(검색 결과는 precheck.go:41 주석 1 건뿐) 오류 문자열이 그대로 찍히는 경로에서만 보인다. 그래서 패턴은 보수 쪽: 대소문자 무시 `rate limited` · `rate_limited` ·
  `too many requests` · 단어 경계 `429`. `429` 는 가격 · 수량 · id 에도 걸릴 수 있다 — 거짓 양성은 세션 하나를 버리는 쪽(보수)이다.
- 새 파일 `enginewatch.go` 의 감시기:
  1. 시작(가드 뒤, evidence · 원장 · client 전): `ENGINE_CONFIG_DIR/engine.log` 를 **읽기 전용**으로 열고 일반 파일인지 · 크기를 보고 **끝으로 이동**한다
     (도구 시작 전 줄은 보지 않음). 못 열면 시작 거부(`engine_log_unreadable`).
  2. 250 ms(= `ReadGap`)마다: 새로 붙은 바이트를 읽어 패턴을 본다(줄바꿈 전 조각도 본다 — 보수). 같은 경로를 다시 `Stat` 해서 열어 둔 파일과
     `os.SameFile` 이 아니면 회전(`engine_log_rotated`), 크기가 읽은 위치보다 작으면 잘림(`engine_log_truncated`), 읽기 오류면 `engine_log_unreadable`.
  3. 패턴을 보면 `engine_rate_limited`. 어느 경우든 `context.WithCancelCause` 의 cancel 을 그 사유로 부른다.
  4. 감시기 고루틴이 어떤 이유로든 끝나면 `defer` 로 `watcher_stopped` 취소. 별도 점검 고루틴이 마지막 폴 시각이 2 초보다 오래되면(파일 시스템 멈춤)
     `watcher_stalled` 로 취소한다.
- 취소가 GET 을 멈추는 근거: 도구의 모든 GET 은 `pacedSource.pause` 를 먼저 지나고(`retrying` B3@92 · `quoteRead` 115), `pause` 는 지연이 0 이어도
  `ctx.Err()` 를 돌려준다(B1@78 → 79). 그래서 취소 뒤 **새 GET 은 0** 이고, 이미 나간 GET 하나는 요청 context 로 끊기지만 서버에 닿았을 수 있다(최대 1).
  첫 엔진 429 자체는 막지 못한다 — 감시기는 그 다음을 막는다.
- 기록: `runCommand` 가 `context.Cause` 를 보고 원장에 `abort` 행(`reason` = 위 사유 하나, 로그 줄 내용 · 매칭 문자열은 **안 남김**)을 쓰고 rc 1 로 끝난다.
  그 세션은 `report-universes` 의 판 합계 둘 다에서 빠지고 `aborted: <reason>` 수로만 보인다. 같은 날 다시 띄우면 원장에 오늘(KST 날짜로 만든
  `KRX:<date>`)의 `abort` 행이 있으므로 달력 GET 전에 거부한다(사람이 원인을 본 뒤 다른 날 다시).
- 엔진을 고치지 않는다: 감시기는 엔진이 이미 쓰는 파일을 읽기만 한다. 엔진 쪽에 신호 · 중단 권한을 새로 만들 필요는 없었다(멈추고 보고할 조건 아님).
- 한계 **[가정]**: 운영 엔진이 콘솔의 `startEngine` 으로 떴다는 것(그래야 stdout 이 `engine.log`). 다른 방식으로 뜨면 그 파일이 자라지 않는다 —
  runbook §14-5 가 세션 중 `engine.log` 크기가 느는지 사람이 확인하고, 안 늘면 감시가 눈먼 것이므로 측정에 안 센다. 파일 권한(0600, 컨테이너 UID)이
  도구를 돌리는 호스트 사용자와 맞는지도 [가정] — 안 맞으면 1 단계가 시작을 거부한다(fail-closed).

**한도가 키별인가 계좌별인가 — [미측정]**

공식 문서는 그룹별 초당 한도만 적고(open-api.md:67~78) 그 단위(키 · 계좌 · IP)를 적지 않는다. 두 키가 같은 계좌에 묶여 있으므로 계좌별이면 별도 키도
MARKET_DATA 예산을 나눈다. 그래서 별도 키는 **토큰 전쟁은 구조로 닫고, 예산 공유는 실시간 감시기로 끊고, 리허설로 판별한다**:

1. 실시간: 위 감시기 — 엔진 rate limit 줄 하나면 그 세션 중단.
2. 도구 쪽: 도구가 받은 429 는 결손(`gap`) 행의 사유 문구(`probe.gap` 141 `err.Error()` 그대로)에 `official: rate limited` 로 남고, MARKET_DATA 면
   `quoteRead` B3@120 이 정지를 걸어 그 뒤 읽기가 B1@112 → 113 `quote reads are suspended…` 사유로 이어진다(source.go:39). 리허설 원장에서 이 두 문구가 든 `gap` 행을 센다.
   도구는 초당 4 GET 을 넘지 못하므로(§7.1) MARKET_DATA(10/s) 429 를 받았다면 그 예산을 다른 소비자가 같이 쓴다는 신호다. **〔3 req/s 계좌 공유로 재계산 필요 — 보류, issues §6〕**
3. 판단: 리허설이 `abort`(어떤 사유든)로 끝났거나 도구 MARKET_DATA 429 > 0 → 측정 0, 사람에게 묻는다(선택지 예: 공유를 받아들이고 a140 처럼
   D11 로 진행 · 도구 속도 낮춤 · 중단 — 이 design 은 고르지 않음). 둘 다 아님 → §7.4 순서대로 측정. 측정 세션에서도 감시기는 같은 규칙으로 돈다.
- 이것은 **키별의 증명이 아니다** — 엔진이 그 창에 MARKET_DATA 를 얼마나 불렀는지에 따라 0 이 우연일 수 있다. 응답 헤더 `X-RateLimit-Remaining`
  (`(*Client).RateBudgets()`, ratebudget.go:217~220)을 원장에 남기면 GET 0 으로 공유 여부를 더 직접 볼 수 있지만, `official.RateBudget` 이름을 쓰려면
  기존 시험 `TestNoMutatingAccountOrWriteSurfaceIsNamed`(static_test.go:70~115)의 허용 목록을 고쳐야 해 FLM 대상이 하나 는다 — issues §4 에 선택지로 남긴다.
- 보조(대체물 아님): 엔진 보호 호가가 의미 있는 것은 엔진이 포지션을 들고 있을 때뿐이라, runbook 은 리허설을 엔진 보유 0 인 날에 하기를 **권장**한다
  (사람 결정 §12-7). 감시기 · 사후 로그 계수를 빼도 된다는 뜻이 아니다.

## 8. 원장 — 행 종류 추가 (schema 문자열 불변)

| 종류 | 언제 | 칸 |
|---|---|---|
| `universe`(기존) | 09:00:30 · U2 선정 뒤 | `symbols` = 그때까지의 **측정 종목 합집합**(재시작 복원용, a140 의미 그대로) |
| `cohort`(새) | 선정마다(U0 · U1 · A140 · U2) | `cohort` · `target_at` · `selected_at` · `measured` · `symbols`(순위 순) · `short` · 랭킹 영수증(`type` · `exclude_caution` · `ranked_at` · 받은 시각 · 항목 수 · 상위 100 항목) · 후보별 제외 사유 · (U2) §4.3 칸 · (U2) `status`(`on_time` · `late`) · `late_reason` · `topup_needed` · `topup_done` · `ready_at`(§4.5) · (U0) `separate_key_checked`(§7.5) |
| `types`(새) | U2 선정 뒤 1 회(+ 재시도) | 받은 시각 · 호출 수 · 물은 종목 · `{symbol: 원문 \| null}` |
| `breakout`(새) | 세션 끝 | `symbol` · `setup_id` · `breakout_index` · `breakout_bar_open_at` · `breakout_bar_close_at` · 장중 대조(§5.3): 장중 행의 `at` · `bars` · 판정(`match` · `lagged` · `mismatch`) · 지연 분 수 |
| `numerator_check`(새) | 세션 끝 | `symbol` · `V` · `lower` · `upper` · `pass` · 기준 시각 `R` 과 출처 · 불합격 크기(§4.4) |
| `budget`(기존) | 10 분마다 | 계수에 `daily_prefetch` · `daily_topup` · `types_calls` 추가 |
| `gap`(기존) | 읽기 포기 | 새 단계 이름 `ranking_u2` · `ranking_a140` · `daily` · `types` |
| `abort`(새) | 감시기 취소로 끝날 때 1 회 | `at` · `session` · `reason`(`engine_rate_limited` · `engine_log_rotated` · `engine_log_truncated` · `engine_log_unreadable` · `watcher_stopped` · `watcher_stalled`, §7.5). 엔진 로그 줄 · 매칭 문자열은 넣지 않음 |

- 모든 새 칸은 `omitempty` 이고, 구조체 칸은 **포인터**로 둔다 — `encoding/json` 의 `omitempty` 는 구조체 값을 빼지 않으므로(포인터 · 슬라이스 · 맵 ·
  기본형만 빠짐) 값 구조체로 두면 a140 모드 행에 빈 객체가 생긴다(review §4 P3-5). 그래서 a140 모드 행의 바이트는 그대로다. 행 하나의 최대 크기는 랭킹 100 항목 × 2 정도 — `readLedger` 의 줄 상한 4 MiB
  (ledger.go:145) 안이다.
- 원장에 들어가는 설정값(`effective_policy.params` 의 가상 위험 예산 · 명목 상한)은 a140 원장과 같은 칸이다 — 계좌 데이터 · 잔고 · 보유는 0.

## 9. 보고서와 판정 (Q8 · Q10)

- 새 하위 명령 `report-universes --ledger <path>`(새 파일). 판(lot)마다, 종목군(U0 · U1 · U2) × 갈래(STOCK · ETF · ETN) **9 칸**, 기타 · 미분류는 개수.
- 칸마다: 퍼널(종목 → RANGE_LOCKED → BREAKOUT_CLOSED(귀속) → RETEST_WAIT → ARMED → PROPOSED), 「선정 전 돌파 — 귀속 제외」 · 「귀속 미정」 ·
  「귀속 불일치」 수, 라벨 수 n · 결과 분포, 그리고 **n · 평균 순 R · 표준편차(표본, n − 1) · 표준오차**(review §4 P2-4). 표준편차는 `r_ppm` 정수에서
  math/big 으로 계산하고 표시할 때만 소수로 바꾼다.
- 지표 정의는 a140 과 같다: 승률 = 목표 선착 비율(report.go:130 · 147), 평균 순 R = Σ `r_ppm` / n / 10⁶(report.go:131 · 148), `r_ppm` 은 왕복
  비용을 뺀 순 R(`rMultiplePPM` sizing.go:96~106).
- **판정(Q8)** — 판 합계(같은 판의 세션 합침, a140 design §2.1-14 ⑪)에서:
  - n < 30(`minVerdictSamples`, report.go:12) → 「보류」
  - n ≥ 30 이고 평균 순 R > 0 → 「후속 검토 후보」
  - n ≥ 30 이고 평균 순 R ≤ 0 → 「중단 근거 유지」
  - 세션끼리 상관이 있어 통계 판정이 아니라 사람 판단의 입력이다(보고서 머리 문구).
  - **칸 옆 표기(review §4 P2-4)**: 판정 셋을 낸 칸마다 「통계 판정 아님 · 9 칸 동시 비교」를 붙인다. 기대값 0 근처에서는 9 칸 중 몇 칸이 우연히
    > 0 이 나올 수 있다(다중 비교). 결정된 임계(Q8)는 바꾸지 않는다 — 표기와 표준오차는 읽는 사람이 그것을 보게 하려는 것이다.
- 판 합계는 두 번 낸다(Q10): (1) 정시 선정 종목군-세션 전부, (2) 그중 N 미만 종목군-세션을 뺀 것. 두 판정이 다르면 그 칸에 「표본 경계 —
  사람 판단」을 붙인다. 늦은 선정 종목군-세션(§2 — U2 마감 미완 §4.5 포함, 사유별 수)과 「귀속 불일치」 종목-세션(§5.3)은 두 합계 모두에서 빼고 따로 보인다.
  `abort` 행이 있는 세션(§7.5 감시기)은 세션 전체를 두 합계에서 빼고 `aborted: <reason>` 수로만 보인다.
- **예상(측정 아님) — 5 세션이면 9 칸 거의 전부 「보류」(review §4 P2-3)**: a140 은 30 종목 × 5 세션 = 종목-세션 150 에서 라벨 22 였다(proposal Why).
  U0 는 a140 과 거의 같은 정의(투자유의 제외 · 1,000 원 하한만 다름)라 U0 전체 라벨도 20 개 안팎으로 예상되고, 이를 STOCK · ETF · ETN 셋으로
  나누면 칸마다 30 미만이다. U1 도 비슷하고, U2 는 첫 30 분 돌파를 못 세므로(§5.2) 더 적다. D3 (a) 로 20 판이 되면 더 적다. 그래서 측정 전에
  사람에게 확인받는다(§12-5): 대부분 칸 보류일 때 (i) 세션을 늘릴지(몇 세션), (ii) 갈래를 합친 종목군별 보조 집계(U0 · U1 · U2 × 전 갈래 3 칸)를
  같이 낼지 — 보조 집계를 내더라도 판정 칸은 9 칸 그대로이고 보조 집계에는 같은 「통계 판정 아님」 표기를 붙인다. 이 design 은 고르지 않는다.
- 함께 내는 것: 겹침 표 · 고유 종목 부분 집계(Q5), U0 대 A140 차집합과 사유(Q7), 호출 예산(그룹별 GET · 10 분 창 수집 시간 · 미리 읽기 수),
  `numerator_check` 합격 수(Q3), a140 비교 기준(인용 — 종목-세션 150 → PROPOSED 22 · 라벨 22, n < 30 보류, proposal Why). a140 수치를 다시
  계산하지 않는다.
- 함께 내는 보조 열: U0@T_U2 · U1@T_U2(§5.2 — 선정 시각 대조, GET 0).
- 한계 문구: U2 는 후보 안의 순위 · 첫 30 분 돌파를 못 셈(§5.2) · U2 와 U0 의 차이에 선정 시각 효과가 섞임(보조 열로 크기를 보임) · 분자 · 분모의
  결측이 0 과 구분되지 않음(§4.2) · 체결 가능성 미측정(label.go:17).

## 10. 구현 표면

### 10.1 새 파일(gate ⑤ 대상 아님)

| 파일 | 내용 |
|---|---|
| `tools/a140-breakout-signal-shadow/profile.go` | `checkSeparateProfile` — §7.5 의 경로 정규화 · G1~G6(경로 · E 신원 · 환경 · 파일 모드 · 좁은 reader 키 비교 · 파싱한 토큰 비교), 오류에는 경로와 가드 번호만 |
| `tools/a140-breakout-signal-shadow/enginewatch.go` | 엔진 `engine.log` 읽기 전용 감시기(§7.5) — 끝으로 이동 · 250 ms 폴 · 패턴 · 회전/잘림/오류 · 정지 점검 → `context.CancelCauseFunc`, 오늘 `abort` 행 검사 |
| `tools/a140-breakout-signal-shadow/universe.go` | `universePolicy` · `universeSource` 인터페이스(`Candles` · `Stocks`) · U2 마감(§4.5) · `pacedSource` 의 `Candles` · `Stocks`(그룹 `MARKET_DATA_CHART` · `STOCK`, `retrying` 재사용) · 공통 필터 · U0/U1/U2 선정 · U2 식 · 일봉 캐시 · 미리 읽기 · 유형 영수증 · 재시작 복원 |
| `tools/a140-breakout-signal-shadow/attribution.go` | 돌파 봉 번호(§5.3) · `numerator_check`(§4.4) |
| `tools/a140-breakout-signal-shadow/report_universes.go` | §9 보고서 |
| 새 시험 파일(`a169_*_test.go`) | 아래 10.3 — 기존 시험 함수 · 가짜는 고치지 않고 새 가짜를 새 파일에 둔다 |

구조체 칸 추가(함수 아님): `probeConfig.Universes` · `effectivePolicy.Universes` · `pacedSource` 의 두 번째 표면 칸 · `ledgerRow` 새 칸 · 상수.

### 10.2 편집하는 기존 함수 = FLM 대상 5

| 함수 | 줄(현재) | 편집 | AST(편집 전) |
|---|---|---|---|
| `runCommand` | main.go:68~165 | `--universes` · `--openapi-dir` · `--profile-preflight` 플래그 · a169 에서 `--openapi-dir` · `--config-dir` 필수(preflight 아닐 때 `--out` 도), `--top` ≠ 30 거부(§1), 기본 모드에서 `--openapi-dir` · `--profile-preflight` 거부 · config.json · evidence · 일반 loader(92~131) 앞에 `checkSeparateProfile`, preflight 면 거기서 끝 · 감시기 시작(실패 시 거부) · 오늘 `abort` 행이면 거부 · `p.run` 뒤 `context.Cause` 면 `abort` 행 · 일반 loader 뒤 키 · 비밀 빈 값 거부 · a169 에서 `apppaths.OpenAPI(openapiDir)` · `cfg.Universes` · `official.New` 값을 두 번째 표면에도 넘김 | `analysis/evidence/…main.go--runCommand.json` |
| `dispatch` | main.go:53~65 | `report-universes` 갈래 하나 | `…main.go--dispatch.json` |
| `effectivePolicyFor` | policy.go:54~69 | 리터럴에 `Universes: cfg.Universes` | `…policy.go--effectivePolicyFor.json` |
| `probe.startSession` | shadow.go:164~261 | B9@200: a169 모드면 `selectUniverse` 대신 U0/U1 선정(+ A140 재구성). B22@249 시작 호가는 편집 안 함(D2 — §7.3) | `…shadow.go--probe.startSession.json` |
| `probe.run` | shadow.go:539~566 | 시작 뒤 cohort 복원 · 분 폴 뒤 일봉 미리 읽기 · 개장 + 30 분 30 초 U2 선정 · 유형 영수증 · `finish` 뒤 `breakout` · `numerator_check` | `…shadow.go--probe.run.json` |

편집하지 않는 것(근거로만): `selectUniverse`(A140 재구성에 그대로 호출) · `step` · `stepSymbol` · `armedAttempt` · `finish` · `label` · `evaluateSetup` ·
`breakoutlane` · `breakoutsignal` · `officialbars` · `internal/official`(2.1 로 끝).

### 10.3 시험(RED 먼저)

1. 판 digest 불변(§1): 아카이브 a140 원장의 판 재현 · a169 판은 다름.
2. 기본 모드 무변화: 가짜 브로커로 한 세션 — `Candles` · `Stocks` GET 0, 새 행 종류 0, 시작 호가 읽기 있음(a140 그대로). a169 모드도 시작 호가를
   U0 ∪ U1 종목마다 읽음(D2). `universePolicy` 를 원장에 쓰고 다시 읽어 digest 가 같음(§1 · P3-5), a140 모드 행에 `universes` · 새 구조체 칸 키 없음.
3. 공통 필터 경계: 999 · 1,000 · 상한 · 상한 + 1 · NaN · 비 KRW · 영문 섞인 심볼 → 사유.
4. U2 식: 정수 비 · 동률 3 단계 · `history_lt_20` · `baseline_zero` · 일봉 경계(§4.2 — review §4 P1-1). 경계 시험은 **실측 표기**로 고정한다:
   오늘 00:00 KST 봉(a128 영수증의 표기)은 제외, 오늘 09:00 봉도 제외, 어제 00:00 KST · 어제 09:00 · 어제 15:30 봉은 포함 — 어제 봉이 셋 중 어느 표기여도
   같은 `S20` · 같은 `last_prior_bar_at` · `next_bar_at`. 월요일 세션에서 금요일 봉이 마지막 직전 봉. `zero_volume_days` 계수.
5. 귀속: 09:29 에 닫힌 돌파(제외) · 09:31 에 닫힌 돌파(귀속) · 돌파 없음(퍼널 귀속) · U0 은 언제나 귀속 · `setup_id` 불일치 표시 ·
   돌파 봉 번호가 호가 둘(서로 다른 값)에서 같고 전체 묶음 평가의 전이와 맞음. 장중 대조(§5.3, P2-6): `N = i + 1` 일치 · `N < i + 1`(15 번째 뒤 봉
   정정 — setup ID 는 같음) 불일치 · `N > i + 1` 이 T_U2 의 같은 쪽이면 지연 일치, 다른 쪽이면 불일치 · 장중 돌파 없음 + 재계산 돌파 → 불일치.
   귀속 대상(P3-1): RANGE_LOCKED → INVALIDATED 로 바로 간 종목(전이에 `BREAKOUT_CLOSED` 있음)이 「돌파 없음」으로 세지지 않음.
   선정 시각 대조 열(P2-5): U0 setup 중 돌파가 T_U2 이하인 것은 보조 열에서만 빠짐.
6. 같은 신호: 같은 가짜 시계 · 봉 묶음 · 호가에서 U0 ∪ U1 종목의 a169 모드 신호 · 라벨 행이 a140 모드와 **바이트 동일**(`snapshot_digest` · `quote_digest`
   포함 — D2 로 시작 호가 경로가 같음).
7. 재시작: U2 선정 뒤 재시작 → 합집합 · cohort 복원, U2 를 다시 고르지 않음 · 선정 전 재시작 → 정시 선정 · 60 초 넘게 늦은 선정 → `late`.
8. 예산 규칙: 미리 읽기가 다음 분 폴 10 초 전에 멈춤(가짜 시계).
9. 유형 갈래 골든 + OpenAPI enum 완전성(§6).
10. 정적 가드(새 시험): `universeSource` 메서드 집합 = {`Candles`, `Stocks`}, 비시험 소스의 인터페이스 이름 집합 = {`marketSource`, `universeSource`} —
    기존 `TestMarketSourceIsTheFiveReadOnlyGets`(static_test.go:46~68)를 고치지 않고 두 번째 표면을 못 박는다. 금지 이름 · import 허용 목록 시험
    (static_test.go:70~136)은 새 파일까지 그대로 훑는다.
11. 보고서: 9 칸 판정 셋 · 두 판 합계 · 표본 경계 표시 · 기타/미분류 개수만 · 칸마다 n · 평균 · 표준편차 · 표준오차와 「통계 판정 아님 · 9 칸 동시 비교」
    표기(P2-4) · 보조 열 U0@T_U2 · U1@T_U2(P2-5) · 늦은 선정 · 귀속 불일치 제외.
12. `selected_at` = 두 랭킹 중 늦게 받은 시각(P3-2) · `numerator_check` 기준 시각 `R` = `rankedAt`(파싱 실패 시 `selected_at`)과 불합격 크기(P3-3).
13. 별도 프로필 가드(§7.5, review §6 P1-3 · §8) — 임시 HOME(`XDG_CONFIG_HOME` · `XDG_CACHE_HOME` · `HOME` 을 `t.TempDir()` 로), **비기본 엔진 경로 fixture**
    `E = <tmp>/engine-profile`(기본 `~/.config/tossctl` 이 아님 — `engine.lock` · `engine.log` · 가짜 `openapi-credentials.json` · `openapi-token.json`),
    HTTP 요청을 세는 가짜 전송(`httptest` 서버 + 호출 계수)으로, 각 가드가 **HTTP 요청 0 · 원장 파일 0 · evidence 파일 0** 으로 거부하고 오류에 그 가드
    번호가 있음을 단언한다(`run` 과 `--profile-preflight` 둘 다):
    - G1: 빈 `--openapi-dir` · 빈 `--config-dir`.
    - G2(a)·P2-3: `E` 와 같음 · `E` 의 하위 · `E/../engine-profile/sub` 처럼 `../` 를 섞은 하위 · `E` 를 가리키는 심볼릭 링크 · 기본 캐시를 상대 경로로 준 것 ·
      기본 설정의 하위. 경계 양성 대조: `E` 가 `/x/engine` 일 때 `/x/engine2` 는 통과.
    - G2(b)·P1-7: **틀린 E** — `engine.lock` 없음 · `engine.log` 없음 · 자격증명 없음인 디렉터리를 `--config-dir` 로 줌 → 거부. 기본 경로에 다른 엔진 파일이
      있어도 `--config-dir` 의 E 만 비교 대상임(기본 `~/.config/tossctl` 에 같은 키를 두고 E 에는 다른 키 → G5 는 E 기준으로 판정).
    - G3: 환경 변수 하나만 · 둘 다.
    - G4: 링크 파일 · `0644`.
    - G5·P2-4: 엔진 파일 없음 · `apiKey` 칸 없음 · 같은 키(**같은 키 경우 HTTP 요청 0 을 따로 단언** — review §8 P1-5). 엔진 파일의 `secretKey` 값은 시험이
      심은 표식 문자열이고, 도구 프로세스 출력 · 원장 · 반환 오류 어디에도 그 표식이 없음.
    - G6·P1-4 넷: (i) 엔진 토큰을 공백 · 필드 순서 · 무관 필드 · `expires_at` 만 바꿔 **재직렬화한 같은 `access_token`** → 거부, (ii) 하드링크 → 거부,
      (iii) 심볼릭 링크 → 거부, (iv) **다른 `access_token`**(양성 대조) → 통과. 그리고 엔진 토큰 파싱 실패 → 거부, 도구 토큰 파싱 실패 → 통과.
    - 양성 대조: 다른 키 · `0600` · 토큰 없음 → `--profile-preflight` 는 `preflight ok` · rc 0 · HTTP 0 · 원장 0, `run` 은 통과하고 `cohort` U0 행에
      `separate_key_checked: true`. 기본 모드에 `--openapi-dir` · `--profile-preflight` → 거부.
    - G7 정적 시험(AST): `profile.go` · `enginewatch.go` 가 키 · 비밀 · 토큰 · 읽은 로그 바이트를 담은 식별자를 `fmt` · `log` · `os.Stderr` · `led.write`
      의 인자로 넘기지 않음, `profile.go` 에 `official.Credentials` · `LoadCredentials` 선택자 · `"secretKey"` 문자열 · `SecretKey` 필드 이름이 없음.
      `engineLogName` 값(cmd/tossctl/engineproc.go:67)과 `enginelock.LockFileName`(enginelock.go:82)이 `profile.go` 의 두 이름과 같음(시험 파일이 AST ·
      import 로 대조 — 이름이 바뀌면 시험이 깨짐).
14. `--top` 계약(§1, review §6 P2-1): a169 모드 `--top 20` · `--top 31` → GET 0 거부, `--top` 생략 · `--top 30` → 통과. 가짜 브로커 한 세션에서
    A140 cohort 30 개 · U0 · U1 각 `per_universe`(30 판 30, 20 판 20) · `effective_policy.top` 30 을 함께 단언. 기본 모드 `--top 20` 은 지금처럼 통과.
15. U2 마감(§4.5, review §6 P2-2) — 가짜 시계 + GET 마다 250 ms 를 쓰는 가짜 브로커:
    (i) 미리 읽기 0 · 보충 200 → `D`(09:30:55) 앞에서 멈춤, U2 cohort `status: late` · `late_reason: daily_topup_deadline` · `topup_done < topup_needed` ·
    `measured: false`, universe 행 다시 안 씀, 09:31:05 폴이 정시이고 U0 ∪ U1 결손 행 0, 그 세션 U2 가 보고서 판 합계 둘 다에서 빠지고 「늦은 선정 — U2 마감」 1.
    (ii) 보충 10 → 정상 선정, `ready_at ≤ D`, U2 새 종목이 09:31:05 폴에 듦, 귀속 T_C = `selected_at`.
    (iii) 경계: 보충 GET 시작이 `now + ReadGap = D` → 시작 안 함. rate limit 재시도 대기가 `D` 를 넘김 → 기다리지 않음.
    (iv) 유형 GET 이 마감 안에 못 들어가면 09:31:05 폴 뒤에 읽힘, 늦은 선정이어도 U0 ∪ U1 유형은 읽힘.
    (v) `D` 뒤 재시작 · cohort 행 없음 → 늦은 선정 행만, 다시 고르지 않음.
16. 엔진 rate limit 감시기(§7.5, review §8 P1-6) — 가짜 시계 · 가짜 브로커(호출 계수) · `t.TempDir()` 의 가짜 `engine.log`:
    - 도구 시작 **전** 로그에 있던 `official: rate limited` 줄 → 무시(끝으로 이동 확인).
    - 시작 뒤 그 줄을 덧붙임 → 다음 감시 폴 이후 **새 GET 0**(가짜 브로커 계수 불변), 원장 `abort` 행 `reason: engine_rate_limited`, 줄 내용 · 매칭
      문자열이 원장 · 출력에 없음, rc 1. 패턴 넷(`rate limited` · `RATE_LIMITED` · `Too Many Requests` · 단어 `429`) 각각 · 줄바꿈 없는 조각도 잡음,
      `14290` 같은 숫자 안의 429 는 안 잡음.
    - 로그를 못 엶(없음 · 권한 0) → 시작 거부, HTTP 0 · 원장 0.
    - 회전(같은 경로에 새 파일로 바꿈) → `engine_log_rotated` 중단, 잘림(크기 감소) → `engine_log_truncated` 중단.
    - 감시기 고루틴 종료 → `watcher_stopped`, 폴 시각이 2 초 넘게 안 바뀜(가짜 파일 시스템이 막힘) → `watcher_stalled`.
    - 같은 날 다시 시작 → 오늘 `abort` 행 때문에 달력 GET 전에 거부(HTTP 0).
    - `report-universes`: `abort` 세션은 두 판 합계에서 빠지고 `aborted` 수로만 나옴.

## 11. 안전 경계 (a140 §7 그대로 + 추가 둘)

- 브로커 호출은 읽기 GET 뿐 — 이 change 가 더하는 것은 `/api/v1/candles`(1d) · `/api/v1/stocks` 둘. 주문 · 정정 · 취소 · 토글 · 활성화 ·
  설정 쓰기 · 엔진 `evidence.db` · 저널 쓰기 · 서명 매니페스트 · Jev 0(정적 시험이 이름을 금지).
- 엔진 프로세스 밖이고 손절 · 비상 청산과 공유하는 코드 0. **Open API 키도 공유하지 않는다** — a169 모드는 별도 앱 키만 쓰고, 엔진 프로필 ·
  엔진 키 · 엔진 토큰 · 환경 변수 덮어쓰기를 시작 전에 거부한다(§7.5 G1~G6). 남는 공유 가능성은 한도가 계좌별일 때의 예산 하나이고 **[미측정]** —
  엔진 `engine.log` 를 읽기만 하는 실시간 감시기가 엔진 rate limit 줄 하나에 세션을 끊고(§7.5, 엔진 편집 0), 리허설 판단(§7.5)과 §7 의 계수 · 리허설 WARNING 규칙(§7.4 재리허설 포함) · D11 재승인으로 다룬다. 시작 호가는 a140 과 같이 읽는다
  (§7.3 — 생략은 GET 을 줄이지 않음).
- 자격증명 값은 문서 · 원장 · 로그 · 기억 어디에도 넣지 않는다. 도구는 경로와 가드 번호만 출력하고, 원장에는 `separate_key_checked` 불리언 하나만 남는다.
- **수익성 결론을 기억에 남기지 않는다(안전 부트스트랩 8 · review §4 P2-4)**: change 종결 때 ai-memory 페이지 · Hindsight retain 에는 「후속 검토 후보」
  칸 이름을 수익성 결론으로 적지 않는다. 남기는 것은 판정 라벨 그대로 + 「미검증 · 통계 판정 아님」 표기와 n 뿐이고, 평균 순 R · 승률 수치와
  「이 종목군이 낫다」 류 문장은 넣지 않는다(보고서 파일에만 있다).
- 측정 실행 자체는 사람이 시작한다(장중 실행 · 공유 키 사용). 이 change 의 대화형 에이전트는 운영 토글 · 설정 · 컨테이너를 만지지 않는다.

## 12. 남은 사람 확인 (측정 전)

1. ~~U2 시각 = 「개장 + 30 분 30 초」 해석(§2)~~ — D1 로 닫힘.
2. 리허설 WARNING 이면 D3 (a) → 재리허설 순서(§7.4). 리허설 2 도 WARNING 이면 다음 걸음, 그리고 측정 판의 N 기준(≤ 90 또는 ≤ 60) D11 재승인.
3. Q12 — 리허설 영수증에 `symbol_format` 제외 후보가 있으면 거름을 넓힐지(D4 로 지금은 유지).
4. Q3 — `numerator_check` 가 어긋나면 대체식(1 분봉 합)과 그 예산 증가(판이 바뀜).
5. **새(review §4 P2-3)**: 5 세션이면 9 칸 거의 전부 「보류」가 예상된다(§9). 측정 전에 (i) 세션 연장 여부 · 세션 수, (ii) 갈래 합친 종목군별 보조
   집계를 낼지.
6. 리허설 영수증의 일봉 경계(§4.2)가 예상(직전 거래일 00:00 KST · 오늘 00:00 KST)과 다르면 측정 전에 멈추고 묻는다.
7. **새(review §6 P1-3 · §8)**: 별도 키 등록과 `--profile-preflight` 통과(§14), `ENGINE_CONFIG_DIR` 확인(§14-1), 리허설을 엔진 보유 0 인 날에 할지(권장 —
   감시기의 대체물 아님), 리허설이 `abort` 로 끝나거나 도구 rate limit 이 0 이 아닐 때의 다음 걸음(§7.5), 세션 중 `engine.log` 가 안 자랄 때(감시 눈멂)의 처리.
8. **새(review §6 P2-2)**: 리허설에서 U2 가 「늦은 선정 — U2 마감」이면 측정 전에 묻는다(09:2x 후보 갱신 · 미리 읽기 늘림 등, issues §3 P3-2).

(처음 초안의 5 「시작 호가 생략」은 D2 변경으로 닫힘 — §7.3.)

## 13. 사람 결정 (2026-10-28, 사용자 「권장대로」 — Manager 경유)

§12 의 남은 확인 중 넷에 답이 왔다. D2 는 review §4 P1-2 가 전제(GET 절감)가 틀렸음을 보여 같은 날 다시 물었고, 사용자가 「권장대로」로
「생략 안 함」을 골랐다. 본문은 review §4 반영 때 이 표에 맞췄다.

| # | 결정 | 해당 절 | 비고 |
|---|---|---|---|
| D1 | U2 선정 시각 = 개장 + 30 분 30 초(정상일 09:30:30 KST) 수락 | §2 · §12-1 | — |
| D2 | **개장 직후 시작 호가 일괄 조회를 생략하지 않는다 — a140 기본 동작 유지(리뷰 §4 P1-2 반영, 2026-10-28)**. spec 의 「a140 과 바이트 동일」 범위는 원래대로(setup · 단계 전이 · 라벨) | §7.3 · spec | 처음 결정(「생략 채택 — GET 120 절감」)은 생략이 GET 을 줄이지 않고 첫 평가 분(stepSymbol B13@399)으로 옮길 뿐임이 드러나 대체됨 |
| D3 | 리허설 WARNING(10 분 창 수집 300 초 초과) 시 §7.4 의 **(a) 종목군당 20** 으로 줄인다 | §1 `per_universe` · §7.4 | 20 판은 새 판 · 재리허설 1 세션(§7.4, review §4 P2-1 · P2-2) |
| D4 | 심볼 거름 `^[0-9]{6}$` 유지. 영문 혼합 코드(예 `0101N0`)가 빠지는지 리허설 영수증(`symbol_format` 제외 사유)으로 확인 | §3.1 · §12-3 | Q12 |
| D5 | **측정 도구 전용 별도 Open API 앱 키 사용**(2026-10-28 사용자 「4는 가능」, 같은 날 발급 — Manager 경유). a169 모드는 그 키만 쓰고 엔진 키 · 프로필 · 토큰을 거부 | §0 · §7.5 · §11 · §14 | review §6 P1-3. 한도가 키별인지는 [미측정] |

남은 §12 항목: 2 · 3 · 4 · 6 · 8 은 리허설 결과에 따름, 5 는 측정 전 사람 확인, 7 은 리허설 전 사람 확인.

## 14. runbook — 별도 키를 측정 프로필에 등록(사람이 함)

에이전트는 이 절을 실행하지 않는다(시크릿 · 운영 접근). 키 값은 이 문서 · 원장 · 기억 · 셸 기록 어디에도 남기지 않는다.
출처: 저장소 루트 `.env`(gitignore 됨)의 `TOSSOS_AUTH_CLIENT_ID2`(앱 키) · `TOSSOS_AUTH_CLIENT_SECRET2`(비밀). 엔진 키는 같은 파일의
`TOSSOS_AUTH_CLIENT_ID` · `TOSSOS_AUTH_CLIENT_SECRET` 이다. 이 변수 이름들을 읽는 TossOS 코드는 없다(Manager grep 실측) — `.env` 는 사람의 보관 장소일 뿐이다.

1. **엔진 프로필을 한 변수로 고정(review §8 P1-7)**: 운영 compose override(`~/.local/share/tossos-deploy/`)의 `TOSSOS_CONFIG_DIR` 값을 사람이 보고,
   그 값을 아래 모든 단계가 같이 쓰는 **한 변수**로 둔다(비밀 아님 — 경로):

   ```bash
   export ENGINE_CONFIG_DIR="<확인한 TOSSOS_CONFIG_DIR 값>"   # 예: ~/.config/tossctl 이라고 가정하지 말고 확인한 값을 그대로
   ls -l "$ENGINE_CONFIG_DIR/engine.lock" "$ENGINE_CONFIG_DIR/engine.log"   # 둘 다 일반 파일이어야 함(G2(b)) — 내용은 열지 않음
   ```

   이후 `--config-dir` · G2/G5/G6 의 비교 대상 · 감시기가 읽는 `engine.log` · 사후 mtime 확인이 전부 이 변수 하나다. `~/.config/tossctl` 을 직접 쓰지 않는다.
2. **허용 IP 등록**: 토스증권 WTS 설정 > Open API > 허용 IP 관리에서 새 앱 키에 이 호스트의 IP 를 등록한다(허용 IP 는 client 마다 — OpenAPI 403
   `IP address not allowed`). 안 하면 아래 4-5 의 `openapi test` 가 `forbidden_ip` 류로 실패한다.
3. **측정 프로필 만들기 — 값이 화면 · 인자 · 기록에 안 나오게**. 프로필은 홈 파일 시스템에 둔다(`/mnt/D` 는 fuseblk 라 권한이 무의미할 수 있어 G4 가
   거부함). 이름은 예시이고 `$ENGINE_CONFIG_DIR` · `~/.config/tossctl` · `~/.cache/tossctl` 의 안이면 안 된다(G2).

   ```bash
   set +o history                      # 이 셸의 이후 명령을 기록에 남기지 않음
   A169_DIR="$HOME/.config/tossctl-a169-measure"
   install -d -m 700 "$A169_DIR"
   # .env 를 source 하지 않는다(파일 전체가 셸로 실행되고 엔진 키까지 환경에 들어옴). 두 이름만 골라 읽어 0600 파일로 씀.
   python3 -I - /mnt/D/Axipient/workspace/TossOS/.env "$A169_DIR/openapi-credentials.json" <<'PY'
   import json, os, sys
   want = ("TOSSOS_AUTH_CLIENT_ID2", "TOSSOS_AUTH_CLIENT_SECRET2")
   vals = {}
   for line in open(sys.argv[1], encoding="utf-8"):
       key, sep, value = line.strip().removeprefix("export ").partition("=")
       if sep and key.strip() in want:
           vals[key.strip()] = value.strip().strip('"').strip("'")
   if sorted(vals) != sorted(want) or not all(vals.values()):
       sys.exit("두 이름 중 없는 것이 있음 (값은 출력하지 않음)")
   fd = os.open(sys.argv[2], os.O_WRONLY | os.O_CREAT | os.O_EXCL, 0o600)  # 이미 있으면 덮지 않고 실패
   with os.fdopen(fd, "w") as f:
       json.dump({"apiKey": vals[want[0]], "secretKey": vals[want[1]], "label": "a169-measure"}, f)
   print("written; values not shown")
   PY
   set -o history
   ```

   - 쓰지 말 것: `tossctl openapi login --key … --secret …`(값이 명령 인자라 셸 기록 · `ps` 에 남음), `.env` 를 `source` · `cat`, 엔진 프로필 복사(`cp -r ~/.config/tossctl …`
     — G5 · G6 가 거부하지만 처음부터 하지 않음).
   - 대안: `tossctl --config-dir "$A169_DIR" init` 의 숨김 입력(cmd/tossctl/init.go:72~79)에 붙여 넣기 — 값을 화면에 띄워 복사해야 하므로 위 방법을 권장.
   - 위 `.env` 파서는 `KEY=value` 한 줄 형식만 읽는다(줄 끝 주석 · 여러 줄 값은 못 읽음) — 형식이 다르면 실패 문구만 내고 멈춘다.
4. **확인 — 네트워크 없이 먼저, 그다음 키 시험(review §8 P1-5)**:
   1. `env | grep -c '^TOSSCTL_OPENAPI_'` → `0`(1 이상이면 G3 가 거부. 그 셸에서 `unset TOSSCTL_OPENAPI_KEY TOSSCTL_OPENAPI_SECRET`).
   2. `stat -c '%a %F' "$A169_DIR/openapi-credentials.json"` → `600 regular file`.
   3. **preflight(HTTP 0 · 원장 0)** — 엔진과 다른 키 · 다른 토큰 · 맞는 엔진 프로필을 네트워크 전에 증명한다:

      ```bash
      go run ./tools/a140-breakout-signal-shadow run --universes a169 --profile-preflight \
        --openapi-dir "$A169_DIR" --config-dir "$ENGINE_CONFIG_DIR"
      ```

      `preflight ok: …` 와 rc 0 이 아니면 **여기서 멈춘다**(가드 번호와 경로만 나옴 — 키 값은 안 나옴). 아래 5 의 `openapi test` 는 preflight 통과 뒤에만.
   4. `stat -c '%y' "$ENGINE_CONFIG_DIR/openapi-token.json"` 을 적어 둔다(엔진 토큰 시각).
   5. `tossctl --config-dir "$A169_DIR" openapi test` → `ok`. 새 키로 토큰을 받아 `$A169_DIR/openapi-token.json` 에만 쓴다(openapi.go:368~395 — 출력에 키가
      없음, config 를 load 하지 않음 — review §8 「확인한 것」). 허용 IP 가 빠졌으면 여기서 실패한다.
   6. 사후 이상 징후(증명 아님): 엔진 토큰 시각을 다시 본다. 바뀌었으면 멈추고 알린다 — 다만 **안 바뀐 것은 아무것도 증명하지 않는다**(엔진이 아직 다시
      부르지 않았을 수 있음). 키가 다르다는 증명은 3 의 preflight 다.
5. **리허설 실행 — 감시기가 엔진 rate limit 에 세션을 끊는다(review §8 P1-6)**: 장 시작 전에 사람이 띄운다.

   ```bash
   stat -c '%s' "$ENGINE_CONFIG_DIR/engine.log"   # 시작 크기 START 를 적어 둠
   go run ./tools/a140-breakout-signal-shadow run --universes a169 \
     --openapi-dir "$A169_DIR" --config-dir "$ENGINE_CONFIG_DIR" \
     --evidence-db <도구 전용 새 파일> \
     --out openspec/changes/a169-breakout-universes-are-compared-in-shadow/analysis/rehearsal-ledger
   ```

   - 도구는 시작 때 `$ENGINE_CONFIG_DIR/engine.log` 를 읽기 전용으로 열고(못 열면 시작 거부), 그 뒤 붙는 줄에서 rate limit 을 보면 즉시 멈추고 원장에
     `abort` 행을 쓴다(§7.5). 이때 사람은 도구를 다시 띄우지 않는다(같은 날 재시작은 도구가 거부) — Manager 에게 알린다.
   - 세션 중 한두 번 `stat -c '%s' "$ENGINE_CONFIG_DIR/engine.log"` 가 START 보다 **커지는지** 본다. 장중 내내 안 자라면 엔진이 그 파일에 안 쓰는 것이라
     감시가 눈멀었다 — 그 세션은 측정에 안 세고 알린다(§7.5 [가정]).
   - 사후 보조 계수(값 없는 개수만): `tail -c +$((START + 1)) "$ENGINE_CONFIG_DIR/engine.log" | grep -ciE 'rate limited|rate_limited|too many requests|\b429\b'`.
     0 이 아니면 그 줄을 눈으로 보고 가른다(줄을 문서에 붙이지 않음). 리허설 원장의 `gap` 행 중 `official: rate limited` · `quote reads are suspended` 사유 수도
     적는다. `abort` 가 있었거나 이 둘 중 하나라도 0 이 아니면 측정하지 않는다(§7.5).
   - 이 단계는 파일 읽기뿐 — 컨테이너 · 토글 · 설정은 만지지 않는다. 엔진 보유 0 인 날 권장(§12-7)은 보조이고 감시기를 대신하지 않는다.
6. 끝나면 `$A169_DIR` 는 그대로 둔다(측정 5 세션에 다시 씀). 키를 폐기할 때는 WTS 에서 키를 지우고 `$A169_DIR` 를 지운다.
