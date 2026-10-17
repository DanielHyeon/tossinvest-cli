## Why

The existing shadow observations show inconsistent selection differences across market sessions and incomplete input coverage; they do not establish why those differences occurred. An offline, reproducible separation of input quality and a separately versioned J2 question can make the next comparison interpretable without changing r1 or trading behavior.

## What Changes

- Define an immutable snapshot and derived input-quality sidecar, then compare existing selection and fixed alternatives within the same cycle and market session.
- Specify a separate r2 lot with the same historical inputs and model provenance, unchanged J1 question and frozen original J1 scores, changing only the J2 wording for the isolated selection comparison.
- Require an explicit blocked result when compatible saved r2 responses are unavailable; synthetic fixtures verify plumbing only. Any new model request requires separate authorization and is outside this design task.
- Preserve existing r1, a128 evidence, operational configuration and all order paths. Findings are observational, not a trading or safety approval.

## Capabilities

### New Capabilities

- `semantic-judgment-shadow`: Add narrowly scoped offline input-quality and paired question-comparison requirements to the capability currently introduced by active change a128. The main spec does not yet contain this capability; formal synchronization must account for both changes and their order.

### Modified Capabilities

None.

## Impact

Future offline analysis artifacts and tests only; implementation location is deliberately undecided. No model calls, network collection, order integration, operational changes, or modification of the existing evidence are authorized by this proposal. Story `STORY-TOS-a145` tracks this change independently of a128.

## Validation infrastructure repair

An advisory CodeGraphContext stats timeout currently aborts index synchronization before the successful hard CodeGraph state is recorded. Handle only that timeout as an advisory failure; retain hard failure and all gate checks.
