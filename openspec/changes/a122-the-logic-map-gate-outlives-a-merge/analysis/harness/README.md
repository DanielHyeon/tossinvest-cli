# a122 측정·검증 하네스

**이 저장소의 도구가 아니라 이 change 의 일회용 하네스다.** 커버리지나 안전을 주장하지
않는다. 여기 있는 이유는 하나다 — `branch-test-map.md` 와 `review.md` 가 이 스크립트들의
출력을 **증거로 인용**하는데, 예전 task 들의 하네스(`722_mut.py` · `726_mut.py` · `76_mut.py`)는
세션 스크래치패드에 살다가 사라져서 **인용만 남고 재현이 불가능**해졌다.

루트는 세지 않고 **유도**한다 — 체크아웃 이름이 바뀌어도 고아가 안 된다
([[renamed-checkout-strands-absolute-path-state]]). 런타임 산출물은 `_work/` 이고 커밋하지 않는다.

| 파일 | 무엇을 재나 | 실행 |
|---|---|---|
| `75_census.py` | change 마다 후보 수 · 번들 수 · 예측 spawn (전수, 걷지 않는다) | `python3 75_census.py` |
| `75_levers.py` | 지렛대 넷의 단가 (`show` 하나씩 · `cat-file --batch` · `ls-tree` · 번들 재파싱) | `python3 75_levers.py` |
| `75_attrib.py` | spawn 을 **호출자 함수**로 귀속 + 한 개짜리 fetch 두 방식 | `python3 75_attrib.py <change-dir-name>…` |
| `75_time.py` | `compute_landing` 의 벽시계와 spawn 수 | `python3 75_time.py <change-dir-name>…` |
| `75_ab.py` | 판정 A/B 전수 (사본 대조군 + 이어 달리기) | `python3 75_ab.py [<before-sha>] [<초 예산>]` |
| `75_mut.py` | 변이 T · U · V · W · Y · Z (사본 대상 · 무변이 대조군 선행 · 스위트 전체 · 생존하면 양성 대조) | `python3 75_mut.py [<구간 lo:hi>]` |
| `751_check_ab.py` | `check()` 반환 A/B 전수 (7.5.1) | `python3 751_check_ab.py [<before-sha>] [<초 예산>]` |
| `752_reads.py` | a099 `check()` 한 번의 `ast.json` 읽기 수 · git 수 (7.5.2) | `python3 752_reads.py` |
| `7521_inputs.py` | 진입점에서 닿는 **모든 I/O** 를 AST 로 센다 — "묶었다" 고 적기 전의 입력 목록 (7.5.2.1) | `python3 7521_inputs.py [check record_landing main]` |
| `61_forgeries.py` | 4.4 의 위조 셋(+ (c) 의 선언 값 셋 P · L · E)을 지금 코드에서 — 대조군 · 기록 명령 · 선언 뒤 (6.1 부모 · 6.4 보수가 L 행 추가) | `python3 61_forgeries.py` |
| `64_base_ab.py` | 해소기 · base 해소의 답을 저장소 **모든** change 디렉터리에서 before/after 로 (6.4(b)(c)) | `python3 64_base_ab.py [<before-rev>]` |
| `64_gate_mut.py` | `tools/gate.sh` 해소기 변이 G1~G5 — 두 스위트(`test_check_analysis` 전체 + shell 자기 시험) (6.4(f) · 6.4 보수가 G5 추가) | `python3 64_gate_mut.py` |
| `64r_census.py` | 6.4 보수의 새 가드(옮긴 change 의 base 를 같은 id 의 다른 자리와 대조)가 거절할 정상 입력 — 저장소 change 디렉터리 전수, git 으로 직접 | `python3 64r_census.py` |
| `64r_head_calls.py` | `check` 한 번의 `_head_commit` 호출 수, 활성 change 전수 (`_head_commit` docstring 의 근거) | `python3 64r_head_calls.py` |
| `7521_main_ab.py` | `main()` 출력 전체 A/B · 순서를 번갈아 · 결정적 계수(git 수 · `ast.json` 읽기 수) (7.5.2.1) | `python3 7521_main_ab.py [<before-sha>] [<초 예산>]` |
| `759_census.py` | 시험 색인 · 인용 해소를 추적 파일로 좁히면 달라지는 번들 — 저장소 번들 **전수**의 인용 판정을 디스크 · 추적 거름 입힌 편집 전 · 편집 뒤 세 판으로 (7.5.9 — 거부하는 정상 입력) | `python3 759_census.py [<before-rev>]` |
| `759_ab_verify.py` | `7521_main_ab.py` 가 DIFFERENT 로 찍은 change 를 전체 출력으로 다시 돌려, 7.5.9 의 거절 문장 꼬리만 정규화하고 견준다 | `python3 759_ab_verify.py <before-rev> <done.json>` |
| `759_flm_rows.py` | FLM 표의 행을 `ast.before-*/after-*` 열거에서 `difflib` 정렬로 찍는다(손 재번호 금지) | `python3 759_flm_rows.py <함수> <전 꼬리표\|-> <후 꼬리표>` |
| `7510_collide.py` | 원장 지문이 서로 다른 두 디스크 상태에 같은 값을 내는가 — 목록 · 순회(편집 전) · 추적 목록(편집 뒤) (7.5.10) | `python3 7510_collide.py [<rev>]` |
| `7511_isdir.py` | `Path.is_dir()` · `rglob` 이 **이 인터프리터에서** 무엇을 삼키는가 — 판본 영수증 (7.5.11) | `python3 7511_isdir.py` · `uv run --no-project --python 3.14 7511_isdir.py` |
| `7515_mut.py` | `execution_baseline.py` 시한 다섯 자리의 변이 EB1~EB5 — `75_mut.py` 의 헬퍼와 규율을 빌린다 (7.5.15) | `A122_HARNESS_WORK=<ext4> python3 7515_mut.py` |
| `7516_order.py` | 기록 명령이 "움직인 입력" 과 "그 편집이 부른 거절" 중 무엇을 먼저 말하는가 — 같은 모양을 `5a54f78d` · `1d1e5ca7` · 워킹트리의 코드로 (7.5.16) | `python3 7516_order.py [<rev> ...]` |
| `759r_resolver_ab.py` | 보수 P1-2(아카이브 이름 먼저)의 해소기 A/B — 활성 + 아카이브 change id 전수, 기준 리비전의 세 모듈 대 워킹트리 | `python3 759r_resolver_ab.py [<before-rev>]` |

