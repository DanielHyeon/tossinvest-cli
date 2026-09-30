# Branch Test Map: `Retrier.escalateCredentialFailure`

- Source: `internal/execgw/retry.go` (409-419); **편집 전** 측정 — `analysis/harness/coverage-pre-unit5-retry.json`(연결 워크트리 `b01e0cd0`, 시험 15개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 410:2 | 승격 수단 없음 | `TestAuthFailureLatchesEntryImmediately` | 편집 전(기준선) | 블록 410.64-412.3을 시험 4개가 실행, PASS |
| B2 | if at 413:2 | 승격 오류 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
