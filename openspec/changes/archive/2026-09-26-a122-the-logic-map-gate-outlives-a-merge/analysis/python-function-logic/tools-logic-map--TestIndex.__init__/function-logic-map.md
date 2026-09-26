# Function Logic Map: `TestIndex.__init__` (Python, a122 task 7.5.9 — 새 메서드, 보수에서 번들 추가)

## 새 함수 (2026-09-26)

편집 후 `tools/logic-map/check_analysis.py:2246-2248` · 분기 0 · 반환 0 · raise 0 · 호출 2 (`ast.after-759r.json`, source sha `ff590b79db6e`)

| id | 줄 | 종류 | 소스 |
|---|---|---|---|

`TestIndex(dict)` 의 생성자 — `files`(색인이 읽은 추적 `*_test.go` 의 `frozenset`)를 붙인다. 분기 0. 7.5.9 첫 판이 FLM 을 빠뜨렸다(독립 주장정확성 리뷰 F10). `enumerate.py` 가 `Class.method` 를 받게 했다 — 이름만으로 찾으면 파일의 **첫** `__init__`(`ReadLedger`)이 잡힌다.
