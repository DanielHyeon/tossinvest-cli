# Branch Test Map: `Runtime.runAuxiliary`

- Source: `internal/app/engine/auxiliary.go` (87-114); **편집 전** 측정 — `analysis/harness/coverage-pre-unit5-aux.json`(연결 워크트리 `b01e0cd0`, 시험 19개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 89:2 | 정상 종료 | `TestEveryAuxiliaryExecutorIsStarted` | 편집 전(기준선) | 블록 89.30-98.3을 시험 2개가 실행, PASS |
| B2 | if at 110:2 | OnStop 없음 | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` | 편집 전(기준선) | 블록 110.23-112.3을 시험 2개가 실행, PASS |
