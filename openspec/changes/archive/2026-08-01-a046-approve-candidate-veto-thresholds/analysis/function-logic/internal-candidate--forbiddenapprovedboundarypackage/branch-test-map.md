# Branch Test Map: `forbiddenApprovedBoundaryPackage`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all authoritative roots (`execgw`, `orderintent`, `orderlineage`, `journal`, `domain`, `risk`, `app/engine`, and the remaining isolation roots) are rejected | `TestApprovedCandidateBoundaryDetectsAllAuthorityRoots` | compile RED: authority helper/bridge not yet implemented; prior implementation also omitted `journal` and other roots | focused GREEN |
| B2 | an unrelated pure package is not classified as an authority root | `TestApprovedCandidateBoundaryDetectsAllAuthorityRoots` | same focused compile RED | focused GREEN |
