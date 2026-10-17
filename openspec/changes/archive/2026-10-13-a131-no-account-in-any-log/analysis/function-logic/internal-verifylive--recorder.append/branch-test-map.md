# Branch Test Map: `Recorder.Append`

- Source: `internal/verifylive/record.go` (349-383) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 350:2 | `if e.FormatVersion == 0 {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B2 | if at 353:2 | `if e.Kind == "" {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B3 | if at 357:2 | `if err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B4 | if at 362:2 | `if err != nil {` | `TestA131VerifyRecordSealsTheAccount` · `TestA131SealedVerifyProofStillBinds` | 편집 영향 |
| B5 | if at 369:2 | `if r.f == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B6 | if at 373:2 | `if err := reopenIfReplaced(r); err != nil {` | `TestA131VerifyAppendReopensAReplacedRecord` | 편집 영향 |
| B7 | if at 376:2 | `if _, err := r.f.Write(line); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B8 | if at 379:2 | `if err := r.f.Sync(); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |

## a131 시험

- `TestA131VerifyRecordSealsTheAccount`
- `TestA131SealedVerifyProofStillBinds`
- `TestA131VerifyAppendReopensAReplacedRecord`
