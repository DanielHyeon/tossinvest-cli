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

## 편집 후 (task 1.8 — `ast.after-1.8.json`)

선언을 읽는 앞부분을 `_declared_landing` 으로 뽑아서 **편집 지점 뒤의 모든 번호가
넷씩 당겨졌다.** 번호를 손으로 옮기지 않고 옛·새 열거의 소스를 `difflib` 로 정렬해
얻었다 — [[positional-branch-ids-break-hand-renumbering]].

| 옛 ID | 새 ID | 소스 |
|---|---|---|
| B1·B2·B3·B4·B5 | — | `_declared_landing` 으로 이동 (그 함수의 B1~B5) |
| — | **B1** | `if candidate is None:` (**신규** — 파일 없음과 빈 선언을 가른다) |
| B6~B20 | B2~B16 | 넷씩 당겨짐. `if not pinning:` = B19 → **B15**, `if mismatched:` = B20 → **B16** |

덮는 테스트는 그대로다(위 표의 짝을 새 번호로 읽는다). 새로 선 **B1** 은
`test_an_empty_record_is_still_a_declaration` 이 거짓 갈래를,
`test_without_a_landing_record_the_target_is_still_the_worktree` 가 참 갈래를 덮는다.

## 이 편집을 재는 변이 (task 1.8, 2026-09-10 실측)

| 변이 | 빨개진 시험 |
|---|---|
| M1 빌린 증거의 착지 **공유** 판정 삭제 | **3** — 빌리는 쪽만 선언 · 창을 좁힘 · 빌려주는 쪽만 선언 |
| M2 빈 선언을 "선언 없음"으로 읽기 (`is None` → `not`) | **1** — `test_an_empty_record_is_still_a_declaration` |
| M3 `if not pinning:` 거절 삭제 | **2** — 번들 0 · `revision: base` |
| M4 해시 불일치 수집 삭제 | **2** — 위조 착지 둘 (**빌린 쪽 포함**) |
| M5 한쪽만 선언한 창을 눈감기 (`None` 이면 통과) | **2** — 빌리는 쪽만 · 빌려주는 쪽만 |

**M4 가 이 태스크의 핵심 측정이다.** 새 공유 판정은 `resolve_landing` **앞에** 서므로,
3.2.3.1 회귀 시험(`test_borrowed_evidence_refuses_a_landing_it_does_not_describe`)이
공유 규칙에 먼저 걸리면 고정 판정을 통째로 지워도 초록으로 남는다 —
[[two-judgements-cover-for-each-other]] 의 모양 그대로다. 그래서 그 시험의 픽스처를
"양쪽이 **같은** 바닥을 선언"으로 고쳐 고정 판정이 홀로 판정하게 두었고, M4 는 그
뒤에도 2건을 빨갛게 한다. 가려지지 않았다는 것을 **재서** 확인했다.

M2 는 이 편집이 **만들어 낸** 자리다. 옛 판본에는 선언을 읽는 자리가 하나뿐이라
`None`/`""` 구분이 필요 없었다. 뽑아내면서 생긴 구분이므로 시험을 같이 세웠다 —
[[surviving-mutant-may-mean-accidental-safety]].

## 편집 후 (task 1.12) — 번호가 안 바뀐다

정규화는 분기가 아니라 **호출 한 줄**이라 `ast.after-1.12.json` 의 분기 16개가 옛
번호 그대로다(`B11·B12` 의 소스 문자열만 `source` → `raw_source`). 번호가 안 움직이는
편집이라 재번호가 필요 없다는 것도 열거로 확인했다.

| 변이 | 빨개진 시험 |
|---|---|
| N4 번들 경로 정규화 제거 | **1** — `test_an_absolute_bundle_path_still_pins_a_landing` |

분기가 안 늘었으므로 **행동 시험 하나가 이 편집을 재는 전부**다. 저장소 번들 3048개가
전부 상대경로라 실데이터 A/B 로는 영원히 안 보인다 — 시험이 없으면 이 줄은 지워도
초록이다.
