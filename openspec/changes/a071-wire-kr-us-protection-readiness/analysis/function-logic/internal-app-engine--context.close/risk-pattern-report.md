# Risk Pattern Report: `Context.Close`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/app/engine/engine.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| early-return leak | a second durable store must close on all startup/close exits | defect-prevention | wire explicit ownership and error joining |
