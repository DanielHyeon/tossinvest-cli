# Function Logic Map: `_git_header_path` (Python, a122 task 7.5.13 — 새 함수)

## task 7.5.13 — 새 함수 (2026-09-27)

편집 후 `tools/logic-map/check_analysis.py:468-487` · 분기 7 · 반환 2 · raise 0 · 호출 5 (`ast.after-7513.json`, source sha `c9a58a69af03`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 477 | BoolOp | `byte < 32 or byte in (34, 92, 127)` |
| B2 | 477 | If | `if not any((byte < 32 or byte in (34, 92, 127) for byte in raw)):` |
| B3 | 477 | comprehension | `for byte in raw` |
| B4 | 480 | For | `for byte in raw:` |
| B5 | 481 | If | `if byte in _C_LETTER:` |
| B6 | 483 | BoolOp | `byte < 32 or byte == 127` |
| B7 | 483 | If | `if byte < 32 or byte == 127:` |

`core.quotePath=false` 인 git 이 통합 diff 머리 줄에 적을 글자. **대조에만** 쓴다(이름을 풀지 않는다). 인용하는 바이트 `< 0x20` · `"`(0x22) · `\\`(0x5C) · `0x7f`
가 하나라도 있으면 전체를 `"` 로 감싸고, 글자 탈출 아홉(`\\a \\b \\t \\n \\v \\f \\r \\" \\\\`)은 그 글자로, 나머지는 세 자리 8진수로. `0x80` 이상은 그대로.
git 2.43.0 과 같은 글자를 내는지는 시험 `test_the_header_renderer_matches_git_for_every_ascii_byte` 가 ASCII 1~127(`/` 제외)과 비ASCII 둘로 잰다.
