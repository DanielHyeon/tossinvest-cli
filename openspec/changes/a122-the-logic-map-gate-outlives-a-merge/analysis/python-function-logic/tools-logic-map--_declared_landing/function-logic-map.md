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


---

## 편집 계획 — task 6.2.1 (코드보다 먼저, `ast.before-6.2.json` = HEAD `fb4e8f92`)

호출 자리 열거(AST, `check_analysis.py` 전체): 이 함수를 부르는 자리 **다섯**.

| 줄 | 함수 | 묻는 것 | 둘러싼 try |
|---|---|---|---|
| 486 | `resolve_landing` | 값 | 없음 — 호출자 `check` 의 try(B23)가 받는다 |
| 891 ×2 | `check`(빌린 증거) | 값(두 선언이 같은가) | `except ValueError` |
| **898** | `check`(이관 probe) | **있는가** | **없음** |
| **1027** | `record_landing` | **있는가** | **없음** — 6.1.2 가 만든 자리, 4.4 뒤라 4.4 가 못 셌다 |

터지는 두 자리가 둘 다 **"있는가"** 를 묻는다. 그 질문에 해독하는 함수를 부른 것이
기전이다 — try 로 감싸는 것은 증상을 받는 것이고, 질문을 맞는 함수에 거는 것이 근본 수정이다.

- 새 함수 `_landing_record(change_dir, root) -> tuple[str, bytes] | None` 가 B1·B2
  (`relative_to` 실패 → None)와 `_committed_bytes` 를 가져간다. 해독하지 않는다.
- 이 함수는 그것을 불러 B3(없음)·B4·B5(해독)만 남긴다. 반환 3 → 2, raise 1 → 1,
  분기 5 → 3. 오류 문장 `landing point is not UTF-8: {relative}` 는 한 글자도 안 바꾼다.
- 898 · 1027 은 `_landing_record(...) is not None` 으로 묻는다.

spec 이 가른 답: 이관 경로의 착지 기록에는 "이관 경로가 그 기록을 받지 않는다는 것을
이름으로 말한다". 못 읽는 기록도 기록이므로, 해독 실패를 오류로 돌려주는 것
(4.4 처방)보다 **있으니 거절**이 spec 에 맞다.


## 편집 결과 — task 6.2.1 (`ast.after-6.2.json`)

분기 5 → 3 · 반환 3 → 2 · raise 1 → 1. **계획대로다.** B1·B2(`try` / `except ValueError`)가
`_landing_record` 로 옮겨 갔고 `if raw is None` 이 `if record is None` 이 됐다. 오류 문장
`landing point is not UTF-8: {relative}` 는 그대로다.

변이 M9(해독을 `errors="replace"` 로 관대하게) CAUGHT —
`test_the_value_path_still_names_an_undecodable_record` 가 빨개진다. **그 문장을 재는 시험은
이 편집 전에 저장소에 0 개였다.**
