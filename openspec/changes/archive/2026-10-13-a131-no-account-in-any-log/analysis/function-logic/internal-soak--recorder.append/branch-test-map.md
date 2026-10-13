# Branch Test Map: `Recorder.Append`

- Source: `internal/soak/record.go` (201-235) — 편집 뒤

| Branch | Anchor | Scenario | Test | a131 |
|---|---|---|---|---|
| B1 | if at 202:2 | `if c.FormatVersion == 0 {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B2 | if at 205:2 | `if c.Kind == "" {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B3 | if at 209:2 | `if err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B4 | if at 214:2 | `if err != nil {` | `TestA131SoakRecordSealsTheAccount` · `TestA131SealedSoakRecordStillAttestsTheSameAccount`(성공 경로 — 봉인 실패 경로는 키를 OpenRecorder 가 준비하므로 생산에서 도달 불가에 가까움, 시험 없음) | 편집 영향 |
| B5 | if at 221:2 | `if r.f == nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B6 | if at 225:2 | `if err := reopenIfReplaced(r); err != nil {` | `TestA131SoakAppendReopensAReplacedRecord` | 편집 영향 |
| B7 | if at 228:2 | `if _, err := r.f.Write(line); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |
| B8 | if at 231:2 | `if err := r.f.Sync(); err != nil {` | a131 무편집 분기 — 패키지 기존 시험 전체(`go test` 영수증 analysis/verify) | 무편집 |

## a131 시험

- `TestA131SoakRecordSealsTheAccount`
- `TestA131SealedSoakRecordStillAttestsTheSameAccount`
- `TestA131MissingSealKeyFailsClosed`
- `TestA131SoakAppendReopensAReplacedRecord`
