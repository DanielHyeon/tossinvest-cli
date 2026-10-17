## ADDED Requirements

### Requirement: MCP honors effective backend policy

MCP SHALL expose official-read and official-trading dependencies only when credentials exist, OpenAPI is enabled, and resolved backend preference is not WTS, matching existing CLI eligibility. Credentials alone MUST NOT enable a forbidden official dependency. Existing engine official-only enforcement, journal, risk checks and shared-token lifecycle MUST remain unchanged.

#### Scenario: WTS is effective with credentials present
- **WHEN** MCP starts with valid-looking stored official credentials and effective backend WTS
- **THEN** official read and trading dependencies are not exposed and official constructors or requests are not invoked through a bypass

#### Scenario: OpenAPI is disabled
- **WHEN** OpenAPI is disabled, including with an explicit openapi/official backend override
- **THEN** MCP does not expose official capabilities regardless of credential presence

#### Scenario: Official or auto authorizes official access
- **WHEN** OpenAPI is enabled, required credentials exist and the resolved preference is auto/openapi
- **THEN** established official MCP functionality remains available with existing trading safeguards

#### Scenario: Credentials or override are invalid
- **WHEN** credentials are absent or backend selection conflicts with configuration
- **THEN** MCP follows existing resolver validation and fallback semantics without creating a new official-access exception

### Requirement: Truncation preserves exact JSON numbers

MCP SHALL preserve the exact numeric value and JSON number type of retained fields while truncating large responses, including integers greater than 2^53, negative integers and long decimal numbers.

#### Scenario: Large nested numeric output
- **WHEN** JSON exceeding the truncation threshold contains retained nested numeric values such as 9007199254740993 and 0.1234567890123456789
- **THEN** shortened output remains valid JSON and retains those numeric tokens without rounding or string conversion

#### Scenario: Ordinary output and truncation metadata
- **WHEN** a small output or a large array response is processed
- **THEN** existing small-output content and large-output omission metadata remain compatible

#### Scenario: Repeated trimming and pre-existing omission markers
- **WHEN** numeric decoding preserves json.Number and a response already contains an omission marker or needs multiple trim passes
- **THEN** cumulative omission counts remain correct and malformed marker values do not panic or overflow
