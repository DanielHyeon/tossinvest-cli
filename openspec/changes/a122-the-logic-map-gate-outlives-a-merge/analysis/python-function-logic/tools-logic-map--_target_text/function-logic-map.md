# Function Logic Map: `_target_text`

`tools/logic-map/check_analysis.py:353-361` · Python · 기존 함수, **편집 없음**.
task 6.3 은 시험만 바꾸지만 그 review 가 이 함수의 **갈래**를 근거로 쓰므로 `ast.json` 을 먼저
뽑았다(`enumerate.py`, 기계 열거).

## Inputs and invariants

`landing`(비교 대상 쪽 끝, 빈 문자열이면 워킹트리) · `audited`(이관 감사로 고정됐는가).
창 줄에 넣을 사람의 말을 돌려준다. 불변식: **무엇이 그 끝을 고정했는지**까지 말한다 —
이관의 끝을 착지 기록과 같은 말로 적으면 있지도 않은 파일을 가리킨다.

## Branches (분기 2 · 반환 2)

| ID | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 359 | If | `if not landing:` |
| B2 | 361 | IfExp | `f'audited source-commit {landing}' if audited else f'landed-commit {landing}'` |

반환: `360` 'working tree (no landed-commit.txt)' · `361` f'audited source-commit {landing}' if audited else

## 갈래마다 **라벨까지** 보는 시험 (task 6.3 뒤)

| 갈래 | 출력 | 시험 |
|---|---|---|
| `not landing` | `working tree (no landed-commit.txt)` | `test_a_failure_prints_the_comparison_window` |
| `audited` 참 | `audited source-commit <sha>` | `test_the_adoption_window_ends_at_the_audited_source_commit` · `test_an_archived_adoption_is_rechecked_by_its_id` |
| `audited` 거짓 | `landed-commit <sha>` | `test_empty_required_with_bundles_is_reported_and_still_passes` — **6.3 전에는 SHA 만 봤다** |

변이 T1(거짓 갈래를 감사 라벨로 바꿈): 6.3 전 SURVIVED → 후 CAUGHT.

## Calls · State · Safety

호출 없음(f-string 뿐), 상태 변경 없음. 출력 문자열만 만들고 게이트 판정에 들어가지 않는다.
주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. Go 변경 0.

## 편집 — task 7.7 (조언이 기록 명령의 판정에 묻는다) · 분기 2 → 2 · 반환 2 → 2 · raise 0 → 0

편집 전 `ast.before-7.7.json`(HEAD `3da639a9` blob) · 편집 후 `ast.after-7.7.json`(워킹트리, L373-384). 아래 표는 두 열거의 `source` 를 스크립트가 줄 단위로 대조한 것이다.

워킹트리 대상 문장을 `(no landed-commit.txt)` → `(no landed-commit.txt in HEAD)` 로. 게이트는 기록을 **커밋에서** 읽으므로 기록이 디스크에만 있을 때 옛 문장은 거짓이었고, 바로 옆 조언 줄이 권한 `--record-landing` 은 그 파일이 "이미 있다"고 거절했다(리뷰 I8). 분기 구조는 그대로다. 이 문장은 `check` 의 `missing Function Logic Map … between base X and <대상>` 오류에도 들어가므로 그 오류 **문자열**이 바뀐다 — 판정 A/B 가 그 치환 하나만 정규화해 비교한다(review.md `## VERIFY — task 7.7`).

| | 종류 | 소스 |
|---|---|---|
| 사라짐 | 반환 | `'working tree (no landed-commit.txt)'` |
| 생김 | 반환 | `f'working tree (no {LANDING_FILE} in HEAD)'` |

호출 — 사라짐 없음 · 생김 없음
