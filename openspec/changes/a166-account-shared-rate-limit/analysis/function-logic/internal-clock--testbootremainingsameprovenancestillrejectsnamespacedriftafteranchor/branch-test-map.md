# Branch Test Map: `TestBootRemainingSameProvenanceStillRejectsNamespaceDriftAfterAnchor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | now=saved 의 ns inode+1, anchor 는 다른 boot(B) — 편집 전: `ErrBootStamp` 단언(B3 ns 항) | 자기 자신 | overlay 실측(`../../time-namespace-lot/selected-tests-baseline-vs-overlay.log`)에서 편집 전 단언 실패 확인 | yes(`../../time-namespace-lot/green-after-edit.log`) |
| B2 | now=saved, uptime-1 — `ErrBootStamp` 단언(B3 uptime 항). 유지 | 자기 자신 | 같음 | yes |
| 변이(실측, `../../time-namespace-lot/mutation-run-2.log`) | M4(B3 에 ns 항 복원) · M5(B7 다른 boot 차감) — 수락 판 `{10s, NeedsAnchor}`; M1(B3 uptime 항 삭제) — 역행 판 | `../../time-namespace-lot/mutation-ledger.md` | — | — |
