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
