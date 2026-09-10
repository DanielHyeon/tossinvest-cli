# Branch Test Map: `main`

번호는 같은 디렉터리 `ast.json` 의 열거 순서다 —
[[positional-branch-ids-break-hand-renumbering]]. 소스 한 줄을 같이 적는다.

## 편집 후 (ast.worktree.json)

| ID | 소스 | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| B1 | `if base:` (참) | `test_a_failure_prints_the_comparison_window` · `test_the_success_line_is_unchanged` | yes |
| B1 | `if base:` (거짓 — base 해소 자체가 실패) | — 픽스처 없음 | **no** |
| B2 | `if not landing:` (참) | `test_a_failure_says_how_much_of_the_window_is_not_this_change` | yes |
| B2 | (거짓 — 착지 기록 있음) | `test_empty_required_with_bundles_is_reported_and_still_passes` | yes |
| B3 | `landed_after or '?'` (왼쪽) | 위 두 시험 | yes |
| B3 | (오른쪽 `'?'` — `git rev-list` 실패) | — 픽스처 없음 | **no** |
| B4 | `if errors:` (참) | `test_a_failure_prints_the_comparison_window` | yes |
| B4 | (거짓) | `test_the_success_line_is_unchanged` | yes |
| B6 | 실행 기준선 이관 문구 | 기존 a063 시험 | yes |

덮이지 않는 둘을 여기 적는 이유는 이 표가 커버리지를 **주장하지 않기** 때문이다.

## 이 편집을 재는 변이 (2026-09-10 실측)

| 변이 | 빨개진 시험 |
|---|---|
| N1 창 출력을 실패 반환 뒤로 (early return 복원) | **2** |
| N2 창 크기 줄 삭제 | 1 |
| N3 메시지 1 을 옛 문구로 되돌림 | 1 |
| N4 커밋 수를 재지 않고 상수 `"3"` 으로 | **1** |

N4 가 [[generated-evidence-must-be-measured]] 의 자리다. 처음 판본의 시험은 픽스처의
정답이 마침 3 이라 상수 `"3"` 을 통과시켰다. 커밋을 하나 더 얹어 **4가 되는지**까지
보게 고쳤고, 그 뒤 상수는 죽는다.
