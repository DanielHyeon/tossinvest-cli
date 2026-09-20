# Function Logic Map: `resolve_landing`

`tools/logic-map/check_analysis.py:345-408`(HEAD `52c761de`) → `:345-424`(편집 후) · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json`(HEAD)과
> `ast.worktree.json`(편집 후)을 `enumerate.py` 가 기계로 열거했고, 아래는 그 열거를
> 옮긴 것이다. Go 용 `extract_go_ast.go` 는 Python 함수에 쓸 수 없으므로 a120 의
> `analysis/python-function-logic/` 선례를 따른다.

## Inputs and invariants

`change_dir` · `root` · `base`(해소된 비교 기준) · `analysis`(그 change 가 딛는 증거
디렉터리)를 받아 착지 커밋 SHA 하나를 돌려준다. 기록이 없으면 빈 문자열이다.

불변식은 `analysis/landing-point.md` 가 적은 그대로다: **저자는 자기 증거가 기술하는
리비전밖에 못 고른다.** 그 문장에는 전제가 있다 — 고정할 번들이 **있어야** 한다.

## Branches and early returns — 편집 전 (ast.json 열거 그대로, 분기 19)

| ID | 줄 | 종류 | 소스 | 무엇을 묻나 |
|---|---|---|---|---|
| B1 | 358 | Try | `try:` (경로가 root 밖) | — |
| B2 | 360 | ExceptHandler | `except ValueError:` | → `return ""` |
| B3 | 363 | If | `if raw is None:` | 기록이 커밋에 없다 → `return ""` |
| B4 | 365 | Try | `try:` (디코드) | — |
| B5 | 367 | ExceptHandler | `except UnicodeDecodeError as exc:` | → raise |
| B6 | 369 | If | `if not FULL_SHA.fullmatch(candidate):` | 40자리 hex 인가 |
| B7·B8 | 380 | BoolOp·If | `process.returncode or process.stdout.strip() != candidate` | 실재하는 커밋인가 |
| B9 | 382 | If | `if not _is_ancestor(root, candidate, "HEAD"):` | 착지한 적 있는 이력인가 |
| B10 | 387 | If | `if not _is_ancestor(root, base, candidate):` | base 보다 뒤인가 |
| B11 | 392 | For | `for ast_path in sorted(analysis.glob('*/ast.json')) …` | **고정 순회** |
| B12 | 392 | IfExp | `… if analysis.is_dir() else ()` | 증거 디렉터리가 있나 |
| B13·B14 | 394 | BoolOp·If | `revision != 'current'` | 이 번들이 착지를 말하나 |
| B15·B16 | 398 | BoolOp·If | `not source or not digest` | 값이 있나 |
| B17·B18 | 401 | BoolOp·If | 해시 대조 | 이 커밋의 내용인가 |
| B19 | 403 | If | `if mismatched:` | 어긋난 것이 있나 |

반환 셋 · raise 여섯: `361 return ''` · `364 return ''` · `368/373/381/383/388/404 raise` ·
`408 return candidate`.

## 결함 (task 3.2.3.1)

**B1~B10 은 전부 "이것이 어느 커밋인가"만 묻는다.** 어느 커밋인지를 *고르지 못하게*
하는 것은 B11 의 순회 하나뿐이다. 그리고 그 순회가 **0회 돌면**

1. `mismatched` 는 빈 채로 남고,
2. B19 가 거짓이 되고,
3. `408 return candidate` 가 **무조건** 참이 된다.

즉 번들이 0 인 change 에서는 착지 선언이 아무것에도 걸리지 않는다. 저자가 구간의
바닥(=base)을 고르면 `changed_existing_functions(base, base)` 가 ∅ 이고, 호출자
`check` 는 `analysis` 가 없으므로 면제 표식 경로로 통과한다. 이것이 3.2.3.1 이
적은 구멍이고, 위 세 줄이 그 기제다.

같은 이유로 **`revision: base` 번들만 있는 change 도 고정하지 못한다.** 그 번들의
`source_sha256` 은 base 를 기술하고 `validate_target` 도 그것을 해싱하지 않는다.

## 편집 — 분기 19 → 20, raise 6 → 7, 반환 3 → 3

`ast.worktree.json` 열거:

| 새 노드 | 줄 | 소스 |
|---|---|---|
| B19 | 410 | `if not pinning:` |
| raise | 415 | `landing point … is pinned by no `revision: current` evidence` |

반환은 **셋 그대로**다(`366 ''` · `369 ''` · `424 candidate`). 이 편집은 경로를
지우지 않고 마지막 반환 **앞에** 조건을 하나 더 세운다 — 그래서 "기록이 없으면
오늘과 똑같다"(task 2.2)는 두 early return 이 그대로 남아 구조로 보장된다.

## Calls and live bindings

`_committed_bytes`(git show) · `_is_ancestor`(git merge-base) · `_ast_value` ·
`FULL_SHA.fullmatch` · `hashlib.sha256` · `subprocess.run(git rev-parse)`.
**쓰기는 하나도 없다.**

## State mutations and fallbacks

없다. 로컬 `mismatched` · `pinning` 만 바꾼다. fallback 도 없다 — 실패는 전부 raise 이고
호출자 `check` 가 `cannot derive modified Go functions: …` 로 돌려준다.

## Safety conclusion

주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 유일한
production 호출자는 사람이 부르는 완료 게이트(`tools/gate.sh:321`)다. 실패 방향은
**게이트가 안 열리는 쪽**이므로 보수적이다. Go 파일 변경 0.

## 편집 (task 1.8) — 분기 20 → 16, 반환 3 → 2, raise 7 → 6

`ast.after-1.8.json` 열거. **줄어든 것은 사라진 것이 아니라 옮겨간 것이다.**
선언을 읽는 앞부분(B1~B5)을 `_declared_landing` 으로 뽑았다 — 빌린 증거의 착지
공유 판정이 **같은 방법으로** 읽어야 비교가 뜻을 갖기 때문이다.

| 이 함수에서 사라진 노드 | 어디로 갔나 |
|---|---|
| `B1 Try` (경로가 root 밖) · `B2 except ValueError` | `_declared_landing` B1·B2 |
| `B3 if raw is None:` | `_declared_landing` B3 |
| `B4 Try` (디코드) · `B5 except UnicodeDecodeError` | `_declared_landing` B4·B5 |
| `raise ValueError('landing point is not UTF-8: …')` | `_declared_landing` 의 유일한 raise |
| `return ''` 둘 중 하나 | `_declared_landing` 의 `return None` 둘 |

| 이 함수에 새로 선 노드 | 줄 | 소스 |
|---|---|---|
| `if candidate is None:` | 391 | 선언 **파일이 없을** 때만 `return ''` |

`None`(파일 없음)과 `''`(빈 선언)을 가르는 것이 이 한 줄이다. 안 가르면 빈 파일을
커밋한 change 가 조용히 워킹트리를 대상으로 삼는다 — 변이 M2 로 쟀고
`test_an_empty_record_is_still_a_declaration` 이 그것을 막는다.

세는 것은 그대로다: 분기 20 − 5(이동) + 1(새) = 16, 반환 3 − 2 + 1 = 2,
raise 7 − 1 = 6. 판정은 하나도 안 없어졌다.

## 편집 (task 1.12) — 분기 16 → 16, 반환 2 → 2, raise 6 → 6

**구조가 하나도 안 바뀐다.** 고정 순회가 번들의 `file` 을 `normalized_source` 로
정규화하는 한 줄이 들어갔고, 그것은 분기가 아니다. 기계 열거에서 보이는 유일한
차이는 `B11·B12` 의 이름이다(`not source or not digest` → `not raw_source or not digest`).

| | 옛 | 새 |
|---|---|---|
| 순회가 `git show` 에 넘기는 경로 | `str(value["file"])` 그대로 | `normalized_source(...)` 의 상대경로 |

**왜 결함인가.** `validate_target` 은 같은 값을 `normalized_source` 로 정규화한다
(`check_analysis.py:454`). 번들의 `file` 은 절대경로일 수 있고, 그러면
`git show <sha>:/abs/path` 가 **언제나** 실패해서 `mismatched` 에 들어간다 — 즉
**정상 입력이 위조로 몰린다**. 3.2.3.1 이 이 순회를 넣을 때 놓친 자리이고,
a063 이관 픽스처의 번들이 절대경로를 쓰면서 드러났다.

저장소 전수(2026-09-11): 번들 `file` 필드 **상대 3048 · 절대 0**. 실물 영향은 0 이라
126건 A/B 에서 안 보인다. 그래서 변이 N4 와
`test_an_absolute_bundle_path_still_pins_a_landing` 이 이것을 재는 유일한 장치다 —
[[surviving-mutant-may-mean-accidental-safety]].

계약은 안 바뀐다. spec 은 이미 "착지 지점에서 그 change 의 모든 `revision: current`
증거 묶음의 source hash 가 일치해야 한다"고 적었고, 코드가 그 말을 못 지키고 있었다.

---

## 편집 — task 6.1.2 (분기 16 → 10, raise 6 → 8, 반환 2 → 2)

전은 `ast.before-6.1.2.json`(HEAD `cac190d8`), 후는 `ast.worktree.json`. 같은 열거기다.

**빠진 8** — 고정 순회를 `_pinning_bundles` · `_pinning_at` 으로 뽑았다. 이 함수에서
사라졌을 뿐 판정은 그대로이고, 같은 순회를 `compute_landing` 도 쓴다. 규칙이 두 집에
살면 각자가 상대의 시험을 통과시킨다 — [[two-judgements-cover-for-each-other]].

| 종류 | 소스 |
|---|---|
| For · IfExp | `for ast_path in sorted(analysis.glob('*/ast.json')) if analysis.is_dir() else ()` |
| BoolOp · If | `not isinstance(value, dict) or value.get('revision','current') != 'current'` |
| BoolOp · If | `not raw_source or not digest` |
| BoolOp · If | `blob is None or hashlib.sha256(blob).hexdigest() != digest` |

**더한 2** — 저자가 못 고르는 하한. 둘 다 `mismatched` 판정 **뒤에** 선다.

| 소스 | 무엇을 묻나 | 왜 이 자리인가 |
|---|---|---|
| `if not floor:` | 고정 번들이 역사에 들어온 적이 있나 | 게이트 뒤에 지울 수 있는 파일은 아무것도 고정하지 못한다 |
| `if not _is_ancestor(root, floor, candidate):` | 선언이 자기 증거보다 앞서나 | 저자는 오늘 만든 번들을 과거 커밋에 넣을 수 없다 |

**앞이 아니라 뒤에 세운 이유**는 기존 시험 하나 때문이다.
`test_a_landing_the_evidence_does_not_describe` 의 바늘은
`"is not the revision this evidence describes"` 인데, 하한을 앞에 세우면 그 픽스처가
새 문장으로 거절돼서 **그 시험이 재던 것을 더는 재지 못한다.** 그 바늘은 6.3 이
"이 파일이 같은 함정을 이미 한 번 발견하고 한 자리만 고쳤다"고 적은 바로 그 자리다.

### 이 편집을 재는 변이 (실측, 원복은 sha256 동일성으로 확인)

| 변이 | 1차 | 시험 보강 후 |
|---|---|---|
| M1 하한 판정을 통째로 삭제 | **CAUGHT** | CAUGHT |
| M2 `if not floor:` 거절 삭제 | SURVIVED | **CAUGHT** |
| M3 `-M --diff-filter=MA` → 평범한 `git log` | SURVIVED | **CAUGHT** |

M2·M3 이 1차에서 살아남은 것이 시험 둘을 더 쓴 이유다 —
[[surviving-mutant-may-mean-accidental-safety]].

---

## 공백 기록 — 7.2.1 · 7.3.1 · 7.4 는 이 함수를 FLM 없이 바꿨다 (2026-09-13, 7.6 이 적음)

이 세 task 는 내부를 바꾸면서 열거도 `not-applicable` 사유도 남기지 않았다. 7.6 이 7.2.1 직전
(`eaf536d2`, 7.1 과 소스 동일)과 HEAD(`2b5b05c1`)를 같은 열거기로 뽑아 `ast.before-7.2.1.json` ·
`ast.before-7.6.json` 으로 남긴다. 판정 근거(뮤테이션·A/B)는 각 task 의 VERIFY 절에 있다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `if unheld:` |
| 생김 | 분기 | `if candidate != computed:` |
| 생김 | 분기 | `computed[:12] if computed else f'none — {why}'` |
| 생김 | raise | `raise ValueError(f'landing point {candidate[:12]} does not hold the evidence this verdict read: ' + ', '.join(` |
| 생김 | raise | `raise ValueError(f"landing point {candidate[:12]} is not the landing this change's evidence computes ({named})` |

---

## 편집 — task 7.6 (규칙 한 집) · 분기 13 → 8 · 반환 2 → 2 · raise 10 → 5

편집 전 `ast.before-7.6.json`(HEAD `2b5b05c1`) · 편집 후 `ast.after-7.6.json`(워킹트리). 아래 표는 두 열거의
`source` 를 스크립트가 줄 단위로 대조한 것이다.

가드 여섯(base · 고정 0 · 불일치 · 하한 없음 · 하한 순서 · 미보유)을 `_landing_refusal` 로 옮겼다. 남은 것은 선언 **글자**에 대한 판정 셋(40-hex · 커밋 실재 · HEAD 조상)과 7.3.1 의 등식이다. 규칙의 순서는 이 함수가 쓰던 순서 그대로다 — 선언 경로의 시험들이 가드마다 그 문장을 못 박고 있기 때문이다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `if refusal:` |
| 사라짐 | 분기 | `if not _is_ancestor(root, base, candidate):` |
| 사라짐 | 분기 | `if not pinning:` |
| 사라짐 | 분기 | `if mismatched:` |
| 사라짐 | 분기 | `if not floor:` |
| 사라짐 | 분기 | `if not _is_ancestor(root, floor, candidate):` |
| 사라짐 | 분기 | `if unheld:` |
| 생김 | raise | `raise ValueError(refusal)` |
| 사라짐 | raise | `raise ValueError(f'landing point precedes the comparison base {base[:12]}: {candidate}')` |
| 사라짐 | raise | `raise ValueError(f'landing point {candidate[:12]} is pinned by no `revision: current` evidence: a declared lan` |
| 사라짐 | raise | `raise ValueError(f'landing point {candidate[:12]} is not the revision this evidence describes: ' + ', '.join(s` |
| 사라짐 | raise | `raise ValueError(f'landing point {candidate[:12]} is pinned by evidence that never entered this history: commi` |
| 사라짐 | raise | `raise ValueError(f"landing point {candidate[:12]} precedes the evidence that pins it ({floor[:12]}): a declare` |
| 사라짐 | raise | `raise ValueError(f'landing point {candidate[:12]} does not hold the evidence this verdict read: ' + ', '.join(` |

## 편집 — task 7.2.2 (착지는 고정 소스를 바꾼 커밋이어야 한다) · 분기 8 → 8 · 반환 2 → 2 · raise 5 → 5

편집 전 `ast.before-7.2.2.json`(HEAD `9692b8d1` blob, L655-719) · 편집 후 `ast.after-7.2.2.json`(워킹트리, L680-744). 표는 두 열거의 `source` 를 스크립트가 대조한 것이다.

**내부 논리 변경 없음** (AST 덤프 동일). 부르는 규칙 `_landing_refusal` 에 조건이 하나 늘어서 선언 경로의 거절이 하나 는다 — `test_a_declared_landing_that_changes_none_of_its_pinned_sources_is_refused` · `test_a_change_whose_work_precedes_its_base_gets_no_landing` 가 그 경로를 잰다.

열거의 분기·반환·raise 가 **같다**.

호출 — 사라짐 없음 · 생김 없음

## 편집 — task 7.2.6 (거절이 복구 경로를 말한다) · 분기 8 → 8 · 반환 2 → 2 · raise 5 → 5

편집 전 `ast.before-7.2.6.json`(HEAD `e9f86820` blob) · 편집 후 `ast.after-7.2.6.json`(워킹트리, L781-853).

같은 깃발 집합을 `_landing_refusal` 에 넘기고(`_self_repair_commits` 호출 하나 생김), 거기서 온 거절을 올릴 때 `LANDING_RECOVERY` 를 붙인다. 이 경로의 모든 거절은 "적힌 기록이 지금 규칙으로는 착지가 아니다"이고 돌아가는 길은 하나다 — 번들 갱신 → 기록 삭제 커밋 → 재기록. 문장은 **한 곳**(상수)에 산다: 기록 명령의 "이미 있다" 거절이 같은 상수를 쓴다([[two-judgements-cover-for-each-other]]).

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | raise | `raise ValueError(refusal)` |
| 생김 | raise | `raise ValueError(f'{refusal} — {LANDING_RECOVERY}')` |
| 생김 | 호출 | `_self_repair_commits` |

### 7.2.6 수리 — 적대 리뷰 F3 (2026-09-16) · 분기 8 → 8 · raise 5 → 5

계산값 대조 거절(`… is not the landing this change's evidence computes …`)만 복구 경로를
빠뜨리고 있었다. spec 이 "거절된 기록을 옮기는 경로를 사유 문장이 말해야 한다(SHALL)"를 요구하는데
이 문장 하나가 그것을 안 지켰고, 게다가 `--record-landing` 을 권하면서 그 명령은 "이미 있다"로
거절했다 — 7.7 이 닫은 **서로를 가리키는 두 문장**과 같은 모양이다. `LANDING_RECOVERY` 를 잇는다.
기존 시험 `test_a_later_commit_that_also_matches_is_refused` 에 단언 한 줄을 더해 못 박았다.

## task 7.5 — 번들 목록을 한 번 재서 넘긴다

`ast.before-7.5.json`(= `8091e6c4` 의 소스)과 `ast.after-7.5.json` 을 같은 열거기로 뽑아
**순서 있는 배열로** 대조했다: 분기 **순서열 8개가 바이트 동일** · 반환 **순서열 2개가 바이트 동일**.

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

## task 7.5.1 — gstack 리뷰의 permissive 결함 수리

`ast.before-7.5.1.json`(= git revision **`b29e1f4e`**, 소스 해시가 그 커밋과 일치)과
`ast.after-7.5.1.json` 을 같은 열거기로 뽑아 **순서 있는 배열**을 difflib 으로 정렬했다
(분기 8 · 반환 2 · raise 5 → 분기 8 · 반환 2 · raise 5).

입력을 `_measure_landing_inputs` 로 한 번 재고 `compute_landing` 에 넘긴다(F2). 분기 순서열은 **바이트 동일**이다.

**분기 순서열 바이트 동일.**

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1057-1147` · 분기 9 · 반환 2 · raise 5 (편집 전 L978-1052 · 분기 8 · 반환 2 · raise 5, `ast.before-7.5.2.json` = revision `e9f905bd`).

거절을 내기 전에 입력이 그대로인지 본다 — 입력을 **잴 때** 저자가 쓰던 중이었으면(재리뷰 적대 F5 재현) 복구 조언 대신 "다시 돌려라". 선언값과 계산값이 다른 자리에는 재확인을 두지 않았다: 선언값이 위를 통과했으면 같은 입력의 계산은 그 값 이하에서 **받고**, 받는 순간 `compute_landing` 이 재확인하므로 그 갈래는 도달 불가다(처음엔 넣었다가 뺐다 — 시험이 닿을 수 없는 갈래). 받으면 판정한 지문을 `facts` 로 넘긴다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1087 | If | `if candidate is None:` |
| B2 | 1089 | If | `if not FULL_SHA.fullmatch(candidate):` |
| B3 | 1100 | BoolOp | `process.returncode or process.stdout.strip() != candidate` |
| B4 | 1100 | If | `if process.returncode or process.stdout.strip() != candidate:` |
| B5 | 1102 | If | `if not _is_ancestor(root, candidate, 'HEAD'):` |
| B6 | 1112 | If | `if refusal:` |
| B7 | 1135 | If | `if candidate != computed:` |
| B8 | 1138 | IfExp | `computed[:12] if computed else f'none — {why}'` |
| B9 | 1145 | If | `if facts is not None:` |

| raise 줄 | 소스 |
|---|---|
| 1093 | `raise ValueError(f'landing point must be a full 40-hex commit id, not {candidate!r}')` |
| 1101 | `raise ValueError(f'landing point is not a commit in this repository: {candidate}')` |
| 1103 | `raise ValueError(f'landing point never landed on this history: {candidate}')` |
| 1122 | `raise ValueError(f'{refusal} — {LANDING_RECOVERY}')` |
| 1139 | `raise ValueError(f"landing point {candidate[:12]} is not the landing this change's evidenc` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1117-1207` · 분기 7 · 반환 2 · raise 5 (편집 전 L1057-1147 · 분기 9 · 반환 2 · raise 5, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`head` · `evidence` 를 받고 `facts` 를 안 받는다 — 착지가 판정한 바이트와 대상 판정이 읽는 바이트가 **같은 값**이라 넘길 지문이 없다. 선언값 ≠ 계산값 갈래의 주석을 **참인 이유**로 고쳤다: 7.5.2 는 "움직인 입력으로 여기 오는 길은 없다" 고 적었는데 거짓이었다(`HEAD` 이동 · git 결함이 둘 다 왔다 — 레드팀 repro_a). 이제 역사는 고정이고 결함은 예외라, 여기 오는 것은 같은 역사 · 같은 증거에서 더 낮은 수락 후보뿐이다 — 그래서 `computed` 가 빈 값인 경우의 문장 갈래(`none — …`)도 지웠다(도달 불가).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1147 | If | `if candidate is None:` |
| B2 | 1149 | If | `if not FULL_SHA.fullmatch(candidate):` |
| B3 | 1160 | BoolOp | `process.returncode or process.stdout.strip() != candidate` |
| B4 | 1160 | If | `if process.returncode or process.stdout.strip() != candidate:` |
| B5 | 1162 | If | `if not _is_ancestor(root, candidate, head):` |
| B6 | 1170 | If | `if refusal:` |
| B7 | 1193 | If | `if candidate != computed:` |

| raise 줄 | 소스 |
|---|---|
| 1153 | `raise ValueError(f'landing point must be a full 40-hex commit id, not {candidate!r}')` |
| 1161 | `raise ValueError(f'landing point is not a commit in this repository: {candidate}')` |
| 1163 | `raise ValueError(f'landing point never landed on this history: {candidate}')` |
| 1180 | `raise ValueError(f'{refusal} — {LANDING_RECOVERY}')` |
| 1201 | `raise ValueError(f"landing point {candidate[:12]} is not the landing this change's evidenc` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:1150-1240` · 분기 8 · 반환 2 · raise 5 (편집 전 L1117-1207 · 분기 7 · 반환 2 · raise 5, `ast.before-7.5.2.2.json` = revision `908a8a36`).

빈 계산값이면 `none — <사유>` 를 다시 말한다. 7.5.2.1 은 "빈 값은 못 온다" 며 지웠는데 Codex 가 얕은 복제 경계를 실행 중 바꿔 빈 값을 만들었다 — sha 를 고정해도 git 이 그 역사를 **읽는 방식**은 고정되지 않는다. 주석을 참인 만큼으로("거의 언제나 진짜 불일치").

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1180 | If | `if candidate is None:` |
| B2 | 1182 | If | `if not FULL_SHA.fullmatch(candidate):` |
| B3 | 1193 | BoolOp | `process.returncode or process.stdout.strip() != candidate` |
| B4 | 1193 | If | `if process.returncode or process.stdout.strip() != candidate:` |
| B5 | 1195 | If | `if not _is_ancestor(root, candidate, head):` |
| B6 | 1203 | If | `if refusal:` |
| B7 | 1226 | If | `if candidate != computed:` |
| B8 | 1233 | IfExp | `computed[:12] if computed else f'none — {why}'` |

| raise 줄 | 소스 |
|---|---|
| 1186 | `raise ValueError(f'landing point must be a full 40-hex commit id, not {candidate!r}')` |
| 1194 | `raise ValueError(f'landing point is not a commit in this repository: {candidate}')` |
| 1196 | `raise ValueError(f'landing point never landed on this history: {candidate}')` |
| 1213 | `raise ValueError(f'{refusal} — {LANDING_RECOVERY}')` |
| 1234 | `raise ValueError(f"landing point {candidate[:12]} is not the landing this change's evidenc` |
