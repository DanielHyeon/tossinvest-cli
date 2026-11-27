# a209 review

## Pre-Edit Gate (2026-11-28)

- change id / task id: a209-console-exit-evidence-counts-arming-rows-only / 2.2
- 대상 심볼: `journal.ReadOnly.BrokerOrderExitLinks`(internal/journal/account_views.go)
- CodeGraph 1.6.0: callers — 생산 `Console.ledgerOrderEvidence`(internal/console/orders.go:716) 하나, 나머지는 journal 시험
  (`account_exit_orders_test.go` 16 · `a207_compact_evidence_test.go` 1). `affected --filter '*_test.go'` 는 파일 단위라 1,267(journal 패키지 전체)로
  퍼져 판단에 쓰지 않음. a166 census `shapes.json` 에 account_views.go 없음.
- CodeGraphContext: not-applicable — 대상이 한 함수의 SQL 문자열 두 줄이고 호출자가 하나라 문맥 확장 후보가 없음. 증거 조정은 design §1 · §4.
- 기존 동작 근거: 편집 전 AST `analysis/ast-evidence/journal--ReadOnly.BrokerOrderExitLinks.base.ast.json`(28 갈래), RED 출력 `analysis/tdd/red-journal.txt`.
- FLM / BTM: `analysis/function-logic/internal-journal--readonly.brokerorderexitlinks/`
- upstream 상속 시험 영향: no(TossOS 고유 리더)
- 실패 시험 선행: yes(RED 3 실패 — FILLED · REFUSED · CANCELLED)
- 설정 · DB · journal 변경: 없음(읽기 SQL 만, 스키마 0)
- §0 안전 불변식: 통과 — 표시 전용 리더, 주문 · 손절 · 원장 쓰기 · 대사 경로 무변경

## 비례 원칙

- 변이 원장: not-applicable — High-risk 경로 · 게이트 판정을 바꾸지 않는 표시 리더(docs/WORKFLOW.md 「비례 원칙」).
- 다중 적대 리뷰: not-applicable — 같은 사유. 경량 gstack `review` 1회(상한 2회).
