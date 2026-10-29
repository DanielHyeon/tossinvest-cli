# a169 review

## 1. 정의 경량 리뷰 (2026-10-23)

- 범위: 정의만(proposal · tasks · spec delta 초안 · AST 산출물 · Story). design · 코드 없음.
- Function Logic Map: not-applicable — 정의 단계, 기존 함수 편집 0. 분기를 근거로 쓴 함수 넷은 `analysis/ast-evidence/` AST 를 먼저 만들었다
  (`probe.run` · `probe.selectUniverse` · `probe.startSession` · `adaptStocks`, HEAD `10377cb4b`).
- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only, HEAD `10377cb4b`). 비례 원칙상 경량.
- 판정: P1 0 · P2 5 · P3 9. 코드 줄 인용 · AST B-id · OpenAPI 주장 · a140 수치(원장 sha256 재계산 포함) · a128 D2 §7 인용은 전부 참으로 확인됨.

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P2 | Q1 근거 「분모 ≈ 0」이 틀림 — 선정 시각이 하나면 경과 비율은 상수라 순위 불변, 실제 문제는 개장 단일가 잡음 | **수용** — Q1 근거 교체, 비율은 보고용이라고 명시 |
| 2 | P2 | Q1 컷오프 · Q5 「모두 귀속」 · spec 「바이트 동일」이 서로 모순 | **수용** — 컷오프를 평가기 밖 종목군별 귀속 필터로, Q5 를 「필터 통과 종목군」으로 |
| 3 | P2 | Q4 에 그룹 점유율(90 종목 = 30%) · 09:30 일봉 버스트 충돌 · D11 승계 안 됨이 없음 | **수용** — Q4 에 셋 추가 |
| 4 | P2 | Q6(a) 가 미종결 a140 의 gate 창과 충돌 | **수용** — 「a140 archive 뒤 착수」 전제 |
| 5 | P2 | 종목군 30 미만 규칙 없음 | **수용** — Q10 추가 |
| 6 | P3 | spec 「non-ETF」가 REIT · DR 을 포함하게 읽힘 | **수용** — 「stock/ETF/ETN as mapped, 기타 · 미분류는 개수」 |
| 7 | P3 | spec 「a140 universe on the same session」— 같은 세션이 없음 | **수용** — 시나리오 문장 교체 |
| 8 | P3 | Q7 이 결정된 공통 필터를 다시 물음 | **수용** — 열린 부분(재구성용 기록 · `false` 랭킹 1 GET)만 남김 |
| 9 | P3 | 「간격 초과」 기준이 a140 의 「창 절반 초과 → WARNING」과 다름 | **수용** — a140 기준으로 맞춤 |
| 10 | P3 | tasks 가 Q4 · Q5 권장안을 고정 | **수용** — 「(Q 답에 따라)」 |
| 11 | P3 | `/stocks/all` 대안 미검토 | **수용** — 검토 사실과 권장 사유 기록 |
| 12 | P3 | U2 후보 제한(상위 100 합집합) 한계 미기재 | **수용** — 표에 한계 한 줄 |
| 13 | P3 | a169 가 만드는 것(출력 칸)과 a128 영수증 파일의 경계 · securityType 선행 여부 | **수용** — 경계 명시, Q11 추가 |
| 14 | P3 | ETN 심볼이 `^[0-9]{6}$` 밖이면 ETN 갈래가 조용히 빔 | **수용** — Q12 추가 |

## 2. 2.1 `securityType` 노출 — 구현 경량 리뷰 (2026-10-23)

- 범위: `internal/domain/models.go`(`Quote.SecurityType`) · `internal/official/stock_reads.go`(`adaptStocks` 대입 한 줄) ·
  새 시험 `internal/official/a169_stocks_security_type_test.go`.
- Function Logic Map: `analysis/function-logic/internal-official--adaptstocks/`(편집 뒤 AST, 분기 1 · 반환 1 불변). `check_analysis.py` rc=0.
- 증거 조정: `analysis/code-context/`.
- RED(HEAD `566e1cf7c`, 편집 전): TestA169 넷 중 셋 FAIL(`security_type` 키 없음), `TestA169ThePricesQuoteJSONCarriesNoSecurityType` 은 가드라 PASS.
- GREEN · 영향 범위: `go test ./internal/official/ ./internal/domain/ ./cmd/tossctl/ ./internal/output/ ./internal/ops/ ./tools/a128-jev-shadow-probe/ -count=1`
  무태그 rc=0(PASS 1,628 · FAIL 0 · SKIP 1), `-tags tossos_testseams` rc=0(PASS 1,646 · FAIL 0 · SKIP 1). `go build ./...` · `go vet`(두 구성) rc=0.
- 바이트 동일: `TestA163TheWTSQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime`(WTS) · `TestA163TheOfficialStocksQuoteJSONIsByteIdenticalAndCarriesNoBrokerTime`
  (값 없는 /stocks 행 — 편집 전 골든) · `TestA169ThePricesQuoteJSONCarriesNoSecurityType`(/prices) PASS.
- 보이스: 독립 리뷰어 1(별도 컨텍스트 · read-only). 비례 원칙상 경량(High-risk 밖 — 변이 원장 · 다중 적대 리뷰 not-applicable).

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P3 | 주석이 「소비자는 Go 필드가 아니라 출력 JSON 을 읽음」이라 했으나 읽는 코드는 아직 0 이고 a128 도구는 Go 필드를 인프로세스로 읽는 구조 | **수용** — 「예정 소비자 · 출력 JSON 또는 Go 필드」로 정정(models.go · 시험 머리) |

- 실측(읽기 전용, 새로 빌드한 `./bin/tossctl`, 2026-10-23 06:19 KST): `ops call stocks` `{"symbols":["005930","069500"]}` → rc=0,
  `005930` `security_type: "STOCK"` · `069500`(KODEX 200) `security_type: "ETF"`. ETN 은 후보 심볼 28 개(6 자리 · `Q` 접두 각 14)를
  `/stocks` 2 GET 으로 찔렀으나 0 행, WTS 인기 랭킹 100 에도 ETN 없음 → **ETN 값 실측 못 함**. ETN 심볼 형식은 Q12 대로 첫 세션(또는 리허설)
  영수증에서 확인한다.

## 3. proposal-freeze 경량 리뷰 — 초안 (2026-10-28) · **Manager 판정 대기**

- 범위: tasks 0.3(CodeGraph hard evidence · 증거 조정) · 1.1 `design.md` · 1.2 spec delta(`specs/breakout-retest-strategy-lane/spec.md` 갱신,
  `official-read-compatibility` 는 2.1 그대로) · 이 절. 코드 · 시험 편집 0, 시험 실행 0. 판정 HEAD `0c9bfd272`(worktree `TossOS.worktrees/a169-docs`).
- 선행 조건: Q6 의 「a140 archive 뒤」 충족(archive `a25db35f4`). a140 도구 `shadow.go` sha256 `62c2fe84…` 는 정의 단계 AST 와 같음.
- 보이스: 작성자 셀프리뷰 1(이 문서). **독립 리뷰어 · gstack(autoplan) 보이스는 아직 없다** — WORKFLOW 「리뷰 게이트」의 proposal-freeze 요구
  (gstack 1 회)를 채우려면 Manager 가 보이스를 정한다. 아래 판정은 승인 전 초안이다.
- `openspec validate a169-breakout-universes-are-compared-in-shadow --strict` → 「is valid」, rc=0.

### 3.1 위험 등급과 비례 원칙

- 등급: **Normal(도구 전용 그림자)**. 편집 예정 기존 함수 5(`runCommand` · `dispatch` · `effectivePolicyFor` · `probe.startSession` · `probe.run`)는
  전부 `tools/a140-breakout-signal-shadow/` 안이고, High-risk 경로(주문 · 손절 · 익절 · 사이징 · Guardian · journal · 대사 · retry matrix · 인증 · 체결)
  함수 0. 생산 `internal/` 편집 0(2.1 로 끝).
- 운영 판정 · 게이트 판정을 바꾸는가: **아니오**. 결과는 원장 · 보고서로만 나가고 어떤 발주 경로 · Guardian · 설정에도 닿지 않는다(main spec
  「breakout 그림자 신호 측정은 주문 권위가 없다」). Q8 판정 문구도 사람 판단의 입력이다.
- 그래서 변이 원장 · 다중 적대 리뷰 = not-applicable, FLM 은 gate ⑤ 대로 편집한 기존 함수 5 개에 번들(편집 전 AST 는 이미 `analysis/evidence/`).
- 인접 위험: 엔진과 같은 Open API 키(MARKET_DATA 공유). design §7 이 센다. (초안의 「§7.3 이 개장 직후 MARKET_DATA 묶음을 없앤다」는 틀린 문구 — §4 P1-2 · P3-7, §5 에서 정정.)
- Function Logic Map(이 단계): not-applicable — 문서 단계, 기존 함수 편집 0. 분기를 근거로 쓴 함수 13 개는 design 보다 먼저 AST 를 만들었다
  (`analysis/evidence/claims.md`).

### 3.2 셀프리뷰 발견

