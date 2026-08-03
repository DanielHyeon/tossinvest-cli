# Risk Pattern Report: `Console.handleStrategyRuntime`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/console/strategy_runtime.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none in target function | `strategy_runtime.go:82` | reviewed-safe | GET/HEAD guard and plain reader only |
