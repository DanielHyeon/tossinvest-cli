# Risk Pattern Report: `NewContext`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/app/engine/engine.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| raw mutator exposure | supervisor owns official conditional gateway | reviewed-safe | keep pointer unexported and expose only execgw Gateway |
