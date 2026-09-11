# Function Logic Map: `_base_shaped_bundles`

`tools/logic-map/check_analysis.py:432-454` · Python · **새 함수** (task 7.1).
`ast.worktree.json` 을 `enumerate.py` 가 기계로 열거했다.

## Inputs and invariants

`root` · `base` · `analysis`. 답하는 질문은 하나다 — **이 change 의 `revision: current`
번들 중 base 의 소스를 적고 있는 것이 무엇인가.** 이름 목록을 돌려주고, 없으면 빈 목록이다.

불변식은 **아무것도 거절하지 않는다**는 것이다. 이 함수는 게이트 판정에 들어가지 않고
`main` 의 조언 한 줄만 고른다. 판정을 바꾸는 것은 7.2 의 일이다.

## Branches and early returns (분기 5 · 반환 1 · raise 0)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 447 | For | `for ast_path, source, digest in _pinning_bundles(root, analysis):` |
| B2 | 449 | BoolOp | `path.is_file() and hashlib.sha256(path.read_bytes()).hexdigest() == digest` |
| B3 | 449 | If | `if path.is_file() and hashlib.sha256(path.read_bytes()).hexdigest() == digest:` |
| B4 | 452 | BoolOp | `at_base is not None and hashlib.sha256(at_base).hexdigest() == digest` |
| B5 | 452 | If | `if at_base is not None and hashlib.sha256(at_base).hexdigest() == digest:` |

`B2`·`B3` 이 **신선한 번들을 먼저 버린다**. 이 `continue` 가 빠지면 base 이후 안 바뀐
파일의 신선한 번들이 전부 걸린다 — `sha256(오늘) == sha256(base)` 이기 때문이다.
[[fail-closed-must-name-what-it-rejects]] 가 말하는 "거절할 정상 입력"이 정확히 그것이고,
`test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice` 가 그 자리를 잡는다.

`B4`·`B5` 가 base 의 blob 과 대조한다. `_committed_bytes` 가 `None` 이면(그 커밋에 그 파일이
없으면) 세지 않는다 — 없는 것을 "base 를 기술한다"고 부르지 않는다.

## Calls and live bindings

`_pinning_bundles`(번들 선별의 **유일한** 자리) · `hashlib.sha256` · `_committed_bytes`
(`git show <base>:<path>`) · `path.is_file` · `sorted`.

번들 선별을 다시 쓰지 않고 `_pinning_bundles` 를 부른 이유는 [[two-judgements-cover-for-each-other]]
다 — 선별이 두 벌이면 한쪽만 고쳐도 양쪽 시험이 초록이 된다.

## State mutations and fallbacks

없다. 읽기만 한다. fallback 도 없다 — 못 읽으면 그 번들을 **안 센다**(빈 목록 쪽으로
기울고, 그쪽이 조언을 살리는 방향이다).

## Safety conclusion

읽기 전용 게이트 도구. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지
않는다. Go 변경 0. 게이트의 rc 에 닿지 않는다 — `check` 의 반환 14개가
편집 전후로 같다는 것이 그 근거다.

## 분기를 묶는 시험 (편집 후 실측)

| 분기 | 묶는 시험 | 그 자리를 재는 변이 |
|---|---|---|
| `B2`·`B3` 신선 번들 건너뛰기 | `test_a_fresh_bundle_for_a_file_unchanged_since_the_base_keeps_the_advice` | M3 |
| `B4`·`B5` base blob 대조 | `test_a_stale_bundle_that_does_not_record_the_base_still_hears_the_advice` | M1 · M4 |
| `B1` 번들 순회 | 위 둘 + `test_a_bundle_that_records_the_base_is_not_told_to_record_a_landing` | M7 |

변이 아홉은 전부 CAUGHT 다(`branch-test-map.md` 의 표 · `71_mutations.py`).
