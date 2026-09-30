# Branch Test Map: `Context.NewRefreshingPairedStrategyEntrySupervisor`

- Source SHA-256: `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`; AST branch locations are authoritative.
- Revision: **modified (태스크 5.6.2.1, 2026-09-30).** (5.6.2.1) 새 B2 `if c.Entry == nil` — 진입 게이트 없는 Context 에서는 생산 감독자를 만들지 않는다(그 조립에서는 중앙 무결성 고장이 진입이 아니라 프로세스를 닫게 되므로). 감독자 옵션에 `EntryGate: c.Entry`. 편집 전 B2 · B3 → B3 · B4. 편집 전 번들은 `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`.
- 측정: `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`(격리 사본, `./internal/app/engine` 시험 109개를 하나씩).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 389:2 — nil Context · nil 시계 → 거절 | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
| B2 | if at 393:2 — (새) 진입 게이트 없음 → `ErrRuntimeUnavailable` | `TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate` | B2: `red-5.6.2.1.log`(`TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate` FAIL) · 변이 E07; 옵션 전달: `TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate` FAIL · 변이 E06 | yes (block 395.20-397.3, 시험 1개) |
| B3 | range at 397:2 — KR · US 권한 갱신 전용 worker 둘 | `TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate` | 해당 없음(분기 불변) | yes (block 399.78-403.48, 시험 1개) |
| B4 | if at 409:2 — 감독자 생성 실패 → 오류 | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
