# Function Logic Map: `collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Current-base source SHA-256: `9e8816b71972be1678026da1c774934c85adc33872eed6e3a616abf9fa73dc2b`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `156:1`–`222:2`
- AST evidence: `ast.json`, generated from frozen base `016da6245feb60e13971388be386c2c2041469a8`.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Inputs/results are the exact AST signature above; this L0 map does not infer undocumented state.
- Any later edit must preserve OFF defaults, the owner key without family/horizon, and zero exposure-raising dispatch while a prerequisite is missing.

## Branches and early returns

- Exact AST return nodes: `159:3, 162:3, 165:3, 168:3, 173:3, 185:4, 197:3, 209:44, 214:3, 220:2`.

| Branch | AST kind | Source location | Required test disposition |
|---|---|---|---|
| B1 | if | 161:2 | planned targeted RED before any edit; not run by L0 |
| B2 | if | 164:2 | planned targeted RED before any edit; not run by L0 |
| B3 | if | 167:2 | planned targeted RED before any edit; not run by L0 |
| B4 | if | 172:2 | planned targeted RED before any edit; not run by L0 |
| B5 | if | 176:2 | planned targeted RED before any edit; not run by L0 |
| B6 | range | 182:2 | planned targeted RED before any edit; not run by L0 |
| B7 | if | 184:3 | planned targeted RED before any edit; not run by L0 |
| B8 | if | 196:2 | planned targeted RED before any edit; not run by L0 |
| B9 | range | 201:2 | planned targeted RED before any edit; not run by L0 |
| B10 | if | 203:3 | planned targeted RED before any edit; not run by L0 |
| B11 | if | 210:2 | planned targeted RED before any edit; not run by L0 |
| B12 | range | 217:2 | planned targeted RED before any edit; not run by L0 |

## Calls and live bindings

| Callee expression | Source location | Current-base evidence/requirement |
|---|---|---|
| len | 159:144 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 161:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 162:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 165:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 168:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 170:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 170:31 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| DecodeString | 171:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| base64.StdEncoding.Strict | 171:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| base64.StdEncoding.EncodeToString | 172:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 172:72 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 173:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 179:12 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 179:30 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 180:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 180:58 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 181:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 181:59 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.approved.Symbol | 183:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| bySymbol.approved.Valid | 184:22 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 185:11 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 188:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.route.Request | 188:97 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.load | 190:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strategyrouter.Market | 191:42 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 191:111 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 191:129 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ed25519.PublicKey | 192:15 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| strings.TrimSpace | 195:23 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| loader.getenv | 195:41 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| batch.ManifestDigest | 196:19 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 197:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| make | 199:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| batch.Len | 199:55 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| batch.For | 202:20 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| ValidProposal | 203:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| authority.Proposal | 203:14 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| append | 207:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| sort.Slice | 209:2 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entries.route.approved.Symbol | 209:51 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entries.route.approved.Symbol | 209:88 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 210:5 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| fail | 211:13 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| sha256.New | 216:7 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| h.Write | 218:10 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| call | 218:18 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.route.approved.Symbol | 218:25 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| entry.authority.Proposal | 218:66 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 221:16 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| len | 221:52 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| hex.EncodeToString | 221:144 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |
| h.Sum | 221:163 | current-base AST call; re-query CodeGraph callers/callees/impact immediately before edit |

## State mutations and fallbacks

- The AST is the exhaustive current-base record of assignments, calls, branches, defers and returns. Before a function body edit, the owning lot must update this map with changed condition semantics and concrete RED/GREEN test evidence.

## Safety conclusion

- L0 status: pre-edit evidence only; no production function was edited and no branch test is claimed as run by L0.
- A named targeted RED or explicit evidence-backed not-applicable rationale is required for every edited branch before GREEN.
