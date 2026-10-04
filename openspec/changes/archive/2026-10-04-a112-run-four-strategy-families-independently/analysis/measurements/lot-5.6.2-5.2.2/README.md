# 로트 5.6.2 · 5.2.2 — 착수 측정 (2026-09-30, HEAD `2f698db6`)

## check_analysis 기준선(델타 판정의 입력 — Manager 판정 2026-09-30)

`check-analysis-baseline-2f698db6.log` — `python3 tools/logic-map/check_analysis.py --change a112-…` 를 편집 **전** HEAD `2f698db6` 에서 돌린 전문(rc 1, `[logic-map]` 줄 129).
- missing evidence 115: 거의 전부 형제 착지(a092 · a066 · a108 · soak 등)가 base `aeeb209e` 뒤에 바꾼 함수 — 창(436 커밋)이 넓어서 잡힘.
- stale 6 × 2: a112 자기 번들인데 형제 편집으로 소스가 바뀐 것.

이 로트의 판정은 **델타**다: 로트 뒤 같은 명령의 발견 집합에서 이 기준선을 뺀 나머지가 이 로트가 편집한 함수의 번들 문제뿐이어야 한다(a065 선례). `check_analysis` 는 완료 게이트에서만 참이라(HANDOFF (2)) 진행 중 로트의 rc 1 자체는 판정이 아니다.

## base 재고정 영수증 — 재고정하지 않음

`repin-receipt-aeeb209e-2f698db6.tsv`(`analysis/harness/a112_repin_receipt.py aeeb209e HEAD`): 자기 비병합 Go 커밋 26, 그 커밋들이 바꾼 함수 200 중 옛 창에서 요구되는 것 28 = **FRESH 20 · STALE 8**(나머지 172 는 base 뒤 새 함수거나 착지에서 불변).
STALE 8 은 전부 자기 번들이 형제 편집으로 낡은 것이다(`strategy_dispatch_cycle_test.go` 4 · `strategy_proposal_ambiguity_test.go` 1 · `console/strategy_runtime_multimarket.go` 1 · `journal/schema_test.go` 1 · `strategyprojectionrpc/transport_unix_test.go` 1).

Manager 판정(2026-09-30)의 조건 「자기 몫 전부 fresh 대응」이 성립하지 않으므로 **재고정하지 않는다** — 완료 게이트에서 결정 (5) 의 `revision: base` 번들 15 재추출과 함께 STALE 8 을 재추출한다.
