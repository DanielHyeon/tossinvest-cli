# Function Logic Map: `_safe_changed_go_paths` (Python, a122 task 7.5.22 · 7.5.23)

편집 전(`197a0355`) `ast.before-7.5.22.json` · 7.5.22 뒤(`d7d494ee`) `ast.after-7.5.22.json` ·
7.5.23 전 `ast.before-7.5.23.json` · 7.5.23 뒤 `ast.after-7.5.23.json`(헬퍼는 `…-_numstat_records.json`).
분기: 가드 **11 → 14 → 14** · raise **3 → 4 → 4** · 헬퍼 **6**. **범위는 그 번들에만 적는다** — 산문에 절대 줄을 쓰면
같은 커밋의 수리가 밀어 버린다(7.5.22 가 정확히 그렇게 틀렸다). 지문 검증은
`analysis/harness/7523_coords.py`.

## 답하는 질문

**"`changed_existing_functions` 에게 이 diff 를 흘려도 되는가."** 반환은 `--numstat` 레코드이고,
통과 못 하면 `RuntimeError` 다. 올린 것은 `_judged` 가 `cannot derive modified Go functions: …`
한 줄로 바꾸고 `judged=False` 로 내려보낸다(fail-closed).

불변식: **바뀐 `*.go` 파일마다 판정 diff 가 본문을 낸다.** `changed_existing_functions` 는
훅(`@@`)으로만 "바뀐 기존 함수" 를 센다. 본문이 없으면 훅이 0 개이고, 그 파일의 요구가
**판정 줄 없이** 사라진다.

## task 7.5.22 — git 이 **이진으로 다루는** 파일

앞 판본의 가드는 `--name-only` 를 읽고 **이름만** 봤다. 워킹트리에 `.gitattributes` 한 줄
(`*.go binary` 또는 `*.go -diff`)이면 — **추적조차 안 해도** — git 이 `Binary files … differ` 를 내고
훅이 0 개가 되는데, `--name-only` 에는 그 파일이 **평범하게** 보인다. 실물 게이트(a112, 격리 worktree):
판정 줄 **36 → 0** · required **64 → 0** · `evidence complete` · rc **1 → 0**.

`--numstat` 으로 바꿔 `-`/`-` 를 거절한다. "훅이 0 개면 거절" 은 답이 아니다 — 정상인 mode-only
변경(`chmod +x`)도 훅이 0 개이고 numstat 은 `0`/`0` 이다. 덤으로 `--numstat` 은 rename 의 **양쪽**
이름을 내므로, 파서가 `base_file` 에 넘기는 **옛** 이름이 처음으로 가드 범위에 들어왔다
(`--name-only` 은 새 이름만 냈다 — 7.5.13 의 절반이 여기서 닫혔다).

## task 7.5.23 — git 이 **본문을 지우는** 파일

**7.5.22 가 적은 "`--numstat` 이 정확히 가른다" 는 거짓이었다.** `--numstat` 이 가르는 것은
"git 이 본문을 냈는가" 가 아니라 **"git 이 이것을 이진으로 다루는가"** 다. `diff=<드라이버>` 의
**textconv** 는 원본 blob 을 세는 `--numstat` 에 `1`/`1` 을 내면서, 판정 diff(textconv **출력**을
읽는다)의 본문을 통째로 지운다. 실측 — 추적 안 된 `.gitattributes` 한 줄 + config 키 하나:

| | 가드가 읽는 것 | 판정이 읽는 것 | 실물 게이트(a112) |
|---|---|---|---|
| `*.go binary` | `-`/`-` → **거절** | 훅 0 | 7.5.22 가 닫았다 |
| `*.go diff=nop` + `textconv` | **`1`/`1` — 정상으로 보인다** | 훅 **0** | **required 0 · evidence complete · rc 0** |
| `*.go diff=ext` + `command` | `1`/`1` | 훅 1 | `--no-ext-diff` 가 막고 있었다 |

가드와 판정이 **여전히 다른 투영**을 봤다 — 7.5.22 가 진단한 그 결함이 다른 문으로 살아 있었다.

**수리는 둘이다.**

1. 판정 diff 가 `--no-textconv` 를 준다(`--no-ext-diff` 옆). 그러면 textconv 저장소에서도
   거절이 아니라 **올바른 판정**이 나온다.
2. **두 투영을 맞춰 본다.** numstat 이 내용이 바뀌었다고(`0`/`0` 도 `-`/`-` 도 아니라고) 한 파일이
   훅을 하나도 안 냈으면 거절한다. 문을 하나씩 세는 대신 **어긋남**을 본다 — textconv · 외부 diff ·
   아직 모르는 git 기능이 같은 그물에 걸린다. 가드가 `--find-renames` 를 같이 줘야 판정과 같은
   짝을 보고 대조가 성립한다.

