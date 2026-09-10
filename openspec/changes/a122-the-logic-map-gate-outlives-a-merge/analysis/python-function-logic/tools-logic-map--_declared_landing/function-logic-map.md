# Function Logic Map: `_declared_landing`

`tools/logic-map/check_analysis.py:359-377`(편집 후) · Python · **새 함수** (task 1.8)

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json` 을
> `enumerate.py` 가 기계로 열거했고 아래는 그 열거를 옮긴 것이다.

## Inputs and invariants

`change_dir` · `root` 를 받아 **HEAD 커밋에 적힌 착지 선언**을 돌려준다.
반환은 셋 중 하나다: `None`(선언 파일이 없다) · `""`(빈 선언) · 적힌 문자열.

불변식 하나: **값을 판정하지 않는다.** 40자리인지·커밋인지·조상인지·증거가
고정하는지는 전부 `resolve_landing` 이 묻는다. 여기가 답하는 것은 "선언이 있는가,
있다면 무엇이라 적혀 있는가" 하나다. 두 change 의 선언을 **같은 방법으로** 읽어야
빌린 증거의 창 대조가 뜻을 갖기 때문에 이 자리가 생겼다.

## Branches and early returns (ast.json 열거 그대로, 분기 5)

| ID | 줄 | 종류 | 소스 | 무엇을 묻나 |
|---|---|---|---|---|
| B1 | 367 | Try | `try:` | 경로가 root 안인가 |
| B2 | 369 | ExceptHandler | `except ValueError:` | 밖이다 → `return None` |
| B3 | 372 | If | `if raw is None:` | 그 커밋에 파일이 없다 → `return None` |
| B4 | 374 | Try | `try:` | 디코드 |
| B5 | 376 | ExceptHandler | `except UnicodeDecodeError as exc:` | → raise |

반환 셋: `370 None` · `373 None` · `375 raw.decode('utf-8').strip()`.
raise 하나: `377 ValueError('landing point is not UTF-8: …')`.

`None` 과 `""` 를 가르는 것이 이 함수의 계약이다. 합치면 빈 선언이 "선언 없음"이
되어 워킹트리 대상으로 조용히 떨어진다 — `resolve_landing` 쪽 표의 M2 를 볼 것.

## Calls and live bindings

`_committed_bytes`(git show HEAD:…) 하나. **워킹트리를 읽지 않는다** — 읽으면
untracked 파일로 게이트를 통과한 뒤 지울 수 있다(task 1.5).

## State mutations and fallbacks

없다. 순수 읽기다. fallback 도 없다 — 디코드 실패는 raise 이고 호출자 둘
(`resolve_landing` · `check` 의 빌린 증거 블록)이 각자 사유를 붙여 돌려준다.

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 읽기
전용이고 실패 방향은 게이트가 **안 열리는** 쪽이다. Go 파일 변경 0.
