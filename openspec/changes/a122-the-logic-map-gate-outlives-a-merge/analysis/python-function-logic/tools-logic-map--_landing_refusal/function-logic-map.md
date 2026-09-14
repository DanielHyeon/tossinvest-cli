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

## 편집 — task 7.2.2 (착지는 고정 소스를 바꾼 커밋이어야 한다) · 분기 6 → 10 · 반환 7 → 8 · raise 0 → 0

편집 전 `ast.before-7.2.2.json`(HEAD `9692b8d1` blob, L584-652) · 편집 후 `ast.after-7.2.2.json`(워킹트리, L584-677). 표는 두 열거의 `source` 를 스크립트가 대조한 것이다.

맨 뒤(미보유 판정 뒤, 수락 반환 앞)에 조건 하나를 더했다: 고정 번들의 소스(`_pinning_bundles` 의 source 집합) 중 **하나 이상**이 base 와 이 후보에서 바이트가 달라야 한다(`_committed_bytes` 두 번 — `all(...)` 이 첫 바뀐 소스에서 멈춘다). 전부 같으면 `landing point … changes none of the sources its evidence pins since the comparison base …: <소스 셋까지> — evidence that still describes the base cannot tell where this change's Go work landed` 를 둘째 값 `[]` 과 함께 돌려준다 — 미보유가 아니므로 계산 경로의 `unheld` 에 안 섞인다. 사람이 2026-09-14 에 고른 규칙(리뷰 C1 · C2)이고, 편집 전 전수에서 착지 76 중 8 이 잃는 것을 알고 골랐다. 맨 뒤인 이유는 앞 가드들이 자기 문장으로 못 박혀 있기 때문이다 — 하한 판정 앞으로 옮기는 변이 K-D 를 다섯 시험이 잡는다. 고정 번들 0 은 둘째 판정이 먼저 돌려보내므로 `all` 이 빈 표본 위에서 참이 되는 자리는 없다. 맨 위 주석("base 와 같아도 된다 — a074·a079·a075 의 정답")은 틀린 말이 되어 고쳤다(AST 동일 확인).

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | ` for _, source, _ in _pinning_bundles(root, analysis)` |
| 생김 | 분기 | `if all((_committed_bytes(root, base, source) == _committed_bytes(root, candidate, source) for source in sources)):` |
| 생김 | 분기 | ` for source in sources` |
| 생김 | 분기 | `f', and {len(sources) - 3} more' if len(sources) > 3 else ''` |
| 생김 | 반환 | `(f"landing point {candidate[:12]} changes none of the sources its evidence pins since the comparison base {base[:12]}: {listed} — evidence that still describes the base c` |

호출 — 사라짐 없음 · 생김 ['_committed_bytes', '_pinning_bundles', 'all', 'len']
