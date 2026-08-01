# Branch Test Map: `verifyProtectionMatrix`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty parsed matrix | malformed parsed state test | whole parsed state could proceed | rejected |
| B2 | account/profile/market/session/type/order/trigger/quantity/tool mismatch | scope matrix | exact matching existed | passes and returns only matched row |
| B3 | missing/extra/swapped/tampered evidence | evidence table | digest validation existed | passes before final trust read |