| # | 등급 | 발견 | 처리 |
|---|---|---|---|
| 1 | P2 | 90 종목이면 10 분 창 수집 시간이 a140 실측(30 종목 중앙 95.3~196.0 초) 비례로 약 290~590 초 — WARNING(300 초) 가능성이 높다 | design §7.4 에 예상으로 적고, 선택지 (a)(b)(c)는 사람에게(고르지 않음). 리허설 1 세션이 판정 |
| 2 | P2 | Q1 은 「09:30:30 KST」(시계 시각), design 은 「개장 + 30 분 30 초」 | 정상일 동치. 개장이 늦은 날 대비 해석 — §12-1 사람 확인 |
| 3 | P2 | §7.3 시작 호가 생략은 a140 과 다른 동작 — ARMED 전 신호 행의 `snapshot_digest` · `quote_digest` 가 a140 과 다를 수 있음 | spec 「바이트 동일」을 setup ID · 단계 · 전이 · 라벨로 좁힘. 생략 여부는 Manager 판정(§12-5) → **폐기(2026-10-28)**: 생략은 GET 을 줄이지 않고 옮길 뿐(§4 P1-2), D2 를 「생략 안 함」으로 바꾸고 spec 범위를 되돌림(§5) |
| 4 | P2 | U2 는 첫 30 분 돌파를 구조적으로 못 셈(한 세션 한 돌파 — `evaluateFresh` B6@57 → break 66) — 증가율 상위 종목일수록 이미 돌파했을 수 있어 U2 표본이 작아질 수 있다 | 결정된 대가(Q1). 보고서가 「선정 전 돌파 — 귀속 제외」 수를 칸마다 보임 |
| 5 | P3 | 돌파 봉 번호를 세션 끝에 계산 — 도구가 끝까지 못 가면 U2 귀속 미정 | 「귀속 미정」으로 따로 셈. `stepSymbol` · `finish` 무편집과 맞바꿈(design §5.3) |
| 6 | P3 | Q3 대조 `lower ≤ V ≤ upper` 는 랭킹 집계 지연이 있으면 어긋날 수 있다 | `ranked_at` 을 함께 적고, 어긋나면 대체식 · 예산을 사람에게(§12-4) |
| 7 | P3 | Q12 정황: OpenAPI 가 KRX 코드 예로 `0101N0` 을 든다 — `^[0-9]{6}$` 이 영문 섞인 코드를 버린다 | 거름은 그대로(Q12 결정), 제외 사유 `symbol_format` 을 선정 행에 남겨 리허설에서 셈 |
| 8 | P3 | a140 기본 `--out`(main.go:42)은 아카이브로 사라진 경로 — a140 모드로 그냥 돌리면 `openspec/changes/a140-…/` 를 새로 만든다 | a169 모드는 `--out` 필수. a140 기본값은 이 change 범위 밖이라 그대로(되살리지 말 것 — 남은 위험) |
| 9 | P3 | 기존 정적 가드 `TestMarketSourceIsTheFiveReadOnlyGets` 는 `marketSource` 만 본다 — 두 번째 인터페이스가 그 가드를 비켜 갈 수 있다 | 새 시험이 `universeSource` 메서드 집합과 「비시험 소스의 인터페이스 이름 집합」을 못 박음(design §10.3-10). 기존 시험 무편집 |
| 10 | P3 | 원장에 설정 유래 가상 위험 예산 · 명목 상한이 들어간다 | a140 원장과 같은 칸(선례). 계좌 · 잔고 · 보유 0 |

### 3.3 증거 단계 기록 (단계 건너뛰기 금지)

- memory recall: ai-memory(`a169 breakout universe…`) · Hindsight(`tossos-dev-episodes`) — a169 직접 기록 없음, 관련 교훈만(gate ⑤ 는 시험 파일
  편집도 셈 → design 은 기존 시험 · 가짜 무편집).
- CodeGraph: 1.6.0, 기존 색인(`make sdd-sync` 미실행 — Manager 지시, `.sdd/index-state.json` codegraph 지문 HEAD `d4dca006` stale, 대상 파일은 그 뒤
  변경 0). 호출 자리 표 `analysis/code-context/codegraph-baseline.md` §0.3.
- CodeGraphContext: not-applicable(색인 HEAD `9eea8bc2` stale, 미사용). GBrain: MCP 연결 실패 — 그 채널은 안 봤음.
- Go AST: `analysis/evidence/` 13 개(HEAD `0c9bfd272`).
- `codegraph affected`: 구현 착수 때 `--filter '*_test.go'` 로(문서 단계라 편집 파일 없음).

### 3.4 남은 위험 · 사람 결정 (측정 전)

1. 리허설 WARNING 시 §7.4 (a)(b)(c) 와 N ≤ 90 기준 D11 재승인.
2. U2 시각 해석(개장 + 30 분 30 초).
3. Q12 심볼 형식 · Q3 분자 — 리허설 영수증에 따라.
4. ~~§7.3 시작 호가 생략 채택 여부~~ — D2 「생략 안 함」으로 닫힘(§5).
5. proposal-freeze 보이스(gstack/독립 리뷰어) 구성과 이 초안의 승인 — **Manager 판정 대기**.

## 4. 독립 proposal-freeze 리뷰 (2026-10-28, Opus)

- 보이스: 독립 리뷰어 1(design 작성자 아님 · 별도 컨텍스트 · read-only — 소스 편집 0 · 시험 실행 0 · 운영 접근 0). 대상은 worktree
  `TossOS.worktrees/a169-docs` 의 **미커밋 작업 트리**(base `0c9bfd272`): design.md · spec delta · `analysis/evidence/`(AST 13 + claims.md) ·
  `analysis/code-context/` · review §3 · tasks.md, 그리고 인용된 코드 줄 전부를 다시 열어 대조함.
- **결론: FAIL(blocker 2 — P1 둘)**. 둘 다 문서 수정(+ 사람 재확인 한 번)으로 닫히고 구조는 그대로 둬도 됨. P0 0 · P1 2 · P2 6 · P3 7.
- 위험 등급 **Normal 판정은 맞음**(항목 8): 편집 예정 함수 5 개가 전부 `tools/a140-breakout-signal-shadow/`(package main) 안이고, 주문 ·
  손절 · 익절 · 사이징 · Guardian · journal · 대사 · 인증 · 체결 함수 0, 운영 · 게이트 판정 변경 0.

### 4.1 재확인한 것 (발견 아님)

| 항목 | 확인 방법 | 결과 |
|---|---|---|
| AST 증거 13 개 | 각 JSON `source_sha256` 을 현재 파일 바이트로 재계산 | 13/13 일치. design 이 인용한 B-id · 줄(startSession B3@182 · B5@186 · B6@189 · B9@200 · B22@249 · B23@250, evaluateFresh B4@52 · B6@57 · B19@114, Derive B13@245, OpeningRangeATR B2@125, retrying B2@89 · B3@92, stepSymbol B12@396 · B13@399, evaluateSetup B9@124 · B10@126, run B3@547 · B4@550 · B5@553) 전부 JSON `at.line` 과 같음 |
| 앞부분 평가 = 전체 평가(항목 2) | signal.go:120~149 · 153~173 · 238~251, machine.go:52~66 · 205~206 | ATR 은 첫 15 봉(i < 14 + 마지막 봉 index 14), RVOL 은 `bars[index−15:index]`, 윗꼬리 · VolumeExpanded 는 그 봉만, tick 은 첫 15 봉 고가, 돌파는 `bars[15:]` 의 **첫** 만족 봉에서 break. 앞부분의 봉 i 지표는 뒤 봉과 무관 — 주장 참. 뒤 봉 하나가 Derive 를 거절시키면 전체는 실패하고 앞부분은 성공하지만, 그것이 장중에 도구가 실제로 본 것이라 오히려 충실함 |
| 호가가 돌파 판정에 안 들어감 | machine.go:111~116, types.go:127 | `validateQuote` 는 ARMED 뒤에만. bid = ask = last 고정 호가는 `NewQuoteSeal` 을 통과(bid ≤ ask) |
| 판 digest 기대값의 출처(항목 3) | 아카이브 원장의 universe 행 `effective_policy` 를 compact JSON 으로 재직렬화해 sha256 | `3ff46b8b…01ff` 재현(5 행 모두 같은 digest, 원장 sha256 `0b802a92…740c` 일치). 기대값은 커밋된 a140 측정 산출물에서 오고, 코드는 `defaultProbeConfig()`(shadow.go:51~56) + 상수에서 계산 — 작성자 증거 순환 아님 |
| omitempty `Universes` nil 의 바이트(항목 4) | policy.go:25~52 · 71~78, ledger.go:45 | `*universePolicy` 포인터 + `omitempty` 는 nil 일 때 키가 빠지고, 빠진 칸은 위치와 무관 — a140 판 JSON 바이트 동일 참. 단 P3-5 |
| a140 `report` 가 a169 원장을 읽음 | ledger.go:153(schema 등식), report.go:59 switch | 모르는 행 종류는 switch 에서 떨어짐 — 참. 단 P3-6 |
| 미래 정보 누출(항목 6) | design §5.2, run 루프 554 | 선정은 T 이하 정보(랭킹 · 직전 일봉)만 쓰고, 귀속되는 돌파는 T 뒤에 닫힌 봉 — 인과적. T 가 든 봉(09:30 봉)의 앞 30 초 거래량이 분자 V 에 섞이는 것은 실시간에도 알 수 있는 값이라 누출 아님 |
| 도구 전체 속도 | source.go:85~126 | 모든 GET 이 직렬 · 호출 앞 250 ms 대기 → 도구 합계 ≤ 4 GET/s 참. MARKET_DATA 는 429 한 번에 5 분 정지(엔진 몫 ≥ 6/s). 생산 코드에서 `PollClosedBars` 호출자는 `internal/officialbars` · 도구 밖에 없음(rg, internal · cmd) — 엔진은 지금 MARKET_DATA_CHART 를 1 분봉 폴로 쓰지 않음 |

### 4.2 발견

