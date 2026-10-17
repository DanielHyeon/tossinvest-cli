# Function Logic Map: `_landing_record`

`tools/logic-map/check_analysis.py:373-386` · Python · **새 함수** (task 6.2.1).
`ast.worktree.json` 을 `enumerate.py` 가 기계로 열거했고, 호출자 표는 편집 후 AST 를 스크립트가 셌다.

## Inputs and invariants

`change_dir` · `root`. HEAD 커밋에 착지 기록이 **있는가**를 답한다 — `(경로, 바이트)` 또는
`None`. **해독하지 않는다.** 그것이 이 함수가 따로 선 이유다: "있는가"와 "무엇이 적혔나"는
다른 질문이고, 앞의 것에 해독 함수를 부르면 못 읽는 기록 앞에서 질문이 터진다.

## Branches and early returns (분기 3 · 반환 2 · raise 0)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 381 | Try | `try:` |
| B2 | 383 | ExceptHandler | `except ValueError:` |
| B3 | 386 | IfExp | `None if raw is None else (relative, raw)` |

옛 `_declared_landing` 의 B1·B2(`relative_to` 실패 → `None`)와 `_committed_bytes` 호출을
**그대로** 옮겼다.

## Calls and live bindings

`_committed_bytes`(git show HEAD:…). 워킹트리를 읽지 않는다(task 1.5 의 이유 그대로).

## 호출자 — 편집 후 AST 로 셈

| 줄 | 함수 | 둘러싼 try |
|---|---|---|
| 397 | `_declared_landing` | 없음 |
| 915 | `check` | 없음 |
| 1044 | `record_landing` | 없음 |

`_declared_landing` 의 남은 호출 자리(모두 **값**을 묻는다):

| 줄 | 함수 | 둘러싼 try |
|---|---|---|
| 503 | `resolve_landing` | 없음 |
| 908 | `check` | ValueError |
| 908 | `check` | ValueError |

`resolve_landing` 안의 호출은 그 함수 안에 try 가 없지만 호출자 `check` 가 `resolve_landing`
을 `except (OSError, RuntimeError, ValueError, json.JSONDecodeError)` 안에서 부른다.
**"있는가"를 묻는 자리에는 이제 해독이 없고, 해독하는 자리는 전부 받는 쪽이 있다.**

## State mutations and fallbacks

없다.

## Safety conclusion

읽기 전용 게이트 도구. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지
않는다. Go 변경 0.

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:593-609` · 분기 3 · 반환 2 · raise 0 (편집 전 L560-573 · 분기 3 · 반환 2 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

상징 `HEAD` 대신 명령이 **한 번** 푼 sha(`head`)를 받는다. 7.5.2 는 `HEAD` 를 열 자리에서 따로 읽고 지문에 표본 하나만 넣었다 — 기록은 표본 **전에** 읽혔고 뒤의 git 호출은 살아 있는 `HEAD` 를 다시 읽어서 가지 전환 한 번(레드팀 repro_b)도, 떠났다 돌아온 `HEAD`(이 세션이 재현)도 rc 0 이었다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 604 | Try | `try:` |
| B2 | 606 | ExceptHandler | `except ValueError:` |
| B3 | 609 | IfExp | `None if raw is None else (relative, raw)` |
