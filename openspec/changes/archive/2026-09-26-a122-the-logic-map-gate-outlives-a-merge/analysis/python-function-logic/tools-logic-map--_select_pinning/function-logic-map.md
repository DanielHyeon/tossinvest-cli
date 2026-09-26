# Function Logic Map: `_select_pinning` (Python, a122 task 7.5.2)

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:626-649` · 분기 5 · 반환 1 · raise 0 (새 함수).

옛 `_pinning_bundles` 의 선별 본문 그대로 — 차이는 디스크를 읽지 않고 **이미 읽은 바이트**를 `_parsed` 로 푼다는 것 하나다. 지문과 판정 목록이 같은 읽기에서 나오게 하려는 것이다(Codex P1 재현: 두 번 읽던 판본은 둘째 읽기의 실패로 번들을 판정에서 뺀 채 두 지문이 같다고 답했다).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 635 | For | `for ast_path, raw in reads:` |
| B2 | 637 | BoolOp | `not isinstance(value, dict) or value.get('revision', 'current') != 'current'` |
| B3 | 637 | If | `if not isinstance(value, dict) or value.get('revision', 'current') != 'current':` |
| B4 | 641 | BoolOp | `not raw_source or not digest` |
| B5 | 641 | If | `if not raw_source or not digest:` |

## task 7.5.2.1 — 판정이 읽는 입력을 전부 세고 하나씩 묶는다 (7.5.2 재리뷰, 2026-09-20)

`tools/logic-map/check_analysis.py:673-696` · 분기 4 · 반환 1 · raise 0 (편집 전 L626-649 · 분기 5 · 반환 1 · raise 0, `ast.before-7.5.2.1.json` = revision `fc35eb2d`).

`Evidence` 를 받는다(바이트 목록 대신). `_parsed` 가 이제 언제나 사전을 돌려주므로 사전인지 묻는 갈래가 빠졌다(분기 5 → 4).

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 682 | For | `for ast_path, raw in evidence.held.items():` |
| B2 | 684 | If | `if value.get('revision', 'current') != 'current':` |
| B3 | 688 | BoolOp | `not raw_source or not digest` |
| B4 | 688 | If | `if not raw_source or not digest:` |
