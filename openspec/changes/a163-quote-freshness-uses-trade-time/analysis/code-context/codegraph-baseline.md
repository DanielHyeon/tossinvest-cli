# a163 CodeGraph baseline (SDD 3단계)

- 기준: HEAD `f7d80c348f612cb7d981cea0a958c81b40d8b7c8` = `base-commit.txt`. 대상 production 파일은 이 HEAD 에서 워킹트리 무편집.
- 도구: **codegraph 1.6.0**, 공유 체크아웃 `/mnt/D/Axipient/workspace/TossOS` 의 색인(Files 2,929 · Nodes 51,218 · Edges 189,573 —
  `codegraph-status.txt`). 질의 원문: `codegraph-callers-callees.txt` · `codegraph-impact.txt`(2026-10-21).
- `make sdd-sync` 는 이 단계에서 돌리지 않았다(proposal 문서 단계 · production 편집 0). 편집 착수 전 tasks 0.6/1.x 에서 sync 후 재질의한다.

## 비시험 호출 관계 (질의 결과에서 `Test*`·`*_test.go` 를 뺀 것)

| 심볼 | 비시험 callers | 비시험 callees |
| --- | --- | --- |
| `official.adaptPrices` (`market_reads.go:168`) | `Client.Prices` `:142` | `parseDecimal` `:15` |
| `ExitObserver.observe` (`exitloop.go:842`) | `ObserveOnce` `:457` (나머지 3 = 이름 충돌 거짓 엣지, 아래) | `Prices`(인터페이스 `:120`) · `clock.LeaseAnchor` · `decimalOf` · `execgw.NewEvidenceInvalidError` · `Journal.MaxExitObservationCycle` (나머지 4 = 거짓 엣지) |
| `ExitObserver.ObserveOnce` (`:457`) | `ExitObserver.Run` `:395` · `Tracer.Run` `tracer.go:273` | `endExitOrderCycle` · `flushStrategyProtectionAlerts` · `Now` · `FillDetectionBehind` · `checkOutage` · `workingSet` · `settleUnobserved` · `observe` · `symbolsOf` · `noteUnobservedCause` · `quoteUsable` · `noteJudged` · `judge` |
| `ExitObserver.quoteUsable` (`:1179`) | `ObserveOnce` · `judge` · `judgeRatchet` · `judgeLadder` · `refreshObservation` · `record` | `clock.Now` · `clock.LeaseElapsed` |
| `stableObservationID` (`:1202`) | `snapshotContext` `:1190` | `riskcalc.CanonicalDecimal` |
| `Prices`(이름 전체, 비시험 18) | `analysis/quote-time-evidence.md` §3 표 C1~C16 의 출처 | — |

거짓 엣지(이름 충돌, a087 `gate-0.4-codegraph.md` 와 같은 종류): `observe` 의 callers `filldetect.pollLocked` · `cmd/tossctl/engineready.go`
`recoverThenReady` · `awaitEngineReady`, callees `requestCancelled` · `endpointAnswered` · `reconcile.Observe` · `i18n.T` — 원문상 `observe`
본문(`exitloop.go:842-913`)에 그 호출이 없다(`analysis/ast/engine--exitobserver.observe.ast.json` 의 `calls` 에 없음).

## 시험 영향

`codegraph affected <file> --filter '*_test.go'`: `market_reads.go` 1,197 · `exitloop.go` 1,216(`../codegraph-affected-*.txt`). 전이 범위가
넓어 선택 근거로 쓰지 않는다. 직접 시험 파일: `internal/official/market_reads_test.go` · `internal/app/engine/exitloop_test.go` ·
`exit_e2e_test.go` · `exit_observation_identity_test.go` · `exit_identity_concurrency_test.go` (affected 목록 안에 있음).
