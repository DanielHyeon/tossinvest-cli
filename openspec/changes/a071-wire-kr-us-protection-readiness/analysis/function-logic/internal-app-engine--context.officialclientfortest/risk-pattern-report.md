# Risk Pattern Report: `Context.OfficialClientForTest`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/app/engine/export_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none | test-only accessor | reviewed-safe | absent from production build |
