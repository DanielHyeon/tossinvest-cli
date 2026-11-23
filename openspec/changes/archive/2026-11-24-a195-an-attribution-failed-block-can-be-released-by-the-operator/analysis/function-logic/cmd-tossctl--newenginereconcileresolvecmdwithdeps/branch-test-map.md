# Branch Test Map: `newEngineReconcileResolveCmdWithDeps` (편집 뒤 판)

- Source: `cmd/tossctl/engine_reconcile.go` (76-110), 작업 트리. AST 분기 0.
- 커버 측정: `../../coverage/pre-edit-runEngineReconcileResolve-and-cmd-blocks.txt` — 76.104·93.52·97.2 블록 hit=1.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path(분기 없음) — 명령 생성·플래그·Annotations | `TestEngineReconcileResolveIsRegisteredAsOfficialReadOnly` · `TestEngineReconcileResolveProductionWiringHasNoBrokerMutationOrHTTPDoor`(텍스트) | 해당 없음(기존 본문) | yes — 76.104·97.2 hit=1 |
| (RunE return@99) | 실행 위임 | `executeReconcileResolve` 를 쓰는 시험 전부(예: `TestEngineReconcileResolveRecordsDurableReleaseBeforeSuccess`) | 해당 없음 | yes — 93.52 hit=1 |
| (Long 문장) | 도움말이 원인 중립 | `TestA195HelpAndSuccessOutputAreCauseNeutral` | yes — RED 에서 실패(red.txt, `help still names only quantity mismatch`) | yes |
