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

## 편집 — task 7.2.4 (병합 안에서만 들어온 증거의 한계를 사실대로 말한다) · 분기 2 → 2 (소스 동일) · 반환 3 → 3 · raise 0 → 0

편집 전 `ast.before-7.2.4.json`(HEAD `f9811236` blob) · 편집 후 `ast.after-7.2.4.json`(워킹트리, L1179-1198).

하한 없음 사유 문장을 "the pinning evidence never entered this history (commit the bundles)" → "no ordinary commit on this history adds the pinning evidence — commit the bundles in an ordinary commit (a merge commit's own changes are not read)" 로. 리뷰 H7: 병합을 마치며 번들을 처음 커밋하면 `_evidence_floor` 의 `git log`(`-m` 없음)가 그 병합의 변경을 안 읽어 여기로 오는데, 옛 문장은 번들이 커밋돼 있는데도 "역사에 들어온 적 없다, 커밋하라"고 했다. 동작(하한 없음 → 기록 안 함)은 사람이 2026-09-14 에 한계로 두기로 골랐다. 분기 · 반환 수 · 호출 불변(열거 대조). 한계 자체를 없애는 변이 H-C(`git log -m`)는 `EvidenceFirstCommittedInsideAMergeIsAKnownLimit` 가 잡는다 — 한계가 **선택**임을 못 박는다.

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 반환 | `('', 'the pinning evidence never entered this history (commit the bundles)')` |
| 생김 | 반환 | `('', "no ordinary commit on this history adds the pinning evidence — commit the bundles in an ordinary commit (a merge commit's own changes are not read)")` |

## task 7.5 — 번들 목록을 한 번 재서 넘긴다

`ast.before-7.5.json`(= `8091e6c4` 의 소스)과 `ast.after-7.5.json` 을 같은 열거기로 뽑아
**순서 있는 배열로** 대조했다: 분기 순서열 2개 중 **1개만 다르고** 그것은 번들 목록을 어디서 받는지다(`_pinning_bundles(root, analysis)` → `bundles`). **판정 조건과 그 순서는 바이트 동일** · 반환 순서열은 **바뀌었다** — 계약이 3-튜플에서 3-튜플로 넓어졌다(`(floor, '', bundles)`). 수만 보면 안 보인다.

> **처음에는 개수로 적었다가 정정했다** (2026-09-18 독립 리뷰 P2). `분기 N · 반환 M` 은
> multiset 크기라 **순열에 불변**이다 — 가드를 서로 바꿔도 같은 수가 나오므로 "순서가 안
> 움직였다"를 못 받친다. 순서 있는 배열은 이미 같은 JSON 안에 있었고, 그것을 인용해야 한다
> ([[a-new-guard-unpins-the-guards-behind-it]] 가 지키려는 것이 바로 순서다).

바뀐 것은 고정 번들 목록이 **어디서 오는가** 하나다. 후보마다 디렉터리를 다시 순회하던
것을 호출자가 한 번 재서 넘긴다 — `floor` 와 `repairs` 가 이미 그렇게 넘어오고 있었고
(7.2.6 · 7.6), 7.5 는 셋째를 같은 방식으로 묶었다. 번들은 걷는 동안 안 변한다: 이 도구는
번들을 **읽기만** 한다.

값: 2026-09-18 프로파일에서 `_pinning_bundles` 가 a071 walk 하나에 341회 돌아
13.07s 중 9.30s 였다. 묶은 뒤 a071 이 10.97s → **3.69s**.

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1756-1783` · 분기 3 · 반환 3 · raise 0 (편집 전 L1534-1559 · 분기 2 · 반환 3 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

이미 읽은 고정 목록을 받을 수 있다(`bundles`, 안 주면 스스로 읽는다 — `_recording_refusal`). 반환이 `(하한, 사유)` 둘로 줄었다: 셋째(목록)는 이제 호출자가 넘긴 것이다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1769 | If | `if bundles is None:` |
| B2 | 1771 | If | `if not bundles:` |
| B3 | 1774 | If | `if not floor:` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1831-1853` · 분기 2 · 반환 3 · raise 0 (편집 전 L1756-1783 · 분기 3 · 반환 3 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`(root, bundles, head)` — 목록이 필수(스스로 읽는 갈래 삭제, 분기 3 → 2).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1841 | If | `if not bundles:` |
| B2 | 1844 | If | `if not floor:` |
