# Function Logic Map: `_changed_existing_functions` (Python, a122 task 7.5.25 — 떼어 낸 몸통)

`changed_existing_functions` 의 몸통을 옮긴 것이다. 분기 대조표는 그쪽 FLM 의 "task 7.5.25" 절에 있다
(`../tools-logic-map--changed_existing_functions/function-logic-map.md`). `ast.after-7.5.25.json` — 분기 39 · raise 4.

## task 7.5.31

분기 39 → 40(B40 `if placed is not None:`). `changed_existing_functions` 의 FLM "task 7.5.31" 절을 보라.

## task 7.5.34 — 두 쪽 바이트를 비교에서 읽는다 (2026-09-24)

`ast.before-7.5.34.json`(분기 40) → `ast.after-7.5.34.json`(분기 **36**). raise 4 · 반환 4 그대로. difflib 정렬:

| 편집 전 | 편집 후 | 뜻 |
|---|---|---|
| B5 `if temporary is None:`(`base_file` 의 `git show` 실패) | B5 `if old_source not in comparison.old:` | 옛 쪽은 **검증한** base 바이트다. 없으면(gitlink) 같은 문장으로 멈춘다 |
| B13–B17 (`if target:` · `base_file(target…)` · `contents[…]` · `current and current.exists()`) | B13 · B14 | 현재 쪽은 대상과 무관하게 `comparison.new` 의 바이트다 — 커밋 대상도 검증한 바이트 |
| B40 `if placed is not None:`(git 밖 대조) | — | 지웠다: 판정 blob 을 실제 저장소에서 찾지 않으므로 대조할 거짓말이 없다. 그 대조가 만들던 헛거절(빈 새 `.go` · pathspec 변수)도 같이 없어졌다 |

나머지(B1–B4 · B6–B12 · 편집 전 B18–B39 = 편집 후 B15–B36)는 순서 그대로다. 분기 밖: 판정 diff 가
`*comparison.trees` 를 `env=comparison.environment` 에서 견주고 pathspec 을 안 받는다.

## task 7.5.35 — 판정 diff 의 형식을 게이트가 정한다 (2026-09-25)

`ast.before-7.5.35.json` → `ast.after-7.5.35.json`: 분기 36 · raise 4 · 반환 4 **그대로**, difflib 정렬도 전부 같다 — 바뀐
것은 분기 밖, 판정 diff 의 명령줄이다: `--src-prefix=a/` · `--dst-prefix=b/`(파서가 떼는 글자를 우리가 정한다 —
`diff.noprefix` 면 `b/x.go` 의 요구가 `x.go` 로 가고 `a/x.go` 는 헛거절이었다) · `--no-color`(색 설정은 머리 줄을 못 읽게
해 교차 검사가 헛거절했다) · `--inter-hunk-context=0`(`diff.interHunkContext` 가 훅을 합쳐 편집 안 한 함수까지 요구했다).
접두사와 훅 합치기는 이 change 이전부터(7.5.26); 색은 이 change 의 base 에서 요구를 **조용히** 비웠고(`[]`) 7.5.23 의 교차
검사가 헛거절로 바꿨다(7.5.35 정정 — 첫 판은 "넷 다 이 change 이전부터" 라 적었다). `--unified=0` 을 이기는 `GIT_DIFF_OPTS` 는 `_isolated_comparison` 이 자식 환경에서 지운다.
