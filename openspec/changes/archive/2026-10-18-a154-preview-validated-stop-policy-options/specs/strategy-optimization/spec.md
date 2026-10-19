## ADDED Requirements

### Requirement: Hypothetical initial-stop comparison remains separate from operating policies
The console SHALL distinguish initial stops, quantity sizing, strategic exits and common profit protection. A manually supplied hypothetical KR/KRW or US/USD long integer-share comparison SHALL NOT represent an approved strategy or US Parker runtime support. Existing policy defaults and selectors SHALL remain unchanged. Validated selection and new-policy CAS apply are deferred, not delivered by this comparison.

#### Scenario: Unapproved research alternatives
- **WHEN** the operator views the three research stop alternatives
- **THEN** the UI identifies hypothetical inputs, unavailable approval and no apply action without registering writable policy fields

### Requirement: Numerical preview uses exact independent calculations
The calculator SHALL validate bounded decimal inputs and a user-declared bounded uniform zero-origin tick grid. Entry SHALL already lie on that grid. The independent stops SHALL be floor(entry × 0.993), floor(signal-bar low) minus one tick, and floor(entry minus 1.5 × ATR14), where floor is to that grid. Valid stops SHALL be positive, below entry and within bounds. Supplied budgets and caps SHALL be positive. Quantity SHALL equal the minimum of floor(price-risk budget / distance), maximum shares and floor(maximum notional / entry). Computed zero SHALL mean no position. Prices and amounts SHALL remain exact; distance percent SHALL use six-place half-up display rounding with positive values below 0.000001% labelled below that threshold. Display rounding SHALL NOT affect arithmetic.

#### Scenario: Partial or invalid policy-specific data
- **WHEN** signal low or ATR is missing or invalid
- **THEN** only its corresponding arm is unknown or invalid and other valid arms remain calculable

#### Scenario: Missing sizing inputs
- **WHEN** common geometry is valid but sizing inputs are incomplete or invalid
- **THEN** valid stop prices and distances remain visible while quantity, risk and notional are unknown or invalid with reasons

#### Scenario: Common geometry error
- **WHEN** entry or the supplied grid is invalid
- **THEN** all numerical arms are blocked with field-linked errors rather than substituted zeros

### Requirement: Hypothetical computation does not create authority or persist state
Calculation SHALL use a separate bounded compute-only POST with existing console session-security semantics, CSRF and origin protections. It SHALL reject query input, unknown or duplicate form fields including duplicate csrf and oversized bodies. Responses including refusals SHALL be no-store. It SHALL NOT call stored preview/apply, issue capabilities, persist scenario inputs, query account/market data or mutate candidate/audit/settings/position state. Costs, taxes, spread, gap/slippage, FX, target RR and filled risk SHALL remain explicitly excluded or unknown. Planned price risk SHALL NOT be labelled a total-loss guarantee.

#### Scenario: Numerical comparison
- **WHEN** valid hypothetical values are submitted
- **THEN** exact results and nearby limitations are displayed without candidate, audit, setting, position or operating-toggle writes

#### Scenario: Forged operating request
- **WHEN** a client submits a research-only stop key to existing optimization mutation paths
- **THEN** the request is refused without partial state changes

#### Scenario: Accessible calculation form
- **WHEN** users navigate with keyboard or a narrow viewport
- **THEN** labelled inputs, visible focus, associated errors and readable result units remain available without an apply control
