# a168 Issues

> Manager 결정 2026-11-25(design §10b 「Manager 결정」). 둘 다 이 change 를 막지 않는다 — 기록용 잔여.

## I1 — 만료 모드에서 대상 아닌 CLOSED 행의 `first.status` · `type` 미검사 (F-A1) [후속 change 후보]

- **분류**: non-blocking, 이 change 범위 밖(해독 모호함이 아니라 a164 부터의 판정 의미 공백).
- **내용**: 다른 id 의 `{"status":"EXPIRED","first":{"status":"WATCHING"}}` · `"type":"OCO"` 행이 top-level `EXPIRED` 만으로 통과한다
  (`internal/verifylive/reconcile_expired.go:120-124`, leg 검사는 대상 행만 `:106-109` · `:145-150`). 출처: 리뷰 2단계 1회차 Claude 적대 · Codex.
- **후속 change 후보**: 목록 행에 `Type` 노출 + 모든 CLOSED 행 leg 어휘 검사. **먼저 브로커가 그런 조합을 실제로 내는지 영수증으로 확인**한다.
- 새 change 디렉터리는 만들지 않았다.

## I2 — 병적 본문 비용 (R5 · F-P1) [잔여]

- **분류**: non-blocking 잔여. 깊이 · 크기 상한이나 O(n) 재작성은 넣지 않는다.
- **근거**: 호출자가 `tossctl verify reconcile` 하나뿐이고 엔진 · 주문 경로 영향 0, 정상 100행 53 KB ≈ 5 ms(`analysis/measure/r5_cost-output.txt`).
  병적 입력(깊이 4,000 → 566 ms · 649 MB, 한 객체 키 20,000 → 0.66 s)은 design §9 R5 에 기록.
- **재고 조건**: 대사 읽기가 엔진 경로로 옮겨지거나 비공식 입력을 받게 될 때 — 별도 change.
