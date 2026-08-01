# Branch Test Map: `scanPositionPolicyScope`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | row scan failure is returned | position policy query suite | yes | yes |
| B2 | durable exit state defaults scope to MANAGED | `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration` | yes | yes |
| B3 | persisted lifecycle generation/version/status overrides the virtual row | `TestPositionPolicyOverrideCASAndAuditAreAtomic` | yes | yes |
| B4 | RELEASED projection hides effective exit policy | `TestPositionPolicyReleaseAndReadoptCreateFreshGeneration` | yes | yes |

Typed external-adoption versus engine-entry provenance is asserted by `TestPositionPolicyReleaseAndReadoptRequireExternalAdoptionAtJournalBoundary`.
