# Function Logic Map: `main`

`tools/logic-map/check_analysis.py:817-840`(HEAD `848b9ba3`) → `:836-872`(편집 후) · Python

> **이 표는 손으로 읽어서 만들지 않았다.** 같은 디렉터리의 `ast.json`(HEAD)과
> `ast.worktree.json`(편집 후)을 `enumerate.py` 가 기계로 열거했다. Go 용
> `extract_go_ast.go` 는 Python 함수에 쓸 수 없으므로 a120 선례를 따른다.

## Inputs and invariants

`--change` 와 `--root` 를 받아 `check` 를 돌리고 결과를 출력하고 종료 코드를 준다.
불변식은 **판정과 설명이 같은 실행에서 나와야 한다**는 것이다 — 어떤 창으로 요구했는지가
없으면 요구된 함수 이름은 판정이 아니라 소음이다.

## Branches and early returns — 편집 전 (분기 4 · 반환 2)

| ID | 줄 | 소스 |
|---|---|---|
| B1 | 824 | `if errors:` → `825 for error in errors:` → **`827 return 1`** |
| B2 | 825 | `for error in errors:` |
| B3 | 828 | `if context.get('execution_baseline_adoption'):` |
| B4 | 835 | `if landing:` → `836 print(landed-commit … required N function(s))` |

## 결함 (task 3.3)

**`B1` 의 `827 return 1` 이 `B3`·`B4` 를 통째로 건너뛴다.** 착지·요구 수를 찍는
`B4` 는 **실패 경로에서 도달 불가**다. 그래서 task 1.9 가 적은 "항상 출력한다"는
성공할 때만 참이었고, 실측이 그것을 확인한다(2026-09-10, HEAD `848b9ba3`):

| change | 출력 줄 | 가장 긴 줄 | 창을 말하는 줄 |
|---|---|---|---|
| a074 | 324 (`missing evidence for modified function` **316**) | 183자 | **0** |
| a076 | **1** | **21,838자** | **0** |

3.2.4 와 **같은 모양**이다 — early return 이 뒤의 판정을 건너뛴다. 그때는 아카이브를
안 세게 했고, 여기서는 설명을 안 찍게 한다.

## 편집 — 분기 4 → 6, 반환 2 → 2

| ID | 줄 | 소스 |
|---|---|---|
| B1 | 848 | `if base:` → `850 print(base … → 대상 … required N function(s))` |
| B2 | 854 | `if not landing:` → `856 print(창에 커밋 K개가 더 있다)` |
| B3 | 858 | `landed_after or '?'` — 못 재면 숫자를 **지어내지 않는다** |
| B4 | 862 | `if errors:` → `865 return 1` |
| B5 | 863 | `for error in errors:` |
| B6 | 866 | `if context.get('execution_baseline_adoption'):` |

`print` 호출 줄이 `[850, 856, 864, 867, 869]` 이고 `return 1` 이 `865` 다. 즉 창 줄
둘은 **실패 반환보다 앞**에 있으므로 두 경로 모두에서 찍힌다 — 구조가 그것을 말한다.
반환은 **둘 그대로**(`865: 1`, `872: 0`)라 이 편집은 경로를 더하거나 지우지 않는다.

옛 `B4 if landing:` 는 없어졌다. 한 줄이 성공·실패 양쪽을 담당하므로 두 벌을 두지
않는다 — 두 벌이면 한쪽만 조용해지는 오늘의 상태로 되돌아간다.

## Calls and live bindings

`check` · `_target_text` · `_commits_after`(`git rev-list --count <base>..HEAD`) ·
`print` · `argparse`. 쓰기는 없다.

## State mutations and fallbacks

없다. `context` 는 `check` 가 채우고 여기서는 읽기만 한다. fallback 은 하나 —
커밋 수를 못 재면 `'?'` 를 찍는다(`B3`). **숫자를 지어내지 않는다.**

## Safety conclusion

판정을 한 줄도 바꾸지 않는다. 이 태스크의 불변식은 **126개 id 의 rc 가 하나도
안 바뀌는 것**이고 그것으로 잰다. 성공 줄
(`evidence complete or diff-proven exempt`)은 기록 스무 곳이 인용하므로 손대지 않았다.
Go 파일 변경 0줄.

## 편집 (task 4.1) — 분기 6 → 7, 반환 2 → 2 (값도 동일)

`ast.before-4.1.json`(= 커밋 `c91dc484`) → `ast.after-4.1.json`.

