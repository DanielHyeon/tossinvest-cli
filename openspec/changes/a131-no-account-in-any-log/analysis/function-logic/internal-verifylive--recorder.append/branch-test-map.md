# Branch Test Map: `Recorder.Append`

- Source: `internal/verifylive/record.go` (335-360) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 336:2 | `if e.FormatVersion == 0 {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B2 | if at 339:2 | `if e.Kind == "" {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B3 | if at 343:2 | `if err != nil {` | `TestA131VerifyRecordSealsTheAccount` · `TestA131SealedVerifyProofStillBinds` | 편집 영향 |
| B4 | if at 350:2 | `if r.f == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B5 | if at 353:2 | `if _, err := r.f.Write(line); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B6 | if at 356:2 | `if err := r.f.Sync(); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |

## a131 시험

- `TestA131VerifyRecordSealsTheAccount`
- `TestA131SealedVerifyProofStillBinds`
