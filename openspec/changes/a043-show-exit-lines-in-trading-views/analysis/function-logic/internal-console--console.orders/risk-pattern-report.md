# Risk Pattern Report: `Console.orders`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/console/orders.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none | `Console.orders` | reviewed-safe | target contains only bounded in-memory loops and read seams; no panic/go/defer/live mutation finding |
