# Branch Test Map: `TestTheCacheIntervalsAreTheirOwnConstants`

- 편집 후(a205 GREEN). RED 원문 `../../verify/red.txt`, GREEN 원문 `../../verify/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | lifecycle 간격 ≠ 30s | `TestTheCacheIntervalsAreTheirOwnConstants` | 해당 없음(무변경 단언, 거짓 갈래) | yes — 거짓 갈래 |
| B2 | runtime 간격 ≠ 2s | 같은 시험 | yes — red.txt `position_policy_cache_test.go:656: engineRuntimeInterval = 5s, want 2s` | yes — green.txt console ok |
| B3 | 두 간격 같음 | 같은 시험 | 해당 없음(거짓 갈래) | yes — 거짓 갈래 |
| B4 | 리터럴 순회 | 같은 시험 | yes(순회함) | yes |
| B5 | 리터럴 없음 | 같은 시험 | yes — red.txt `:667: "const engineRuntimeInterval = 2 * time.Second" is no longer a literal` | yes — 거짓 갈래 |
