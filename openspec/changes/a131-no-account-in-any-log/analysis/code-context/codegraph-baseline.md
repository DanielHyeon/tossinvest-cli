# CodeGraph baseline — a131

- 날짜: 2026-10-13 · base `233d5d9c` · codegraph **1.6.0** · 질의 직전 `codegraph sync .` · 모든 질의 `--limit 1000`, 원자료 `raw/`.

| 질의 | 결과(비시험 / 전체) | HEAD 대조 |
|---|---|---|
| `callers NewLogger` | 2 / 52: `runEngineRun`(`engine.go:212`) · `buildFlattenWiring`(`flatten.go:253`) | `git grep NewLogger` 비시험 2 자리와 일치 — 생산 obs 로거 구성은 이 둘뿐 |
| `callers accountsLocked` | 2 / 2: `Client.Accounts`(`reads.go:28`) · `Client.ensureAccountSeq`(`client.go:244`) | 계좌 원문 입구가 이 한 함수임을 확인(design D2 R0) |
| `callers runEngineRun` | 1: `newEngineRunCmd` | 일치 |
| `callers Record` | 11 비시험 — 이름 충돌(`execgw`·`trading`·`strategydispatch`·`measure` 의 다른 `Record`) 섞임. `audit.Log.Record` 의 직접 호출자는 `RecordChange` · `RecordAction` | `git grep '\.Record(audit\.Entry'` 비시험 0 — 외부는 둘을 거침 |
| `callers LoadCycles` | 2 / 16: `loadSoakSummary`(`soak.go:602`) · `readSoak`(`console/data.go:234`) | 일치 |
| `callers LoadEntries` | 7 / 30: `supervisedProofs` · `readVerify` · `redoSet` · `report`(console) · `consoleVerifyStarter` · `runVerifyRun` · `loadVerifyRecord` | 일치 — 열린 `AccountRef` 를 쓰는 자리: `supervisedProofs`(→`SucceededEndpoints`) |
| `callers appendLine` | 1 / 2: `Reconcile`(`reconcile.go:168`) | 일치 |
| `callers readRecordStrictNoTail` | 1: `readRecordStrict` ← `Reconcile` | 일치 |
| `callers runSoakAttest` | 1: `newSoakAttestCmd` | 일치 |
| Recorder.Append 쓰는 자리(`git grep`) | soak `soak.go:389` · verify `abort.go:137` · `cleanup.go:321` · `runner.go:544` · `:680` | 모두 `Append` 하나를 거침 — 봉인은 `Append` 안 한 곳 |
| `affected <편집 9 파일>` 기본 | 시험 0~1(`raw/affected-default.txt`) | 하네스 주석의 `isTestPath` 결함 재현 — 결론으로 안 씀 |
| `affected … --filter '*_test.go'` | 1119 줄 — `internal/app/engine` 170 · `internal/journal` 140 · `cmd/tossctl` 90 · `internal/console` 62 · `internal/verifylive` 42 · `internal/official` 36 · `internal/obs` 20 · `internal/soak` 12 … | 검증은 해당 패키지 전체 `go test`(+ race) 로 함 |
