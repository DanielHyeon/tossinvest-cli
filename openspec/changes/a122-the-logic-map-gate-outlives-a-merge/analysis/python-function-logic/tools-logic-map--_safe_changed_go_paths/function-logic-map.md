# Function Logic Map: `_safe_changed_go_paths` (Python, a122 task 7.5.22)

## task 7.5.22 — git 이 본문을 안 내면 그 파일의 요구가 조용히 사라진다 (7.5.2.4 재리뷰 2026-09-22, 보안)

`tools/logic-map/check_analysis.py:104-125` (편집 전) · 분기 **11** · 반환 0 · raise 3 · 호출 11.
`enumerate.py` 가 기계로 열거했다 — `ast.before-7.5.22.json` · `ast.after-7.5.22.json`
(새 헬퍼는 `ast.after-7.5.22-_numstat_records.json`).

이 결함은 이 change 의 어느 로트도 만들지 않았다. `_safe_changed_go_paths` 는 7.2.2 부터 글자 그대로였고,
재확인 원장은 여기를 **원리상 못 본다** — 입력을 자식 프로세스(`git diff`)가 읽기 때문이다(→ task 7.5.8).

## Inputs and invariants

`root` · `base` · `target`. 답하는 질문은 **"`changed_existing_functions` 에게 이 diff 를 흘려도 되는가"** 다.
반환값이 없다 — 통과하거나 `RuntimeError` 를 올린다. 올린 것은 `_judged` 가
`cannot derive modified Go functions: …` 한 줄로 바꾸고 `judged=False` 로 내려보낸다(fail-closed).

불변식: **바뀐 `*.go` 파일마다 git 이 본문을 낸다.** `changed_existing_functions` 는 훅(`@@`)으로만
"바뀐 기존 함수" 를 센다. 본문이 없으면 훅이 0 개이고, 그 파일의 요구가 **판정 줄 없이** 사라진다.

## 편집 전의 논리와 그것이 못 보는 자리

가드는 `git diff --no-ext-diff --name-only -z` 를 읽고 **이름만** 본다(B6~B11).

| 갈래 | 줄(편집 전) | 하는 일 |
|---|---|---|
| B2 | 113 | git 이 죽으면 결함 |
| B8 · B9 | 118 · 120 | UTF-8 아닌 이름을 거절 |
| B11 | 124 | `\n` · `\r` · `\t` 가 든 이름을 거절 |

못 보는 것 둘 (**둘 다 진짜 git 으로 실측**):

1. **본문 억제.** 워킹트리에 `.gitattributes` 한 줄(`*.go binary` 또는 `*.go -diff`)이면 git 은
   `Binary files a/x.go and b/x.go differ` 를 내고 훅을 **0 개** 낸다. 그런데 `--name-only` 에는
   `x.go` 가 **평범하게** 보이므로 이 가드는 아무것도 안 한다. 결과: 그 change 의 Go 요구가
   **0 건**이 되고 판정 줄은 안 나간다 — 게이트의 중심 요구가 조용히 꺼진다.
   그 `.gitattributes` 는 **추적될 필요조차 없다**(워킹트리에 놓기만 하면 git 이 읽는다).
2. **rename 의 옛 이름.** `--name-only` 은 rename 에서 **새** 이름 하나만 낸다. 그런데 파서가
   `base_file` 에 넘기는 것은 `--- a/<옛 이름>` 에서 뽑은 **옛** 이름이다. 옛 이름에 탭이 있으면
   git 이 인용해서(`"a/a\tb.go"`) `removeprefix("a/")` 가 무효가 되고
   `cannot load existing base file <인용된 이름>` 이라는 말이 안 되는 문장으로 **거짓 차단**된다
   (7.5.13 이 적은 모양이 rename 갈래에서 실제로 살아 있다).

## 왜 "훅이 0 개면 거절" 이 답이 아닌가

**mode-only 변경도 훅이 0 개다**(`chmod +x x.go`). 그것은 정상 입력이고 요구가 0 건인 것이 맞다.
`--numstat` 이 그 둘을 정확히 가른다 — 실측:

| 입력 | `--name-only` | `--numstat` | 훅 |
|---|---|---|---|
| 평범한 편집 | `x.go` | `1\t1\tx.go` | 있음 |
| mode-only (`chmod +x`) | `x.go` | **`0\t0\t`**`x.go` | 없음 |
| `*.go binary` / `-diff` | `x.go` | **`-\t-\t`**`x.go` | 없음 |

## 수리 (편집 후)

가드가 `--name-only` 대신 `--numstat -z` 를 읽는다. **호출은 늘지 않는다** — 같은 자리에서
형식만 바꿨다. 레코드 해독은 `_numstat_records` 로 뺐다(`<더함>\t<지움>\t<경로>\0`, rename 은
경로 칸이 비고 다음 두 칸이 옛·새 이름). 모양이 다르면 지어내지 않고 결함으로 올린다 —
못 읽은 표를 "바뀐 파일 없음" 으로 읽으면 그것이 이 task 가 닫는 바로 그 구멍이다.

**새 거절은 가장 뒤에 선다.** 이름 검사를 **전부** 마친 뒤에 본문을 묻는다
([[a-new-guard-unpins-the-guards-behind-it]]: 앞에 세우면 뒤의 가드가 못에서 뽑힌다).

**새 거절이 오늘 거절하는 정상 입력: 0 (측정이다, 유도가 아니다).** `base-commit.txt` 를 가진
change **117** 전부에서 base→HEAD 의 `*.go` 파일 줄 **97,235** 를 셌다 — `-`/`-` **0** · `0`/`0` **0** ·
rename 쌍 **0** · 이름 거절 **0**. 가드 단독 A/B 도 **234/234 SAME · DIFFERENT 0**
(117 change × target 둘). 충돌 중인 워킹트리에서도 `numstat` 은 평범한 3칸 레코드를 낸다(실측).

편집 후: `:135-179` · 분기 **14** · 반환 0 · raise **4** · 호출 13.
헬퍼 `_numstat_records`: `:104-132` · 분기 6 · 반환 1 · raise 2 · 호출 9.

| id | 줄 | 종류 | 무엇을 정하나 |
|---|---|---|---|
| B6 · B7 | 164 · 165 | For | 레코드마다 · 경로마다 — rename 은 **양쪽** 이름 |
| B8 · B9 | 166 · 168 | Try | UTF-8 아닌 이름 (앞 로트와 같음) |
| B11 | 172 | If | `\n` · `\r` · `\t` 이름 (앞 로트와 같음) |
| **B12** | 174 | For | **새 갈래** — 이름을 다 본 **뒤에** 레코드를 다시 돈다 |
| **B13 · B14** | 175 | If | **새 갈래** — `-`/`-` 면 파일 이름을 대고 거절한다 |

관련: [[fail-closed-must-name-what-it-rejects]] · [[a-silent-skip-is-a-door]] ·
[[existence-check-is-not-a-role-check]] · [[parsing-needs-the-grammars-state]].