| 옛 | 새 |
|---|---|
| `B1 if base:` | `B1 base and 'landing' in context` (BoolOp) · `B2 if …:` (If) |

경로는 하나도 안 지운다. 반환 둘의 **값까지 글자 그대로 같다**. `print` 자리는
`[914, 920, 928, 931, 933]` 이고 창 줄 둘(914·920)은 여전히 실패 반환보다 **앞**이다 —
3.3 이 세운 "실패해도 창을 찍는다"가 그대로다. 달라진 것은 **못 잰 창을 안 찍는다**
하나다.

**왜 조건 하나가 판정인가.** 착지 해소가 실패하면 `check` 는 `facts["landing"]` 에
닿기 전에 돌아간다. 그래서 `"landing" in context` 는 "이 창을 실제로 쟀는가"와 같은
말이다. 옛 조건은 그 빈칸을 `context.get("landing", "")` 의 기본값으로 메워
`working tree (no landed-commit.txt) required 0 function(s)` 라고 찍었다 — 대상도
아니고 세지도 않은 값이다.

| 변이 | 빨개진 시험 |
|---|---|
| O1 조건 되돌리기(`if base:`) | **1** — `test_a_window_that_could_not_be_derived_is_not_printed` |
| O2 창 줄 통째로 삭제 | **4** — 3.3 의 창 시험 둘 · 빈 요구 집합 · 1.12 이관 창 |

둘이 **겹치지 않는다**. 조건과 줄이 각각 독립으로 묶였다는 뜻이고, 하나만 묶으면
나머지 변이가 산다 — [[surviving-mutant-may-mean-accidental-safety]].

## 편집 (task 7.1) — 분기 9 → 13, 반환 3 → 3

`ast.before-7.1.json`(= 커밋 `60150803`) → `ast.after-7.1.json`. 표는 두 열거를
스크립트가 대조해 만들었다 — 손으로 옮겨 적지 않았다.
`check_analysis.py:1109-1184` · Python

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1120 | If | `if args.record_landing:` |
| B2 | 1122 | For | `for line in lines:` |
| B3 | 1135 | BoolOp | `base and 'landing' in context` |
| B4 | 1135 | If | `if base and 'landing' in context:` |
| B5 | 1142 | If | `if not landing:` |
| B6 | 1147 | BoolOp | `landed_after or '?'` |
| B7 | 1153 | BoolOp | `context.get('base_shaped_bundles') or []` |
| B8 | 1153 | comprehension | ` for name in context.get('base_shaped_bundles') or []` |
| B9 | 1154 | If | `if base_shaped:` |
| B10 | 1158 | If | `if len(base_shaped) > 3:` |
| B11 | 1174 | If | `if errors:` |
| B12 | 1175 | For | `for error in errors:` |
| B13 | 1178 | If | `if context.get('execution_baseline_adoption'):` |

**새로 생긴 넷**은 전부 `B5 if not landing:` 안에 있다:
- `B7` BoolOp — `context.get('base_shaped_bundles') or []`
- `B8` comprehension — ` for name in context.get('base_shaped_bundles') or []`
- `B9` If — `if base_shaped:`
- `B10` If — `if len(base_shaped) > 3:`

번호가 밀렸다 — 옛 `B7·B8·B9`(`if errors:` · `for error in errors:` · 이관 문구)가
새 `B11·B12·B13` 이다. [[positional-branch-ids-break-hand-renumbering]] 가 말하는
자리이고, 그래서 이 표는 옛 표를 고쳐 쓰지 않고 **다시 열거해서** 적었다.

**반환은 셋 그대로다**(`1096, 1127, 1134` →
`1124, 1177, 1184`). 값도 같다. 이 편집은 경로를 더하거나
지우지 않는다 — `if not landing:` 안에서 **출력 두 갈래**가 생겼을 뿐이고, 판정(rc)에
닿는 분기는 하나도 안 바뀌었다.

### 왜 사실과 조언을 갈랐나

옛 판본은 창의 크기(사실)와 `--record-landing`(조언)이 **한 f-string** 이었다. 조언만
막으려면 사실까지 같이 죽는다. 그래서 `window` 를 먼저 만들고 두 갈래가 그것을 공유한다 —
3.3 이 세운 "실패해도 창을 찍는다"는 양쪽에서 글자 그대로 남는다.

### 왜 `check` 가 재고 `main` 이 읽나

