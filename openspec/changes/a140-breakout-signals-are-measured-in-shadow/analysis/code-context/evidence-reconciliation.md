# a140 증거 조정 (CodeGraph ↔ HEAD ↔ 기억)

| # | 주장 | CodeGraph 1.6.0 | HEAD(grep · AST) | 결론 |
|---|---|---|---|---|
| R1 | 생산 제안 적재기는 생산에서 쓰인다 | `callers LoadProductionAuthorityBatch` 생산 0 | strategy_proposal_authority.go:254 함수 값 참조 · collectMarket 364 줄 `loader.load(...)` | **HEAD 가 맞다** — 함수 값 참조는 호출 엣지가 아님(a137 R4 와 같은 모양) |
| R2 | 「4/8 레인은 있으나 진입에 살아 있는 신호 생산자가 없다」(Manager 브리프 — 기억 출처) | `ProductionWorkers` 생산 1 → 여덟 인스턴스 | worker.go:133~152 여덟 항목 · 네 가족 순수 평가기 존재 | **정정**: 인스턴스는 8/8, 가족 평가기는 4/4 가 있다. 살아 있는 신호 생산자는 **0/8** 이다 — 아래 R3~R5 |
| R3 | 생산 신호는 서명 매니페스트에서 온다 | — | collectMarket B8@370(env 핀 digest 대조) · LoadProductionAuthorityBatch B3@321 · B4@325(서명) | 신호 = 사람 서명 · 환경 핀 파일. 주기마다 새 신호를 실을 수 없다(새 digest = 재시작) |
| R4 | 증거 DB 를 채우는 생산자가 없다 | `PollClosedBars` · `PollQuoteL1` 생산 호출 0 | `strategyevidence.Open(` 생산 0 · LoadProductionAuthorityBatch B5@329 읽기 전용 | 생산 `evidence.db` 는 쓰는 자가 없다 — 재생은 실패하거나 빈 DB |
| R5 | breakout 은 증거가 있어도 입력을 못 만든다 | `breakoutlane.Evaluate` 생산 호출 0 | buildLaneInput B1@439 → @440 무조건 | 파생 지표(ATR · RVOL · 윗꼬리 · 거래량 확장) 생산자가 0 이라는 것을 코드가 스스로 적는다(production.go:57~61) |
| R6 | KR continuation · reversal 은 공식 데이터로 증거를 만들 수 있다 | — | 두 레인 모두 `KindKRNetFlow` 요구(continuation production_proposal.go:85 · reversal :87) | **불가로 판단** — 종목별 투자자 매매는 공식 표면에 없다(a128 tasks 1.7 영수증: investor-trading = KOSPI/KOSDAQ 시장 단위만). a140 의 첫 가족 선택 근거(design D1) |
| R7 | 생산에 서명 매니페스트 0 건 | — | 이 세션에서 운영 디렉터리를 재지 않음(라이브 접근 금지) | **미검증 · advisory**(기억 2026-09-04 「측정」 기록). a140 은 이 값에 기대지 않는다 — 도구는 매니페스트를 읽지 않는다 |
| R8 | Jev seam 자리는 매수만 지난다 | `callers dispatchStrategyMarketHandoffs` 생산 1 | strategyFirstLegPlaceIntent 268 줄 `Side: "buy"` 리터럴 · dispatch 151 줄 이전은 원장 쓰기 없음 | 확인 — S3 설계 근거(이 change 의 구현 대상 아님) |

불일치 R1 은 HEAD 로 해소했고, R2 는 브리프의 수치를 정정했다. 편집을 막는 미해소 불일치는 없다(이 change 는 기존 파일을 편집하지 않음).
