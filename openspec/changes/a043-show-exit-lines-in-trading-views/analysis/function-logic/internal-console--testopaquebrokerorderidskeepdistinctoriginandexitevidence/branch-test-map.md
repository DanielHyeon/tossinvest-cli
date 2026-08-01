# Branch Test Map: `TestOpaqueBrokerOrderIDsKeepDistinctOriginAndExitEvidence`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | whitespace-distinct rows remain visible and counted | this function | one row survived RED | yes |
| B2 | exact row collection | this function | spaced ID was trimmed in evidence key | yes |
| B3 | spaced row owns engine origin and exit intent | this function | RED labelled it other/unlinked | yes |
| B4 | unspaced row remains other and unlinked | this function | RED row was hidden | yes |
