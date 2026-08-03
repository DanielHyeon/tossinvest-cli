# Branch Test Map: `Journal.LinkCampaignOrder`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | unbacked lineage + stable retry/replay | `TestLinkCampaignOrderUsesAuthoritativeAttemptAndUniqueOrderScope` | yes | yes |
| B2 | scoped order reused by two legs | same test | yes | yes |
| B3 | predecessor has two successors | `TestReplacementPredecessorHasOneAuthoritativeSuccessor` | yes | yes |
| B4 | zero cap | `TestCampaignPlanAndOrderCapRejectZeroQuantity` | yes | yes |
| B5 | exact replacement lineage | late-predecessor and cap tests | yes | yes |
| B6 | request validation group | link tests | yes | yes |
| B7 | request validation group | link tests | yes | yes |
| B8 | transaction/retry group | retry tests | yes | yes |
| B9 | transaction/retry group | retry tests | yes | yes |
| B10 | transaction/retry group | retry tests | yes | yes |
| B11 | header/version group | link tests | yes | yes |
| B12 | header/version group | link tests | yes | yes |
| B13 | admission query | EXIT FIRST tests | yes | yes |
| B14 | admission result | EXIT FIRST tests | yes | yes |
| B15 | campaign block | link tests | yes | yes |
| B16 | leg lookup | link tests | yes | yes |
| B17 | immutable intent | lineage tests | yes | yes |
| B18 | authority lookup | lineage tests | yes | yes |
| B19 | refusal latch | refusal replay/retry test | yes | yes |
| B20 | refusal commit | refusal replay/retry test | yes | yes |
| B21 | scoped duplicate query | unique order test | yes | yes |
| B22 | scoped duplicate latch | unique order test | yes | yes |
| B23 | scoped duplicate commit | unique order test | yes | yes |
| B24 | successor count query | successor test | yes | yes |
| B25 | successor conflict latch | successor test | yes | yes |
| B26 | successor conflict commit | successor test | yes | yes |
| B27 | predecessor lookup | replacement tests | yes | yes |
| B28 | predecessor terminal update | replacement tests | yes | yes |
| B29 | leg transition | transition tests | yes | yes |
| B30 | terminal predecessor recovery | late-fill tests | yes | yes |
| B31 | transition refusal | transition tests | yes | yes |
| B32 | campaign transition | transition tests | yes | yes |
| B33 | order insert | link tests | yes | yes |
| B34 | leg update | link tests | yes | yes |
| B35 | campaign update | link tests | yes | yes |
| B36 | command append | replay tests | yes | yes |
| B37 | event append | replay tests | yes | yes |
| B38 | commit | crash tests | yes | yes |
| B39 | final order read | link tests | yes | yes |
| B40 | result version projection | link tests | yes | yes |
| B41 | authoritative replacement edge | successor tests | yes | yes |
| B42 | caller ambiguity durably refused and digest-bound | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/caller_ambiguity_is_durable_and_in_command_digest` | yes | yes |
| B43 | initial intent quantity differs from requested cap | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/intent_quantity_mismatch` | yes | yes |
| B44 | replacement edge quantity differs from requested cap | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/replacement_edge_quantity_mismatch` | yes | yes |
| B45 | successful successor remaining derives from successor cap | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
