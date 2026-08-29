# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go` (159-236)
- Function: `strategyProposalAuthorityLoader.collectMarket` in package `engine`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- File SHA-256: `88e06b6c841ba30cb1c3107fba33c134c82b34f871dec646ee92b739a2e58c94`
- Pinned revision: `current` — the AST and the SHA-256 above are this worktree's file.
- AST evidence: `ast.json` — AST branches 14.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

Turns one market's routed entries into sealed proposal entries. It now refuses the whole market before building the entry list when any routed symbol carries more than one family proposal, because dropping that one symbol would shrink the list to one and thereby *satisfy* the `len(entries) != 1` gate that three downstream readers share (review finding C2).

The signature above is the exhaustive input/result record; this map does not infer state the AST does not show.

## Branches and early returns

- Measurement regime: Go coverage profiles, count mode.
- untagged proposal suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal ./internal/strategyproposal/`
- tagged proposal suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/strategyproposal/`
- tagged engine suite: `go test -count=1 -tags tossos_testseams -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- untagged engine suite: `go test -count=1 -covermode=count -coverpkg=./internal/strategyproposal,./internal/strategyflow,./internal/strategyrouter,./internal/app/engine ./internal/app/engine/`
- Measured entry: the function body was not executed (untagged proposal suite); not executed (tagged proposal suite); executed 4x (tagged engine suite); not executed (untagged engine suite).

Exact AST return positions: 162:3, 165:3, 168:3, 171:3, 176:3, 188:4, 200:3, 210:4, 223:44, 228:3, 234:2.

| Branch | AST kind | Position | Measured disposition |
|---|---|---|---|
| B1 | if | 164:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B2 | if | 167:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B3 | if | 170:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B4 | if | 175:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B5 | if | 179:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 2x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B6 | range | 185:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 4x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B7 | if | 187:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B8 | if | 199:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 1x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal` |
| B9 | range | 206:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B10 | if | 207:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B11 | range | 215:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |
| B12 | if | 217:3 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B13 | if | 224:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm not entered (tagged engine suite); arm not entered (untagged engine suite); no per-test profile in the attribution set entered it |
| B14 | range | 231:2 | arm not entered (untagged proposal suite); arm not entered (tagged proposal suite); arm entered 3x (tagged engine suite); arm not entered (untagged engine suite); entered by `TestStrategyProposalAuthorityKeepsMarketFailureLocal`, `TestStrategyProposalAuthorityLoadsKRUSConcurrently` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 162:144 |
| `len` | 164:31 |
| `fail` | 165:10 |
| `fail` | 168:10 |
| `fail` | 171:10 |
| `strings.TrimSpace` | 173:13 |
| `loader.getenv` | 173:31 |
| `DecodeString` | 174:14 |
| `base64.StdEncoding.Strict` | 174:14 |
| `base64.StdEncoding.EncodeToString` | 175:19 |
| `len` | 175:72 |
| `fail` | 176:10 |
| `strings.TrimSpace` | 182:12 |
| `loader.getenv` | 182:30 |
| `make` | 183:13 |
| `len` | 183:58 |
| `make` | 184:14 |
| `len` | 184:59 |
| `entry.approved.Symbol` | 186:13 |
| `bySymbol.approved.Valid` | 187:22 |
| `fail` | 188:11 |
| `append` | 191:13 |
| `entry.route.Request` | 191:97 |
| `loader.load` | 193:16 |
| `strategyrouter.Market` | 194:42 |
| `strings.TrimSpace` | 194:111 |
| `loader.getenv` | 194:129 |
| `ed25519.PublicKey` | 195:15 |
| `strings.TrimSpace` | 198:23 |
| `loader.getenv` | 198:41 |
| `batch.ManifestDigest` | 199:19 |
| `fail` | 200:10 |
| `batch.Ambiguous` | 207:6 |
| `entry.approved.Symbol` | 207:22 |
| `fail` | 208:14 |
| `make` | 213:13 |
| `batch.Len` | 213:55 |
| `batch.For` | 216:20 |
| `ValidProposal` | 217:14 |
| `authority.Proposal` | 217:14 |
| `append` | 221:13 |
| `sort.Slice` | 223:2 |
| `entries.route.approved.Symbol` | 223:51 |
| `entries.route.approved.Symbol` | 223:88 |
| `len` | 224:5 |
| `fail` | 225:13 |
| `sha256.New` | 230:7 |
| `h.Write` | 232:10 |
| `(unnamed)` | 232:18 |
| `entry.route.approved.Symbol` | 232:25 |
| `entry.authority.Proposal` | 232:66 |
| `len` | 235:16 |
| `len` | 235:52 |
| `hex.EncodeToString` | 235:144 |
| `h.Sum` | 235:163 |

## State mutations and fallbacks

- AST assignments: 25. Defers: 0. Goroutine statements: 0.

## Safety conclusion

The new guard only adds a refusal; it can never admit a proposal that was previously refused. All three `len(entries) != 1` readers — `ResultAuthority`, `strategyAccountAuthorityLoader.collectMarket` and `strategyProjectionFromAssembly` — read the list this function produces, so this is the single place the correction belongs.
