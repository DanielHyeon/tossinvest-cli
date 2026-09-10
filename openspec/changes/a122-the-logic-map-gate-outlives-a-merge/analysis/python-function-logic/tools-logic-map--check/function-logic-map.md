# Function Logic Map: `check` — 해소 호출 두 자리만

`tools/logic-map/check_analysis.py:684-765` · Python ·
분기 34 · 반환 9 · 호출 45 (`ast.json` 기계 열거)

이 change 는 `check` 의 판정 내용을 바꾸지 않는다. 바꾸는 것은 **해소 실패를
어떻게 받느냐** 한 곳이다. 그래서 전체 34개 분기를 옮겨 적지 않고, 열거에서
`resolve_referenced_change` 를 부르는 자리와 그것을 받는 handler 만 적는다.

| 자리 | 호출 | 받는 분기 | 실패하면 |
|---|---|---|---|
| 게이트 대상 | `:690` | B1 `try:`(689) → **B2 `except ValueError:`**(691) | **삼키고** `openspec/changes/<id>` 로 되돌아간다 |
| 빌린 증거 | `:715` | B14 `try:`(714) → B15 `except (OSError, RuntimeError, ValueError, json.JSONDecodeError) as exc`(717) | `:718` `return ["function-logic reference base is invalid: …"]` |

**두 자리가 같은 예외를 정반대로 다룬다.** 아래쪽은 오류로 돌려주고 위쪽은 삼킨다.
그러므로 해소기를 fail-closed 로 만드는 것만으로는 게이트 대상 경로가 안 바뀐다 —
B2 가 새 실패를 **기존 fallback 과 구분**해야 한다.

구분의 근거는 문구가 아니라 **타입**으로 둔다. 문구로 가르면 메시지를 고치는 순간
조용히 뚫린다 — [[existence-check-is-not-a-role-check]].

## Safety conclusion

`check` 의 나머지 32개 분기, 요구 함수 집합, base·landing 규칙, 번들 판정은
한 줄도 바뀌지 않는다. Go 파일 변경 0.
