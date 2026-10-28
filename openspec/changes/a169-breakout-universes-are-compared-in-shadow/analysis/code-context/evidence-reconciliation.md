# 증거 조정 — a169 2.1

- 기준: base `2bbee345f`, 판정 HEAD `566e1cf7c`(+ 작업 트리 편집).
- 변경 표면: `domain.Quote.SecurityType`(새 칸, `security_type,omitempty`) · `adaptStocks` 리터럴 대입 한 줄.
- CodeGraph callers 와 `rg 'adaptStocks|\.Stocks\('` 결과 일치(소비자 3 · 시험 3). 불일치 0.
- 직렬화 · digest 영향: `domain.Quote` 를 통째로 JSON 으로 내는 곳은 `output.WriteQuote` · `ops call` 반환뿐이고 둘 다 omitempty 라
  값 없는 생산자(WTS `GetQuote` · `/prices`)는 바이트 동일. a128 도구는 Quote 를 통째로 직렬화하지 않아 상태 digest 불변.
  `reflect` 로 Quote 칸을 세는 시험 0(`rg`).
- 결론: 편집 허용 — 읽기 전용 경로, High-risk 밖.

## 0.3 — design 단계 (2026-10-28)

- 기준: base `2bbee345f`, 판정 HEAD `0c9bfd272`. 대상 파일은 색인 지문 HEAD `d4dca006` 뒤로 바뀐 커밋이 0 이다.
- CodeGraph 와 `grep -nF` 불일치 둘 — 둘 다 도구 특성이고 결론을 바꾸지 않음:
  1. 수신자 메서드(`probe.selectUniverse` · `probe.startSession`)의 `callers` 가 0 → 자리는 grep 으로 셈(baseline 표).
  2. `effectivePolicyFor` 「5 callers」는 함수 단위, 자리는 생산 1 · 시험 7.
- proposal 줄 번호 재확인(2.1 착지 뒤):
  - `adaptStocks` 는 이제 `stock_reads.go:85~103`(proposal 의 `81~97` 은 base `2bbee345f` 기준 — 2.1 의 주석 · 대입으로 밀림).
    분기 1(range@87) · 반환 1(102) 불변(FLM `../function-logic/internal-official--adaptstocks/ast.json`, `source_sha256` `011792c5…` = 현재 파일).
  - `domain.Quote` 는 `models.go:395~430`(proposal `395~426`), `SecurityType` 칸 429.
  - `stock_reads.go:37` `SecurityType` 디코드 칸 · `read_operations.go:198` `d.Client.Stocks` — 그대로.
  - a140 도구 줄(`shadow.go:23~25` · `:33` · `:52` · `:196` · `:200~201` · `:265` · `:271` · `:275` · `:281` · `:547` · `:550`, `main.go:99` · `:103`)
    — 그대로(`shadow.go` sha256 `62c2fe84…` 가 정의 단계 AST 와 같음).
- 새로 확인된 사실(design 에 반영):
  - `Client.Stocks` 의 새 소비자 `tools/a128-d2-receipts/types.go:89`(a128 영수증 도구) — 이 change 의 유형 영수증과 같은 API · 같은 원문.
  - 일봉 20 일 평균 거래량은 a128 `dailyFeatures`(features.go:216~247)에 선례가 있다(재사용 불가 — `package main`).
  - OpenAPI `/api/v1/candles` `symbol` 설명: 「KRX: 6자리 종목코드 (숫자 또는 영문·숫자 조합; 예: 005930, **0101N0**)」 — a140 거름
    `^[0-9]{6}$`(shadow.go:33)이 영문 섞인 코드를 버린다는 정황(Q12 에 붙임 — 판단은 리허설 영수증).
- 결론: 문서 단계 — 편집 0. 구현 착수 전 차단 불일치 0.
