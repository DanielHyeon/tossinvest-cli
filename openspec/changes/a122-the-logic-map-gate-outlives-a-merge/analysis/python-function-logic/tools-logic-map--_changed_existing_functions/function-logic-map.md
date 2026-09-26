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

## task 7.5.13 · 7.5.26 잔여 — 이름은 numstat 레코드에서, 머리 줄은 대조만 (2026-09-27)

> 편집 전 `ast.before-7513.json`(HEAD `26e5bb3f`) · 편집 후 `ast.after-7513.json`, `759_flm_rows.py` 로 정렬. 안쪽 함수 `section_name` 의 분기가
> 이 열거에 같이 든다(B5~B7 — 따로 번들 `tools-logic-map--section_name/`).

편집 후 `tools/logic-map/check_analysis.py:597-797` · 분기 41 · 반환 5 · raise 6 · 호출 67 (`ast.after-7513.json`, source sha `c9a58a69af03`)
편집 전 `tools/logic-map/check_analysis.py:568-748` · 분기 36 · 반환 4 · raise 4 · 호출 60 (`ast.before-7513.json`, revision `26e5bb3f`, source sha `ff590b79db6e`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 630 | If | `if process.returncode:` | 같음 |
| B2 | B2 | 632 | BoolOp | `process.stderr.decode('utf-8', 'replace').strip() or f'git diff failed for base {base}'` | 같음 |
| — | B3 | 652 | comprehension | `for _, _, paths in records` | **새** |
| — | B4 | 652 | comprehension | `for raw in paths` | **새** |
| — | B5 | 658 | BoolOp | `index < 0 or index >= len(named)` | **새** |
| — | B6 | 658 | If | `if index < 0 or index >= len(named):` | **새** |
| — | B7 | 664 | If | `if header != _git_header_path(prefix + name):` | **새** |
| B3 | B8 | 673 | BoolOp | `not old_source or not hunks` | 번호만 |
| B4 | B9 | 673 | If | `if not old_source or not hunks:` | 번호만 |
| B5 | B10 | 678 | If | `if old_source not in comparison.old:` | 번호만 |
| B6 | B11 | 684 | Try | `try:` | 번호만 |
| B7 | B12 | 686 | comprehension | `for function in old_functions` | 번호만 |
| B8 | B13 | 687 | For | `for function in old_functions:` | 번호만 |
| B9 | B14 | 690 | If | `if any((intersects(start, end, old, old_count) for old, old_count, _, _ in hunks)):` | 번호만 |
| B10 | B15 | 690 | comprehension | `for old, old_count, _, _ in hunks` | 번호만 |
| B11 | B16 | 691 | BoolOp | `new_source or old_source` | 번호만 |
| B12 | B17 | 693 | BoolOp | `new_source or old_source` | 번호만 |
| B13 | B18 | 700 | IfExp | `_temporary_go(comparison.new[new_source]) if new_source in comparison.new else None` | 번호만 |
| B14 | B19 | 701 | If | `if current is not None:` | 번호만 |
| B15 | B20 | 702 | For | `for function in go_functions(current, root):` | 번호만 |
| B16 | B21 | 709 | If | `if qualified(function) not in old_qualified:` | 번호만 |
| B17 | B22 | 713 | If | `if any((intersects(start, end, new, new_count) for _, _, new, new_count in hunks)):` | 번호만 |
| B18 | B23 | 713 | comprehension | `for _, _, new, new_count in hunks` | 번호만 |
| B19 | B24 | 725 | If | `if current is not None:` | 번호만 |
| B20 | B25 | 740 | If | `if match is None:` | 번호만 |
| B21 | B26 | 745 | BoolOp | `match.group(2) or 1` | 번호만 |
| B22 | B27 | 747 | BoolOp | `match.group(4) or 1` | 번호만 |
| B23 | B28 | 752 | For | `for line in diff_lines:` | 번호만 |
| B24 | B29 | 753 | If | `if line.startswith('diff --git '):` | 번호만 |
| B25 | B30 | 759 | If | `if in_body:` | 번호만 |
| B26 | B31 | 762 | If | `if hunk(line):` | 번호만 |
| B27 | B32 | 766 | If | `if bodied:` | 번호만 |
| B28 | B33 | 768 | If | `if line.startswith('--- '):` | 번호만 |
| B29 | — | — | IfExp | `'' if value == '/dev/null' else value.removeprefix('a/')` (옛) | **빠짐** |
| — | B34 | 770 | IfExp | `'' if value == '/dev/null' else section_name(value, 'a/', 0)` | **새** |
| B30 | B35 | 771 | If | `if line.startswith('+++ '):` | 번호만 |
| B31 | — | — | IfExp | `'' if value == '/dev/null' else value.removeprefix('b/')` (옛) | **빠짐** |
| — | B36 | 773 | IfExp | `'' if value == '/dev/null' else section_name(value, 'b/', -1)` | **새** |
| B32 | B37 | 783 | If | `if len(bodied) != len(records):` | 번호만 |
| B33 | B38 | 788 | For | `for (added, deleted, paths), had_body in zip(records, bodied):` | 번호만 |
| B34 | B39 | 789 | BoolOp | `added in (b'-', b'0') and deleted in (b'-', b'0')` | 번호만 |
| B35 | B40 | 789 | If | `if added in (b'-', b'0') and deleted in (b'-', b'0'):` | 번호만 |
| B36 | B41 | 791 | If | `if not had_body:` | 번호만 |

**결함.** 옛 B29 · B31 은 머리 줄 값에서 `removeprefix("a/")` · `removeprefix("b/")` 로 이름을 **유도**했다. 이름에 `< 0x20` · `"` · `\\` · `0x7f`
바이트가 있으면 git 은 머리 줄을 C-인용한다(`core.quotePath=false` 로도 — `analysis/harness/7513_quoted.py`, git 2.43.0 ASCII 전수 34 개, 비ASCII 는
안 인용). 인용된 글자는 `"a/…` 로 시작해서 `removeprefix` 가 안 먹고 그 글자가 이름이 됐다. RED(편집 전 코드, 새 시험):

| 모양 | 편집 전 | 편집 뒤 |
|---|---|---|
| 인용되는 이름의 편집(`"` · `\\` · `\x01` · `\x7f`) | `cannot load existing base file <base>:"a/we\\"ird.go"` — **헛거절** | `(이름, F)` 요구 · 현재 쪽 해시 |
| 삭제 · 인용되는 이름**에서의** rename | 같은 헛거절 | base 쪽 요구 |
| 인용되는 이름**으로의** rename | 요구 키가 `('"b/we\\"ird.go"', 'F')` · 새 쪽 바이트를 못 찾아 **현재 논리를 안 봄**(permissive) | `('we"ird.go', 'F')` · 현재 쪽 해시 |
| 인용되는 이름의 새 파일 | `{}`(옛 쪽이 없어 우연히 통과) | `{}` |
| numstat 순서를 뒤집음(주입) | 결함 없음 — 본문 대조만 순서를 썼다 | `the two views of the same diff disagree` |

**수리.** 구역마다 이름을 레코드(`--numstat -z`, 날 바이트 · 가드가 UTF-8 확인)에서 고르고(새 B3 · B4 가 레코드를 푼다), 머리 줄이 그 이름을 git 의
규칙으로 인용한 글자(`_git_header_path`)와 **같아야** 받는다(새 B5~B7 = `section_name`). `/dev/null` 은 머리 줄로 가린다 — 실제 경로는 언제나 `a/` ·
`b/` 가 붙으므로 모호하지 않다. 7.5.24 가 **순서**로 짓던 짝은 이제 구역마다 이름으로도 대조된다.

**왜 (c) 이고 (a) · (b) 가 아닌가.** (a) 인용을 **푸는** 해독기는 틀리면 조용히 다른 이름을 판정한다. (b) 인용된 이름을 이름 대고 거절하면 7.5.24 가 못 박은
정상 입력(인용되는 이름의 새 파일 — `test_a_name_git_has_to_quote_is_not_called_a_vanished_body`)을 **새로 거절**한다. (c) 는 이미 날 바이트로 받고 있는 레코드를
이름으로 쓰고, 새 코드(렌더러)는 **대조에만** 쓴다 — 틀리면 두 글자가 달라 이름 댄 결함(막는 쪽)이다. `--numstat -z` 의 rename 표기(빈 경로 칸 + 옛 · 새
이름이 NUL 로 따로)는 7.5.22 가 이미 재서 `_numstat_records` 가 읽는다.

**거부 · 수용이 바뀌는 정상 입력(편집 전에 셌다).** `7513_quoted.py` 센서스(HEAD `26e5bb3f`): 역사 전부 + 추적의 고유 `*.go` 이름 **1,816** 중 git 이 인용하는
바이트가 든 것 **0**. 새 거절(머리 줄 ≠ 렌더링)이 정상 git 출력에서 서지 않는다는 것은 `main()` A/B 전수가 잰다(review.md). 가드의 `\n` · `\r` · `\t` 거절은
**그대로 둔다** — 받으면 이름 댄 판정 줄에 줄바꿈 · 탭이 들어간다. 받을지는 열린 task 7.5.41.
