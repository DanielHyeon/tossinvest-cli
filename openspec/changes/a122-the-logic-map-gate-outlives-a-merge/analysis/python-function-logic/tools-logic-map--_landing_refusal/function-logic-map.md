# Function Logic Map: `_landing_refusal`

`tools/logic-map/check_analysis.py:581-649` · Python · **task 7.6 (리뷰 I2)** ·
분기 6 · 반환 7 · raise 0 (`ast.after-7.6.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 스크립트가 옮긴 것이다.
> 이 함수는 7.6 이 새로 만들었다 — 편집 전 판본은 `resolve_landing/ast.before-7.6.json` 의
> raise L614·L623·L628·L640·L645·L655 와 `compute_landing/ast.before-7.6.json` 의 B7·B8·B10 이다.

## Inputs and invariants

`candidate` 하나를 착지로 받을지 판정한다. `floor` 는 호출자가 **한 번** 재서 넘긴다.
돌려주는 것은 `(거절 사유, 미보유 번들 이름)` — 받으면 `("", [])`. 예외를 만들지 않는다:
선언 경로는 사유를 `ValueError` 로 올리고 계산 경로는 다음 후보로 넘어간다. 그 차이는
**호출자**의 몫이고, 규칙은 여기 한 곳이다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L605 | If | `if not _is_ancestor(root, base, candidate):` |
| B2 | L608 | If | `if not pinning:` |
| B3 | L617 | If | `if mismatched:` |
| B4 | L628 | If | `if not floor:` |
| B5 | L633 | If | `if not _is_ancestor(root, floor, candidate):` |
| B6 | L643 | If | `if unheld:` |

| 반환 줄 | 값 |
|---|---|
| L606 | `(f'landing point precedes the comparison base {base[:12]}: {candidate}', [])` |
| L613 | `(f'landing point {candidate[:12]} is pinned by no `revision: current` evidence: a declared landing m` |
| L618 | `(f'landing point {candidate[:12]} is not the revision this evidence describes: ' + ', '.join(sorted(` |
| L629 | `(f'landing point {candidate[:12]} is pinned by evidence that never entered this history: commit the ` |
| L634 | `(f"landing point {candidate[:12]} precedes the evidence that pins it ({floor[:12]}): a declared land` |
| L644 | `(f'landing point {candidate[:12]} does not hold the evidence this verdict read: ' + ', '.join(unheld` |
| L649 | `('', [])` |

**순서가 거절 지점이다.** base → 고정 0 → 불일치 → 하한 없음 → 하한 순서 → 미보유. 이것은
편집 전 `resolve_landing` 의 raise 순서와 같다(열거 대조). `compute_landing` 의 옛 순서(싼 조상
판정 먼저)를 쓰지 않은 이유는 review.md `## Pre-Edit Gate — task 7.6` 의 I2 절이다.

## Calls and live bindings

`_is_ancestor` ×2 · `_pinning_at` · `_unheld_bundles` · `sorted`/`set`/`join`. git 을 부르는 것은
앞의 셋이다(`merge-base --is-ancestor` · 번들마다 `git show`). 시각·환경을 안 읽는다.

## State mutations and fallbacks

없다. 디스크·git 을 **읽기만** 한다. fallback 도 없다 — 판정 못 하는 입력(`_pinning_at` 의
`ValueError`, git 의 `TimeoutExpired`)은 그대로 올라가 7.4 의 경계(`GATE_FAULTS`)가 받는다.

## Safety conclusion

생산 Go 코드 변경 0. 이 함수는 판정을 **옮겼을 뿐** 새 조건을 만들지 않았다 — 여섯 조건의
문장은 편집 전 `resolve_landing` 의 raise 문장과 글자 단위로 같다(열거의 `source` 대조).
