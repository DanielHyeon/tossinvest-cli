# Risk Pattern Report: `TestAmbientMotionContract`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/console/display_primitives_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| `go-float64-cast` | `display_primitives_test.go:702` | not-applicable — mapped function은 640행에서 끝나며 RGB 검사 보조 함수의 정수 채널 변환임 | `function-logic-map.md` |
