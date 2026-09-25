# a123 design — 유도의 조건과 검증 순서

## Context

a122 가 `landed-commit.txt` 로 5단계의 창을 좁혔고, 착지가 없는 change 는 워킹트리까지를 창으로 둔다.
착지 규칙 K2 는 `revision: current` 번들이 base 와 다른 소스를 적어야 한다고 요구한다 — 그래서 base 가 작업
뒤로 옮겨진 change(모양 B)와 번들이 없는 change(모양 A)는 착지를 얻지 못한다. proposal 「Why」의 표가
2026-09-25 측정이다.

## D1 — R1 은 세 조건 전부, 조건 3 은 둔다

세 조건은 각각 다른 저자 행위를 막는다.

| 조건 | 없으면 뚫리는 것 |
|---|---|
| 1 번들 전부 base 소스 | 번들이 없어도 빈 창(5.6 이 걱정한 "증거가 없어서 0") |
| 2 base 뒤 자기 Go 커밋 0 | 작업 뒤에 base 를 옮기고 그 뒤 리뷰 수리를 창 밖에 두기 |
| 3 base 를 옮긴 비병합 커밋이 번들도 재추출 | `base-commit.txt` 한 줄 커밋으로 빈 창 만들기 |

조건 3 은 재기준화 관례(ast 재추출)를 강제한다. 그것은 이미 `840b3377` 의 모양이고, 그 뒤의 재기준화도 같은
모양이어야 한다는 것은 새 부담이 아니라 기존 관례의 기계화다.

## D2 — R2 의 귀속은 규칙 8 과 같은 함수를 쓴다

새 귀속 규칙을 만들지 않는다. `_self_repair_commits` 는 "그 change 디렉터리를 만지면서 Go 파일도 고친 비병합
커밋"이고, 그 구멍(디렉터리를 안 만진 Go 커밋)은 a122 가 한계로 열거했다. R2 는 그 구멍을 **같은 방향**으로
가진다 — 놓치는 쪽은 요구를 줄이므로, 픽스처 (c) 로 못 박고 고치지 않는다(비목표).

## D3 — 출력은 사유를 이름으로 말한다

a122 의 거절 문장 관례를 따른다. 유도된 빈 창은 `derived empty window — rebaseline <sha> re-extracted <n> bundles
at base`, R2 는 `no current bundles — required set from <k> attributed commits: <n> functions` 모양이다.
정확한 문구는 골든 픽스처에 얼린다.

## 검증 순서 (tasks 1.x)

1. **census 먼저(RED)**: a122 `722_census.py` 방식으로 R1·R2 를 **사본**에 넣고 활성+아카이브 전수에 대어
   "같음/획득/상실" 표를 만든다 — 상실 0 이어야 한다(더 받기만 하는 규칙). 획득 목록이 proposal 표와 일치해야 한다.
2. **위조 픽스처**: (a) base 만 옮긴 커밋(조건 3 위반) 거절 · (b) 번들 하나를 base 와 다르게 위조 → 오늘의 K2
   경로로 감 · (c) 디렉터리를 안 만진 Go 커밋 → 요구 누락(알려진 구멍, 문서화·시험으로 **못 박되 고치지 않음**) ·
   (d) 병합으로만 들어온 번들 → 창 전체. 병합 픽스처가 최소 하나 있어야 한다(기억: 선형 픽스처는 DAG 가드를
   안 건드린다).
3. **뮤테이션**: 조건 1·2·3 각각을 빼면 픽스처가 빨갛다. 변이는 사본 대상 · 무변이 대조군 선행 · `GOFLAGS=-trimpath`
   + 전용 `GOCACHE`(a122 하네스 관례).
4. **수용 측정**: a077 · a079 · a074 게이트 5단계 PASS, a067 · a068 은 "자기 함수 N 요구" 로 바뀜(N 실측),
   align 은 그대로(사유 출력).

## Python FLM — 편집 전 필수 (a122 관례)

`_landing_refusal` · `compute_landing` · `_self_repair_commits` · `check` · `changed_existing_functions` — a122 의
`analysis/python-function-logic/enumerate.py` 로 편집 전 AST 를 먼저 뽑는다(tasks 0.3). 기억: Python 내부 편집은
게이트가 안 요구해서 a122 에서 세 task 연속 조용히 빠졌다.

## Risks

- 이 편집은 게이트가 **받는** 것을 늘린다. 그래서 census 의 "상실 0" 과 위조 픽스처가 GREEN 보다 먼저다.
- 같은 파일을 a122(tossos-42)가 편집 중이다. a122 아카이브 전에는 착수하지 않는다.
