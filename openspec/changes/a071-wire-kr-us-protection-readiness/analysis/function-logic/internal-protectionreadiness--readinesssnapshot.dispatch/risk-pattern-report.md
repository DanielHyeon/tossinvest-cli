# Risk Pattern Report: `ReadinessSnapshot.Dispatch`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/dispatch.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| authority scope truncation | dispatch compared only account/profile/market while signed quantity/order/session/trigger/replace/capability were discarded | defect | scope substitution could authorize unsupported protection; fixed in 3.5 with exact sealed dispatch scope |
