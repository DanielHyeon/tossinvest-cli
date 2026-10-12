# a130 issues

## I-1 (blocking, 2026-10-13) — 사용자 결정 1항의 전제가 사실과 다름: 늘어난 수량의 「브로커 평균가」 가 원장에 없음

- 결정 1항: 「편입 뒤 늘어난 수량의 매입가 = 그 대사 조정의 브로커 평균가(`new_avg_price` 등)」.
- 실측(HEAD 233d5d9c): 편입 포지션의 수량을 늘리는 경로는 대사 수량 수렴 `internal/reconcile/converge.go:209-226` 이고,
  `NewAvgPrice: ""` 를 **항상** 보낸다(주석: 「계좌의 cost basis 는 수량 비교의 일부가 아님」). `ApplyPositionAdjustment` 는
  `firstNonEmpty(req.NewAvgPrice, target.AvgPrice)` 를 저장하므로 `new_avg_price` = **프로젝션의 기존 평균가** = fold-in 때의 브로커
  `averagePurchasePrice`(편입 전 원 보유분의 cost_basis, `reconcile/external.go:233`). 사건 11건이 전부 `prev_avg_price = new_avg_price` 인 이유다.
- 그대로 구현하면 늘어난 수량을 원 보유분의 cost_basis 로 매입한 것으로 계산 — adopt-external-positions A7 round-3 의 「cost_basis 는 어느
  다리에도 나오지 않는다(SHALL NOT)」 위반이고, 동결 행은 UPDATE 금지 트리거 때문에 나중에 고칠 수 없다.
- 선택지(리뷰어 제시, 사람이 고름):
  - **(a)** 늘어난 수량도 편입 관측가 p_a 로 매입 처리 — 주당 위험 r 과 같은 시점(t0). 분자 · 분모 모두 편입 시점 기준이 되어 A7 의 「같은
    t0」 원칙과 일치. 단점: 늘어난 수량의 실제 매입가와 다를 수 있음(그 값은 원장에 없음).
  - **(b)** 편입 뒤 수량 증가가 있으면 계속 거절(행 없음) + 사유 기록만 — 조작 위험 0, 사건 11건은 행 없이 남음(사유는 보임).
  - **(c)** converge 가 앞으로는 브로커 평균가를 기록하게 고침(새 동작, 과거 11건은 복구 불가) — (a)/(b) 와 병행 가능.
- 구현 팀메이트 권고: **(a)**(A7 의 같은 t0 원칙 · 11건 복구 가능) + 잔여로 명명. (c) 는 별도 change.

## I-2 (P1, 사람 결정 또는 거절) — 재편입은 exit 상태의 최초 위험을 다시 쓴다

- `resetExitStateForReadoptTx`(apply_hook.go:699, 호출 position_policy.go:145)가 `entry_price` · `initial_risk` 를 재설정하고
  `position_adoptions` 는 그대로다. 재편입 포지션에서 r(재편입 시점) 과 매수 다리(최초 편입 관측가)의 시점이 갈린다.
- 권고: 재편입 이력(`READOPTED` exit 이벤트 또는 `lifecycle_generation > 1`)이 있으면 거절(`readopted`) — 새 사유 코드.

## 진행 상태

- 구현 착수 전 정지(WORKFLOW 「스펙 결함 — blocking」). production 파일 편집 0.
- 초안(새 파일 · 시험 · 기존 함수 패치 스크립트)은 작업 트리 밖 `refs/heads/wip/a130-1.2` 에 보존(작업 트리 · 인덱스 무접촉으로 만든 커밋).
- 재개 순서: I-1 · I-2 결정 → design 2판(F-1~F-13 반영) · spec 2판 · F-5 번들 → proposal-freeze 재리뷰 → RED → GREEN.
