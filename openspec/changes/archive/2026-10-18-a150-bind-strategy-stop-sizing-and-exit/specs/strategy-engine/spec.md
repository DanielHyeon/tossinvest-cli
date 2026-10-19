## ADDED Requirements

### Requirement: Diagnostic risk inputs have bounded decimal syntax
Parker RR diagnostics SHALL validate caller-supplied numeric strings as bounded decimal literals before numeric parsing. They SHALL reject exponent, fraction, radix and oversized representations without changing the frozen runtime parser or existing valid diagnostic results.

#### Scenario: Unsupported numeric representation
- **WHEN** a supplied planned or realized RR numeric input contains a binary exponent, fraction, nondecimal radix or exceeds the documented length bound
- **THEN** the diagnostic returns invalid without parsing that representation or granting trading authority
