# Function Logic Map: `_walk_floor`

`tools/logic-map/check_analysis.py:1147-1159` · Python · **task 7.7 (리뷰 I8)** ·
분기 2 · 반환 3 · raise 0 (`ast.after-7.7.json`)

## Inputs and invariants

`root` 와 착지를 고정할 증거 디렉터리 `analysis`. 후보를 **걷기 전에** 정해지는 것만 답한다:
순회가 설 하한(`_evidence_floor`)과, 하한이 설 수 없으면 그 사유. base 를 받지 않는다 — 하한은
증거의 함수이지 비교 기준의 함수가 아니다.

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L1154 | If | `if not _pinning_bundles(root, analysis):` |
| B2 | L1157 | If | `if not floor:` |

| 반환 줄 | 값 |
|---|---|
| L1155 | `('', 'no `revision: current` evidence pins a landing for this change')` |
| L1158 | `('', 'the pinning evidence never entered this history (commit the bundles)')` |
| L1159 | `(floor, '')` |

## Calls and live bindings

`_pinning_bundles` · `_evidence_floor`. 부르는 쪽 둘: `compute_landing`(순회 전 조기 종료)과
`_recording_refusal`(기록 명령·조언 줄의 걷기 전 판정). 편집 전에는 이 두 조건이 `compute_landing`
안에만 있었고, 조언 줄은 아무것도 묻지 않았다.

## State mutations and fallbacks

없다. `_pinning_bundles` 가 저장소 밖을 가리키는 번들에서 `ValueError` 를 낸다 — 여기서 받지 않고
호출자의 결함 경계(`record_landing` 의 `GATE_FAULTS`, `main` 조언 갈래의 `GATE_FAULTS`)로 올린다.

## Safety conclusion

두 조건과 두 문장은 `compute_landing` 에 있던 것을 글자 그대로 옮겼다(편집 절 표: `compute_landing/function-logic-map.md`).
판정이 받는 후보는 안 바뀐다 — 받는 규칙은 `_landing_refusal` 에 그대로 있다. 바뀐 것은 이 두 사유를 묻는
자리의 수(1 → 2)이고, 둘은 **같은 함수**를 부른다. 두 갈래를 지우는 변이(R17·R18)가 계산 경로와 조언 경로
양쪽의 시험에서 빨개지는지는 review.md `## VERIFY — task 7.7` 에 있다.
