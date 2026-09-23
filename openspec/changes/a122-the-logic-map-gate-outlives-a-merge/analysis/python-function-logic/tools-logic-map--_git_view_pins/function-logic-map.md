# Function Logic Map: `_git_view_pins` (Python, a122 task 7.5.27, 새 함수)

`ast.after-7.5.27.json` 이 범위를 적는다 — 산문에는 절대 줄을 안 쓴다(`7523_coords.py` 가 지문을 검증한다).

## 답하는 질문

**"판정의 두 `git diff` 가 워킹트리를 다시 쓰는 설정 없이 돌려면 명령줄에 무엇을 줘야 하나."**
반환은 `["-c", "core.fsmonitor=false", "-c", "filter.<d>.clean=", …]` 이고 호출자가 두 호출에 **같이** 펼친다.

## 갈래

| 갈래 | 무엇 | 왜 |
|---|---|---|
| `git config -z --name-only --get-regexp ^filter\.` 의 rc 가 0·1 이 아님 | 결함으로 올린다 | 1 은 "맞는 키 없음" 이라 정상 |
| 키가 `filter.` 로 시작하지 않거나 점이 둘 미만 | 건너뛴다 | 드라이버 키가 아니다 |
| 드라이버 이름 = `filter.` 뒤, **마지막** `.키` 앞 | 모은다 | 이름에 점이 들어갈 수 있다(`filter.a.b.clean` → `a.b`) |
| 이름에 `=`·공백 | 거절 | `-c name=value` 는 첫 `=` 에서 가른다 — 정확히 못 적으면 중화가 조용히 실패한다 |
| 드라이버마다 `clean=` · `process=` · `required=false` | 붙인다 | `clean` 만 비우면 `required=true` 에서 git 이 rc 128 로 죽는다(실측) |

`core.fsmonitor=false` 는 늘 붙는다 — 거짓말하는 fsmonitor 가 편집을 감춘다(실측).

## 이것이 닫지 **않는** 것

아직 모르는 워킹트리 재작성 문. 이 함수는 **알려진** 설정 문을 끈다 — class 는 사람 결정 7.5.25 다.
그리고 **정상적인** clean 필터(git-lfs · git-crypt)도 끈다 — 그런 저장소에서는 판정이 raw 바이트로 바뀐다
(오늘 노출 0).

시험: `TheWorktreeIsNotRewrittenUnderTheGate` 의 **여섯** — 행동 넷(clean · process(진짜 v2 프로토콜) ·
required+점 이름 · fsmonitor)과 계약 둘(`…every_configured_driver_gets_all_three_pins` ·
`…driver_name_that_cannot_be_pinned_is_refused`). `clean=` 을 못 박는 것은 **계약 시험 하나뿐**이다 — 행동 시험은
git 의 우연으로 초록이 된다. 변이 `AF1~AF5`. (7.5.28 정정: 앞 판본은 "넷" 이라 적고 계약 둘을 빠뜨렸다.)

## task 7.5.28

`core.checkStat=default` · `core.trustctime=true` 를 더했다 — `checkStat=minimal` 이 같은 크기 편집을 감춘다
(7.5.27 의 "재현 안 됨" 은 mtime 을 과거로 안 돌린 픽스처 탓의 **거짓**이었다). 드라이버 이름은 `=` 만 거절한다 —
공백은 `-c` 로 적힌다(7.5.27 의 공백 거절은 헛거절이었다). 시험: `…minimal_stat_check…` · `…driver_name_with_a_space…`.
변이 `AG2` · `AG7`.
