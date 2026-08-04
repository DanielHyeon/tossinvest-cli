# Risk Pattern Report: `TestProductionEngineAssemblesPairedUnwiredReadinessProvider`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/app/engine/a071_readiness_assembly_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none | test-only production assembly assertion | reviewed-safe | no live transport; isolated official stub |
