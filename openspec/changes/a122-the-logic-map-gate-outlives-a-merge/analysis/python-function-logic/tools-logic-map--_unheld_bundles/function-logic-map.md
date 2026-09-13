# Function Logic Map: `_unheld_bundles`

`tools/logic-map/check_analysis.py:532-567`(HEAD `2b5b05c1`) · Python ·
분기 9 · 반환 1 · raise 0 (`ast.before-7.6.json`)

> **공백 기록.** 7.2.1 이 새로 만든 함수인데 FLM 이 없었다. 7.6 이 HEAD 에서 기계로 열거해 남긴다.
> 7.6 은 이 함수의 **내부를 바꾸지 않는다** — 부르는 자리만 `_landing_refusal` 한 곳이 된다.

## Inputs and invariants

`candidate` 커밋이 판정이 읽은 `ast.json` 바이트를 그대로 들고 있는지. 등식은 **워킹트리**와 세운다
(판정이 워킹트리를 읽으므로). 범위는 `ast.json` 하나 — 산문까지 넓히면 76건 중 3건이 착지를 잃는다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L550 | For | `for ast_path, _, _ in _pinning_bundles(root, analysis):` |
| B2 | L551 | Try | `try:` |
| B3 | L553 | ExceptHandler | `except ValueError:` |
| B4 | L555 | Try | `try:` |
| B5 | L557 | ExceptHandler | `except OSError:` |
| B6 | L561 | If | `if committed is None:` |
| B7 | L563 | If | `if before:` |
| B8 | L565 | BoolOp | `committed is None or committed != judged` |
| B9 | L565 | If | `if committed is None or committed != judged:` |

## Calls and live bindings

`_pinning_bundles` · `Path.relative_to` · `Path.read_bytes`(심링크를 따라간다 — 판정과 같은 읽기) ·
`_committed_bytes` ×2(지금 경로, 없으면 옮기기 전 경로 `_pre_archive_path`).

## State mutations and fallbacks

없다. 읽을 수 없는 번들(`OSError`)은 **미보유**로 센다 — fallback 이 아니라 거절 쪽이다.

## Safety conclusion

7.6 에서 불변. 호출자는 7.6 전 `resolve_landing` · `compute_landing` 둘, 후 `_landing_refusal` 하나.
