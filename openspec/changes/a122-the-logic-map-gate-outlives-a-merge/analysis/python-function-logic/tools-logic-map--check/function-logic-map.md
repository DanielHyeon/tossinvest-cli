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

---

## 두 번째 편집 — 해소 순서 (task 3.2.3.1, 2026-09-10)

열거 파일이 셋이 됐다. **세 파일은 사슬이다** — 가운데 파일의 `source_sha256` 이
`52c761de` 의 `check_analysis.py` 해시와 같다는 것을 대조해 확인했다(`b6ca421654cf`).

| 파일 | 리비전 | 분기 | 반환 |
|---|---|---|---|
| `ast.json` | 3.2.4 편집 **전** HEAD | 34 | 9 |
| `ast.worktree.json` | 3.2.4 편집 **후** = `52c761de` = 3.2.3.1 편집 **전** | 35 | 10 |
| `ast.worktree-after-3.2.3.1.json` | 3.2.3.1 편집 **후** | 37 | 11 |

이번 편집은 판정을 하나도 안 바꾸고 **호출 순서**만 바꾼다. 그것이 기계 열거로
보이는 자리는 `calls` 의 줄 번호다.

| | `resolve_landing` | 빌린 증거 해소(`resolve_referenced_change`+`resolve_base`) |
|---|---|---|
| 편집 전 | `L728` | `L742-743` — **뒤** |
| 편집 후 | `L764` | `L754-755` — **앞** |

**왜 순서가 판정이 되는가.** `resolve_landing` 은 넘겨받은 `analysis` 로 착지를
고정한다. 빌린 증거를 쓰는 change 는 그 시점의 `analysis` 가 **지역** 디렉터리이고
거기 번들이 0 이다(a073 이 a072 의 231개를 빌린다). 그래서 뒤에서 풀면 고정할 것이
하나도 없는 채로 판정이 끝난다 — 3.2.3.1 의 새 거절이 그 change 를 통째로 죽인다.
순서를 바꾸면 빌린 번들이 고정한다.

분기 35→37, 반환 10→11 은 `try` 하나를 둘로 가른 결과다(`resolve_base` 의 실패와
착지·diff 의 실패를 각자 받는다). 두 handler 의 문구는 **일부러 같게** 뒀다 —
바깥에서 본 실패 종류가 갈리지 않아야 오늘의 판정이 그대로 남는다.
