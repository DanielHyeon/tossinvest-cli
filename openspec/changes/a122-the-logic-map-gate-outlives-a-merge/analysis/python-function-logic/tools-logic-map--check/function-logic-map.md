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

---

## 세 번째 편집 — 창의 끝 공유 (task 1.8, 2026-09-10)

열거 파일이 다섯이 됐다. `ast.before-1.8.json` 은 3.3 커밋(`73cc6bdc`)에서 뽑았다.
3.3 이 `check` 에서 바꾼 것은 반환 **값** 하나(실패 메시지)뿐이라 구조는 그대로다 —
`ast.worktree-after-3.2.3.1.json` 과 분기 37 · 반환 11 이 같고 분기 **종류 열도
글자 그대로 같다**. 값만 갈렸다는 것을 열거로 확인했다.

| 파일 | 리비전 | 분기 | 반환 |
|---|---|---|---|
| `ast.worktree-after-3.2.3.1.json` | 3.2.3.1 편집 후 | 37 | 11 |
| `ast.before-1.8.json` | `73cc6bdc` = 1.8 편집 **전** | 37 | 11 |
| `ast.after-1.8.json` | 1.8 편집 **후** | **40** | **13** |

새로 선 노드는 셋이고 전부 빌린 증거 블록 안이다:

| 새 ID | 줄 | 소스 |
|---|---|---|
| B17 | 793 | `try:` (양쪽 선언을 같은 방법으로 읽는다) |
| B18 | 795 | `except ValueError as exc:` → `cannot derive modified Go functions: …` |
| B19 | 797 | `if not shared:` → `must share the exact landing point` |

반환 둘이 늘었다(11 → 13): `L796` 은 기존 문구를 **그대로** 쓴다(빌리는 쪽의
UTF-8 오류가 오늘과 같은 말을 하게), `L798` 이 새 사유다.

**자리가 왜 여기인가.** 이 판정은 `L799` 의 `analysis = referenced_dir / …` **앞**,
그리고 `L803` 의 `resolve_landing` **앞**에 선다. 뒤에 두면 저자가 고른 착지가 이미
고정 판정을 통과한 뒤라 늦다. 3.2.3.1 이 해소 순서를 앞으로 당겨 둔 덕에 이 판정이
설 자리가 이미 있었다.

**대신 이 자리는 3.2.3.1 의 고정 판정을 가릴 수 있다.** 공유 규칙이 먼저 걸리면
`resolve_landing` 이 아예 안 불린다. 그래서 3.2.3.1 회귀 시험의 픽스처를 "양쪽이
같은 값을 선언"으로 고쳐 고정 판정이 홀로 판정하게 두었고, 변이 M4(해시 대조 삭제)가
그 시험을 여전히 빨갛게 하는지 **재서** 확인했다 — 2건 FAIL.

`resolve_referenced_change` 의 호출 자리는 여전히 **둘**(`L756` 게이트 대상 ·
`L781` 빌린 쪽)이고 이 편집이 닿는 것은 뒤의 하나다. 앞의 자리는 `AmbiguousChange`
와 `ValueError` 를 타입으로 갈라 받는 3.2.4 의 모양 그대로다.

---

## 네 번째 편집 — 이관 예외의 창 끝 (task 1.12, 2026-09-11)

`ast.after-1.8.json`(= 커밋 `52d5cb2c` 의 상태, `source_sha256` 을 커밋된 파일 해시와
대조해 확인) → `ast.after-1.12.json`. **분기 40 → 43, 반환 13 → 14.**

| 새 ID | 줄 | 소스 | 무엇을 하나 |
|---|---|---|---|
| B5 | 785 | `{} if context is None else context` | 문맥을 **항상** 채운다 |
| B21·B22 | 820 | `if adopted and _declared_landing(…) is not None:` | 이관의 두 번째 손잡이 거절 |
| B24 | 836 | `str(facts.get('adoption_source','')) if adopted else resolve_landing(…)` | 이관이면 감사된 값, 아니면 해소 |

사라진 것 하나: `if context is not None:`(옛 B22). 문맥 기록이 조건부가 아니게 됐다.
그것이 **판정**이었다는 것이 이 편집의 발견이다 — 호출자가 문맥을 안 주면 `check` 가
"이것이 이관인가"를 스스로 모르게 되고, 같은 입력에 다른 문장을 만든다. 변이 N5
(`facts` 를 `context` 로 되돌림)가 시험 하나를 빨갛게 한다.

반환 하나가 늘었다(13 → 14): `L825` 의 이관 거절. `L848` 은 새 반환이 아니라
`missing Function Logic Map …` 의 **문자열이 바뀐** 것이다(`_target_text` 에 감사
여부를 넘긴다).

**`resolve_landing` 을 이관 경로에서 부르지 않는다.** 그 함수의 판정 전부는 **저자가
고른 값**을 위한 것이고 감사된 source 는 고른 값이 아니다. `validate` 가
`ancestry(P,E)` · `ancestry(E,source)` · `ancestry(source,head,strict=True)` ·
`source^{tree}` 대조 · digest 셋으로 이미 묶는다 — 같은 판정을 두 번 하지 않는다
([[two-judgements-cover-for-each-other]]).


---

## 편집 계획 — task 6.2 · 6.2.1 (코드보다 먼저, `ast.before-6.2.json` = HEAD `fb4e8f92`)

분기 43 · 반환 14 는 **그대로** 여야 한다. 조건의 피연산자와 호출 인자만 바뀐다.

| 자리 | 전 | 후 | task |
|---|---|---|---|
| L865 `resolve_base(change_dir, root, facts)` | 신원 = 디렉터리 이름 | `change_id=change` | 6.2 |
| L879 `resolve_base(referenced_dir, root)` | 같음 | `change_id=referenced_change` | 6.2 |
| B21·B22 L898 | `_declared_landing(...) is not None` — 해독, try 밖 | `_landing_record(...) is not None` — 해독 안 함 | 6.2.1 |

B21·B22 가 기전이다: 이관 probe 는 "착지 기록이 **있는가**"를 묻는데 해독하는 함수를
불렀고, 다섯 호출 자리 중 이것만 try 밖이었다. 비-UTF-8 기록이면 `ValueError` 가 `check()`
를 뚫었다(4.4 Testing 전문가 실측).


## 편집 결과 — task 6.2 · 6.2.1 (`ast.after-6.2.json`)

분기 43 → 43 · 반환 14 → 14. 열거 diff 는 B21·B22 의 callee 하나(`_declared_landing` →
`_landing_record`)뿐이다. `resolve_base` 두 호출은 인자만 늘었다. 변이 M7(probe 가 다시 해독)
CAUGHT — `test_an_undecodable_landing_record_in_the_adoption_path_is_refused_not_raised` 가
`ValueError` 로 **에러**가 된다(편집 전의 증상 그대로).
