# Branch Test Map: `consoleBroker.instrumentMetadata`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | request expires while another account read owns the broker gate | `TestA061MetadataResolverHonorsCancellationWhileAccountResolutionOwnsTheBroker` | yes | yes |
| B2 | no client exists, so the account-resolved builder path is taken | `TestA061MetadataAndAccountReadsShareOneOfficialClient` | yes | yes |
| B3 | missing account builder fails explicitly | `TestA061InstrumentMetadataCapabilityFailuresAreExplicit` | yes | yes |
| B4 | cold builder cancellation/error is propagated without caching a client | `TestA061ColdMetadataResolverHonorsRequestCancellation` | yes | yes |
| B5 | a broker without `Stocks` cannot be widened | `TestA061InstrumentMetadataCapabilityFailuresAreExplicit` | yes | yes |
| tail | a cached metadata-capable broker returns its narrow `Stocks` surface | A061 chunk, market, and shared-client tests | yes | yes |
