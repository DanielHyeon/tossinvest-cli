# Branch Test Map: `_safe_changed_go_paths` (a122 task 7.5.22)

편집 **후** 분기 14 (`ast.after-7.5.22.json`) + 헬퍼 `_numstat_records` 분기 6.
시험은 전부 `tools/logic-map/test_check_analysis.py` 의
`AGoFileWithNoTextualDiffIsNotSilentlyEmpty` 이고, 픽스처는 **진짜 저장소 · 진짜 git** 이다.

## 이 로트가 새로 못 박은 갈래

| 분기 | 줄 | 무엇 | 시험 |
|---|---|---|---|
| B12 | 174 | 이름을 다 본 뒤 레코드를 다시 돈다 | `test_a_suppressed_body_is_refused_instead_of_silently_requiring_nothing` |
| B13 · B14 | 175 | `-`/`-` 면 거절 (`binary`) | 같은 시험 + `test_the_minus_diff_attribute_suppresses_the_body_too` |
| B13 · B14 (거짓 갈래) | 175 | `0`/`0` 은 **안** 거절 | `test_a_mode_only_change_is_not_refused` — **거절의 경계다** |
| B6 · B7 | 164 · 165 | rename 의 **양쪽** 이름 | `test_the_guard_reads_both_names_of_a_rename` |
| `_numstat_records` B4 | 123 | 평범한 레코드 대 rename 레코드 | 위 둘이 양쪽 갈래를 각각 돈다 |

## 픽스처가 허구가 아님을 먼저 못 박는다

`test_git_really_suppresses_the_body_for_a_binary_marked_go_file` 이 **양성 대조군**이다 —
git 이 정말 `Binary files` 를 내고 `@@` 를 **안** 내며 `--numstat` 이 `-\t-\t` 를 낸다는 것을
단언한다. 이것이 틀리면 아래 넷은 아무것도 재지 않는다
([[mutation-must-reach-the-thing-under-test]]).

`test_an_uncommitted_gitattributes_is_enough_to_suppress_the_body` 는 픽스처의
`.gitattributes` 가 정말 **추적되지 않는지**(`git status --short` 가 `??`)를 단언한 뒤에 거절을 잰다.
그 단언이 없으면 "추적 안 해도 된다" 는 주장이 시험에 없다.

## RED (수리 전, `197a0355`)

`Ran 6 tests · FAILED (failures=4)`. 통과한 둘은 **통과해야 하는 둘**이다:
양성 대조군(문법 사실)과 mode-only 경계(거절하면 안 되는 정상 입력).
rename 시험의 편집 전 실패 메시지는
`cannot load existing base file <sha>:"a/a\tb.go"` — 7.5.13 이 적은 **거짓 차단**이 그대로 나왔다.

## 앞 로트의 갈래 (이 로트가 안 건드린 것)

| 분기 | 줄 | 무엇 | 어디서 |
|---|---|---|---|
| B2 | 159 | git 이 죽으면 결함 | 앞 로트부터 |
| B8 · B9 | 166 · 168 | UTF-8 아닌 이름 | 앞 로트부터 |
| B10 · B11 | 172 | `\n` · `\r` · `\t` 이름 | 앞 로트부터 + 이제 rename 옛 이름으로도 닿는다 |
| B1 · B3 · B4 · B5 | 154~162 | 인자 조립 · bytes/str 방어 | 행동 시험 없음(앞 로트와 같음) |

## 안 닫은 것

- `_numstat_records` 의 결함 갈래 둘(`cannot read git diff --numstat record` · `… rename record`)에
  행동 시험이 없다. 오늘 그 모양을 내는 git 입력을 못 만들었다(충돌 중 워킹트리도 평범한 3칸이었다).
  **지어낸 바이트로 시험을 쓰면 그것은 git 의 증거가 아니라 내 상상의 증거다** — 안 쓴다.
- `--find-renames` 를 가드 호출에 안 넣었다. `diff.renames` 기본이 참이라 오늘은 같은 집합을 보고,
  거짓이면 가드가 rename 을 **삭제+추가 둘**로 보므로 이름을 **더** 본다(초집합이라 안전).
