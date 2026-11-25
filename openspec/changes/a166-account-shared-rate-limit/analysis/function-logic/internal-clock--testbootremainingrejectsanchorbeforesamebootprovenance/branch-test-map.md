# Branch Test Map: `TestBootRemainingRejectsAnchorBeforeSameBootProvenance`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | anchor=saved 이고 uptime 0(앞섬) — `ErrBootStamp` 단언(B2 uptime 항). 유지 | 자기 자신 | overlay 실측(`../../time-namespace-lot/selected-tests-baseline-vs-overlay.log`)에서 편집 전 단언 실패 확인 | yes(`../../time-namespace-lot/green-after-edit.log`) |
| B2 | anchor=saved 의 ns inode+1 — 편집 전: `ErrBootStamp` 단언(「conflicting same boot anchor」, B2 ns 항) | 자기 자신 | 같음 | yes |
| B3(편집 뒤) | 다른 ns 의 앞선 anchor 거절 | 자기 자신 | 새 단언(편집 전 계약에서도 참) | yes |
| 변이(실측, `../../time-namespace-lot/mutation-run-2.log`) | M3(B2 uptime 항 삭제) — 앞 판; M4a · M4b · M4c(B2 · B4 · B8 에 ns 항 복원) — 뒤 판(수락 `{9s,false}`) | `../../time-namespace-lot/mutation-ledger.md` | — | — |
