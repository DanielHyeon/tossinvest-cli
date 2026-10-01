# Function Logic Map: `Context.NewRefreshingPairedStrategyEntrySupervisor`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- AST evidence: `ast.json` — **편집 뒤**, :388–416, 분기 4, source_sha256 `64f1cc0b85ec…`.
- Risk scan: `risk-pattern-report.md`
- 편집: (5.6.2.1) 새 B2 `if c.Entry == nil` — 진입 게이트 없는 Context 에서는 생산 감독자를 만들지 않는다(그 조립에서는 중앙 무결성 고장이 진입이 아니라 프로세스를 닫게 되므로). 감독자 옵션에 `EntryGate: c.Entry`. 편집 전 B2 · B3 → B3 · B4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 편집이 더한 입력 | 진입 게이트(`*execgw.EntryGate`) 또는 없음 | 엔진 조립(`Context.Entry`) | 위 편집 설명 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:389) | nil Context · nil 시계 → 거절 | — | (측정 표본 0) |
| B2 | if (:393) | (새) 진입 게이트 없음 → `ErrRuntimeUnavailable` | — | `TestTheProductionStrategySupervisorRefusesAContextWithoutAnEntryGate` |
| B3 | range (:397) | KR · US 권한 갱신 전용 worker 둘 | — | `TestTheProductionStrategySupervisorBlocksOnTheEnginesOwnEntryGate` |
| B4 | if (:409) | 감독자 생성 실패 → 오류 | — | (측정 표본 0) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `errors.New` | 390:15 |
| `fmt.Errorf` | 394:15 |
| `make` | 396:13 |
| `append` | 399:13 |
| `c.runProductionStrategyMarketCycle` | 402:12 |
| `NewStrategyEntrySupervisor` | 406:21 |
| `c.strategyProjectionMu.Lock` | 412:2 |
| `c.strategyProjectionMu.Unlock` | 414:2 |

## State mutations and fallbacks

- 위 편집 설명 외 없음.

## Safety conclusion

- Safe edit boundary: 생산 기동 순서에서 이 생성자 앞의 `Recovery` 가 이미 같은 게이트를 요구하므로(runtime_wiring.go) 생산 기동 동작 변화 0.
- High-risk impact: yes(진입 게이트 경로) — 편집은 진입을 닫는 방향만 더함.

> **5.2.2.1(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 편집으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

> **5.2.2.1 리뷰 수리(2026-09-30) 재추출** — 이 함수 본문 · 분기 종류 불변(같은 파일 `runProductionStrategyMarketCycle` 의 몸통 이동으로 줄 이동 · 파일 해시만 바뀜, `analysis/harness/shift_same_file_bundles.py`). 분기 좌표는 `ast.json` 이 정본.

a112 5.2.2.2: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 5.2.2.2 리뷰 수리: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