A/B 셋의 이어 달리기 기록(`_work/*_done.*.json`)은 **양쪽 소스**에 묶는다 — 기준만으로 묶으면 워킹트리가
바뀐 뒤 옛 after 와 새 after 의 줄이 한 표에 섞인다(7.5.2 에서 `751`, 7.5.2.1 에서 `75_ab` 를 고쳤다).

## 읽는 법 두 가지

**`75_ab.py` 의 비교 기준은 인자다.** 기본값은 7.5 직전 커밋(`8091e6c4`)이다. `HEAD` 로
굳혀 두면 수리가 랜딩한 다음 날 "before" 사본에 수리가 들어가 대조군이 조용히 오염된다 —
그래서 사본에 `_committed_many` 가 있으면 **멈춘다**
([[a-recorded-boundary-stops-being-rechecked]]). 한 번에 다 못 도므로(before 쪽 합계 약 2000초)
초 예산을 주고 여러 번 부르면 `_work/75_ab_done.<기준 12자리>.<after 소스 해시 12자리>.json` 에 이어 달린다(위 문단).

**하네스는 도구를 따라 낡는다.** `75_census.py` 는 `_walk_floor` 가 7.5 에서 값 셋을 돌려주게
되자 전수 126건을 전부 "못 걸음"으로 찍었다 — 빈 결과가 발견처럼 보였다
([[missing-tool-reports-clean]]: 0 은 "위반 0"이 아니라 "검사 0"이다). 여기 있는 것들을
쓰기 전에 **작은 표본 하나로 먼저 돌려 보고** 숫자가 말이 되는지 확인할 것. 7.5.2.2 에서도 같았다: 증거
읽기가 `_read_regular`(`os.open`)로 옮기자 `Path.read_bytes` 만 세던 `7521_main_ab.py` · `752_reads.py` 의
`ast.json` 읽기 계수가 **0** 을 찍었다(표본 `48 → 0`) — 두 곳 다 그 길도 세게 고쳤다.

