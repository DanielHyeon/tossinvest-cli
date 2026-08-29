# Branch Test Map: `TestOpenReadOnlyReadsWhatTheEngineIsWriting`

본문 무변경이므로 RED 없음. GREEN은 `go test ./internal/journal/...` 통과로 확인.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 첫 read가 오류 없이 돌아온다 | 자체 실행 | — (동작 무변경) | yes |
| B2 | writer가 쓴 행이 read-only handle에 보인다 | 자체 실행 | — | yes |
| B3 | 두 번째 read가 오류 없이 돌아온다 | 자체 실행 | — | yes |
| B4 | handle을 연 뒤의 쓰기도 보인다(스냅샷 비고정) | 자체 실행 | — | yes |
