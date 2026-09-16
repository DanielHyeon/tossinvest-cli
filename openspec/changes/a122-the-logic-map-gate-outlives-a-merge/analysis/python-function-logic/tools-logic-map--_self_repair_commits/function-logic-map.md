# Function Logic Map: `_self_repair_commits`

`tools/logic-map/check_analysis.py:601-691` · Python · **task 7.2.6 (H4)** ·
분기 15 · 반환 2 · raise 3 (`ast.after-7.2.6.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** 아래 표는 `enumerate.py` 의 열거를 그대로 옮긴 것이다.
> 7.2.6 이 새로 만든 함수라 편집 전 판본이 없다.

## Inputs and invariants

`(root, analysis)` 를 받아 **이 change 자신의 나중 Go 작업** 커밋들을 오래된 것부터 돌려준다.
세는 조건은 하나다: 그 change 의 디렉터리를 만지면서 Go 파일도 고친 **비병합** 커밋(HEAD 까지).
판정하지 않는다 — 어느 후보를 거절할지는 `_landing_refusal` 이 정한다. 예외를 만들지 않고,
못 세는 경우(저장소 밖 증거 · git 실패 · 만진 커밋 0)는 **빈 목록**이다.

`floor` 와 같은 모양으로 호출자가 **한 번** 재서 `_landing_refusal` 에 넘긴다. 두 호출자
(`resolve_landing` · `compute_landing`)가 각자 조건을 들고 있으면 갈린다
([[two-judgements-cover-for-each-other]]).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | L628 | Try | `try:` |
| B2 | L630 | ExceptHandler | `except ValueError as exc:` |
| B3 | L637 | If | `if change_dir.startswith(ARCHIVE_PREFIX):` |
| B4 | L641 | If | `if before:` |
| B5 | L654 | If | `if touching.returncode:` |
| B6 | L660 | IfExp | `touching.stderr.strip().splitlines()[0] if … else 'git log failed'` |
| B7 | L663 | If | `if not hashes:` |
| B8 | L679 | If | `if listing.returncode:` |
| B9 | L682 | IfExp | `listing.stderr.strip().splitlines()[0] if … else 'git log --no-walk failed'` |
| B10 | L685 | For | `for block in listing.stdout.split('\x00'):` |
| B11 | L686 | comprehension | ` for line in block.splitlines() if line` |
| B12 | L687 | BoolOp | `lines and any(name.endswith('.go') for name in lines[1:])` |
| B13 | L687 | If | `if lines and any(name.endswith('.go') for name in lines[1:]):` |
| B14 | L687 | comprehension | ` for name in lines[1:]` |
| B15 | L691 | comprehension | ` for commit in reversed(hashes) if commit in flagged` |

| 반환 줄 | 값 |
|---|---|
| L664 | `[]` — 그 디렉터리를 만진 비병합 커밋이 **없다**(정상) |
| L691 | `[commit for commit in reversed(hashes) if commit in flagged]` |

| raise 줄 | 무엇을 못 쟀나 |
|---|---|
| L635 | change 디렉터리를 저장소 안 경로로 못 적는다 (오늘 도달 불가) |
| L658 | 디렉터리를 만진 커밋 목록 조회 실패 |
| L680 | 그 커밋들의 파일 목록 조회 실패 |

## Calls and live bindings

`subprocess.run` ×2 (`git log`) · `_archived_change_id` · `ARCHIVE_PREFIX` ·
`analysis.parent.parent.relative_to` · 순수 문자열 처리. 시각·환경·워킹트리를 안 읽는다 —
HEAD 까지의 커밋 역사만 본다.

**판정은 사람의 git 설정의 함수가 아니다** — 세 가지를 명령줄에서 못 박는다
(2026-09-16 적대 리뷰 F1 · F4 · F5):

| 못 박은 것 | 안 박으면 |
|---|---|
| `--full-history` | 경로 조회의 기본 단순화가 병합 반대편 가지의 자기 수리를 **통째로 버린다**(실측 재현). 오늘 이 저장소 126건에서 더 세는 커밋은 **0** 이다 — 지금 넣으면 공짜다 |
| `-c diff.renames=false` | `.go` 를 비-`.go` 이름으로 옮긴 커밋의 옛 이름이 사라져 깃발이 안 선다 |
| `-c core.quotePath=false` | 비ASCII 이름이 인용돼 `.go` 로 안 끝난다 |

셋 다 **안 박으면 깃발이 덜 서는**(= 창이 안 넓어지는) 방향이다.

**왜 조회가 둘인가.** 경로 제한을 건 `git log` 는 **그 경로의 파일만** 적어 주므로
"이 커밋이 Go 도 고쳤나"를 같은 조회로 못 본다. 그래서 디렉터리를 만진 커밋을 먼저 고르고,
`--no-walk --stdin` 한 번으로 그 커밋들의 **전체** 목록을 읽는다. 커밋마다 `git diff-tree` 를
돌리는 측정 판본과 착지 있는 68건 전수에서 깃발 224개 · 불일치 0 이다(`h4b_ab_flagged.py`).

## State mutations and fallbacks

없다. 읽기만 한다. git 이 **실패하면 판정이다**(`RuntimeError` → 경계가 오류 줄로 바꾼다).
빈 목록으로 물러나지 않는다 — 빈 목록은 "거절할 것이 없다"라서 실패가 **위반 0** 으로 읽히고
가드가 조용히 꺼진다([[missing-tool-reports-clean]], 2026-09-16 적대 리뷰 F2: 주입 실험에서
rc 1 하나로 거절돼야 할 입력이 초록이었다). 옆의 `_evidence_floor` 도 실패하면 거절로 간다 —
두 자리가 같은 방향이어야 한다. **빈 목록은 하나뿐이다**: 그 디렉터리를 만진 커밋이 정말 없을 때.

## Safety conclusion

생산 Go 코드 변경 0. 신원(`openspec/changes/<id>/`)을 읽지만 **거절에만** 쓴다 — 1.12 가 막은
것은 신원으로 착지를 받아들이는 판정이다. 오늘 저장소에서 기록을 가진 change 는 a099 하나이고
네 시점 전부 값이 같다(2026-09-16 전수).

**한계 둘** [[fail-closed-must-name-what-it-rejects]]:
- Go 수리와 문서를 **다른 커밋**으로 쪼개면 깃발이 안 선다 — 망각 가드이지 위조 가드가 아니다.
- 병합 커밋 **자신의** 변경은 안 읽힌다(`--no-merges`, 그리고 `git log` 의 기본이 병합 diff 를
  안 낸다 — git 2.43 에서 어떤 설정으로도 안 바뀐다). 7.2.4 의 H7 과 같은 부류다.

두 한계는 **놓치는** 쪽이므로 창이 넓어지지 않는다 — 그 자리는 이 규칙 이전 상태로 남는다.
거절하는 쪽으로 틀리면(넓은 교차 커밋) 창이 넓어진다. 방향이 둘이고 다르다는 것을 spec 이 적는다.
