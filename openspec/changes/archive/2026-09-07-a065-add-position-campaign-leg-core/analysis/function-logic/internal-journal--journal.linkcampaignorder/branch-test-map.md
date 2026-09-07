# Branch Test Map: `Journal.LinkCampaignOrder`

2026-09-07 (task 6.4): 이 함수를 편집해 분기가 45 → 46 개가 됐다. 재번호는 옛/새 `ast.json`
의 분기 열을 difflib 으로 **정렬해서** 했다. 측정 결과: B1–B11·B16–B34 동일, 옛 B12–B15
(EXIT FIRST 거절 갈래 넷)가 같은 자리의 새 넷으로 교체됐고(거절이 이제 증거를 남기고
commit 한다), 750 행에 successor 링크 remaining 계산 갈래가 하나 삽입돼 옛 B35–B45 가
새 B36–B46 으로 밀렸다.

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
| B12 | `campaignExposureBlockedInTx` 조회 오류 (661) | EXIT FIRST tests | yes | yes |
| B13 | `exposureBlocked || blocked` — EXIT FIRST 거절 (664) | `TestExposureRefusalAtLinkLeavesRefusalEvidence` | yes | yes |
| B14 | 거절 증거 latch 실패 (665) | `TestExposureRefusalAtLinkLeavesRefusalEvidence` | yes | yes |
| B15 | 거절 commit 실패 (668) | `TestExposureRefusalAtLinkLeavesRefusalEvidence` | yes | yes |
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
| B35 | successor 링크 시점 remaining(`StoredOrderRemaining`) 계산 실패 (750) | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
| B36 | campaign update | link tests | yes | yes |
| B37 | command append | replay tests | yes | yes |
| B38 | event append | replay tests | yes | yes |
| B39 | commit | crash tests | yes | yes |
| B40 | final order read | link tests | yes | yes |
| B41 | result version projection | link tests | yes | yes |
| B42 | authoritative replacement edge | successor tests | yes | yes |
| B43 | caller ambiguity durably refused and digest-bound | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/caller_ambiguity_is_durable_and_in_command_digest` | yes | yes |
| B44 | initial intent quantity differs from requested cap | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/intent_quantity_mismatch` | yes | yes |
| B45 | replacement edge quantity differs from requested cap | `TestLinkCampaignOrderAmbiguityAndQuantityAuthorityLatch/replacement_edge_quantity_mismatch` | yes | yes |
| B46 | successful successor remaining derives from successor cap | `TestApplyPositionCampaignFillPreservesLatePredecessorExactlyOnce` | yes | yes |
