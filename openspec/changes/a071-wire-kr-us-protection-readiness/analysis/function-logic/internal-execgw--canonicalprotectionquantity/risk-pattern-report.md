# Risk Pattern Report: `canonicalProtectionQuantity`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/execgw/protection.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| float-to-integer authority boundary | `canonicalProtectionQuantity` | reviewed-safe | canonical round-trip and 2^53-1 ceiling |
