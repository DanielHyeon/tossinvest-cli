# Risk Pattern Report: `Console.strategyRuntimeSummary`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/console/settings_tabs.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| unrelated float casts | `settings_tabs.go:280-345` | reviewed-safe/not-applicable | target summary performs no numeric accounting |
