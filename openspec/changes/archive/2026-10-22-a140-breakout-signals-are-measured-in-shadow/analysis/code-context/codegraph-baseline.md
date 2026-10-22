# a140 CodeGraph baseline

- 도구: codegraph **1.6.0**(`raw/codegraph-version.txt`), 기준 HEAD `050d092f`(base 미고정 — proposal 단계, READY 때 `capture_change_base.py`).
- 원문은 `raw/`. Go 시험 귀속은 `affected … --filter '*_test.go'`(기본 판별식은 Go 시험을 못 찾음 — .claude/CLAUDE.md 하네스 주석).

| 질의 | 결과(raw) | 결론 |
|---|---|---|
| `callers PollClosedBars` | 20 — 전부 `internal/officialbars/producer_test.go` | 닫힌 1분봉 생산자는 있으나 **생산 호출 0** |
| `callers PollQuoteL1` | 10 — 전부 `quote_test.go` | L1 호가 생산자도 **생산 호출 0** |
| `callers LoadProductionAuthorityBatch` | 11 — 전부 시험 | **불일치(R1)**: 생산은 함수 값 참조(`load: strategyproposal.LoadProductionAuthorityBatch`, strategy_proposal_authority.go:254)라 CodeGraph 가 호출로 세지 않음 |
| `callers NewDormantSnapshotReadPort` | 생산 1 — production.go(LoadProductionAuthorityBatch) | 생산의 증거 읽기는 이 읽기 전용 재생 port 하나 |
| `callers ProductionWorkers` | 생산 1 — strategyworker/lane.go:110 | 여덟 레인 인스턴스는 생산 런타임 안에 선다(a112) |
| `callers dispatchStrategyMarketHandoffs` | 생산 1 — strategy_entry_supervisor.go:531(`runProductionStrategyMarketCycle` 마지막 문장) | 한 시장 주기의 주문 전달 자리는 하나 |
| `callers deliverEachStrategyHandoff` | 생산 1 — strategy_market_handoff_delivery.go:32 | 위 함수의 몸통 루프 |
| `callers Evaluate`(이름 겹침) | 63 — breakoutlane.Evaluate 생산 호출 없음(시험 · a112 리뷰 사본만) | breakout 순수 코어는 생산에서 도는 일이 없음 |
| `affected strategy_market_handoff_delivery.go --filter '*_test.go'` | 1025 시험 파일 | 의존 깊이 5 라 넓다 — S3 착수 때 패키지 직접 시험으로 좁힐 것 |
| `affected officialbars/producer.go` · `breakoutlane/machine.go` · `strategyproposal/production.go` | 1118 · 1104 · 1014 | 같음 — a140 은 이 셋을 **편집하지 않는다**(새 파일만) |

HEAD grep(`raw/grep-production-refs.txt`, 비시험):

- `strategyevidence.Open(`(쓰기 열기) 생산 0 · `strategyevidence.OpenReadOnly(` 생산 1(production.go:328).
- `KindKRNetFlow` · `KindUSParticipation` 를 **만드는** 코드 0 — 소비자(continuation · reversal production_proposal)와 모델 상수뿐.
- `ParkerConservativeLane`(strategyengine 첫 레인) 생산 생성 0.
- `ErrBreakoutEvidenceUnavailable` 반환 자리 1(production.go:440) — 분기 열거는 `../ast-evidence/claims.md`.
