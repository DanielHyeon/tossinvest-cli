# Branch Test Map: `Console.routes`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | local/remote route set | existing remote access route tests | yes | yes |
| B2 | optional release route wiring | existing release route tests | yes | yes |
| B3 | autostart exact route is session+CSRF guarded | `TestAutostartRequiresCSRF` + static route guard | no | no |
| B4 | no order/credential route added | existing static forbidden-route tests | no | no |