**앵커는 소스를 따라 낡는다.** 7.5.2.2 에서 수리가 한 줄을 두 줄로 바꾸자 변이 셋의 앵커가 `0회` 가 됐다 —
하네스는 그때 `assert` 로 멈춘다(조용히 건너뛰지 않는다). 돌리기 전에 앵커 수를 한 번 세어 볼 것:

```python
ca = Path("tools/logic-map/check_analysis.py").read_text()
[(n, ca.count(o)) for n, e in MUTATIONS.items() for o, _ in e if ca.count(o) != 1]
```

그리고 **소스를 고치면 A/B 와 변이를 다시 돌린다** — 둘 다 그 소스의 증거이지 이름의 증거가 아니다.

**`resolve_base` 에 `head=` 가 필수가 된 뒤 (6.4(b), 2026-09-26)** `75_ab.py` · `75_attrib.py` · `75_census.py` · `75_time.py` 는
각자 `head=<모듈>._head_commit(ROOT)` 를 넘기게 한 줄씩 고쳤다(배관만 — 무엇을 재는지는 안 바뀌었다). 안 고치면 네 스크립트가
`TypeError` 로 전부 "base 못 풂" 을 찍는다 — 위 "하네스는 도구를 따라 낡는다" 의 한 예다.

**작업 자리 (6.4 보수).** `75_mut.py` · `64_gate_mut.py` 는 `A122_HARNESS_WORK=<디렉터리>` 가 있으면 사본과 전용 `GOCACHE` 를 거기에
둔다(없으면 `_work/`). 이 저장소는 /mnt/D(ntfs-3g) 위라 `_work/` 에서는 스위트 한 판이 8~10 분이었다 — ext4 스크래치에서 잰
판의 시간은 review.md 6.4 보수 절에 있다. 사본 · 캐시 자리만 바뀌고 판정 규율(무변이 대조군 창 양끝 · 사본에 pid · 한 번에 한 판)은 같다.

**앵커 재조준 (7.5.9 · 7.5.10 · 7.5.16, 2026-09-26).** `75_mut.py` 의 AA3 · AA7 · AA19 는 겨누던 줄이 바뀌어 앵커가 0 회가 됐다 — 같은 뜻의
새 줄로 다시 걸었다(원장이 트리 목록을 잊는다 → 추적 목록을 잊는다 · 기록 직전 거절을 다시 안 묻는다 · 목록 지문에서 종류가 빠진다). 이 로트의
새 변이는 AM1~AM9 · AN1~AN3 · AO1~AO3 · AP1 · AQ1~AQ2(창 `228:246`)이고 `execution_baseline.py` 쪽은 `7515_mut.py` 다.

**`7521_main_ab.py` 의 기준 사본은 `check_analysis.py` 만 그 리비전이다 (보수 정정, 2026-09-26).** `build()` 가 다른 모듈(`execution_baseline.py` ·
`role_check.py`)은 워킹트리에서 복사한다 — 그 두 파일을 바꾼 로트의 A/B 는 "before `<rev>`" 가 아니다. 세 모듈을 다 그 리비전에서 꺼내는 것은
`759_ab_verify.py` · `759r_resolver_ab.py` · `7516_order.py` 다. 보수 변이 창은 `246:253`(MZ1 · MR1~MR6)이다.