| # | 등급 | 발견 (file:line) | 고칠 것 |
|---|---|---|---|
| P1-1 | **P1** | **직전 일봉 경계가 오늘 봉을 넣는다.** design §4.2 · spec(「bars before today's open only」)은 a128 `dailyFeatures`(features.go:219~222, `Time.Before(open)`) 경계를 빌리며 「여는 시각이든 닫는 시각이든 오늘 봉이 섞이지 않음」이라 쓰지만, 바로 그 함수의 주석(features.go:214~216)이 「자기 날짜의 자정 표기라면 오늘 봉이 직전으로 섞인다」고 적고 있다. 그리고 커밋된 a128 영수증이 KR 일봉이 **자기 날짜 자정(KST)** 으로 찍힘을 보인다: `a128-…/analysis/shadow-ledger/states/00b48ddf69da91b2…json`(003670) `observed_at` 2026-10-19T00:20Z(월 09:20 KST) · `daily.last_prior_bar_at` 2026-10-18T15:00Z = **2026-10-19 00:00 KST** — 전날(일요일)은 거래일이 아니므로 이 봉은 그날(월) 봉이다(10-14 · 10-15 · 10-16 · 10-20~10-23 표본도 모두 「그날 00:00 KST」). 결과: S20 = 오늘 부분 거래량 + 직전 19 일, 게다가 부분 거래량은 그 종목을 **미리 읽은 시각**(09:01~09:30, §7.2)마다 달라 U2 순위에 종목별 잡음이 들어간다. 시험 10.3-4(「오늘 자정 · 오늘 개장 · 어제 마감 셋 모두 같은 답」)는 적힌 경계로는 통과할 수 없다 — design 이 자기 시험과 모순 | 경계를 「KST 세션 날짜의 00:00 보다 앞선 봉」(또는 달력의 직전 거래일 이하)으로 바꾸고 spec 문장도 같이 고칠 것. 시험 4 의 「오늘 자정」 경우를 실측 표기(그날 00:00 KST)로 고정. 선정 행에 마지막 직전 봉 시각과 그 다음 봉(있으면 오늘 봉) 시각을 같이 남겨 리허설 영수증에서 확인. (a128 의 같은 경계는 이 change 범위 밖 — §4.4) |
| P1-2 | **P1** | **시작 호가 생략은 GET 을 줄이지 않고 옮길 뿐이다 — 사람 결정의 전제(「GET 120 절감」)가 코드와 다르다.** startSession B22@249 → 251 에서 읽던 호가가 없으면 stepSymbol B13@399 → 400 이 첫 평가 직전에 **같은 2 GET** 을 읽는다. `state.quote` 를 nil 로 되돌리는 경로가 없으므로(shadow.go:256 · 404 · 449) 세션당 종목마다 정확히 한 번은 어느 쪽에서든 읽힌다 — design §7.1 표도 「종목 수 × 2 + 시도」로 이미 같은 수를 센다. 또 §7.3 의 「대개 09:17 이후로 흩어짐」은 틀림: U0 ∪ U1 의 모든 종목이 같은 분(첫 15 봉이 모이는 step — 대개 09:15:05 또는 09:16:05, B12@396 의 15 봉 조건)에 첫 평가에 들어가므로 60 종목 × 2 = 120 GET 이 **그 한 분에 몰린다**(봉 GET 과 번갈아 서서 MARKET_DATA 순간 속도는 4/s → 약 2.7/s 로 낮아지고, 개장 첫 분을 피하는 효과는 실제로 있음). U2 새 종목 30 은 09:31:05 에 같은 식으로 몰린다. a140 실측 MARKET_DATA 70~135 GET/세션(supplement 13~17 줄) 중 시작 호가 몫은 최대 60(30 종목 × 2) — a169 는 그 몫이 없어지는 게 아니라 시각만 바뀐다 | design §7.3 문장을 「줄임」이 아니라 「개장 첫 분 → 첫 평가 분(대개 09:15~09:16 · 09:31:05)으로 옮기고 봉 GET 과 섞음」으로 고치고, 사람에게 **다시 묻는다**: 절감 0 인 대신 ARMED 전 신호 행의 `snapshot_digest` · `quote_digest` 가 a140 과 달라지는 것(spec 「바이트 동일」 축소)을 그래도 받을지. 받지 않으면 a169 모드도 시작 호가를 읽고 spec 의 바이트 동일 범위를 되돌린다(둘 다 안전상 문제 없음 — 판단은 사람) |
| P2-1 | P2 | `universePolicy` 의 칸이 design 어디에도 없다(§1 · §10.1 은 이름만). 판(lot)을 가르는 값은 종목군당 N(30 → (a) 면 20) · U2 경과 시각(30 분 30 초) · 가격 하한 1,000 · 시작 호가 생략 여부 · 미리 읽기 분당 8 · 귀속 규칙 판본인데, 이것이 `universePolicy` 에 안 들어가면 (a) 로 20 으로 바꿔도 같은 판이 되어 30 과 20 의 세션이 한 판 합계로 섞인다. 반대로 N 을 `cfg.Top`(policy.go:60 `Top`) 으로 넣으면 A140 재구성(`selectUniverse` 271 이 `p.cfg.Top` 사용)까지 20 이 돼 a140 비교가 깨진다. spec 도 「top 30」을 박아 두어 (a) 를 고르면 spec 이 거짓이 된다 | design §1 에 `universePolicy` 칸 목록을 적고 N 은 `cfg.Top` 이 아니라 그 안에 둘 것(A140 재구성은 30 유지). spec 의 「top 30」을 「top N(30, 리허설 WARNING 이면 20 — 2026-10-28 사람 결정)」으로 |
| P2-2 | P2 | 결정 (a) 의 다음 걸음이 없다. a140 비례 추정으로 90 종목 중앙 290~590 초면 60 종목은 약 190~390 초 — 20 으로 줄여도 WARNING 이 남을 수 있다. 그리고 리허설은 90 종목 판에서 돌았으므로 20 판의 예산은 잰 적이 없게 된다 | design §7.4 에 「(a) 적용 뒤 20 판으로 리허설 1 세션을 다시 돌림(측정에 안 셈) · 그래도 WARNING 이면 다시 사람에게」를 한 줄로 |
| P2-3 | P2 | 5 세션으로 판정 칸이 n ≥ 30 에 닿을 가능성이 낮은데 design 이 기대치를 안 적는다. a140 은 30 종목 × 5 세션 = 종목-세션 150 → 라벨 22(proposal Why). U0 는 a140 과 거의 같은 정의이므로 U0 전체가 라벨 ~20 대이고 이를 STOCK · ETF · ETN 셋으로 나누면 칸마다 30 미만 — 9 칸 거의 전부 「보류」가 예상된다(20 으로 줄이면 더 적음) | §9 에 「예상(측정 아님): 5 세션이면 대부분 칸 보류」와 그때의 선택(세션 연장 · 갈래 합친 보조 집계)을 사람 확인 항목으로 올릴 것 |
| P2-4 | P2 | 판정 「n ≥ 30 · 평균 순 R > 0 → 후속 검토 후보」(design §9, Q8)는 불확실성 표시가 없고 9 칸을 한꺼번에 본다 — 기대값 0 근처에서는 칸마다 우연히 > 0 이 나올 확률이 크다(다중 비교). 결정된 임계는 바꾸지 않되, 안전 부트스트랩 8(검증 안 된 수익성 결론을 기억 · 로그에 저장 금지)과의 경계가 design 에 없다 | §9 · §11 에: 칸마다 n · 평균 · 표준편차(또는 표준오차)를 같이 내고 「통계 판정 아님 · 9 칸 동시 비교」를 칸 옆에 붙임. change 종결 때 ai-memory · Hindsight retain 에는 「후속 검토 후보」 칸 이름을 수익성 결론으로 적지 않는다(판정 라벨 + 「미검증」만) |
| P2-5 | P2 | U2 대 U0 비교에 선정 시각 효과가 섞인다는 한계(§9 한계 문구)를 적기만 하고 줄이지 않는다. U2 는 「09:30:30 뒤 돌파」만 세고 U0 · U1 은 09:16 부터 센다 — 오전 첫 30 분 돌파의 성과가 다르면 종목군 차이로 오독됨. GET 0 으로 만들 수 있는 대조가 있다 | 보고서에 「U0 · U1 중 돌파 봉이 T_U2 뒤에 닫힌 setup 만」 보조 열을 추가(같은 귀속 필터를 U0 · U1 에 T_U2 로 적용) |
| P2-6 | P2 | 「귀속 불일치」 검출이 `setup_id` 비교뿐인데 setup ID 는 봉 판(revision)에 무관하다 — `setupID` 는 첫 봉 · 15 번째 봉의 `LaneBarID`(= 세션:여는 시각, signal.go:281~283)와 config digest 만 쓴다(machine.go:205~206). 15 번째 뒤 봉이 장중 이후 정정돼 세션 끝 재계산의 돌파 번호가 장중에 본 것과 달라져도 「불일치」가 안 뜬다. 귀속은 바로 그 돌파 봉 시각으로 갈리므로 놓치면 U2 귀속이 조용히 틀린다 | `breakout` 행에 장중 신호 행 중 transitions 에 `BREAKOUT_CLOSED` 가 처음 든 행의 `bars`(ledger.go:60) · `at` 을 같이 적고, `breakout_index + 1 ≠ bars`(소급 적재가 없던 경우)면 「귀속 불일치」로 셀 것 |
| P3-1 | P3 | §5.3 대상 「가장 멀리 간 단계가 RETEST_WAIT 이상」은 `phaseRank`(report.go:15)에 INVALIDATED · TIMED_OUT 이 없어 소급 적재 · 재시작 뒤 RANGE_LOCKED → INVALIDATED 로 바로 간 종목을 놓친다 → 「돌파 없음 → 귀속」으로 잘못 셈(라벨은 없으니 퍼널만 틀림) | 대상 = 어느 신호 행이든 transitions 에 `BREAKOUT_CLOSED` 가 있는 종목 |
| P3-2 | P3 | `selected_at` 의 정의가 없다(랭킹 받은 시각인가, 보충 일봉 · 유형 GET 까지 끝난 시각인가). 09:30:05 step 은 60 종목이면 a140 실측 비례 19~39 초라 U2 랭킹 GET 이 09:30:30 보다 늦게 나가고, 미리 읽기 후보가 09:00:30 랭킹 기준이라 09:30:30 보충(§7.2)이 수십 GET 일 수 있다 | `selected_at` = U2 두 랭킹 중 늦게 받은 시각으로 정의(귀속 T_C 도 그것). 보충 수가 크면 09:2x 에 랭킹 2 GET(RANKING 그룹)으로 후보를 한 번 갱신하는 선택지를 리허설 뒤 판단 |
| P3-3 | P3 | `numerator_check`(§4.4)의 lower · upper 가 선정 시각 기준이라 랭킹 `rankedAt` 이 수십 초 늦으면 `V < lower` 로 거짓 불합격. 또 장전 시간외 · NXT 등 정규장 1 분봉 밖 거래가 랭킹 누적에 들어 있으면 `V > upper` 가 체계적으로 날 수 있음(미확인) | 경계를 `rankedAt`(파싱되면) 기준 봉으로 잡고, 불합격 시 `V − upper` · `lower − V` 크기를 행에 남길 것 |
| P3-4 | P3 | 분자 · 분모의 「결측 = 0」이 구분되지 않는다: `parseDecimal` 실패는 0 (market_reads.go:394), `Candle.Volume` 은 `float64 omitempty`(models.go:382)라 빠진 값도 0. design 의 `numerator_invalid` · `baseline_invalid` 는 이것을 못 잡는다 | 한계로 적거나, 0 거래량 직전 봉 수를 선정 행에 같이 남김 |
| P3-5 | P3 | 「모든 새 칸 omitempty」(§8)는 구조체 값 칸에는 듣지 않는다(encoding/json 은 struct 값을 omitempty 로 빼지 않음) — 포인터 · 슬라이스 · 맵이어야 함. 또 `universePolicy` 는 원장에서 읽어 다시 digest 해도 같아야 한다(startSession B6@189 → 190 `row.Policy.digest()`) | 구조체 칸은 포인터로, `universePolicy` 에 float · `time.Time` 을 넣지 말 것(정수 ns · 문자열) |
| P3-6 | P3 | U2 에만 든 종목이 선정 전에 돌파했으면 09:31:05 첫 평가에서 바로 ARMED → `armedAttempt` 가 MARKET_DATA 를 읽고 제안 · 라벨까지 나는데, 그 라벨은 어느 종목군에도 귀속되지 않는다. a140 `report` 를 a169 원장에 돌리면 판 합계에 이 라벨이 섞인다 | 「교차 확인용」 a140 report 결과를 측정치로 인용 금지라고 §1 에 한 줄. (그 ARMED 읽기를 막으면 바이트 동일이 깨지므로 막지는 않음 — 비용으로 적기만) |
| P3-7 | P3 | design §12-5 와 review §3.2-3 은 시작 호가 생략을 「MARKET_DATA 묶음을 없앤다」로 적어 P1-2 의 오해를 사람 결정까지 끌고 갔다 | P1-2 와 함께 문구 정정 |

### 4.3 비례 원칙 · 단계 기록

- FLM(이 단계): not-applicable — 문서 리뷰, 기존 함수 편집 0. 분기 주장은 작성자의 AST 13 개로 대조함(위 4.1).
- 변이 원장 · 다중 적대 리뷰: not-applicable(Normal, High-risk 경로 · 게이트 판정 변경 0).
- 실행하지 않은 것: 시험 · `go vet` · `make lint` · `make sdd-check` · CodeGraph 재색인(지시상 시험 실행 금지). `openspec validate --strict` 만 실행.
- 안 본 채널: GBrain(MCP 연결 실패), CodeGraphContext(색인 stale — 작성자 기록과 같음). 엔진의 RANKING · STOCK 그룹 사용량은 안 셈.

### 4.4 범위 밖이지만 알릴 것

- a128 `tools/a128-jev-shadow-probe/features.go:219~222` 의 같은 경계가 커밋된 KR 영수증에서 **오늘 부분 일봉을 직전 봉으로 넣고 있다**
  (`last_prior_bar_at` = 그날 00:00 KST, 위 P1-1 표본). a128 의 20 일 지표가 그만큼 어긋났을 수 있음 — a128 소유자에게 알릴 것. 이 change 는 고치지 않는다.

### 4.5 구현 착수 전에 고칠 것

1. P1-1 — 일봉 경계(design §4.2 · spec · 시험 4).
2. P1-2 — §7.3 · §12-5 문구 정정 후 사람 재확인(시작 호가 생략 유지 여부).
3. P2-1 · P2-2 — `universePolicy` 칸 목록 · spec 의 N · (a) 뒤 재리허설 규칙.
4. P2-3 ~ P2-6 은 design 문장 추가(보고서 · 원장 칸)로 닫힘 — 착수 전 반영 권장, 구현 중 반영도 가능.

## 5. §4 반영 기록 (작성자, 2026-10-28) · **재리뷰 대기**

- 작성자가 §4 의 발견을 문서에 옮긴 기록이다. **이 절은 리뷰가 아니다** — 판정은 다른 보이스의 재리뷰가 낸다. 소스 · 시험 편집 0, 시험 실행 0.
- 사람 결정(2026-10-28, 사용자 「권장대로」): D2 를 「시작 호가 생략 안 함」으로 바꿈, P1 · P2 반영 뒤 재리뷰.
- 새 AST 2(design 이 새로 분기를 근거로 쓴 함수): `analysis/evidence/…shadow.go--probe.armedAttempt.json`(sha `62c2fe84…`) ·
  `…market_reads.go--parseDecimal.json`(sha `a2c2f016…`), `claims.md` 갱신(합계 15). `stepSymbol` B18@417 은 기존 AST 에 있던 분기를 새로 인용.
- `openspec validate a169-breakout-universes-are-compared-in-shadow --strict` → rc 0(아래 확인 시점).

| # | 반영 위치 | 무엇을 바꿨나 |
|---|---|---|
| P1-1 | design §4.2 · §4.3 · §1 `daily_boundary` · §10.3-4 · §12-6, spec 본문 + 새 시나리오 「today's daily candle is not in the baseline」 | 경계를 「KST 세션 날짜 00:00 보다 앞선 봉」으로. 근거: a128 git 추적 KR state 7,406 개(관측일 8 일) 전부 `last_prior_bar_at` = 관측일 00:00 KST, 표본 `00b48ddf69da91b2…`(003670, 월요일). 시험 4 를 실측 표기로 고정, 후보 행에 `last_prior_bar_at` · `next_bar_at`, 리허설에서 다르면 멈춤 |
| P1-2 | design §0 · §7.1 MARKET_DATA 행 · §7.3(전면 재작성) · §10.2 startSession 행 · §10.3-2 · §10.3-6 · §11 · §12 · §13 D2, spec 시나리오 「same signal」, issues §2 D2 | 시작 호가를 a140 대로 읽음(D2 변경). 「GET 120 절감」 · 「09:17 이후로 흩어짐」 삭제, 생략이 시각만 옮김을 근거(B13@399 · shadow.go:256 · 404 · 449)와 함께 적음. spec 「바이트 동일」을 원문(setup · 단계 전이 · 라벨)으로 되돌림. startSession 편집은 B9@200 하나 |
| P2-1 | design §1 `universePolicy` 칸 표 · §3.2 · §3.3 · §4 · §9, spec 본문 | 칸 목록 명시, N = `per_universe`(`cfg.Top` 과 분리 — A140 재구성 shadow.go:271 은 30 유지), 시작 호가 생략 칸 없음(D2 고정) 명기. spec 「top 30」 → 「top N(30, 리허설 WARNING 이면 20)」 + 「N 이 바뀌면 다른 판」 |
| P2-2 | design §7.4 · §12-2 · issues §2 D3 | (a) 적용 → 20 판 재리허설 1 세션 → 아니면 측정 / 그래도 WARNING 이면 측정 없이 재질문, D11 재승인 기준 N 명기 |
| P2-3 | design §9 「예상(측정 아님)」 · §12-5 · issues §3 | a140 150 → 22 근거로 9 칸 거의 전부 보류 예상, 세션 연장 · 보조 집계를 사람 확인 항목으로 |
| P2-4 | design §9(n · 평균 · 표준편차 · 표준오차, 「통계 판정 아님 · 9 칸 동시 비교」) · §11(retain 규칙) · §10.3-11 | 종결 retain 에 「후속 검토 후보」를 수익성 결론으로 남기지 않음(판정 라벨 + 「미검증」 + n 만) |
| P2-5 | design §5.2 · §9 · §10.3-5 | 보조 열 U0@T_U2 · U1@T_U2(GET 0, 판정 없음) |
| P2-6 | design §5.3 · §8 `breakout` 행 · §9 · §10.3-5 | 장중 첫 `BREAKOUT_CLOSED` 행의 `bars`(N)와 재계산 i 대조: 일치 / N < i+1 불일치 / N > i+1 은 T_C 같은 쪽이면 지연 일치 · 다른 쪽이면 불일치 / 장중 없음 + 재계산 있음 불일치. 재계산은 측정 종목 전부(GET 0), 불일치는 판 합계에서 뺌 |
| P3-1 | design §5.3 귀속 대상 · §10.3-5 | 대상 = 어느 행이든 `transitions` 에 `BREAKOUT_CLOSED` 가 있는 종목 — **닫음** |
| P3-2 | design §3.2 `selected_at` 정의 · §10.3-12 | 정의는 닫음. 09:2x 후보 갱신 선택지는 issues §3(리허설 뒤) |
| P3-3 | design §4.4 · §8 · §10.3-12 | 기준 시각 `rankedAt`(없으면 `selected_at`) · 불합격 크기 기록은 닫음. 장 밖 거래 포함 여부는 issues §3(미확인, 리허설) |
| P3-4 | design §4.2 한계 · §4.3 · §9 한계 문구 | `zero_volume_days` · `V == 0` 기록 + 한계. 생산 코드 수정은 범위 밖 — issues §3 |
| P3-5 | design §1(칸 형식) · §8(포인터) · §10.3-2 | 구조체 칸 포인터, `universePolicy` 에 float · `time.Time` 금지, 원장 왕복 digest 시험 — **닫음** |
| P3-6 | design §1 | a140 `report` 의 a169 원장 출력 인용 금지 — **닫음**. ARMED 읽기 비용은 issues §3 |
| P3-7 | design §7.3 · §12 · §13, 이 문서 §3.1 · §3.2-3 · §3.4-4 | 「MARKET_DATA 묶음을 없앤다」 문구를 정정 · 폐기 표시 — **닫음** |

재리뷰가 꼭 볼 곳: design §1 `universePolicy` 표 · §4.2 경계와 근거 · §5.3 장중 대조 규칙(특히 `N > i + 1` 처리) · §7.3 · §7.4 순서 · §9 예상 · spec 본문.

## §6 재리뷰 (GPT-5.6 Terra, 2026-10-28)

- 보이스: §4 작성자·§5 반영자와 다른 읽기 전용 재리뷰. 대상은 worktree `TossOS.worktrees/a169-docs`의 미커밋 작업 트리, base/HEAD `0c9bfd272`.
- 결론: **FAIL(blocker 1 — P1 1)**. §4의 P1 2·P2 6·P3 7 반영은 아래 표처럼 확인했다. 그러나 공유 Open API 키의 엔진 보호 읽기 예산을 실행 전에 보장하는 계약이 없어 implementation start를 막는다. 새 P0 0 · P1 1 · P2 2 · P3 0.
- 증거: 현재 source·AST 15(`analysis/evidence/claims.md`)·커밋 추적 a128 KR state 표본을 대조했다. `00b48ddf…`(003670, 2026-10-19 09:20 KST)의 `last_prior_bar_at`은 2026-10-19 00:00 KST, 별도 표본 386380·222800·138360도 각각 관측일 00:00 KST였다. §4.2의 자정 경계 반증과 수정은 맞다.
- 기억/조사: ai-memory `a169 OR breakout OR universe`는 결과 0, GBrain은 `No brain configured`. `search-first` quick mode로 현재 저장소 코드·기존 OpenAPI 문서만 확인했다. 새 의존성·외부 구현 채택 없음. `impeccable`: N/A — 웹/콘솔 UI 변경 없음; CLI 보고서 계약은 spec·RED 목록으로만 검토했고, 사용자 지시상 렌더/시험은 실행하지 않았다.
- 미실행: 시험, `go vet`, lint, build, `openspec validate`, gstack. 이 절은 문서 재리뷰만 수행했다.

### 6.1 §4 발견 닫힘 확인

| §4 | 판정 | 재확인 |
|---|---|---|
| P1-1 | 닫힘 | `design.md:126-145`와 spec `:12-13, :33-36`이 세션 날짜 00:00 KST 이전만 S20에 넣고, `:398-400`이 실측 표기 경계를 시험으로 고정한다. 위 a128 표본도 이를 지지한다. |
| P1-2 | 닫힘 | D2가 시작 호가 유지로 바뀌었고(`design.md:283-299, :449`), U0∪U1은 `startSession` 기존 경로를 보존한다. `stepSymbol`의 nil 호가 대체 경로(`shadow.go:399-405`)와의 구별도 맞다. |
| P2-1 | 부분 닫힘, 새 P2-1 | `universePolicy`는 N·시각·필터·일봉 경계·미리 읽기를 판에 넣고(`design.md:29-49`) N과 `cfg.Top`을 분리했다. 다만 `--top`은 현재 `main.go:75, :107`에서 그대로 `cfg.Top`에 대입되고 `selectUniverse`는 `shadow.go:271`을 쓴다. a169 모드에서 non-30 `--top`을 거부하거나 A140 재구성 top을 30으로 고정하는 구현 계약·RED가 없다. |
| P2-2 | 닫힘 | 30 판 WARNING 뒤 20 판을 별도 원장으로 재리허설하고, 다시 WARNING이면 측정하지 않는 순서가 `design.md:301-318`에 정확히 있다. |
| P2-3 | 닫힘 | 5 세션·9칸 대부분 보류 예상과 측정 전 사람 결정이 `design.md:354-358, :435-436`에 있다. |
| P2-4 | 닫힘 | 칸별 n·표준편차·표준오차와 다중비교 표기(`design.md:339-351`), retain 시 미검증 라벨+n만 남기는 경계(`:424-426`)가 있다. 보고서는 관측값만 내고 수익성 결론을 내리지 않으므로 안전 부트스트랩 8에 충분하다. |
| P2-5 | 닫힘 | U0/U1에 T_U2를 적용하는 GET 0 보조열(`design.md:196-198, :362`)이 선정 시각 효과를 분리한다. |
| P2-6 | 닫힘 | `N = i+1`·`N < i+1`·장중 행 없음은 불일치, `N > i+1`은 선택 시각의 같은 쪽일 때만 지연 일치(`design.md:209-220`)다. 후자의 실제 돌파 시각은 덮지 않고 지연 분을 남기며, 귀속 결과가 같은 경우만 합계에 남긴다. |
| P3-1 | 닫힘 | `BREAKOUT_CLOSED` 전이 존재로 대상 판정(`design.md:224-227`). |
| P3-2 | 일부 보류 타당 | `selected_at` 정의는 닫혔다(`design.md:106-107`). 09:2x 랭킹 갱신은 top-up 실측 뒤에만 판단 가능하여 `issues.md:42-44`의 보류 사유가 맞다. |
| P3-3 | 일부 보류 타당 | `rankedAt` 경계와 편차 기록은 닫혔다(`design.md:171-175`). 장외/NXT 포함 여부는 OpenAPI 미기재라 리허설 `numerator_check`로 남긴 것이 맞다. |
| P3-4 | 보류 타당 | 0과 결측을 구별할 수 없는 현재 adapter 범위는 `internal/official` 변경이 필요하다. `V == 0`·`zero_volume_days` 기록과 한계 표기(`design.md:147-149`)는 정직한 축소다. |
| P3-5 | 닫힘 | 포인터·정수/문자열 형식·digest 왕복 시험이 `design.md:29-31, :332-334, :394-396`에 있다. |
| P3-6 | 닫힘 | 귀속 없는 U2 라벨을 a140 report 측정치로 인용하지 않고 비용만 계상한다(`design.md:61-63`, `issues.md:46`). |
| P3-7 | 닫힘 | GET 절감·분산 주장을 폐기하고 D2 근거를 바로잡았다(`design.md:287-293`). |

### 6.2 새 발견

| # | 등급 | 발견 (file:line) | 구현 전 고칠 것 |
|---|---|---|---|
| P1-3 | **P1** | **공유 키에서 엔진 보호 읽기 몫 ≥6/s는 보장되지 않는다.** `design.md:266-299`는 도구 자신의 250 ms 직렬화만으로 10/s 중 엔진 몫을 계산한다. 그러나 `pacedSource`의 mutex·정지는 그 도구 인스턴스 안에만 있고(`source.go:43-125`), 429 뒤에도 도구만 멈춘다. `official.Client`의 rate budget은 client별 마지막 응답을 기록할 뿐 자체 로직이 쓰지 않는다(`internal/official/client.go:46-49, :191-206`). 엔진과 별 프로세스/별 client면 도구 4/s와 엔진 `/prices`·`/orderbook`은 합산되어 보호 읽기가 먼저 429를 받을 수 있다. 리허설 `budget` 행은 사후 관측이라 이 위험을 막지 못한다. | a169 실행 전에 적용되는 **공유 키용 권위 있는 admission/reservation 계약**을 design·spec·RED에 정한다. 엔진의 보호 읽기 예약을 확인 못 하면 a169의 MARKET_DATA 호출을 fail-closed로 시작하지 않아야 하며, tool 자체 4/s·사후 WARNING만으로는 닫히지 않는다. 이 change 범위를 넘는 조율 구현이면 Manager가 별 change/선행 조건으로 분리해야 한다. |
| P2-1 | P2 | **a169의 A140 재구성 30 유지가 사용자 `--top` 입력에 대해 강제되지 않는다.** 설계는 `cfg.Top`/A140이 30이라고만 한다(`design.md:36, :383`), 현행 플래그는 임의 1..100을 허용해 `cfg.Top`에 넣고(`main.go:75, :89-90, :107`) `selectUniverse`가 그대로 쓴다(`shadow.go:271`). | `--universes a169 --top != 30`을 명시적으로 거부하거나 A140 재구성값을 고정하고, U0/U1의 N=20과 A140=30을 함께 단언하는 RED를 추가한다. |
| P2-2 | P2 | **U2 일봉 top-up의 최악 200 GET이 09:31 첫 폴 보장을 깨며 `selected_at`과 실제 편입 시각을 분리한다.** U2는 두 top-100 합집합과 모든 S20을 요구한다(`design.md:123-145`), 새 후보 top-up은 09:30:30 뒤에 한다(`:275-281`). 모든 GET은 250 ms 직렬이므로 최대 200건은 응답 시간 전에도 약 50초다. `selected_at`은 랭킹 수신 시각으로 고정(`:106-107`)되지만 `probe.run`은 동기 작업 뒤 다음 `:05`만 계산한다(`shadow.go:553-563`); 09:31:05를 넘기면 첫 U2 폴은 09:32:05가 된다. spec도 즉시 09:30:30 편입을 요구한다(`spec.md:7-16`). | 실제 S20 완료/편입 시각을 분리해 기록하고 late 판정에 쓰거나, 09:31:05 전 완료하지 못한 U2 cohort를 fail-closed로 제외하는 정책을 정한다. 200 top-up fake-clock RED에서 첫 폴·귀속·late 행을 단언한다. |

### 6.3 구현 착수 전 조치

1. P1-3의 공유 키 admission/reservation 선행 조건 또는 분리 change를 결정·명세한다.
2. P2-1의 a169 `--top` 계약과 시험을 명시한다.
3. P2-2의 top-up 완료 시각·첫 폴·late 처리를 명시하고 spec 시나리오를 보강한다.
4. 그 뒤에만 이 절의 FAIL을 재리뷰한다. tasks 1.3은 체크하지 않는다.

## §7 재리뷰 처분 (작성자, 2026-10-29) · **재리뷰 대기**

- 작성자가 §6 의 발견을 문서에 옮긴 기록이다. **이 절은 리뷰가 아니다** — 판정은 다른 보이스의 다음 재리뷰가 낸다. 소스 · 시험 편집 0, 시험 실행 0,
  운영 · 브로커 접근 0, 자격증명 파일 · `.env` 내용 읽기 0(경로와 파일 이름만 다룸).
- 사람 결정(2026-10-28, 사용자 「4는 가능」 — Manager 경유): 측정 도구 전용 별도 Open API 앱 키. 같은 날 발급됨(`.env` 의 `TOSSOS_AUTH_CLIENT_ID2` ·
  `TOSSOS_AUTH_CLIENT_SECRET2` — 값은 안 봄). design §13 D5.
- 새 AST 5(design §7.5 가 분기를 근거로 쓴 함수, design 을 쓰기 전에 생성): `LoadCredentials` · `apppaths.OpenAPI` · `tokenManager.token` · `quoteRead` ·
  `probe.gap` — `analysis/evidence/claims.md`(합계 20). 다섯 다 편집하지 않는 함수.

| §6 | 처분 | 반영 위치 | 무엇을 바꿨나 |
|---|---|---|---|
| P1-3 | **수용 — 별도 키로 구조를 바꿈**(공유 키 위의 admission 계약 대신). 엔진 쪽 예약은 엔진 코드 변경이라 범위 밖 | design §0 · §7.1 MARKET_DATA 행 · §7.3 비용 문단 · §7.4 끝 · **§7.5(새)** · §8 cohort 행 · §10.1 `profile.go` · §10.2 `runCommand` · §10.3-13 · §11 · §12-7 · §13 D5 · **§14 runbook(새)**, spec 본문 + 시나리오 「the engine's key or profile is refused」, issues §4 | a169 모드는 `--openapi-dir`(자격증명 · 토큰만) 필수, `--config-dir` 는 엔진 프로필(config.json 한도 · evidence 가드) 그대로. 시작 전 fail-closed 가드 G1~G6(빈 경로 · 엔진 프로필/기본 토큰 캐시와 같거나 그 안 · `TOSSCTL_OPENAPI_*` 환경 변수 · 링크/0600 아님/빈 값 · 엔진 키와 같음 또는 엔진 자격증명 못 읽음 · 엔진 토큰 복사본) + G7 정적 시험(값 출력 금지). 「엔진 몫 ≥ 6/s」 문장 정정 |
| P1-3 잔여 | **[미측정]으로 남김** — 한도가 키별인지 계좌별인지 | design §7.5 「한도가 키별인가」 · §14-5, issues §4 P1-3a~c | 리허설 창의 엔진 로그 rate limit 줄 수(기준선 포함, 읽기만)와 도구 `gap` 행의 429 사유 수가 둘 다 0 일 때만 측정. 0 이 아니면 측정 0 · 재질문. 키별의 증명은 아님을 명기, 헤더 기반 판별은 비용(FLM +1) 때문에 선택지로만(issues §4 P1-3e) |
| P2-1 | 수용 — 거부 | design §1 「A140 재구성 30 고정」 · §10.2 · §10.3-14, spec 시나리오 「the a140 reconstruction stays at 30」 | a169 모드 `--top` ≠ 30 → GET 0 거부(무시 안 함), 생략 · 30 통과. A140 30 · U0/U1 `per_universe` · `effective_policy.top` 30 을 함께 단언하는 RED |
| P2-2 | 수용 — 마감 + fail-closed | design §2 표 · 늦은 선정 문단 · **§4.5(새)** · §8 cohort 칸 · §9 · §10.3-15 · §12-8, spec 본문 + 시나리오 「a late volume-increase baseline does not delay the poll」, issues §4 P2-2a | 마감 `D` = 다음 분 폴 10 초 전(09:30:55, 미리 읽기 규칙과 같아 새 판 칸 없음). 보충 GET · 재시도 대기가 `D` 를 넘기면 멈춤. 미완이면 그 세션 U2 를 고르지 않음(`status: late` · `late_reason: daily_topup_deadline` · 판 합계 둘 다 제외 · universe 행 안 씀), 09:31:05 폴은 정시. 완료면 `ready_at ≤ D`. 유형 GET 은 마감 밖이면 폴 뒤로. 가짜 시계 RED (i)~(v) |

§6.1 의 나머지 표(P1-1 · P1-2 · P2-2~P2-6 · P3-*)는 「닫힘」 또는 「보류 타당」이라 손대지 않았다. 새로 생긴 위험: G5 가 엔진 자격증명 파일을 **비교용으로**
읽는다 — a140 기본 모드가 이미 그 파일로 브로커를 부르므로 새 접근은 아니고, 값은 메모리에서만 비교하고 버린다(design §7.5).

- `openspec validate a169-breakout-universes-are-compared-in-shadow --strict` → rc 0(2026-10-29, 이 절을 쓴 뒤).

재리뷰가 꼭 볼 곳: design §7.5 가드 표와 「한도가 키별인가」 판단 규칙(측정 0 조건이 충분한가) · §4.5 마감과 재시작 · §14 runbook 의 `.env` 파서와
확인 단계 · spec 본문의 새 문단.

## §8 재리뷰 (GPT-5.6 Terra, 2026-10-29)

- 보이스: §7 작성자와 분리된 읽기 전용 재리뷰. 지정 worktree `TossOS.worktrees/a169-docs`, HEAD `a98475faa`를 대상으로 했다. `.env`·tossctl 설정/토큰 파일 내용, Docker·브로커·운영 명령은 보지 않았다.
- 결론: **FAIL(blocker 4)**. §6의 P2-1(`--top`)과 P2-2(U2 마감)는 계약·RED 목록까지 닫혔다. 그러나 별도 키의 실제 격리와 계좌/IP 한도 미확정 중 엔진 보호 호가 보장은 아직 fail-closed가 아니다.

### 발견

| # | 심각도 | file:line | 근거 | 처분 제안 |
|---|---|---|---|---|
| P1-4 | **P1** | `design.md:397`, `internal/official/token.go:72-74,178-193` | G6는 `os.SameFile` 또는 파일 바이트 sha256만 비교한다. 그러나 `tokenManager.token`은 JSON의 유효한 `access_token`·`expires_at`만 읽어 그대로 재사용하며 키 식별자는 없다. 엔진 토큰을 공백·필드 순서·무관 필드·만료 시각만 다르게 직렬화한 복사본은 바이트가 달라도 같은 bearer token으로 통과한다. 이는 G6가 막겠다는 토큰 재사용 경로다. | G6를 두 캐시의 파싱된 비공개 `access_token` 동등성(값 비출력, 상수시간 비교)으로 판정하고, 링크·파일 동일성 검사는 보조로 둔다. 재직렬화한 같은 token, 하드링크, 심볼릭 링크, 다른 token 양성 대조 RED를 추가한다. spec의 “a copy of an engine token”과 tasks의 구현 전제도 이 의미로 명확히 한다. |
| P1-5 | **P1** | `design.md:627-629`, `cmd/tossctl/openapi.go:377-396`, `internal/official/token.go:72-78,130-173` | runbook은 G5를 실행하지 않는 `tossctl --config-dir "$A169_DIR" openapi test`를 먼저 네트워크 호출한다. 캐시가 없거나 만료면 새 토큰을 발급하며, 같은 키였다면 엔진 토큰을 이미 무효화할 수 있다. 그 뒤 엔진 토큰 mtime이 그대로인 것은 엔진이 아직 다시 호출하지 않은 경우에도 가능하므로 토큰 전쟁을 잡지 못한다. “mtime 변경이면 중단”은 안전한 단방향 신호일 뿐이다. | 네트워크·원장 접근 0의 `a169 profile preflight`(G1~G6)을 만들고 `openapi test`보다 먼저 실행한다. preflight가 실제 엔진 프로필과 다른 키를 증명한 뒤에만 IP/키 검사를 허용한다. 같은 키 RED에서 HTTP 요청 0을 단언하고, runbook의 mtime 검사는 사후 이상 징후로만 낮춘다. |
| P1-6 | **P1** | `design.md:411-425,639-643`, `tasks.md:42-44`, `internal/official/client.go:191-206` | 한도 단위가 계좌/IP일 수 있는데, 도구의 rate budget은 자기 client 응답만 기록한다. 엔진 429와 도구 `gap`은 리허설이 끝난 뒤 로그에서 세므로 해당 리허설 또는 이후 측정 중 보호 호가가 굶는 것을 막지 못한다. “엔진 보유 0인 날”도 권장일 뿐 검증·강제가 아니다. 429 0은 키별 한도의 증명이 아니다. | 한도 단위가 확인되기 전에는 권위 있는 실시간 interlock을 선행 조건으로 한다: 엔진 429 1건을 감지하면 도구 context를 즉시 중단하고 그 시험을 RED로 검증한다. 그런 신호/중단 권한을 만들 수 없으면 리허설·측정 모두 fail-closed로 막고 별 change로 분리한다. 사후 Docker 로그 계수와 보유 0 권장은 대체물이 아니다. |
| P1-7 | **P1** | `design.md:589-590,626,634`, `internal/app/paths/paths.go:31-36,48-54` | §14-1은 실제 `TOSSOS_CONFIG_DIR`을 확인하라지만, mtime 비교와 실행 예시는 모두 `~/.config/tossctl`로 고정한다. 실제 엔진 override가 다르면 `--config-dir`이 정한 다른 E만 G5/G6가 비교하므로 실제 엔진 키·토큰 보호를 증명하지 못한다. | runbook에서 사람 확인값을 단일 비밀 아닌 `ENGINE_CONFIG_DIR`로 고정하고 mtime·`--config-dir`·G2/G5/G6의 모두에 같은 canonical 경로를 쓴다. 비기본 엔진 경로 fixture와 잘못된 E를 준 거부 RED를 추가한다. |
| P2-3 | P2 | `design.md:393-397`, `internal/app/paths/paths.go:31-36,48-54` | G2는 `--openapi-dir`만 `EvalSymlinks`한다고 적고 E·기본 경로의 절대화·링크 해제·경계 비교 알고리즘은 명시하지 않는다. `Resolve`/`OpenAPI`는 받은 상대 경로를 그대로 `Join`한다. 상대 경로·심볼릭 링크·접두사 형제 디렉터리에서 같은/하위 판정이 문자열 구현에 따라 갈릴 수 있다. | 비교 대상 전부를 `Abs` 후 존재하는 것은 `EvalSymlinks`로 canonicalize하고, 포함 여부는 `filepath.Rel`의 경계로 판정한다고 design·RED에 고정한다. `../`, 엔진 경로 심볼릭 링크, `/x/engine2` 접두사, 기본 cache의 상대 경로를 시험한다. |
| P2-4 | P2 | `design.md:396,400-401`, `review.md:252-253` | G5가 엔진 자격증명 파일을 읽는 것은 a169 별도 프로필 프로세스에는 새 시크릿 노출 표면이다. a140 기본 모드가 읽는다는 사실은 별도 프로필 도구가 엔진 `SecretKey`를 메모리에 가져올 필요의 근거가 아니다. | 최소안은 엔진 `APIKey`만 비교하는 좁은 reader로 `SecretKey`를 `Credentials` 객체에 만들거나 전달하지 않는 것이다. 더 강한 격리는 신뢰된 등록 단계가 남긴 프로필 식별자/지문만 런타임에 읽는 설계다. 어느 쪽이든 G7 정적 시험 범위를 엔진 secret의 보관·출력 금지까지 넓힌다. |

### 확인한 것

- **AST로 확인**: 현행 `runCommand`는 `apppaths.OpenAPI`(121) 뒤 `official.LoadCredentials`(125)와 `official.New`(133)를 호출한다. 설계의 `checkSeparateProfile` 삽입 위치는 이 앞이라 일반 자격증명 로드·client 생성·원장 열기·브로커 GET보다 앞에 둘 수 있다. 다만 G4/G5 자체가 파일을 읽어야 하므로 “자격증명 읽기 전”은 문자 그대로 성립하지 않는다. 계약은 **일반 loader/client/ledger/broker보다 전**, 실패 시 GET·원장 0이어야 한다.
- **AST로 확인**: `LoadCredentials`는 두 환경 변수가 모두 있으면 파일보다 우선하고(`credentials.go:29-48`), token cache는 키 식별 없이 유효하면 그대로 쓴다(`token.go:60-79`). `apppaths.OpenAPI`의 명시 경로는 자격증명·토큰 둘 다 그 디렉터리에 둔다(`paths.go:48-60`).
- **손으로 읽음**: config.json이 없는 `$A169_DIR`에서도 `openapi test`의 명령 본문은 config를 load하지 않고 경로·자격증명·`Accounts()`만 쓴다(`openapi.go:372-396`). root의 config 상태 조회 실패도 무시된다(`root.go:79-81,370-376`). 따라서 이 주장은 정적 코드상 참이다. 실제 네트워크 실행은 하지 않았다.
- **손으로 읽음**: a169 `--top != 30` GET 0 거부와 기본 모드 보존은 `design.md:53-57`, spec 시나리오 `:63-66`, RED `:529-530`이 일치한다. U2는 `D=09:30:55` 전 미완이면 U2만 제외하고 09:31:05 폴을 보존하며, 가짜 시계 (i)~(v)가 이를 단언한다(`design.md:189-212,531-537`, spec `:68-71`). 구현 전 계약으로는 충분하다.
- spec delta·tasks·issues를 대조했다. P1-4~P1-7은 아직 spec/tasks/issues에 없으므로 이 결론을 닫기 전 세 문서에 같은 선행 조건·중단 규칙을 반영해야 한다. `openspec validate a169-breakout-universes-are-compared-in-shadow --strict`는 rc 0이었다.
- UI/UX: N/A — 문서·CLI/도구 계약 검토뿐이고 렌더 UI 변경은 없다.

### 확인하지 않은 것

- 소스·시험 편집, 시험·build·vet·lint·gstack, 실제 키/IP/토큰·환경 변수·파일 권한, Docker/엔진 로그, 브로커 호출, 리허설·측정은 실행하지 않았다.
- `checkSeparateProfile`·U2 구현은 아직 존재하지 않아 AST로 런타임 순서를 인증할 수 없다. 위 AST는 기준 HEAD의 기존 경로이고, 설계상 RED 목록의 구현 전 검증만 했다.

## §9 처분 (작성자, 2026-10-29) · **재리뷰 대기**

- §8 의 발견을 Manager 결정(2026-10-29)대로 문서에 옮긴 기록이다. **리뷰가 아니다** — 판정은 다른 보이스의 다음 재리뷰가 낸다. 소스 · 시험 편집 0, 시험
  실행 0, 운영 · 브로커 · Docker 접근 0, 자격증명 · 토큰 · `.env` · `engine.log` 내용 읽기 0(경로 · 파일 이름 · 코드만).
- 새 AST 6(design §7.5 를 쓰기 전에 생성, 전부 편집하지 않는 함수): `engineJournalDir` · `startEngine` · `spawnDetachedEngine` · `classifyStatus` ·
  `pacedSource.pause` · `tokenManager.loadCache` — `analysis/evidence/claims.md` 합계 26.
- 함께 고친 것: a98475faa 의 §7.5 편집이 `## 8. 원장` 제목을 지웠었다(§7.5 마지막 줄에 붙음) — 되살림. §8 리뷰는 이것을 짚지 않았다.

| §8 | 처분 | 반영 위치 | 무엇을 바꿨나 |
|---|---|---|---|
| P1-4 | 수용(Manager 결정) | design §7.5 G6 · §10.3-13 G6 넷, spec 본문 · 시나리오 「the preflight proves the separate key without the network」 | 판정 = 두 캐시를 **파싱한 `access_token`** 의 상수 시간 동등(빈 값 제외). 링크 · `SameFile` 은 보조 거부. 엔진 토큰 읽기 · 파싱 실패 → 거부, 도구 토큰 파싱 실패 → 통과(`loadCache` B3@190 · `token` B2@72 — 채택 안 됨). RED: 재직렬화한 같은 토큰 · 하드링크 · 심볼릭 링크 · 다른 토큰(양성) |
| P1-5 | 수용 | design §7.5 배선 「`--profile-preflight`」 · §10.2 · §10.3-13 · §14-4, spec 본문 · 같은 시나리오 | 플래그 검사 + G1~G6 만, config · evidence · 원장 · 달력 · 네트워크 0 으로 끝나는 모드. runbook 은 `openapi test` 보다 먼저, 실패면 거기서 멈춤. 같은 키 RED 에서 HTTP 요청 0 단언. 엔진 토큰 mtime 은 「사후 이상 징후 — 안 바뀐 것은 증명 아님」으로 낮춤 |
| P1-6 | 수용 — **엔진 편집 0 으로 가능**(멈출 조건 아님) | design §7.5 「엔진 rate limit 실시간 감시기」 · 「한도가 키별인가」 판단 · §8 `abort` 행 · §9 · §10.1 `enginewatch.go` · §10.2 · §10.3-16 · §11 · §12-7 · §14-5, spec 본문 · 시나리오 「an engine rate-limit line stops the measurement」, tasks 3.1, issues §5 P1-6a~d | 엔진은 이미 `ENGINE_CONFIG_DIR/engine.log` 에 stdout · stderr 를 붙인다(`engineJournalDir` B1@186 · `startEngine` 208 · `spawnDetachedEngine` 440 · 446~447). 도구가 시작 때 그 파일을 읽기 전용으로 열고 끝으로 이동(못 열면 시작 거부), 250 ms 폴로 rate limit 패턴 · 회전 · 잘림 · 읽기 오류 · 감시기 종료/정지를 보면 `CancelCause` → `pause` 가 `ctx.Err()`(B1@78)라 새 GET 0 → `abort` 행(줄 내용 안 남김) · 판 합계 제외 · 같은 날 재시작 거부. 429 문구는 코드에서 찾음: `official: rate limited`(`classifyStatus` B4@54 → 55, HTTP 코드 없음). 엔진이 모든 429 를 로그에 남기는지는 [미측정] → 보수 패턴. 보유 0 권장은 보조로 유지 · 대체물 아님 명시 |
| P1-7 | 수용 | design §7.5 배선 · G1 · G2(b) · §10.3-13 · §14-1 · §14-4 · §14-5 | a169 모드 `--config-dir` 필수(기본값 거부). runbook 은 사람이 확인한 `ENGINE_CONFIG_DIR` 하나를 `--config-dir` · G2/G5/G6 · 감시기 · mtime 에 씀. 틀린 E 거부 = `engine.lock` · `engine.log` · 자격증명이 일반 파일로 없으면 거부. RED: 비기본 엔진 경로 fixture · 틀린 E |
| P2-3 | 수용 | design §7.5 「경로 정규화」 · G2(a) · §10.3-13 | `canon` = `Abs` → 있으면 `EvalSymlinks` → `Clean`, 포함 = `filepath.Rel` 경계(문자열 접두사 금지). 시험 넷: `../` · 엔진 경로 심볼릭 링크 · `/x/engine2` 접두사(통과) · 기본 cache 상대 경로 |
| P2-4 | 수용(최소안) | design §7.5 「좁은 reader」 · G4 · G5 · G7 · §10.3-13 G5 · G7, issues §5 P2-4a | 엔진 파일은 `apiKey` 만 디코드 — `official.Credentials` · `SecretKey` 값을 안 만듦. G7 정적 시험이 `profile.go` 의 `Credentials` · `LoadCredentials` · `secretKey` · `SecretKey` 이름과 값 출력 · 원장 기록을 막음. 빈 키/비밀 검사는 일반 loader 뒤 `runCommand` 로 옮김. 버퍼 잔존 한계와 더 강한 격리(지문)는 issues 에 남김 |

새로 생긴 사람 결정: (1) `ENGINE_CONFIG_DIR` 값 확인(§14-1), (2) 리허설이 `abort` 로 끝났을 때 · 세션 중 `engine.log` 가 안 자랄 때의 다음 걸음(§12-7),
(3) 엔진이 모든 429 를 로그로 남기게 하는 엔진 쪽 계약이 필요하다고 보면 별 change 로 할지(issues §5 P1-6a — 이 change 는 엔진을 고치지 않음).

- `openspec validate a169-breakout-universes-are-compared-in-shadow --strict` → rc 0(2026-10-29, 이 절을 쓴 뒤).

재리뷰가 꼭 볼 곳: design §7.5 감시기(회전 · 정지 · 재시작 거부 · 취소에서 GET 0 까지의 사슬) · G6 판정 규칙 · G2(b) 의 E 신원 조건이 충분한가 · §14-4 순서.

## §10 재리뷰 (GPT-5.6 Terra, 2026-10-29)

### 결론: **FAIL (blocker 3)**

§9가 고친 G6의 파싱 동등 비교·preflight 순서·경로 정규화·좁은 reader와 되살린 §8 원장은 설계상 일관된다. 그러나 한도 단위가 [미측정]인 동안 엔진 보호 호가를 지키는 유일한 근거로 올린 감시기는 아직 (i) 취소 뒤 호출을 원자적으로 막지 못하고, (ii) 엔진 429 전부를 보지 못하며, (iii) 시작 때 실제 살아 있는 엔진의 로그를 보고 있음을 확인하지 않는다. 이 셋이 닫히기 전에는 리허설·측정을 시작하면 안 된다.

| ID | 심각도 | file:line | 근거 | 처분 제안 |
|---|---|---|---|---|
| P1-8 | P1 | `tools/a140-breakout-signal-shadow/source.go:77-82,92-96,115-119`; `design.md:443-445`; `specs/breakout-retest-strategy-lane/spec.md:79-80` | **AST** `pacedSource.pause` B1@78은 그 순간의 `ctx.Err()`만 반환한다. 이어 `retrying` B3@92 또는 `quoteRead` B2@115가 통과한 뒤 각각 `p.count`와 `call(ctx)`가 있다. 따라서 감시기가 `CancelCause`를 부르는 시점이 `pause` 반환 뒤·`call` 전이면 새 호출이 하나 이상 시작할 수 있다. `pacedSource`에는 호출을 직렬화하거나 cancel과 호출 시작을 한 임계구역에 묶는 장치도 없다. “취소 뒤 새 GET 0, 최대 1”은 이 AST에서 나오지 않는다. | `pause` 뒤 검사 하나를 덧붙이는 것으로 닫지 말 것. 감시기가 먼저 닫는 admission gate와 실제 broker `call` 시작을 같은 동기화 경계에 두고, 이미 admission한 호출 수의 상한도 명시할 것. 취소를 `pause`와 `call` 사이에 주입하고 동시 호출을 여럿 대기시킨 RED에서, cancel 완료 뒤 broker 호출 계수가 늘지 않음을 증명하라. 그때까지 spec의 “all further broker calls”와 design의 “새 GET 0”을 주장하지 말 것. |
| P1-9 | P1 | `internal/official/errors.go:54-55`; `internal/execgw/retry.go:347-355,370-377,430-439`; `internal/app/engine/exitloop.go:435-441,864-904`; `design.md:431-434,459-464`; `issues.md:69`; `tasks.md:40-41` | **손으로 읽음.** 429는 `official: rate limited` sentinel이 된다. 엔진 exit loop는 `Retrier.Query`로 가격을 읽고, 429에서 대기 후 재시도가 성공하면 351~355에서 로그·alert 없이 성공 반환한다. `Retry-After`가 없는 429도 430~439의 보통 backoff를 거쳐 똑같이 숨겨진다. 즉 “모든 429가 engine.log에 남는다”는 [미측정]이 아니라 현재 반례가 있는 거짓이며 감시기는 그 429를 볼 수 없다. issues §5 P1-6a가 이 눈먼 칸을 적었지만, design·tasks는 이를 가진 채 감시기만으로 리허설·측정을 허용해 서로 모순된다. | 엔진이 `ErrRateLimited`를 받은 **모든** 경로(재시도 전·재시도 성공·Retry-After 유무·fallback 포함)에 비밀 없는 정규 rate-limit 이벤트를 남기고, a169은 그 이벤트를 감시하도록 별 change/계약을 선행하라. 그 계약이 없으면 한도 단위가 확인될 때까지 a169 리허설·측정을 fail-closed로 막아야 한다. 해당 네 경우가 각각 도구 중단으로 이어지는 통합 RED를 추가하라. |
| P1-10 | P1 | `design.md:435-452,730-733`; `internal/enginelock/enginelock.go:84-93,361-385,487-516` | **손으로 읽음.** Manager가 제공한 사실에 따르면 현재 호스트 사용자는 실제 `ENGINE_CONFIG_DIR/engine.log`를 읽을 수 있다. 하지만 설계의 시작 검사는 일반 파일을 열어 끝으로 이동할 뿐이고, “세션 중 한두 번 크기가 커지는지”를 사람이 뒤늦게 보는 절차다. stale·다른 프로세스·더 이상 쓰지 않는 로그를 시작 뒤 첫 GET부터 감시 대상으로 받아들일 수 있다. 같은 디렉터리에는 이미 1분 갱신·5분 stale 판정의 `engine-run.json` marker가 있으나 새 감시기 계약은 이를 확인하지 않는다. | 감시기 시작 전에 `ENGINE_CONFIG_DIR`의 fresh·ready 엔진 marker(필요하면 현재 콘솔의 process-instance 대조까지)를 비밀 없이 확인하고, 실패·stale·unready면 HTTP/원장 0으로 거부하라. marker와 로그 모두가 같은 canonical E 아래임을 시험하고, 시작 뒤 marker가 stale/사라져도 즉시 abort하도록 정하라. 로그 성장의 사후 사람 확인은 보조로만 남긴다. |

### 확인한 것

- **G6 도구 토큰 파싱 실패 → 통과:** **AST** `tokenManager.loadCache` B3@190은 JSON 파싱 오류를 돌리고, `token`은 `err == nil`인 cache만 재사용한 뒤 실패하면 새 exchange로 간다(`internal/official/token.go:72-78,178-193`). 그래서 G5가 별도 API key를 먼저 증명한다는 전제에서는 malformed 도구 cache가 엔진 bearer token을 재사용하는 경로가 아니다. 이 판정은 설계 수준이며 `profile.go` 구현은 아직 없다.
- **§8 원장 복구:** **손으로 읽음.** `a98475faa`에는 §7 다음이 §9였고, 현재 `design.md:471-488`은 `## 8. 원장`과 `abort` 행을 복구했다. 현재 최상위 절 0~14는 연속하고 `abort` schema는 §7.5·§9·spec의 제외 규칙과 맞는다. `git diff --check a98475faa..c1af3177b -- <a169 change>`도 출력 0이었다. 이 비교에서 다른 제목/절 번호 손상은 찾지 못했다.
- **문서 간 정합:** G6·preflight·E 경로·`--top`/U2 RED 항목은 design·spec·tasks에 남아 있다. 다만 P1-9의 눈먼 429를 “별 change가 필요할 수 있음”으로만 남기면서 현 change의 실행을 허용한 부분은 위 blocker다.

### 안 본 것

- 소스·시험 편집, 시험/build/vet/lint, 브로커·Docker·운영 명령, 리허설·측정은 실행하지 않았다.
- `.env`, tossctl 설정·캐시, 실제 `engine.log`·`engine.lock`·marker의 내용과 권한은 읽지 않았다. Manager가 제공한 경로·소유자·mode·mtime 사실은 독립 검증하지 않았다.
- `profile.go`·`enginewatch.go`는 HEAD에 아직 없어 실제 런타임 순서·로그 follower·원장 abort 구현은 인증하지 못했다. 이 절의 AST는 기존 `pause`·429 분류·토큰 cache·엔진 marker 경로만 대상으로 한다.

## §11 보류 (작성자, 2026-10-29)

- 사람 결정(2026-10-29, 사용자 「권장대로」): a169 리허설 · 측정을 a166(계좌 공유 3 req/s 상한 · 모든 429 기록) 착지 뒤로 보류한다. 사유 · 재개 조건은 `issues.md` §6.
- §10 P1-9(재시도에 삼켜지는 429)는 a166 이 소유한다.
- §10 **P1-8 · P1-10 은 지금 처분하지 않는다** — a166 의 admission · 429 이벤트 계약을 본 뒤 함께 처분한다(감시기 · 취소 사슬과 엔진 경로 권한이 그 계약에 따라 바뀔 수 있음).
- design 의 그룹별 10/s · 5/s 기준 문장에는 「3 req/s 계좌 공유로 재계산 필요 — 보류」 표시만 달았다(본문 재계산은 재개 때).
- 소스 · 시험 편집 0, 실행은 `openspec validate --strict` 뿐.
