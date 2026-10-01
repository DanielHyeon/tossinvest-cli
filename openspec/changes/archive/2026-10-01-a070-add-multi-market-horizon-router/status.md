# Status — a070-add-multi-market-horizon-router

- 처분(2026-10-01 감사 `analysis/disposition-audit.md`, Manager 승인 ② 부분 대체 — 사용자 거부권 항목): `--skip-specs` 아카이브
- 살아 있음(a072 · a112 생산 경로의 기반): `types.go`, `OwnerKey`/`NewOwnerKey`, `Owner`/`OwnerSnapshot` 봉인,
  `MarketRecord`/`EvaluateMarketLifecycle`(서명 매니페스트에서 매 로드 생성), `RouteRequest`/`Candidate`/`RouteDecision`, descriptor registry
- 대체됨: `Route()` 의 원시 점수 단일 승자 선택 → a112 `RouteSet` + 보정 점수 조정자(4.3.1, 생산 폐포 금지 가드 4.3.2)
- **생산 호출자 0(섬)**: durable `SchedulerState`/CAS/rollback/`MarketRecordStore`, `MigrateLegacy`, `QuotaAuthority`
  — 시장 · 가족 활성화는 서명 매니페스트(a072 3.17 · 3.23, a112 8.7.1), 레인 cadence/latch 는 `strategyworker.Lane`(a112 5.3.x).
  예전 이 문서의 "Scheduler: independent KR/US records, CAS/locks …" · "Quota: one physical endpoint …" 줄은 패키지 안 시험만 GREEN 이었고
  생산에 배선된 적이 없다(정정).
- 남은 태스크: 3.1 은 a072 3.23 · a112 가 수행, 3.4 의 quota · migration 대상은 배선된 적 없음, 4.2 · 4.3 은 독립 의미 없음
- spec delta: 적용하지 않음 — 중복(정본 market-aware-scheduler · strategy-runtime · strategy-engine), 충돌(a112 「평가 전 cross-family winner 금지」),
  미구현(durable CAS · legacy migration · quota subscope → a112 7.1)
- 후속: quota 단일화는 a112 7.1 의 설계 입력(`scheduler.BudgetCoordinator` 위에 subscope), 섬 코드 삭제는 7.1 뒤 정리 change 하나(ROADMAP 후속 후보)

No market was selected, no lane was enabled, and no live order/toggle/activation was created.
