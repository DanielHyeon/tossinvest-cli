# Function Logic Map: `relativeLuminance`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | exactly seven-byte `#RRGGBB` | fixed test contract | fatal test failure if malformed |
| channels | three parsed 8-bit sRGB values | CSS token | fatal on parse error; convert to `[0,1]` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | colour is not strict six-digit hex shape | test diagnostic | fatal |
| B2 | iterate R, G, B channels | assign local slice | continue for three channels |
| B3 | channel hex cannot parse | test diagnostic | fatal |
| B4 | normalized sRGB channel is at/below 0.04045 | use linear division | continue |
| B5 | channel exceeds 0.04045 | use WCAG gamma transform | continue |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strconv.ParseUint` | strict base-16 channel parse | failure is fatal | B3 |
| `math.Pow` | WCAG sRGB gamma transform | finite 8-bit input | B5 |

## State mutations and fallbacks

- Allocates and mutates only a three-element local test slice.

## Safety conclusion

- Safe edit boundary: WCAG 2.x relative-luminance formula for static CSS tests.
- High-risk impact: no; float conversions represent colours, not financial values.
