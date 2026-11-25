# Branch Test Map: `TestBootZeroDurationDoesNotHideInvalidClock`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | saved=now, duration 0 → `{0,false},nil` 단언. 유지 | 자기 자신 | overlay 실측(`../../time-namespace-lot/selected-tests-baseline-vs-overlay.log`)에서 편집 전 단언 실패 확인 | yes(`../../time-namespace-lot/green-after-edit.log`) |
| B2 | now ns inode+1, duration 0 — 편집 전: `ErrBootStamp` 단언(「zero masked mismatch」) | 자기 자신 | 같음 | yes |
| B3(편집 뒤) | 같은 boot uptime 역행 + duration 0 거절 | 자기 자신 | 새 단언(편집 전 계약에서도 참) | yes |
| 변이(실측, `../../time-namespace-lot/mutation-run-2.log`) | M8(B5 duration=0 early return 을 B2 앞으로 이동) · M1(B3 uptime 항 삭제) — 새 역행 판; M4 — ns 수락 판 | `../../time-namespace-lot/mutation-ledger.md` | — | — |
