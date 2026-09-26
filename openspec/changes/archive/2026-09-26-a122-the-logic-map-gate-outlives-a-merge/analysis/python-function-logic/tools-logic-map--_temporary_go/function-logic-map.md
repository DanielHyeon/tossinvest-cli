# Function Logic Map: `_temporary_go` (Python, a122 task 7.5.25 — 새 함수)

`ast.after-7.5.25.json` — 분기 0. 스냅숏의 바이트를 `go_functions` 가 읽을 임시 `.go` 로 쓴다. 호출자(`flush`)가
`finally` 에서 지운다. `base_file` 의 같은 다섯 줄을 뽑아 합치지 않은 까닭: `base_file` 내부를 바꾸면 그 FLM 이
필요하고, 이 로트의 범위 밖이다.
