# Branch Test Map: `Recorder.Append`

- Source: `internal/soak/record.go` (187-212) — 편집 전(base)

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 188:2 | `if c.FormatVersion == 0 {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B2 | if at 191:2 | `if c.Kind == "" {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B3 | if at 195:2 | `if err != nil {` | `TestA131SoakRecordSealsTheAccount` · `TestA131SealedSoakRecordStillAttestsTheSameAccount` | 편집 영향 |
| B4 | if at 202:2 | `if r.f == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B5 | if at 205:2 | `if _, err := r.f.Write(line); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B6 | if at 208:2 | `if err := r.f.Sync(); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |

## a131 시험

- `TestA131SoakRecordSealsTheAccount`
- `TestA131SealedSoakRecordStillAttestsTheSameAccount`
- `TestA131MissingSealKeyFailsClosed`
