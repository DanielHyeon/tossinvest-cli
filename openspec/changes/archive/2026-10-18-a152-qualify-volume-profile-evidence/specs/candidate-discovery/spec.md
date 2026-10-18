## ADDED Requirements

### Requirement: Volume profile feasibility separates executions from proxies
The feasibility report SHALL classify actual trades, documented volume bins, OHLCV proxies and unavailable inputs, preserving source, time semantics, coverage and correction handling.

#### Scenario: Only OHLCV exists
- **WHEN** only five-minute candles and orderbook snapshots are provided
- **THEN** true volume profile is unsupported and neither candles nor book size are relabeled as executed price-bin volume

### Requirement: Unsupported volume profile does not block non-profile research
Feasibility SHALL complete with an unsupported result when inputs are unavailable and SHALL supply an explicit exclusion reason to downstream experiments.

#### Scenario: No licensed trade data
- **WHEN** no suitable existing input is available
- **THEN** no new collection starts and the report completes without inventing LVN/HVN evidence

### Requirement: Offline profile qualification preserves epistemic boundaries
The bounded a152-v1 local manifest SHALL distinguish structural eligibility from authentic source, complete coverage, licensed access, profitability and runtime authority. Source and coverage SHALL remain caller assertions. Synthetic fixtures SHALL NOT become empirical VP inputs. Other research SHALL remain unblocked.

#### Scenario: Valid synthetic executed bins
- **WHEN** internally consistent, contiguous, zero-inclusive executed-volume bins are supplied as synthetic data
- **THEN** classification is documented_price_bins with eligible_unverified structure, but empirical vp_input_eligible is false with synthetic_only and production_authorized remains false

#### Scenario: Book-only evidence
- **WHEN** only an orderbook source declaration is supplied
- **THEN** classification is unavailable with book_size_is_not_executed_volume and not ohlcv_proxy or executed volume

### Requirement: Profile time and correction uncertainty fail closed
The qualifier SHALL validate requested prior-session or current-prefix bounds, event and receipt times, source declarations, tick/bin grid, explicit coverage and correction policy. Duplicate identities, future information, incomplete coverage and unsupported corrections SHALL NOT silently produce an eligible profile.

#### Scenario: Future correction and unknown policy
- **WHEN** a trade correction/cancel event or unresolved/unknown correction policy is supplied
- **THEN** the dataset is not eligible and no correction replay or silent deduplication is invented

#### Scenario: Late final bins used for an earlier prefix
- **WHEN** bin build or receipt time exceeds as_of
- **THEN** that dataset is refused for the requested prefix and no retrospective promotion occurs

#### Scenario: Partial session or price-bin holes
- **WHEN** session coverage is partial or price bins overlap or omit declared zero-volume price intervals
- **THEN** the dataset is not eligible and the corresponding exclusion reason is retained

### Requirement: Offline qualifier preserves inputs and bounded execution
The qualifier SHALL use bounded strict JSON, exact decimal arithmetic, input hashing and exclusive output creation, with no network, runtime changes or evidence mutation. Invalid datasets SHALL be reported independently; invalid envelopes SHALL fail the CLI.

#### Scenario: Existing output or symlink destination
- **WHEN** the output already exists or traverses a symlink
- **THEN** writing is refused and original input/output bytes remain unchanged

#### Scenario: One malformed dataset among valid datasets
- **WHEN** a structurally valid envelope includes one malformed dataset and an independent valid dataset
- **THEN** the malformed dataset is reported invalid without suppressing the independent result
