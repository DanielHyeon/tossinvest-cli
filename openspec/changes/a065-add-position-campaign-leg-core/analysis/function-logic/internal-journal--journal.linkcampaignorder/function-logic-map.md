# Function Logic Map: `Journal.LinkCampaignOrder`

- Source: `internal/journal/position_campaign.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| command/order identity | non-empty immutable order, intent and attempt | request + authoritative journal lookup | deterministic invalid-identity refusal |
| execution authority | confirmed PLACE/AMEND, exact decision/account/market/day/symbol/side/order | intents, attempts, scoped lineage | latch RECONCILE + event |
| requested cap | initial cap equals authoritative intent quantity; replacement cap equals immutable edge requested quantity | `intents.quantity` / `scoped_lineage_edges.requested_quantity` | durable invalid-identity latch |
| ambiguity | caller-declared ambiguity is evidence, never a successful false default | request digest + authoritative lookup | durable invalid-identity latch |
| uniqueness | one scoped order owner and one successor per predecessor | unique indexes + precheck | latch RECONCILE + event |
| EXIT FIRST | 묶인 generation 이 CLOSED 가 아니고, CLOSING position 이 없으며, 현재 generation 이 열린 뒤의 미해결 risk-reducing intent 가 없을 것 | journal local state | exposure blocked — 거절도 증거를 남긴다(attempt 는 이미 CONFIRMED 이므로 주문을 되돌릴 수 없다) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | retry command | none | record or stable refusal | retry tests |
| B2 | version/admission block | EXIT FIRST 거절은 RECONCILE/entry-block 으로 격리하고 `ORDER_LINK_REFUSED` 를 남긴 뒤 commit 한다 | typed refusal | version/EXIT FIRST tests · `TestExposureRefusalAtLinkLeavesRefusalEvidence` |
| B3 | missing/mismatched authoritative lineage | refusal command+event+latch | invalid identity after commit | hardening lineage test |
| B3a | caller ambiguity or authoritative quantity mismatch | refusal command+event+latch | invalid identity after commit | independent-review ambiguity/cap tests |
| B4 | duplicate scoped order/predecessor successor | refusal command+event+latch | invalid identity after commit | uniqueness tests |
| B5 | exact lineage | order+leg+campaign+command+event atomically | record | replacement tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `authoritativeCampaignOrderInTx` | bind decision→intent→attempt→order/scope | exact lookup; ambiguity refuses | AST |
| `latchCampaignLinkConflict` | durable stable refusal | append-only command/event in same tx | AST |

## State mutations and fallbacks

- Caller `LineageAmbiguous=true` is included in the command digest and durably refused; a successful stored false is derived only after exact authoritative lineage and quantity checks.
- A successor starts with `remaining_quantity=requested_cap`, independent of the aggregate leg residual.
- No broker submission or runtime activation.

## Safety conclusion

- Safe edit boundary: immutable lineage binding and journal isolation.
- High-risk impact: yes; ambiguity blocks only new exposure.
