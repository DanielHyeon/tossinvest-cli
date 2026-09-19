# Function Logic Map: `_evidence_fingerprint`

`tools/logic-map/check_analysis.py:1501-1517` · 분기 3 · 반환 1 · raise 0 · **task 7.5.1** (`ast.after-7.5.1.json` 기계 열거)

> **손으로 읽어서 만들지 않았다.** `enumerate.py` 의 열거를 옮긴 것이다. 7.5.1 이 새로 만든 함수다.

## Inputs and invariants

판정이 딛는 증거의 지문 — 고정 번들 목록과 각 `ast.json` 의 **바이트** sha256. `compute_landing` 이 수락 직전에 이것이 처음 잰 것과 같은지 본다 (Codex P1-2).

## Branches and early returns — 열거 그대로

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1511 | For | `for ast_path, source, digest in _pinning_bundles(root, analysis):` |
| B2 | 1512 | Try | `try:` |
| B3 | 1514 | ExceptHandler | `except OSError:` |

| 반환 줄 | 값 |
|---|---|
| 1517 | `fingerprint` |

## Calls and live bindings

`_pinning_bundles` · `Path.read_bytes` · `hashlib.sha256`.

## State mutations and fallbacks

없다.

## Safety conclusion

하한을 계산하지 않는다 — `_walk_floor` 가 사는 "하한은 한 곳" 규칙과 겹치지 않는다. 못 읽는 `ast.json` 도 지문(빈 문자열)이다: 다음에 읽히면 달라지므로 변화로 잡힌다.

## task 7.5.2 — 판정은 한 번 읽은 바이트로 선다 (수리한 트리의 재리뷰, 2026-09-19)

`tools/logic-map/check_analysis.py:1709-1729` · 분기 1 · 반환 1 · raise 0 (편집 전 L1501-1517 · 분기 3 · 반환 1 · raise 0, `ast.before-7.5.2.json` = revision `e9f905bd`).

지문 = `Fingerprint(head, read, pins)` — 모든 `ast.json` 의 (경로, 바이트 해시) · 고정 목록 · `HEAD`. 그리고 **그 지문을 만든 같은 읽기**의 고정 목록과 바이트를 같이 돌려준다(옛 판본은 지문만 돌려줬고 판정 목록은 `_walk_floor` 가 따로 읽었다). `HEAD` 는 재리뷰 적대 F9: 수리 신호는 걷기 전 역사로 재는데 후보 순회는 그 뒤의 `HEAD` 를 읽는다 — 이 작업트리는 병행 세션이 같이 쓴다.

| id | 줄 | 종류 | 소스 |
|---|---|---|---|
| B1 | 1728 | comprehension | ` for ast_path, source, digest in bundles` |
