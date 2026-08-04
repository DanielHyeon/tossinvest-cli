# Risk Pattern Report: `brokerCapabilityDigest`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/types.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| hash field omission risk | `brokerCapabilityDigest` | reviewed-safe | all capability fields enumerated and substitution tested |