`main` 에는 `change_dir` 도 `analysis` 도 없다. 여기서 다시 해소하면 디렉터리 해소가
**세 벌**이 된다(7.6 이 이미 두 벌을 결함으로 적었다). `check` 는 그 둘을 이미 손에 쥐고
있고 `facts["landing"]` · `facts["required_count"]` 를 넣는 자리가 바로 거기다.
빌리는 change 면 `analysis` 가 빌려주는 쪽을 가리키는데, 고정 번들이 사는 자리가 거기라서
그것이 맞는 대상이다.

---

## 공백 기록 — 7.2.1 · 7.3.1 · 7.4 는 이 함수를 FLM 없이 바꿨다 (2026-09-13, 7.6 이 적음)

이 세 task 는 내부를 바꾸면서 열거도 `not-applicable` 사유도 남기지 않았다. 7.6 이 7.2.1 직전
(`eaf536d2`, 7.1 과 소스 동일)과 HEAD(`2b5b05c1`)를 같은 열거기로 뽑아 `ast.before-7.2.1.json` ·
`ast.before-7.6.json` 으로 남긴다. 판정 근거(뮤테이션·A/B)는 각 task 의 VERIFY 절에 있다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `try:` |
| 생김 | 분기 | `except GATE_FAULTS as exc:` |
| 생김 | 분기 | `try:` |
| 생김 | 분기 | `except GATE_FAULTS as exc:` |

## 편집 — task 7.7 (조언이 기록 명령의 판정에 묻는다) · 분기 17 → 20 · 반환 3 → 3 · raise 0 → 0

편집 전 `ast.before-7.7.json`(HEAD `3da639a9` blob) · 편집 후 `ast.after-7.7.json`(워킹트리, L1329-1430). 아래 표는 두 열거의 `source` 를 스크립트가 줄 단위로 대조한 것이다.

조언 줄의 권유 갈래(`else`) 안에서 `_recording_refusal` 에 묻는다. 사유가 있으면 ``— `--record-landing` cannot narrow it: <사유>`` 를 찍고 명령을 권하지 않는다. 판정 함수가 결함을 내면(`GATE_FAULTS`) `cannot tell whether it would record: …` 로 **권하지 않는다** — 조언은 판정이 아니므로 그 결함이 판정 줄을 바꾸지 않는다. 권유 문장은 기록을 약속하지 않게 바꿨다("if no commit on this history matches that evidence, the command says so instead of recording") — 걸어야 아는 거절은 예측하지 않기 때문이다(실물 15건, 전부 아카이브). base 모양 갈래(7.1)는 그대로 **앞**에 선다.

| | 종류 | 소스 |
|---|---|---|
| 생김 | 분기 | `try:` |
| 생김 | 분기 | `except GATE_FAULTS as exc:` |
| 생김 | 분기 | `if refusal:` |

호출 — 사라짐 없음 · 생김 ['_recording_refusal', 'resolve_referenced_change']

## 편집 — task 7.2.2 (착지는 고정 소스를 바꾼 커밋이어야 한다) · 분기 20 → 20 · 반환 3 → 3 · raise 0 → 0

편집 전 `ast.before-7.2.2.json`(HEAD `9692b8d1` blob, L1337-1438) · 편집 후 `ast.after-7.2.2.json`(워킹트리, L1364-1465). 표는 두 열거의 `source` 를 스크립트가 대조한 것이다.

권유 갈래의 문장 하나: `if no commit on this history matches that evidence` → `if no commit on this history is accepted as the landing`. 조언이 말하는 조건이 명령이 거절하는 조건과 같아야 한다(7.7). 분기·호출 변화 0 — 7.7 의 R20 은 자리가 사라져 SKIP 이고 새 자리에 건 R20' 를 `test_a_refusal_only_the_walk_finds_is_not_promised_away` 가 잡는다.

열거의 분기·반환·raise 가 **같다**.