(2)가 (1)을 못 박는다: 깃발을 지우면 그 저장소의 본문이 사라지고 교차 검사가 **런타임에** 거절하므로,
픽스처 시험이 빨개진다. 7.5.22 까지 그 두 깃발은 **지워도 313 시험이 전부 초록**인 공짜 삭제였다.

## task 7.5.24 — 짝은 이름이 아니라 **순서**로

7.5.23 의 (2)는 짝을 **이름으로** 지었고, 그것이 정상 입력을 거절하는 회귀였다. 가드의 이름은
`-z` 라 **날 바이트**이고, 파서가 아는 이름은 `removeprefix` 를 거친 **유도값**이며 git 이 인용하면
`"a/we\"ird.go"` 다 — 두 이름 공간은 만나지 않는다. 부모와 A/B 로 잰 거절 셋:
새 파일 `we"ird.go` · `back\slash.go` · `diff.noprefix` 아래의 `b/new.go`. 그 거절은 `check()` 에서
`cannot derive modified Go functions` 가 되므로 **판정 줄이 하나도 안 나간다**.

수리는 **이름을 아예 안 비교하는 것**이다. 두 호출은 같은 `git diff` 에 형식만 다르므로 파일 순서가
같다(여섯 모양 섞은 픽스처로 실측, 레코드 5 · 구역 5 · 정렬 일치). `bodied` 를 구역마다 한 칸인
**목록**으로 두고 `zip(records, bodied)` 로 짝짓는다. 크기가 다르면 그 자체가 어긋남이다 —
깃발이 사라지면 판정 diff 가 구역 자체를 안 내므로 그쪽이 먼저 잡는다.

**"새 거절이 거절하는 정상 입력 0" 은 빈 표본이었다.** 그 수는 오늘 저장소의 **이름**을 센 것이고,
거절 가능한 **집합**은 `.gitattributes` 없이도 닿는다 ([[universal-check-passes-on-an-empty-sample]]).

편집 후 `ast.after-7.5.24.json`: 가드 분기 14 · raise 4 · 헬퍼 6.

## 새 거절이 오늘 거절하는 정상 입력

**0 — 세어서 하는 말이다.** `base-commit.txt` 를 가진 change **117** 전부에서 `*.go` 파일 줄
**97,235**: `-`/`-` **0** · `0`/`0` **0** · rename 쌍 **0** · 이름 거절 **0**
(`analysis/harness/7522_numstat.py`). 저장소에 `.gitattributes` 는 **0 개**다. 교차 검사가 거는 것도
0 이다 — 깃발 둘이 살아 있으면 본문이 안 사라진다. 가드 단독 A/B 는 7.5.22 에서 **234/234 SAME**.

관련: [[a-fingerprint-must-be-the-bytes-judged]] · [[universal-check-passes-on-an-empty-sample]] ·
[[fail-closed-must-name-what-it-rejects]] · [[a-new-guard-unpins-the-guards-behind-it]] ·
[[correction-unit-must-be-the-value]] · [[a-measurement-carries-its-moment]].

## task 7.5.27 — 가드도 판정과 같은 고정을 받는다

인자 `pins` 가 하나 늘었고 `git` 바로 뒤에 펼친다(`*(pins or [])`). 가드와 판정이 같은 시야를 봐야 7.5.23 의
교차 검사가 성립한다 — 가드만 고정이 없으면 clean 필터 아래에서 가드는 레코드를 **안** 내고 판정은 내서
크기가 갈려 거절된다(행동 시험이 잡는다). 분기는 **14 → 15** 다 — `*(pins or [])` 의 `or` 하나가 늘었다
(`ast.after-7.5.27.json`; 7.5.28 정정: 앞 판본은 "분기 수는 그대로다" 라 적었다).

## task 7.5.25 — 가드도 스냅숏 인덱스를 본다 (2026-09-24)

`ast.before-7.5.25.json`(분기 15) → `ast.after-7.5.25.json`(분기 15). 분기 수는 같고 자리가 바뀌었다:
편집 전 `BoolOp pins or []` · `IfExp [target] if target else []` 둘이 빠지고, 맨 앞에
`BoolOp not target and environment is None` · `If` 둘이 들어왔다(raise `ValueError: a worktree comparison needs
the worktree snapshot`). 넷째 인자는 `pins` 에서 `environment` 로 바뀌었다 — 가드와 판정이 **같은 쌍**
(`*_compared(base, target)`)을 **같은 환경**에서 견준다.

