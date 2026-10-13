# a130 issues

## I-1 (해소 — 사용자 결정 2026-10-13 2차: 묶음별 편입가 재계산, design D1) (blocking, 2026-10-13) — 사용자 결정 1항의 전제가 사실과 다름: 늘어난 수량의 「브로커 평균가」 가 원장에 없음

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

## I-2 (해소 — readopted 거절) (P1, 사람 결정 또는 거절) — 재편입은 exit 상태의 최초 위험을 다시 쓴다

- `resetExitStateForReadoptTx`(apply_hook.go:699, 호출 position_policy.go:145)가 `entry_price` · `initial_risk` 를 재설정하고
  `position_adoptions` 는 그대로다. 재편입 포지션에서 r(재편입 시점) 과 매수 다리(최초 편입 관측가)의 시점이 갈린다.
- 권고: 재편입 이력(`READOPTED` exit 이벤트 또는 `lifecycle_generation > 1`)이 있으면 거절(`readopted`) — 새 사유 코드.

## I-3 (해소 — 사용자 결정 2026-10-13 3차: 브로커 공식 1분봉, design D1) (blocking, 2026-10-13) — 사용자 결정 2차의 가격 출처가 원장에 저장되지 않음

- 결정: 늘어난 묶음 k 의 매입가 = 그 조정 시각 뒤 엔진이 관측한 시장가.
- 실측(HEAD): 엔진은 5초마다 관측하지만 **exit 선이 바뀐 관측만** `exit_events` 에 남긴다. 평평한 선의 관측은 a111(882a0b49)부터
  `refreshObservation` → `RefreshExitObservation`(exit_observation_refresh.go:37 — 「exit event 를 붙이지 않음」)으로 `exit_states` 의 현재값만 덮어쓴다.
  수량 증가는 선 변경으로 치지 않는다. 그래서 원장에서 고를 수 있는 「조정 뒤 첫 관측」 은 「조정 뒤 첫 선 변경 관측」(신고가 틱 또는 손절 무장)이고,
  손절 직전이면 손절 무장 가격이 매입가가 되어 손실이 작게 박제된다(동결 행은 고칠 수 없다). 사본의 `no_intake_observation` 8/12 도 이 때문.
- 선택지(리뷰어 제시 + 팀메이트 정리, 사람이 고름):
  - **(A) 앞으로는 편입 관측을 저장**: 포지션 수량이 마지막 저장 판정 때와 달라진 첫 관측에서 전체 판정(또는 가격만 담은 이벤트)을 강제 — exit 루프
    (High-risk) 편집, FLM 필수. 과거 묶음은 「그 이벤트가 조정 뒤 첫 관측임」 을 증명할 수 없으므로 거절(과거 12건 중 행 0 가능성 큼). 별도 change 가 맞음.
  - **(B) 선 변경 이벤트를 그대로 쓰되 좁힘**: 고른 이벤트가 제안을 무장했으면(action · proposed_intent_id 있음) 거절, 허용 간격을 더 좁히고, 남는 상향
    편향을 spec · D5 에 명시. 과거 행 일부 복구 가능.
  - **(C) 2판으로 되돌림**: 늘어난 묶음은 거절 + 사유(가격 지어내지 않음) — a132(converge 평균가) 이후 종결부터 정직한 매입가.
- 팀메이트 권고: **(C) + (A) 를 별도 change 로**. (B) 는 고른 가격이 판정 경로에 의해 선택되어(신고가 · 손절 무장) 통계적으로 편향된 숫자를 고칠 수 없는
  행으로 남긴다.

## I-4 (후속, 범위 밖) — 앞으로 생기는 편입 뒤 증가 종결의 자동 복구

- 4판에서 그런 종결은 체결 트랜잭션에서 `needs_intake_price` 로 거절되고, 사람이 backfill 도구를 돌려야 행이 생긴다.
- 후보: 주기적 읽기 전용 작업(1분봉 GET)이 미리보기를 만들고 운영자에게 알림 — 행 쓰기는 여전히 사람 승인. 별도 change.

## 결정 이력

- 2026-10-13 1차: I-1 = (b) 거절만 · I-2 = readopted → 2판.
- 2026-10-13 2차(10:4x): (b) 철회 — 묶음별 편입가(조정 뒤 첫 엔진 관측, 실측 허용 간격 안)로 재계산, 관측 없으면 `no_intake_observation` 거절,
  backfill 필요(한정 모집단 · 항목별 재대조) → 3판. (a) 편입 관측가 대입은 기각, (c) converge 평균가 기록은 a132.

- 2026-10-13 3차 리뷰(3판 REJECT): R3-1 — 가격 출처가 원장에 없음 → I-3.
- 2026-10-13 3차 결정: (C) 「사유만」 불가(사용자: 재계산 필수), (A) · (B) 기각 → **(D) 브로커 공식 1분봉** — 그 분의 봉 종가, 없으면 `no_intake_candle`,
  체결 트랜잭션 안 네트워크 금지(`needs_intake_price`), 가격 산정 · 행 쓰기는 backfill → 4판.

## 진행 상태 (1판 기록)

- 구현 착수 전 정지(WORKFLOW 「스펙 결함 — blocking」). production 파일 편집 0.
- 초안(새 파일 · 시험 · 기존 함수 패치 스크립트)은 작업 트리 밖 `refs/heads/wip/a130-1.2` 에 보존(작업 트리 · 인덱스 무접촉으로 만든 커밋).
- 재개 순서: I-1 · I-2 결정 → design 2판(F-1~F-13 반영) · spec 2판 · F-5 번들 → proposal-freeze 재리뷰 → RED → GREEN.
