# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json` — **편집 뒤**, :606–684, 분기 18, source_sha256 `c9f398dbc621…`.
- Risk scan: `risk-pattern-report.md`
- 편집: (5.6.2.1) 반환하는 감독자 리터럴에 `entry: opts.EntryGate` 한 칸 — 분기 · 검증 규칙 불변(분기 18 전후 동일). `StrategyEntrySupervisorOptions.EntryGate`(좁은 인터페이스 `StrategyEntryBlocker` — `Block` 하나)를 감독자에 넘긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:608) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine`, `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning` 외 27 |
| B2 | if (:611) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if (:615) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation` 외 15 |
| B4 | if (:618) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if (:621) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if (:625) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` 외 11 |
| B7 | if (:629) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B8 | range (:634) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 30 |
| B9 | if (:635) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if (:638) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if (:641) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if (:644) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B13 | if (:648) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B14 | if (:652) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B15 | if (:658) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if (:661) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range (:674) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 28 |
| B18 | if (:675) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fmt.Errorf` | 612:15 |
| `fmt.Errorf` | 619:15 |
| `len` | 621:5 |
| `errors.New` | 622:15 |
| `clock.System` | 626:9 |
| `clk.Now` | 628:9 |
| `now.IsZero` | 629:5 |
| `errors.New` | 630:15 |
| `make` | 633:13 |
| `validStrategyMarket` | 635:7 |
| `fmt.Errorf` | 636:16 |
| `fmt.Errorf` | 639:16 |
| `fmt.Errorf` | 642:16 |
| `descriptor.AuthorityExpiresAt.IsZero` | 644:70 |
| `now.Before` | 645:5 |
| `validStrategyDigest` | 645:51 |
| `fmt.Errorf` | 646:16 |
| `descriptor.RestartNotBefore.IsZero` | 648:124 |
| `validStrategyWorkerRefusal` | 649:96 |
| `descriptor.RestartNotBefore.IsZero` | 649:151 |
| `fmt.Errorf` | 650:16 |
| `fmt.Errorf` | 656:16 |
| `descriptor.AuthorityExpiresAt.IsZero` | 658:72 |
| `fmt.Errorf` | 659:16 |
| `descriptor.RestartNotBefore.IsZero` | 661:123 |
| `fmt.Errorf` | 662:16 |
| `make` | 666:22 |
| `fmt.Errorf` | 676:16 |
| `make` | 681:62 |
| `make` | 681:91 |

## State mutations and fallbacks

- 위 편집 설명 외 없음.

## Safety conclusion

- Safe edit boundary: 분기 · 검증 불변. 감독자가 진입을 닫을 수만 있고(열 수단 없음) 그 수단은 호출자가 준다.
- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함.

> **5.2.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 편집으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1 리뷰 수리(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 의 몸통 이동으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.
