# Function Logic Map: `NewStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json` — **편집 뒤**, :616–694, 분기 18, source_sha256 `da4fa6d1b572…`.
- Risk scan: `risk-pattern-report.md`
- 편집: (5.6.2.1) 반환하는 감독자 리터럴에 `entry: opts.EntryGate` 한 칸 — 분기 · 검증 규칙 불변(분기 18 전후 동일). `StrategyEntrySupervisorOptions.EntryGate`(좁은 인터페이스 `StrategyEntryBlocker` — `Block` 하나)를 감독자에 넘긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:618) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine`, `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning` 외 27 |
| B2 | if (:621) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B3 | if (:625) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestExpiredAuthorityLatchesBeforeEvaluation` 외 15 |
| B4 | if (:628) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B5 | if (:631) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B6 | if (:635) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety`, `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` 외 11 |
| B7 | if (:639) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B8 | range (:644) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 30 |
| B9 | if (:645) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B10 | if (:648) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B11 | if (:651) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B12 | if (:654) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B13 | if (:658) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B14 | if (:662) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |
| B15 | if (:668) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B16 | if (:671) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestStrategySupervisorRejectsInvalidAssemblies` |
| B17 | range (:684) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 28 |
| B18 | if (:685) | (분기 불변 — 편집 전 번들의 서술 그대로) | — | (측정 표본 0) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `fmt.Errorf` | 622:15 |
| `fmt.Errorf` | 629:15 |
| `len` | 631:5 |
| `errors.New` | 632:15 |
| `clock.System` | 636:9 |
| `clk.Now` | 638:9 |
| `now.IsZero` | 639:5 |
| `errors.New` | 640:15 |
| `make` | 643:13 |
| `validStrategyMarket` | 645:7 |
| `fmt.Errorf` | 646:16 |
| `fmt.Errorf` | 649:16 |
| `fmt.Errorf` | 652:16 |
| `descriptor.AuthorityExpiresAt.IsZero` | 654:70 |
| `now.Before` | 655:5 |
| `validStrategyDigest` | 655:51 |
| `fmt.Errorf` | 656:16 |
| `descriptor.RestartNotBefore.IsZero` | 658:124 |
| `validStrategyWorkerRefusal` | 659:96 |
| `descriptor.RestartNotBefore.IsZero` | 659:151 |
| `fmt.Errorf` | 660:16 |
| `fmt.Errorf` | 666:16 |
| `descriptor.AuthorityExpiresAt.IsZero` | 668:72 |
| `fmt.Errorf` | 669:16 |
| `descriptor.RestartNotBefore.IsZero` | 671:123 |
| `fmt.Errorf` | 672:16 |
| `make` | 676:22 |
| `fmt.Errorf` | 686:16 |
| `make` | 691:62 |
| `make` | 691:91 |

## State mutations and fallbacks

- 위 편집 설명 외 없음.

## Safety conclusion

- Safe edit boundary: 분기 · 검증 불변. 감독자가 진입을 닫을 수만 있고(열 수단 없음) 그 수단은 호출자가 준다.
- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함.
