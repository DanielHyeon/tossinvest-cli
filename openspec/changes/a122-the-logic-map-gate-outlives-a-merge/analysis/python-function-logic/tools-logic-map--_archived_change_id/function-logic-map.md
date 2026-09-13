# Function Logic Map: `_archived_change_id`

`tools/logic-map/check_analysis.py:232-241` · Python · **task 7.6 (리뷰 I6)** ·
분기 1 · 반환 1 (`ast.after-7.6.json`)

## Inputs and invariants

아카이브 디렉터리 **이름** 하나. `<YYYY-MM-DD>-<id>` 전체가 맞을 때만 `<id>` 를 준다
(`ARCHIVED_CHANGE.fullmatch`). 접미사로 고르지 않는다 — `2026-08-29-other-reference` 는
`other-reference` 이지 `reference` 가 아니다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L241 | IfExp | `matched.group('change') if matched else ''` |

## Calls and live bindings

`ARCHIVED_CHANGE.fullmatch` · `matched.group`. 부르는 쪽: `resolve_referenced_change`(id → 디렉터리)와
`_pre_archive_path`(아카이브 경로 → 옮기기 전 경로). 편집 전에는 둘이 각자 정규식을 불렀다.

## State mutations and fallbacks

없다.

## Safety conclusion

해독 규칙 자체는 안 바뀌었다 — 정규식이 같고 두 호출자가 그것을 부르는 방식(`fullmatch` +
`group("change")`)도 같다. 바뀐 것은 그 규칙이 **사는 자리의 수**다(2 → 1). shell 쪽 사본
(`tools/gate.sh`)은 그대로다.
