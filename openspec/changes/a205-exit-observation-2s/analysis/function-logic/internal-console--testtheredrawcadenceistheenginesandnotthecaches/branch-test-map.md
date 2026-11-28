# Branch Test Map: `TestTheRedrawCadenceIsTheEnginesAndNotTheCaches`

- 편집 후(a205 GREEN). RED 원문 `../../verify/red.txt`, GREEN 원문 `../../verify/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 재로드 주기가 엔진 2s 와 다름 | `TestTheRedrawCadenceIsTheEnginesAndNotTheCaches` | yes — red.txt `line_cadence_test.go:103: lineRefreshInterval = 5s, want the engine's 2s` | yes — green.txt console ok(거짓 갈래) |
| B2 | 재로드 주기 ≥ 캐시 TTL | 같은 시험 | 해당 없음(기존 단언, 5s · 2s 모두 거짓 갈래) | yes — 거짓 갈래 |
| B3 | 두 화면 순회 | 같은 시험 | 해당 없음 | yes |
| B4 | 화면 재로드 초 ≠ 기대 | 같은 시험 | 해당 없음(사본 · 화면이 같은 상수를 읽어 RED 에서도 거짓) | yes — 거짓 갈래 |
