# Function Logic Map: `validate_target`

> 손으로 읽어서 만들지 않았다 — `enumerate.py` 열거.

## Inputs and invariants

미리 읽은 것을 받는다. 빠진 경로면 예전처럼 한 번 읽는다 — 미리 읽기는 최적화이고 판정은 그것에 기대지 않는다.


## task 7.5.1 — gstack 리뷰의 permissive 결함 수리

`ast.before-7.5.1.json`(= git revision **`b29e1f4e`**, 소스 해시가 그 커밋과 일치)과
`ast.after-7.5.1.json` 을 같은 열거기로 뽑아 **순서 있는 배열**을 difflib 으로 정렬했다
(분기 37 · 반환 5 · raise 0 → 분기 39 · 반환 5 · raise 0).

미리 읽은 것을 받는다. 빠진 경로면 예전처럼 한 번 읽는다 — 미리 읽기는 최적화이고 판정은 그것에 기대지 않는다.

| | 종류 | 소스 |
|---|---|---|
| + | BoolOp | `prefetched is not None and relative in prefetched` |
| + | IfExp | `prefetched[relative] if prefetched is not None and relative in prefetched else _committed_` |
