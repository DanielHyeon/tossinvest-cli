# Function Logic Map: `_measure_landing_inputs`

`tools/logic-map/check_analysis.py:1520-1531` · 분기 1 · 반환 1 · raise 0 · **task 7.5.1** (`ast.after-7.5.1.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** `enumerate.py` 의 열거를 옮긴 것이다. 7.5.1 이 새로 만든 함수다.

## Inputs and invariants

착지 판정의 입력을 **한 번** 잰다 — `(하한, 사유, 고정 번들, 수리 신호, 지문)`. 선언 경로(`resolve_landing`)와 계산 경로(`compute_landing`)가 같은 한 벌을 쓴다 (서브에이전트 F2 — 7.5 의 "한 번 잰다" 주석은 거짓이었고 실측 `_evidence_floor`·`_self_repair_commits` 각 두 번이었다).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1530 | IfExp | `_self_repair_commits(root, analysis) if not why else []` |

| 반환 줄 | 값 |
|---|---|
| 1531 | `(floor, why, bundles, repairs, fingerprint)` |

## Calls and live bindings

`_pinning_bundles` · `_walk_floor` · `_self_repair_commits`.

## State mutations and fallbacks

없다.

## Safety conclusion

지문을 **먼저** 잰다. 하한 뒤에 재면 그 사이에 생긴 번들이 지문에는 있고 판정 목록에는 없어서, 수락 직전 재확인이 "그대로" 라고 답한다(변이 V10 이 이 순서를 잰다). 사유가 있으면 수리 신호를 안 잰다 — 그 경우 `_landing_refusal` 이 수리 가드(8번)에 닿기 전에 가드 2 나 4 에서 거절하므로 쓰이지 않는다.

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1741-1753` · 분기 1 · 반환 1 · raise 0 (편집 전 L1520-1531 · 분기 1 · 반환 1 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

한 번 읽기: `_evidence_fingerprint` 가 돌려준 목록을 `_walk_floor` 에 넘긴다(목록을 두 번 안 읽는다). `LandingInputs` 는 `NamedTuple`(위치 튜플은 `floor`·`why` 둘 다 `str` 이라 바꿔 넣어도 아무도 못 봤다). 하한이 못 서면 수리 신호는 `None`(잰 적 없음) — `[]`(없음)과 섞지 않는다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1752 | IfExp | `None if why else _self_repair_commits(root, analysis)` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1817-1828` · 분기 1 · 반환 1 · raise 0 (편집 전 L1741-1753 · 분기 1 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`(root, head, evidence)` — 증거를 스스로 읽지 않는다. 여기서 읽으면 착지가 판정한 바이트와 대상 판정이 읽는 바이트가 다른 읽기가 된다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1827 | IfExp | `None if why else _self_repair_commits(root, evidence.directory, head)` |
