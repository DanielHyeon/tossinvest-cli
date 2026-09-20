# Function Logic Map: `normalized_source`

`tools/logic-map/check_analysis.py:969-980` · Python · **task 7.5 (리뷰 I1)** ·
분기 **2 → 2** · 반환 1 · raise 1 (`ast.before-7.5.json` · `ast.after-7.5.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** `enumerate.py` 의 열거를 옮긴 것이다.

## Inputs and invariants

번들이 적은 `file` 값을 저장소 상대경로로 바꾼다. 절대경로도 받는다(a063 이관 픽스처가
그렇게 만들었다). 저장소 **밖**으로 나가면 `ValueError`.

## Branches — 편집 전후 (소스 줄로 짝지음)

| id | 종류 | before | after |
|---|---|---|---|
| B1 | IfExp | `raw if raw.is_absolute() else root / raw` | 같음 |
| B2 | If | `if not resolved.is_relative_to(root.resolve()):` | `if not resolved.is_relative_to(anchor):` |

반환 1개 · raise 1개 — **개수도 조건도 안 바뀐다.** 바뀐 것은 `root.resolve()` 를
한 번 재서 두 자리가 같이 쓴다는 것뿐이다(before 는 B2 와 반환 줄에서 각각 한 번).

## Calls and live bindings

`Path` · `root.resolve` (before ×2 → after ×1) · `path.resolve` ×1 ·
`resolved.is_relative_to` · `resolved.relative_to(...).as_posix`.

## State mutations and fallbacks

없다.

## Safety conclusion

`root.resolve()` 는 이 호출 **안에서 불변**이다 — 같은 인자로 두 번 불러 다른 답이 나오면
그 사이에 파일시스템이 바뀐 것이고, 그때 before 판본은 **두 답을 섞어** 검사와 계산에
서로 다른 닻을 쓴다. 한 번 재서 같이 쓰는 쪽이 더 엄하다.

값은 2026-09-18 프로파일이 잰 것이다: 이 함수가 `_pinning_bundles` 안에서 호출당 resolve
3회를 돌아 a071 walk 하나에 35,805회 · `lstat` 216,876회 · 7.85s 였다. 중복을 없애고
a071 이 27.84s → **10.97s**.

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1210-1227` · 분기 2 · 반환 1 · raise 1 (편집 전 L1150-1161 · 분기 2 · 반환 1 · raise 1, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`Path.resolve()` 대신 `os.path.realpath`. 둘은 같은 풀이인데 3.12 의 `resolve()` 만 심링크 고리에서 `RuntimeError` 를 **더** 낸다(3.13 부터는 안 낸다) — 판정 쪽은 `ValueError` 만 대상 이름을 붙여 받으므로 고리 하나가 판정 전체를 한 줄로 바꿨다(재리뷰 보안 전문가). 처음엔 `except RuntimeError` 로 바꿔 올리려 했는데, 손 복사 예외 목록을 막는 구조 시험(`test_every_fault_handler_uses_the_one_list`)에 걸렸다 — 판본마다 다른 동작을 없애는 쪽이 근본이다. 고리는 풀리지 않은 채 남고 "AST source is missing" 이 된다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1217 | IfExp | `raw if raw.is_absolute() else root / raw` |
| B2 | 1225 | If | `if not resolved.is_relative_to(anchor):` |

| raise 줄 | 소스 |
|---|---|
| 1226 | `raise ValueError('AST source escapes repository')` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1243-1261` · 분기 2 · 반환 1 · raise 1 (편집 전 L1210-1227 · 분기 2 · 반환 1 · raise 1, `ast.before-7.5.2.2.json` = revision `908a8a36`).

기준점도 `os.path.realpath` — 두 경로를 같은 도구로 풀어야 비교가 같은 규칙이다(재리뷰 maintainability).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1251 | IfExp | `raw if raw.is_absolute() else root / raw` |
| B2 | 1259 | If | `if not resolved.is_relative_to(anchor):` |

| raise 줄 | 소스 |
|---|---|
| 1260 | `raise ValueError('AST source escapes repository')` |
