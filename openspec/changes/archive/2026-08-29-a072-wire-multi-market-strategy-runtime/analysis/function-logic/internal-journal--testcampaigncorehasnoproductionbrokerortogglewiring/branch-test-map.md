# Branch Test Map: `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority`

> base 시점 이름은 TestCampaignCoreHasNoProductionBrokerOrToggleWiring 였다. a073 커밋 `8022f578` 이 지금 이름으로
> 바꿨고, `ast.json` 은 `revision: base` 라 옛 이름을 그대로 둔다. 이 표가
> 가리키는 **덮는 테스트**는 현재 트리에 실제로 있는 이름이어야 한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if branch at `internal/journal/position_campaign_hardening_test.go:447` preserves its exact condition and failure path | named `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority` regression plus affected-package suite | source/base behavior recorded | verification required |
| B2 | if branch at `internal/journal/position_campaign_hardening_test.go:451` preserves its exact condition and failure path | named `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority` regression plus affected-package suite | source/base behavior recorded | verification required |
| B3 | if branch at `internal/journal/position_campaign_hardening_test.go:455` preserves its exact condition and failure path | named `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority` regression plus affected-package suite | source/base behavior recorded | verification required |
| B4 | range branch at `internal/journal/position_campaign_hardening_test.go:458` preserves its exact condition and failure path | named `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority` regression plus affected-package suite | source/base behavior recorded | verification required |
| B5 | if branch at `internal/journal/position_campaign_hardening_test.go:459` preserves its exact condition and failure path | named `TestCampaignCoreProductionWiringHasNoBrokerOrToggleAuthority` regression plus affected-package suite | source/base behavior recorded | verification required |
