# Branch Test Map: `consoleBroker.resolve`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | broker gate acquisition fails | background context makes this defensive branch unreachable; lock cancellation is covered through `instrumentMetadata` | n/a | reasoned |
| B2 | cached account-resolved client returns without rebuilding | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce`, shared-client test | existing | existing |
| B3 | cold factory failure is returned and remains retryable | existing console seam failure coverage | existing | existing |
| tail | account-first build caches the broker and trimmed account reference | `TestOpeningEveryConsoleReadScreenResolvesTheAccountOnce` | existing | existing |