호출 — 사라짐 없음 · 생김 없음

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1977-2083` · 분기 21 · 반환 3 · raise 0 (편집 전 L1755-1856 · 분기 20 · 반환 3 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

조언 계산 결함이면 `--record-landing` 을 권하지도 거절하지도 않고 "cannot tell whether …: <git 의 말>" 로 말한다 — 모르는 채로 권하면 7.1 의 이유가 되살아난다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1988 | If | `if args.record_landing:` |
| B2 | 1989 | Try | `try:` |
| B3 | 1991 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B4 | 1993 | For | `for line in lines:` |
| B5 | 2000 | Try | `try:` |
| B6 | 2002 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B7 | 2013 | BoolOp | `base and 'landing' in context` |
| B8 | 2013 | If | `if base and 'landing' in context:` |
| B9 | 2020 | If | `if not landing:` |
| B10 | 2025 | BoolOp | `landed_after or '?'` |
| B11 | 2031 | BoolOp | `context.get('base_shaped_bundles') or []` |
| B12 | 2031 | comprehension | ` for name in context.get('base_shaped_bundles') or []` |
| B13 | 2033 | If | `if fault:` |
| B14 | 2037 | If | `if base_shaped:` |
| B15 | 2041 | If | `if len(base_shaped) > 3:` |
| B16 | 2055 | Try | `try:` |
| B17 | 2059 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B18 | 2061 | If | `if refusal:` |
| B19 | 2073 | If | `if errors:` |
| B20 | 2074 | For | `for error in errors:` |
| B21 | 2077 | If | `if context.get('execution_baseline_adoption'):` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2063-2182` · 분기 22 · 반환 3 · raise 0 (편집 전 L1977-2083 · 분기 21 · 반환 3 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

출력을 `backslashreplace` 로 — 판정 줄의 홀로 선 서로게이트가 엄격한 UTF-8 로캘에서 `UnicodeEncodeError` 로 판정 줄 없이 끝났다(레드팀, 로캘의 함수였다). 창 줄 · 조언은 `check` 가 푼 `head` 를 쓴다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2073 | If | `if reconfigure is not None:` |
| B2 | 2080 | If | `if args.record_landing:` |
| B3 | 2081 | Try | `try:` |
| B4 | 2083 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B5 | 2085 | For | `for line in lines:` |
| B6 | 2092 | Try | `try:` |
| B7 | 2094 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B8 | 2105 | BoolOp | `base and 'landing' in context` |
| B9 | 2105 | If | `if base and 'landing' in context:` |
| B10 | 2112 | If | `if not landing:` |
| B11 | 2120 | BoolOp | `landed_after or '?'` |
| B12 | 2126 | BoolOp | `context.get('base_shaped_bundles') or []` |
| B13 | 2126 | comprehension | ` for name in context.get('base_shaped_bundles') or []` |
| B14 | 2128 | If | `if fault:` |
| B15 | 2132 | If | `if base_shaped:` |
| B16 | 2136 | If | `if len(base_shaped) > 3:` |
| B17 | 2150 | Try | `try:` |
| B18 | 2158 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B19 | 2160 | If | `if refusal:` |
| B20 | 2172 | If | `if errors:` |
| B21 | 2173 | For | `for error in errors:` |
| B22 | 2176 | If | `if context.get('execution_baseline_adoption'):` |

## task 7.5.2.2 — 스냅숏은 끝에서 디스크와 다시 대조한다 (7.5.2.1 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:2171-2291` · 분기 22 · 반환 3 · raise 0 (편집 전 L2063-2182 · 분기 22 · 반환 3 · raise 0, `ast.before-7.5.2.2.json` = revision `908a8a36`).

창 줄 끝에 `— judged at HEAD <sha12>` — 게이트의 PASS 가 어느 역사의 판정인지 출력에 남는다(재리뷰 적대). `head` 를 창 줄 앞으로 올렸다(조언 갈래만 쓰던 값).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 2181 | If | `if reconfigure is not None:` |
| B2 | 2188 | If | `if args.record_landing:` |
| B3 | 2189 | Try | `try:` |
| B4 | 2191 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B5 | 2193 | For | `for line in lines:` |
| B6 | 2200 | Try | `try:` |
| B7 | 2202 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B8 | 2213 | BoolOp | `base and 'landing' in context` |
| B9 | 2213 | If | `if base and 'landing' in context:` |
| B10 | 2224 | If | `if not landing:` |
| B11 | 2229 | BoolOp | `landed_after or '?'` |
| B12 | 2235 | BoolOp | `context.get('base_shaped_bundles') or []` |
| B13 | 2235 | comprehension | ` for name in context.get('base_shaped_bundles') or []` |
| B14 | 2237 | If | `if fault:` |
| B15 | 2241 | If | `if base_shaped:` |
| B16 | 2245 | If | `if len(base_shaped) > 3:` |
| B17 | 2259 | Try | `try:` |
| B18 | 2267 | ExceptHandler | `except GATE_FAULTS as exc:` |
| B19 | 2269 | If | `if refusal:` |
| B20 | 2281 | If | `if errors:` |
| B21 | 2282 | For | `for error in errors:` |
| B22 | 2285 | If | `if context.get('execution_baseline_adoption'):` |
