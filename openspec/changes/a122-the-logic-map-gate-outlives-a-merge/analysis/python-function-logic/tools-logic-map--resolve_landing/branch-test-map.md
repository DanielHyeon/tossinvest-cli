# Branch Test Map: `resolve_landing`

번호는 같은 디렉터리 `ast.json` 의 열거 순서다. **번호는 위치이므로** 함수를 편집하면
편집 지점 뒤가 다른 분기를 가리킨다 — [[positional-branch-ids-break-hand-renumbering]].
그래서 번호만이 아니라 **소스 한 줄**을 같이 적는다.

## 편집 후 (ast.worktree.json)

| ID | 소스 | 덮는 테스트 | 덮이나 |
|---|---|---|---|
| B3 | `if raw is None:` | `test_without_a_landing_record_the_target_is_still_the_worktree` · `test_an_uncommitted_landing_record_is_not_read` | yes |
| B6 | `if not FULL_SHA.fullmatch(candidate):` | `test_a_landing_written_as_a_revision_expression` | yes |
| B8 | 실재하는 커밋인가 | `test_a_landing_that_is_not_a_commit` | yes |
| B9 | HEAD 의 조상인가 | — 버려진 갈래를 만드는 픽스처가 없다 | **no** |
| B10 | base 의 자손인가 | `test_a_landing_before_the_base` | yes |
| B11 | 고정 순회 | `test_recorded_landing_requires_only_this_changes_functions` | yes |
| B12 | `analysis.is_dir()` | `test_a_zero_bundle_change_cannot_declare_a_landing` (거짓 갈래) | yes |
| B14 | `revision != 'current'` | `test_base_revision_bundles_do_not_pin_a_landing` | yes |
| B16 | `not source or not digest` | — 값이 빈 번들을 만드는 픽스처가 없다 | **no** |
| B18 | 해시 대조 | `test_a_landing_the_evidence_does_not_describe` · `test_borrowed_evidence_refuses_a_landing_it_does_not_describe` | yes |
| **B19** | **`if not pinning:`** | `test_a_zero_bundle_change_cannot_declare_a_landing` · `test_base_revision_bundles_do_not_pin_a_landing` | yes |

B1·B2(경로가 root 밖) · B4·B5(UTF-8 아님)는 덮이지 않는다. **덮이지 않는 넷을 여기
적는 이유는 그것이 이 편집의 결론이 아니기 때문이다** — 이 표는 커버리지를 주장하지
않고 열거한다.

## 이 편집을 재는 변이 (2026-09-10 실측)

| 변이 | 빨개진 시험 |
|---|---|
| M1 `if not pinning:` raise 제거 (구멍 복원) | **2** — 번들 0 · `revision: base` |
| M2 빌린 증거 해소를 착지 판정 **뒤로** 되돌림 | **2** — 빌린 증거의 정상 통과 + 빌린 증거의 위조 거절 |
| M3 리비전을 안 가리고 번들 수를 셈 | **1** — `revision: base` |
| M4 해시 불일치 수집 삭제 | **2** — 위조 착지 둘 |

M2 가 이 태스크의 핵심 측정이다. 되돌리면 **정상 입력이 죽는다**(빌린 증거를 쓰는
change 는 지역 번들이 0 이므로 착지 선언이 통째로 막힌다). 즉 "무엇을 거부하게 되는가"
가 시험으로 고정돼 있다 — [[fail-closed-must-name-what-it-rejects]].

M4 는 처음에 **1건**만 빨갛게 했다. 기존 위조 시험이 사유를 `internal/own.go` 로
찾고 있어서 `validate_target` 의 `AST source hash is stale` 도 그 바늘을 만족했기
때문이다. 판정 둘이 서로를 가려 준 모양이라([[two-judgements-cover-for-each-other]])
그 시험의 바늘을 **착지 판정의 문장**으로 좁혔고, 그 뒤 M4 는 2건을 빨갛게 한다.
