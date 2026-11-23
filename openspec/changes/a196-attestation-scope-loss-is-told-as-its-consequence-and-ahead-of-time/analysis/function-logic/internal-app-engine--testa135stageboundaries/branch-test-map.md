# Branch Test Map: `testa135stageboundaries`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 13행 순회 | 자기 자신 | 편집 전 판이 GREEN 구현 뒤 3행 FAIL(`../../tdd/green-before-a135-fix.txt`) | yes — PASS |
| B2 | stage 불일치 보고 | 자기 자신(실패 경로) | 위 FAIL 이 여기서 남 | 통과 실행에서 안 탐(정상) |
| B3 | days 불일치 보고 | 자기 자신(실패 경로) | — | 통과 실행에서 안 탐(정상) |
| B4 | critical 불일치 보고 | 자기 자신(실패 경로) | — | 통과 실행에서 안 탐(정상) |
