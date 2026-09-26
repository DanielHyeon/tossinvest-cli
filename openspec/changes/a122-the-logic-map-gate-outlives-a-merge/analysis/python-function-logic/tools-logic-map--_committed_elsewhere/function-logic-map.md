# Function Logic Map: `_committed_elsewhere` (새 잎 함수, a122 task 6.4 보수)

`tools/logic-map/check_analysis.py:820-842` · Python · 분기 5 · 반환 1 · raise 1 (`ast.json` = 워킹트리 `b28398e2`, `../enumerate.py` 가 기계로 열거).

새 함수라 편집 전 판이 없다. `resolve_base` 의 B5(else) 에서만 불린다 — 지금 경로가 `head` 에 없을 때.

## Inputs and invariants

`root` · `head`(명령이 한 번 푼 sha) · `change_id`(요청받은 id) · `relative`(지금 디스크 경로). 돌려주는 것은 `head` 가 **커밋한**
`base-commit.txt` 중 `relative` 가 아닌 같은 id 의 자리들 `(경로, 바이트)`. 디스크를 보지 않는다 — 옮겨 버린 옛 자리는 디스크에 없다.

## Branches

| id | 줄 | 종류 | 소스 | 뜻 |
|---|---|---|---|---|
| B1 | 834 | If | `if listed.returncode:` | `git ls-tree` 가 못 답함 → `RuntimeError`(결함, 빈 아카이브로 읽지 않는다) |
| B2 · B3 | 836 · 838 | comprehension | NUL 로 끊은 항목 · 빈 항목 거름 | `-z` — 이름의 개행이 항목을 쪼개지 않는다 |
| — | 839 | (B2 의 조건) | `_archived_change_id(entry.removeprefix(ARCHIVE_PREFIX)) == change_id` | 날짜를 벗긴 나머지 **전체**가 id — 해독은 한 곳(셋째 정규식 없음) |
| B4 | 841 | comprehension | `place != relative` | 지금 자리는 이미 물었다 |
| B5 | 842 | comprehension | `value is not None` | `head` 에 없는 자리는 뺀다 |

반환 하나(842). raise 하나(835). `_committed_many` 의 결함 여섯은 그대로 위로 간다(`GATE_FAULTS`).

## Calls and live bindings

`subprocess.run(git ls-tree)` · `_first_line` · `os.fsdecode` · `_archived_change_id` · `_committed_many`. 경로 기준은 `root` 다 —
`_committed_bytes` 와 같은 노출(tasks 6.4(k)).

## Safety conclusion

읽기 전용(git 객체만). 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어디에도 닿지 않는다. 실패 방향은 게이트가 **안 열리는** 쪽이다.

## Branch Test Map

변이 `75_mut.py`(AL10 · AL12 · AL13 · AL14 · AL16) — 결과는 `../tools-logic-map--resolve_base/branch-test-map.md` 의 6.4 보수 절과 review.md.
