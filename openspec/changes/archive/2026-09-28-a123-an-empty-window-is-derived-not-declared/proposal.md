# a123 · 빈 창은 저자가 아니라 역사가 정한다

- **Feature**: `FEAT-TOS-001` — StockOS SDD toolchain adaptation
- **Story**: `STORY-TOS-a123`
- **Spec**: `sdd-workflow` (ADDED 1)
- **위험 등급**: Normal — 게이트 도구 `tools/logic-map/check_analysis.py` 만, 거래 경로 무관.
  단 **게이트가 받는 것을 늘리는** 편집이므로 fail-closed 명세와 위조 픽스처가 먼저다(design).
- **선행**: a122 착지·아카이브(같은 파일). 그 전에는 `tools/logic-map/` 을 만지지 않는다.
- 작성 2026-09-25, Manager(Fable, tossos-d6). 사용자 승인 2026-09-25(결정 ①).

## Why — 오늘 잰 것

완료 게이트 5단계는 `base-commit.txt` → 대상(착지 커밋 또는 워킹트리) 창에서 바뀐 **모든** 기존 Go 함수의
Function Logic Map 을 요구한다. a122 가 착지 기록으로 창을 좁혔지만, 착지는 `revision: current` 번들이
base 와 **다른** 소스를 고정할 때만 유도된다(a122 7.2.2 K2 — 사람이 "기계로 막는다"를 골랐다).
그 결과 두 모양의 change 는 창을 좁힐 수 없고 남의 함수를 요구받는다
(격리 워크트리 `54004f44`, 2026-09-25 측정; a071 `review.md` 「5.2 실측」이 같은 벽을 374 로 적었다):

| 모양 | change | 5단계 요구 | 사유(도구 문장) |
|---|---|---:|---|
| A `current` 번들 0 | align-full-sdd-pm-contract | 546 | `no revision: current evidence pins a landing` |
| A | a067-add-kr-us-continuation-lanes | 338 | 같음(391 커밋 창) |
| A | a068-add-kr-us-reversal-lanes | 338 | 같음 |
| B base 가 작업 뒤(`840b3377` 재기준화) | a077-screens-show-what-they-already-know | 319 중 318 결여 | `landing point … changes none of the sources its evidence pins since the comparison base` |
| B | a079-operator-can-lift-a-quarantine | 319 중 317 결여 | 같음 |
| B | a074 · a089 · a091 · a092 · a094 · a095 (a122 5.7 census) | — | 같은 모양 |

a122 tasks 5.6·5.7 은 이것을 **한계로 선언**했고 닫는 방향만 적었다: "요구 집합이 빈 창을 저자가
**고르지 못하게** 하는 방법(예: 그 조건이 성립하는 가장 늦은 커밋으로 유도)", "freeze 시점의 base 가 그
change 의 Go 작업 앞에 있는지 물을 방법". 이 change 는 그 둘을 **역사에서 유도**한다 — 저자가 고르는 입력은
하나도 늘지 않는다.

**버린 안 — change 마다 사람 재기준화.** 모양 B 는 base 를 작업 **앞**으로 되돌려야 하는데 그러면 그 사이
남의 커밋이 전부 창에 들어와 오늘과 같은 수백 함수를 요구받는다. 도구 없이는 안 풀린다.
align 만은 R2 로도 안 풀리므로(아래) 사람 재기준화 한 건을 따로 한다(선례 `840b3377`).

## What Changes

**R1 (모양 B) — 재기준화된 change 의 빈 창을 유도한다.** 다음이 **전부** 역사에서 참이면 5단계는 창의 끝을
base 로 유도하고(요구 0), 번들 유효성 검사는 계속하며, 출력에 "유도된 빈 창 — 재기준화 커밋 `<sha>`" 를
적는다(SHALL):

1. 그 change 의 `revision: current` 번들 전부의 `source_sha256` 이 base 소스와 같다(a122 가 이미 재는 조건).
2. base 뒤에 그 change 의 자기 Go 작업 커밋이 **0** 이다 — a122 규칙 8 의 귀속 함수 `_self_repair_commits`
   ("그 change 디렉터리를 만지면서 Go 파일도 고친 비병합 커밋").
3. `base-commit.txt` 의 현재 값을 쓴 **비병합 커밋이 역사에 있고** 그 커밋이 그 change 의 번들(`ast.json`)도
   같이 고쳤다 — 즉 재기준화가 증거를 base 로 다시 뽑은 사건이다(`840b3377` 이 정확히 이 모양:
   base-commit.txt + ast.json 다수).

저자는 셋 중 어느 것도 게이트 시점에 고를 수 없다. 5.6 이 걱정한 "증거가 없어서 0" 은 1 에서 갈린다
(번들이 있어야 한다).

**R2 (모양 A) — 번들 0 인 change 의 요구 집합은 자기 커밋이 고친 기존 함수다.** `revision: current` 번들이
하나도 없으면 5단계는 창 전체가 아니라 `_self_repair_commits` 가 귀속한 커밋들이 고친 기존 Go 함수만
요구한다(SHALL). 그 집합이 비면 "자기 Go 작업 0" 을 출력하고 통과, 비지 않으면 그 함수들의 번들을 요구한다 —
a067·a068 은 남의 338 대신 **자기 함수 N**(레인 등록·평가기 등)을 요구받고, 그것은 저자가 원래 만들었어야 할
증거다. align 은 R2 로도 안 풀린다: 귀속 커밋 `c0619279`(콘솔 feature, Go 52 파일)가 align 디렉터리를 같이
만졌다 — 디렉터리 귀속은 남의 작업을 **더** 요구하는 쪽으로 틀리므로 안전하고, align 의 답은 사람 재기준화다.

**R3 — 출력.** 두 유도 모두 판정 줄에 근거(재기준화 sha · 귀속 커밋 목록 · 요구 수)를 적는다(SHALL).
조용한 0 은 면제와 구분되지 않으므로 금지한다.

## Non-goals

- `current` 번들이 있고 base 와 다른 change 의 착지 규칙(K2 · 규칙 1~8)은 **바꾸지 않는다.**
- 저자가 값을 적는 새 파일·플래그를 두지 않는다. `tools/gate.sh` 무변경.
- 귀속의 알려진 구멍(change 디렉터리를 안 만진 Go 커밋)은 규칙 8 과 같은 수준으로 **그대로 둔다** — 이
  change 가 새로 여는 문이 아니다. 픽스처로 못 박되 고치지 않는다.
- 이미 아카이브된 change 를 소급해 다시 판정하는 것.

## 열린 결정 (freeze 리뷰가 답할 것)

- R1 조건 3 의 "번들도 같이 고쳤다" 를 요구할지: 없애면 base-commit.txt 만 옮긴 커밋으로 빈 창을 만들 수
  있다(저자 행위). 두면 재기준화 관례(ast 재추출)를 강제한다 — **보수 쪽은 "둔다"** (design D1).
- R2 에서 귀속 커밋이 **병합**뿐인 change(하한 없음)는 오늘처럼 창 전체를 요구한다(a122 규칙 4 와 같은 한계).

## Impact

- `tools/logic-map/check_analysis.py` (`_landing_refusal` · `compute_landing` · `_self_repair_commits` · `check` ·
  `changed_existing_functions`) 와 `test_check_analysis.py`, `README.md`.
- `sdd-workflow` 스펙 ADDED 1. 런타임 거래·위험·원장·엔진 기동 의미는 건드리지 않는다.
- 수용 측정 대상: a077 · a079 · a074(R1) / a067 · a068(R2) / align(사유 출력만).
