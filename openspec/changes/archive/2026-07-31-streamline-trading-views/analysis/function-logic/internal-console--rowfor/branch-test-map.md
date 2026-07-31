# Branch Test Map: `rowFor`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | requested symbol is absent | existing rowFor callers fail immediately | yes — old row lookup could not isolate new row markup | yes |
| B2 | symbol marker is outside a row | helper malformed-markup guard | n/a structural guard | yes by code inspection |
| B3 | marker is in a closed primary row | portfolio label/adoption/exclusion tests | yes — old helper returned the remainder of the page | yes |
