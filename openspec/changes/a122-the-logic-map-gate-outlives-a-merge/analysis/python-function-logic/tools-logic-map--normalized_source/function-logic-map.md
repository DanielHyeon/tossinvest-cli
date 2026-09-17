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
