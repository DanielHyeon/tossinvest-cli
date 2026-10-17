# Branch Test Map: `TestTheExitObserverDefersToFillDetection`

- Source: `cmd/tossctl/engine_test.go` (383-392). 시험 함수 자체라 분기는 그 단언이다. 반증: 연결 워크트리 `c6e2e3ac` 에서 `engine.go` 의
  `SLO: detectorPressure{detector: detector},` 줄을 지운 사본으로 돌려 B1 이 FAIL(`engine_test.go:387`), 원복 뒤 PASS.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 386:2 | SLO 배선 부재 | 자기 자신 | SLO 줄 삭제 사본에서 FAIL | 작업 트리에서 PASS |
| B2 | if at 389:2 | 감지기 판정 부재 | 자기 자신 | 해당 없음(편집 전과 같은 단언) | PASS |
