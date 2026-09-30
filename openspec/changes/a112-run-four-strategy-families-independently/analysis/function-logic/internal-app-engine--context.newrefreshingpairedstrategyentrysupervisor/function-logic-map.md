# Function Logic Map: `Context.NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json` — **편집 뒤**, :390–418, 분기 4, source_sha256 `da4fa6d1b572…`.
- Risk scan: `risk-pattern-report.md`
- 편집: (5.6.2.1) 새 B2 `if c.Entry == nil` — 진입 게이트 없는 Context 에서는 생산 감독자를 만들지 않는다(그 조립에서는 중앙 무결성 고장이 진입이 아니라 프로세스를 닫게 되므로). 감독자 옵션에 `EntryGate: c.Entry`. 편집 전 B2 · B3 → B3 · B4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:391) | nil Context · nil 시계 → 거절 | — | (측정 표본 0) |
| B2 | if (:395) | (새) 진입 게이트 없음 → `ErrRuntimeUnavailable` | — | `TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate` |
| B3 | range (:399) | KR · US 권한 갱신 전용 worker 둘 | — | `TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate` |
| B4 | if (:411) | 감독자 생성 실패 → 오류 | — | (측정 표본 0) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 392:15 |
| `fmt.Errorf` | 396:15 |
| `make` | 398:13 |
| `append` | 401:13 |
| `c.runProductionStrategyMarketCycle` | 404:12 |
| `NewStrategyEntrySupervisor` | 408:21 |
| `c.strategyProjectionMu.Lock` | 414:2 |
| `c.strategyProjectionMu.Unlock` | 416:2 |

## State mutations and fallbacks

- 위 편집 설명 외 없음.

## Safety conclusion

- Safe edit boundary: 생산 기동 순서에서 이 생성자 앞의 `Recovery` 가 이미 같은 게이트를 요구하므로(runtime_wiring.go) 생산 기동 동작 변화 0.
- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함.
