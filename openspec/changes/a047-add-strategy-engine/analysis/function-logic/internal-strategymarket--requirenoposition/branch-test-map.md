# Branch Test Map: `RequireNoPosition`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | nil/error/stale/nonzero/order/canonical zero | `TestNoPositionProofRequiresFreshAuthoritativeZeroPositionAndOrders` | existing | yes |
| B2 | wrong symbol/caller-claimed source | same test rows | yes | yes |
| B3 | exact official zero pass | same test pass case | yes | yes |
| B4 | nonzero/canonicalization/open-order refusal | same exposure rows | existing | yes |
