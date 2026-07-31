# Branch Test Map: `Console.render`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid template input returns HTTP 500 without a partial success response | inherited renderer/template coverage and full console suite | unchanged | passed in console suite |
| final | rendered normal/shared and restart documents use exact header/meta `same-origin` and contain no `no-referrer` override | `TestConsoleDocumentsUseSameOriginReferrerPolicy` | failed: header and both metas were `no-referrer` | passed |
