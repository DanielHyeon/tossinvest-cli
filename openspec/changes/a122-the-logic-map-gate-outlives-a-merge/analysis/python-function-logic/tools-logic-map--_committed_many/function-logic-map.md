# Function Logic Map: `_committed_many`

`tools/logic-map/check_analysis.py:374-422` · Python · **task 7.5 (리뷰 I1)** ·
분기 9 · 반환 3 · raise 0 (`ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 7.5 가 새로 만든 함수라 편집 전 판본이 없다.

## Inputs and invariants

`(root, ref, relatives)` — 한 커밋에서 여러 경로의 blob 을 **한 프로세스**로 읽어
`{경로: bytes | None}` 로 돌려준다. 키 집합은 중복을 없앤 `relatives` 와 **정확히 같다**
(요청한 경로는 답이 없어도 자리가 있다).

존재 이유는 알고리즘이 아니라 **fetch 단위**다. 도구는 blob 을 `git show` 로 한 프로세스에
하나씩 읽었고, 후보 하나를 판정하는 데 번들 수만큼 프로세스를 썼다 — 2026-09-18 실측으로
아무 후보도 안 받는 walk 에서 spawn 의 **97.1~97.3%** 가 이 fetch 였다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 396 | comprehension | `for relative in wanted` (전부 `None` 으로 시작) |
| B2 | 397 | If | `if not wanted:` |
| B3 | 401 | comprehension | `for relative in wanted` (NUL 로 끊은 입력) |
| B4 | 404 | If | `if process.returncode:` |
| B5 | 407 | For | `for relative in wanted:` |
| B6 | 409 | If | `if end < 0:` — 응답이 요청보다 짧다 |
| B7 | 416 | BoolOp | `len(fields) != 3 or not fields[2].isdigit()` |
| B8 | 416 | If | 같은 줄의 `if` — `missing`·`ambiguous` 는 내용이 안 따라온다 |
| B9 | 419 | If | `if fields[1] == b'blob':` |

| 반환 줄 | 값 |
|---|---|
| 398 | `found` — 요청이 비었다 (git 을 안 부른다) |
| 405 | `found` — 프로세스 실패, 전부 `None` |
| 422 | `found` — 읽은 것 |

raise 없음. 다만 `subprocess.run` 의 `TimeoutExpired` 는 그대로 올라가고 경계의
`GATE_FAULTS` 가 오류 줄로 바꾼다 ([[a-fault-must-become-a-verdict]]).

## Calls and live bindings

`subprocess.run` ×1 (`git cat-file --batch -Z`) · `dict.fromkeys` · `bytes.join` ·
`data.find` · `header.rsplit` · `int`. 워킹트리·시각·환경을 안 읽는다.

## State mutations and fallbacks

없다. fallback 은 **하나** — 프로세스가 실패하면 전부 `None` 이고, 이것은 옛
`git show` 의 rc≠0 과 **같은 방향**이다. 부르는 쪽은 `None` 을 불일치로 세므로 판정이
느슨해지지 않는다.

## Safety conclusion

생산 Go 코드 변경 0. 판정을 바꾸지 않는다 — 옛 `git show` 판본과 **같은 바이트**를 낸다
(시험 `test_many_reads_the_same_bytes_as_one`, 그리고 실물 전수 A/B).

엄해지는 자리가 **하나** 있다: 디렉터리를 물으면 `git show <ref>:<dir>` 는 트리 **목록**을
찍었고 여기서는 `None` 이다. 그 바이트는 파일 내용이 아닌데 판정에 들어오면 내용인 척한다.
거부할 정상 입력을 먼저 셌다 — 오늘 실물 입력(번들의 `file` · 번들 `ast.json` 경로)에
디렉터리는 **0 건**이다 ([[fail-closed-must-name-what-it-rejects]]).

`-Z` 는 장식이 아니다. 입력을 줄로 끊으면 개행이 든 경로가 쪼개져 엉뚱한 blob 이나
`missing` 이 되고, 출력을 줄로 끊으면 tree 의 raw 바이트 안 개행이 응답을 쪼갠다.
크기는 **머리가 선언한 값**으로 자른다 — 내용에서 NUL 을 찾으면 tree 에서 틀린다.
