# Risk Pattern Report: `NewReadinessAdapter`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protection/readiness_adapter.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| implicit supervisor defaults | adapter had no sealed supervisor contract | defect | production constructor must require paired exact contracts; default adapter remains UNWIRED only |
