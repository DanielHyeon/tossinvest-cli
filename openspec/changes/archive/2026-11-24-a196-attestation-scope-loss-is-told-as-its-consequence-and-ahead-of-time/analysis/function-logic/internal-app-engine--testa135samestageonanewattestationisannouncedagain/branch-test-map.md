# Branch Test Map: `testa135samestageonanewattestationisannouncedagain`

- 편집 후. 커버: `../../implementation/after-cover-attestation_watch.out`(`GOFLAGS=-trimpath go test -count=1 -run 'A196|A135' -covermode=set ./internal/app/engine`, rc 0). RED 원문 `../../tdd/red.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전반 보고 | 자기 자신(실패 경로) | 편집 전 판이 GREEN 구현 뒤 FAIL(`../../tdd/green-before-a135-fix.txt`) | 통과 실행에서 안 탐(정상) — PASS |
| B2 | 같은 증거 재발급 보고 | 자기 자신(실패 경로) | — | 통과 실행에서 안 탐 |
| B3 | 증거 시각 변경 | 자기 자신 | — | yes — 실행됨(PASS 경로) |
| B4 | 기한 변경 보고 | 자기 자신(실패 경로) | — | 통과 실행에서 안 탐 |
