# Branch Test Map: `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | iterate capability exemptions | `TestTheSeamsWithVerbExemptionsAreEnumeratedAndArgued` | new seams unlisted | pass |
| B2 | skip seams without exemptions | same | baseline | pass |
| B3 | reject exemption seam missing from expected map | same | updater missing | pass |
| B4 | inspect every exempted name | same | baseline | pass |
| B5 | reject unenumerated name | same | baseline | pass |
| B6 | reject empty justification | same | baseline | pass |
| B7 | inspect expected names | same | baseline | pass |
| B8 | reject stale expected exemption | same | baseline | pass |
| B9 | inspect expected seams | same | baseline | pass |
| B10 | reject expected seam without exemptions | same | baseline | pass |
| B11 | scan every capability | same | baseline | pass |
| B12 | scan every exempt name | same | baseline | pass |
| B13 | reject forbidden spelling outside exemption | same | baseline | pass |
| B14 | reject unnamed unexpected capability | same | baseline | pass |
| B15 | inspect every `Options` field | same | baseline | pass |
| B16 | reject field missing capability entry | same | baseline | pass |
| B17 | reject stale capability entry | same | baseline | pass |
