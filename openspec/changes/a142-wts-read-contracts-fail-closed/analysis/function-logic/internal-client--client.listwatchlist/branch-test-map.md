# Branch Test Map: `Client.ListWatchlist`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | no session still rejects | existing client coverage | no | no |
| B2 | legacy POST/decode error propagates at baseline | replaced by dedicated-route local-server test | no | no |
| B3 | non-WATCHLIST section continues scan | removed with legacy adapter | no | no |
| B4 | WATCHLIST data JSON decode failure returns error | removed with legacy adapter | no | no |
| B5 | matching WATCHLIST section selects first match | removed with legacy adapter | no | no |
| B6 | no matching legacy section errors | removed with legacy adapter | no | no |
| B7 | every legacy group maps | replaced by multi-group dedicated fixture | no | no |
| B8 | every legacy item flattens with money defaults | replaced by optional-field dedicated fixture | no | no |
| planned wrapper | multiple groups flatten in stable numeric order with wire-order ties | planned fixture test | no | no |
| planned wrapper | `Symbol` equals code when ticker differs | planned fixture test | no | no |
| planned wrapper | currency verbatim; nil prices zero | planned fixture test | no | no |
| planned wrapper | watchlists/items missing/null/wrong type error | planned table test | no | no |
| planned wrapper | explicit empty watchlists/items succeeds | planned table test | no | no |

No test executed during pre-edit preparation.
