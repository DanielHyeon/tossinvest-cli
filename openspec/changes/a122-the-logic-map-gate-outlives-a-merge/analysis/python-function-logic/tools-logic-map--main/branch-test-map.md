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

## 편집 후 (ast.after-7.1.json) — task 7.1

번호는 `ast.after-7.1.json` 의 열거 순서다. 옛 표의 `B7·B8·B9` 가 여기서는
`B11·B12·B13` 이다 — [[positional-branch-ids-break-hand-renumbering]].

| ID | 소스 | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| B7·B8 | `context.get('base_shaped_bundles') or []` | 아래 넷 전부 | yes |
| B9 | `if base_shaped:` (참) | `test_a_bundle_that_records_the_base_is_not_told_to_record_a_landing` · `test_only_three_bundles_are_named_and_the_rest_are_counted` | yes |
| B9 | (거짓 — 억제하지 않는다) | `test_a_stale_bundle_that_does_not_record_the_base_still_hears_the_advice` · `test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice` · `test_a_failure_says_how_much_of_the_window_is_not_this_change` | yes |
| B10 | `if len(base_shaped) > 3:` (참) | `test_only_three_bundles_are_named_and_the_rest_are_counted` | yes |
| B10 | (거짓 — 셋 이하) | `test_a_bundle_that_records_the_base_is_not_told_to_record_a_landing` | yes |

`B7` 의 오른쪽(`[]`)은 **문맥에 키가 아예 없을 때**다. 그 상태는 착지 해소가 실패한
경우인데, 그러면 `B4` 가 창 줄 자체를 안 찍으므로 여기 도달하지 않는다(4.1 이 세운
"못 잰 창은 안 찍는다"). 도달 불가를 "덮였다"고 적지 않는다.

## 이 편집을 재는 변이 (2026-09-12 실측, `71_mutations.py` · 원복 sha256 동일)

| 변이 | 빨개진 시험 |
|---|---|
| M1 base 모양 판정 제거 (stale 이면 전부 억제) | 1 — `…_stale_bundle_that_does_not_record_the_base…` |
| M2 억제를 아예 안 함 | 2 |
| M3 신선 번들 건너뛰기(`continue`) 제거 | 1 — `…fresh_bundle_for_a_file_unchanged_since_the_base…` |
| M4 base 가 아니라 HEAD 와 대조 | 2 |
| M5 어느 번들인지 안 말함 | 2 |
| M6 억제 갈래에서 창의 크기(사실)를 뺌 | 2 |
| M7 `check` 가 재지 않는다 | 2 |
| M8 이름을 자르지 않고 다 쏟아냄 | 1 — `…only_three_bundles_are_named…` |
| M9 못 적은 수를 안 셈 | 1 — `…only_three_bundles_are_named…` |

**M5 는 처음에 SURVIVED 였다.** 바늘이 `assertIn("internal--own", printed)` 라 출력
**전체**를 봤는데, 같은 이름이 오류 줄
(`[logic-map] internal--own: AST source hash is stale`)에도 있어서 조언 줄이 이름을
잃어도 통과했다. 6.3 이 `resolve_landing` 에서 고친 것과 **같은 결함**이다
([[existence-check-is-not-a-role-check]]). `_advice_line` 로 조언 줄 하나를 집어내
바늘을 좁히고 나서 CAUGHT 가 됐다. M1·M3·M8·M9 가 각각 **한 시험만** 죽이는 것은
그 넷이 서로 다른 자리를 묶고 있다는 뜻이다.