새 거절이 막는 것: 워킹트리 대상인데 환경이 없으면 `--cached` 가 **실제** 인덱스를 base 와 견준다 — 워킹트리도
스냅숏도 아닌 셋째 시야다. 오늘 호출자는 하나(`_changed_existing_functions`)이고 늘 환경을 건넨다 — 정상 입력 중
이 거절에 걸리는 것은 0 이다. 시험이 이 함수를 직접 부르는 자리 하나(`test_newline_changed_go_path_is_rejected_before_unified_diff_parsing`)는
스냅숏을 열어 건네도록 고쳤다. (7.5.31 정정: 커밋된 트리에서 직접 부르는 자리는 **셋**이다 — 같은 커밋이 뒤에 더한 `…untouched_worktree…` ·
`…guard_without_the_snapshot…` 도 부른다. 같은 실수 일곱 번째의 한 사례.)

## task 7.5.34 — 격리한 비교만 받는다 (2026-09-24)

`ast.before-7.5.34.json`(분기 15 · raise 5) → `ast.after-7.5.34.json`(분기 **13** · raise **4**). difflib 정렬: 편집 전
B1 · B2(`if not target and environment is None: raise ValueError`)가 빠졌다 — 서명이 `(root, comparison)` 이라 스냅숏 없는
워킹트리 비교를 **만들 수 없다**. 편집 전 B5 → 편집 후 B3 은 결함 문장만 바뀌었다(`for base {base}` 를 뗐다 — base 인자가
없다). 나머지 B3–B4 · B6–B15 는 편집 후 B1–B2 · B4–B13 과 같다. 분기 밖: diff 가 `*comparison.trees` 를
`env=comparison.environment` 에서 견주고 **pathspec 을 안 받는다**(`"--", "*.go"` 를 뗐다 — 두 트리에 `*.go` 만 있다).

## task 7.5.15 — 시한 (2026-09-26)

> 편집 전 `ast.before-7515.json`(revision `1d1e5ca7`) · 편집 후 `ast.after-7515.json` — 분기 13 → 13, 정렬 결과 전부 같음.

편집 후 `tools/logic-map/check_analysis.py:493-550` · 분기 13 · 반환 1 · raise 4 · 호출 13 (`ast.after-7515.json`, source sha `ae8931c3f577`)
편집 전 `tools/logic-map/check_analysis.py:493-548` · 분기 13 · 반환 1 · raise 4 · 호출 13 (`ast.before-7515.json`, revision `1d1e5ca7`, source sha `b28398e26f3d`)

| 옛 id | 새 id | 새 줄 | 종류 | 소스(새) | 바뀐 것 |
|---|---|---|---|---|---|
| B1 | B1 | 529 | If | `if process.returncode:` | 같음 |
| B2 | B2 | 530 | IfExp | `process.stderr.decode('utf-8', 'replace') if isinstance(process.stderr, bytes) else process.stderr` | 같음 |
| B3 | B3 | 531 | BoolOp | `stderr.strip() or 'git diff failed'` | 같음 |
| B4 | B4 | 532 | IfExp | `process.stdout if isinstance(process.stdout, bytes) else process.stdout.encode('utf-8')` | 같음 |
| B5 | B5 | 534 | For | `for _, _, paths in records:` | 같음 |
| B6 | B6 | 535 | For | `for raw in paths:` | 같음 |
| B7 | B7 | 536 | Try | `try:` | 같음 |
| B8 | B8 | 538 | ExceptHandler | `except UnicodeDecodeError as error:` | 같음 |
| B9 | B9 | 542 | BoolOp | `'\n' in path or '\r' in path or '\t' in path` | 같음 |
| B10 | B10 | 542 | If | `if '\n' in path or '\r' in path or '\t' in path:` | 같음 |
| B11 | B11 | 544 | For | `for added, deleted, paths in records:` | 같음 |
| B12 | B12 | 545 | BoolOp | `added == b'-' and deleted == b'-'` | 같음 |
| B13 | B13 | 545 | If | `if added == b'-' and deleted == b'-':` | 같음 |

`git diff --numstat` 에 `timeout=30` — 판정 diff(`_changed_existing_functions`, 같은 두 트리)와 같은 값. 멎으면 `TimeoutExpired` 가
`GATE_FAULTS` 로 판정 줄이 된다. 분기 · 반환 · raise 불변. `branch-test-map.md` 의 "`timeout=` 이 없다 → 7.5.15" 는 이것으로 닫혔다.

## 보수 — 최종 파일로 다시 열거 (독립 주장정확성 리뷰 F11, 2026-09-26)

최종 `tools/logic-map/check_analysis.py:495-552` · 분기 13 · 반환 1 · raise 4 · 호출 13 · source sha `ff590b79db6e` (비교 기준 `ast.after-7515.json`, 그 판 `ae8931c3f577` `:493-550`).
위 절들이 "최종" 이라 적은 sha 는 **중간판**이었다(리뷰 지적 — 표기를 고쳤다). 최종 파일에서 `ast.after-759r.json` 을 다시 뽑아 앞 절의 편집 후 열거와 대조했다 — 분기(종류 · 소스) · 반환 · raise · 호출 수가 **같다**(구조 동일). 바뀐 것은 sha 와 줄 좌표다.
