# Function Logic Map: `transitiveDependencyWithStop`

- Source: `internal/candidate/approved_consumer_guard_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `graph` | repository-local package dependency adjacency list | parsed production imports | absent edges end traversal conservatively without inventing reachability |
| `start` | package under audit | boundary audit | start is never exempted merely because its name appears in the sanitizer set |
| `matches` | predicate identifying an unsanitized direct ApprovedCandidate reader | direct-reader set | a match returns its exact path |
| `stop` | predicate identifying an independently audited sanitizer boundary | exact sanitizer allowlist | dependencies beyond that boundary are intentionally not tainted |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | node already seen | none | skip duplicate/cycle | existing graph traversal tests |
| B2 | non-start node is an audited sanitizer, including a sanitizer that directly reads ApprovedCandidate | none | stop before applying the downstream match | new direct-sanitizer stop test and repository audit |
| B3 | node matches a direct reader and is not an audited downstream sanitizer | none | return exact path and true | existing cross-package laundering test |
| B4 | neither stop nor match | enqueue unseen dependencies with copied path | continue breadth-first traversal | existing transitive dependency tests |
| B5 | queue drains | none | nil, false | existing negative guard fixtures |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `matches` predicate | classify direct readers | pure callback with no mutation | boundary audit construction + AST |
| `stop` predicate | classify audited sanitizer packages | applies only after leaving `start`; exact allowlist | boundary audit construction + adversarial tests |

## State mutations and fallbacks

- Traversal state is local only; no repository or runtime state changes.
- Checking stop before match is limited to non-start nodes, so a sanitizer's own direct-reader implementation remains audited as a boundary while callers receive its sealed output without reverse taint.

## Safety conclusion

- Safe edit boundary: reorder the downstream sanitizer check ahead of direct-reader matching while explicitly excluding the start node from that exemption.
- High-risk impact: yes. A broad or start-applicable stop would hide an authority path; leaving match first makes an audited direct sanitizer unusable and falsely taints authority-bearing consumers of its opaque output.
