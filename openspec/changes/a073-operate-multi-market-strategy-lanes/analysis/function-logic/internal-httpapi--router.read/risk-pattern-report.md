# Risk Pattern Report: `router.read`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/httpapi/router.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none in target function | `router.go:114` | reviewed-safe | fixed switch over narrow readers |
