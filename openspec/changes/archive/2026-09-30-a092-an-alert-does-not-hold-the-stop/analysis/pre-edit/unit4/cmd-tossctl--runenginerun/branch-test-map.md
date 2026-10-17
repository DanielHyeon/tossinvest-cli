# Branch Test Map: `runEngineRun`

- Source: `cmd/tossctl/engine.go`; **편집 전** 측정 — `analysis/harness/coverage-pre-unit4-cmd.json`(`./cmd/tossctl` 시험 81개), 연결 워크트리 `22db26e7`.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 188:2 | `if ctx == nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B2 | if at 194:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B3 | if at 200:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B4 | if at 210:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B5 | if at 211:3 | `if clauses := engine.UnmetInterlockClauses(err); clauses != ` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B6 | range at 213:4 | `for _, clause := range clauses` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B7 | if at 222:2 | `if !ectx.Automation.Verified` | `TestAGateOffEngineRefusesWithoutEnumeratingClauses` | 편집 전(기준선) | 블록 222.31-224.3을 시험 1개가 실행, PASS |
| B8 | if at 232:2 | `if lockPath, verr := engineVerifyLockPath(root); verr == nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 232.63-233.81을 시험 2개가 실행, PASS |
| B9 | if at 233:3 | `if fresh, at := runlock.Fresh(lockPath, clk.Now(), runlock.S` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B10 | if at 249:2 | `if token, terr := engineProcInstance(os.Getpid()); terr == n` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 249.65-251.3을 시험 2개가 실행, PASS |
| B11 | if at 252:2 | `if merr != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B12 | else at 256:9 | `} else` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B13 | if at 268:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B14 | if at 272:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B15 | if at 295:2 | `if policyControl != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 295.26-297.3을 시험 2개가 실행, PASS |
| B16 | if at 298:2 | `if policyControlErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | 편집 전(기준선) | 블록 298.29-301.3을 시험 1개가 실행, PASS |
| B17 | if at 303:2 | `if policyRuntime != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 303.26-305.3을 시험 2개가 실행, PASS |
| B18 | if at 306:2 | `if policyRuntimeErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | 편집 전(기준선) | 블록 306.29-309.3을 시험 1개가 실행, PASS |
| B19 | if at 321:2 | `if strategyRuntime != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | 편집 전(기준선) | 블록 321.28-330.3을 시험 1개가 실행, PASS |
| B20 | if at 331:2 | `if projErr != nil` | `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 331.20-333.3을 시험 1개가 실행, PASS |
| B21 | if at 338:2 | `if err != nil` | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B22 | if at 342:2 | `if alertControl != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine`, `TestAFailedStrategyProjectionDoesNotStopTheEngine` | 편집 전(기준선) | 블록 342.25-344.3을 시험 2개가 실행, PASS |
| B23 | if at 345:2 | `if alertControlErr != nil` | `TestAFailedSiblingEndpointDoesNotStopTheEngine` | 편집 전(기준선) | 블록 345.28-348.3을 시험 1개가 실행, PASS |
