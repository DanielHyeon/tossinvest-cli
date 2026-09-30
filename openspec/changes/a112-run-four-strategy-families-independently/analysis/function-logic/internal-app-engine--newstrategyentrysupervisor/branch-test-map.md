# Branch Test Map: `NewStrategyEntrySupervisor`

- Source SHA-256: `da4fa6d1b57217a08a05d0ae57a4f4be1e3c173c35947e06527b02014ae75b23`; AST branch locations are authoritative.
- Revision: **modified (태스크 5.6.2.1, 2026-09-30).** (5.6.2.1) 반환하는 감독자 리터럴에 `entry: opts.EntryGate` 한 칸 — 분기 · 검증 규칙 불변(분기 18 전후 동일). `StrategyEntrySupervisorOptions.EntryGate`(좁은 인터페이스 `StrategyEntryBlocker` — `Block` 하나)를 감독자에 넘긴다. 편집 전 번들은 `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`.
- 측정: `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`(격리 사본, `./internal/app/engine` 시험 109개를 하나씩).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 618:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine`, `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning` 외 27 | 해당 없음(분기 불변) | yes (block 618.16-620.3, 시험 29개) |
| B2 | if at 621:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 621.52-623.3, 시험 1개) |
| B3 | if at 625:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation` 외 15 | 해당 없음(분기 불변) | yes (block 625.21-627.3, 시험 17개) |
| B4 | if at 628:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 628.62-630.3, 시험 1개) |
| B5 | if at 631:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 631.28-633.3, 시험 1개) |
| B6 | if at 635:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` 외 11 | 해당 없음(분기 불변) | yes (block 635.16-637.3, 시험 13개) |
| B7 | if at 639:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
| B8 | range at 644:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 30 | 해당 없음(분기 불변) | yes (block 644.42-645.46, 시험 32개) |
| B9 | if at 645:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 645.46-647.4, 시험 1개) |
| B10 | if at 648:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 648.60-650.4, 시험 1개) |
| B11 | if at 651:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 651.54-653.4, 시험 1개) |
| B12 | if at 654:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
| B13 | if at 658:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
| B14 | if at 662:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
| B15 | if at 668:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 668.180-670.4, 시험 1개) |
| B16 | if at 671:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestStrategySupervisorRejectsInvalidAssemblies` | 해당 없음(분기 불변) | yes (block 671.161-673.4, 시험 1개) |
| B17 | range at 684:2 — (분기 불변 — 편집 전 번들의 서술 그대로) | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 28 | 해당 없음(분기 불변) | yes (block 684.78-685.29, 시험 30개) |
| B18 | if at 685:3 — (분기 불변 — 편집 전 번들의 서술 그대로) | (측정 표본 0) | 해당 없음(분기 불변) | 측정 표본의 시험 0개(블록 좌표 없거나 미실행) |
