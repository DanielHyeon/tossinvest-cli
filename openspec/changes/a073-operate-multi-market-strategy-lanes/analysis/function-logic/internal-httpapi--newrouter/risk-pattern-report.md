# Risk Pattern Report: `NewRouter`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/httpapi/router.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none in target function | `router.go:31` | reviewed-safe | fixed capability construction only |
