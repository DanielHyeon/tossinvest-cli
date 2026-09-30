# Branch Test Map: `ExitObserver.checkOutage`

- Source: `internal/app/engine/exitloop.go` (817-854); **편집 전** 측정 — `analysis/harness/coverage-pre-unit5-outage.json`(연결 워크트리 `b01e0cd0`, 시험 24개).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 819:2 | 관측 이력 없음 | `TestA092CredentialTighteningRecordsWithoutSending` | 편집 전(기준선) | 블록 819.20-821.3을 시험 2개가 실행, PASS |
| B2 | if at 822:2 | 한도 전 | `TestA092CredentialTighteningRecordsWithoutSending` | 편집 전(기준선) | 블록 822.46-824.3을 시험 3개가 실행, PASS |
| B3 | if at 825:2 | 이미 알림 | `TestASustainedOutageBlocksEntriesAndAlertsOnce` | 편집 전(기준선) | 블록 825.20-827.3을 시험 1개가 실행, PASS |
| B4 | if at 843:2 | 승격 수단 없음 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
| B5 | if at 848:2 | 승격 오류 | (미실행) | 편집 전(기준선) | 측정 표본의 시험 0개 |
