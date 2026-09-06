# a063 Design proposal review

2026-09-05. Input: completed `01-ceo-review.md`, restored current D5, `internal/console/templates.go:578-598`, and AST-backed console reader map. UI scope is the existing capability-attestation section only; APP UI classification. Cross-model voices were not run; Manager explicitly accepted substantive in-context perspectives with limitations. The designer binary exists, but visual generation/comparison-board feedback was not run under this bounded in-context review; this is a disclosed deviation from the full visual skill, not a claim that mockups were generated or approved.

## Scope assessment and existing design

Initial score 7/10 before D5, 8/10 after fixed surface/state contract. No root DESIGN.md found; current semantic section/dl/notice/list vocabulary is the reference. Navigation, fonts, colors, new forms, animation and dashboard redesign are outside this change; do not introduce generic cards or a new control button.

## Pass 1: Information architecture — 8/10

Current attestation state remains first, followed by renewal diagnostic status and age, then required recovery information. Startup denial reasons keep their own existing block. This separates “last attempt failed” from “current attestation unusable,” avoiding an incorrect inference that the engine has stopped.

```text
capability attestation
  current file: issued / expiry / existing usable state
  renewal health: issued | refused | failed | unknown
                  last attempt age + pre-expiry advisory
  startup-denial reasons (existing separate block)
  existing note: operator controls gate approval
```

## Pass 2: Interaction state coverage — 8/10

There is no new asynchronous request or loading spinner; existing server rendering produces a complete snapshot. Empty or rejected diagnostics show unknown even without an attestation file. A stale success cannot hide the pre-expiry warning computed from current attestation metadata.

| Feature | Loading | Empty | Error | Success | Partial |
|---|---|---|---|---|---|
| Renewal health | Existing snapshot render | Unknown diagnostic | Fixed failed/refused or invalid message | Issued + attempt age | Unknown on stale or inconsistent issued metadata |
| Expiry horizon | Same snapshot | Unknown if attestation absent | Fixed malformed-file message | Current expiry | Warning still visible when status invalid |

## Pass 3: User journey — 8/10

At five seconds the operator should distinguish current validity from last-attempt health. At five minutes the operator should identify unmet evidence or a service problem without trying an engine restart. Over repeated use, fixed wording and explicit unknown states build trust because diagnostic absence never becomes green success.

| Step | Operator does | Experience | Contract support |
|---|---|---|---|
| 1 | Opens existing verification page | Finds familiar section | No new navigation |
| 2 | Reads current state and renewal health | Distinguishes failure from startup denial | Separate blocks |
| 3 | Reads age/expiry and fixed reason | Can choose an approved recovery window | 72h warning, 12h stale |
| 4 | Returns after successful renewal | Sees failure replaced | Latest attempt wins |

## Pass 4: Specificity and AI-slop risk — 9/10

This is a dense operations interface with a defined section and useful status text. No marketing hero, decorative card grid, gradient, carousel or stock imagery is proposed. Current template hierarchy is reused; adding “renew now” or “restart engine” actions would both exceed scope and confuse the safety boundary.

## Pass 5: Design system alignment — 8/10

Reuse the existing semantic headings, description lists, notice text and ordinary lists. A root design system is absent, but that does not justify restyling this section during an attestation repair. Fixed reason vocabulary and plain text should fit the surrounding Korean operator language without raw code values becoming the sole explanation.

## Pass 6: Responsive and accessibility — 8/10

Status must be expressed in text, not color alone. Keep warning/reason text in normal document order and allow it to wrap on a narrow viewport; 16 bounded codes must not create an unbounded horizontal table. No new interactive control means no new focus or touch-target behavior; post-implementation isolated rendering should still verify readable headings and no raw HTML injection.

## Pass 7: Unresolved design choices — zero

D5 selects the existing surface, fixes missing/stale/error behavior, distinguishes advisory from denial, and specifies bounded escaped content. Implementation must render the diagnostic block outside the existing attestation-present conditional. That requirement is concrete and testable; no further aesthetic choice is needed to release coding.

## Litmus and completion

| Litmus | Primary finding | Claude | Outside Codex | Consensus |
|---|---|---|---|---|
| Product identity | Existing console retained | N/A | N/A | N/A |
| Visual anchor | Section heading | N/A | N/A | N/A |
| Scannable | Current state / renewal health | N/A | N/A | N/A |
| One job per section | Attestation status | N/A | N/A | N/A |
| Cards necessary | No new cards | N/A | N/A | N/A |
| Motion useful | No new motion needed | N/A | N/A | N/A |
| Shadows necessary | No decorative dependency | N/A | N/A | N/A |

Hard rejection patterns proposed: zero. All seven passes examined, overall 8/10, no unresolved design decision, no new deferred TODO. Required implementation checks: diagnostic block visible without attestation, text labels independent of color, bounded escaping, and warning/denial separation. Mockups generated/approved: zero; real browser visual QA remains unperformed and is not implied by this plan review.
