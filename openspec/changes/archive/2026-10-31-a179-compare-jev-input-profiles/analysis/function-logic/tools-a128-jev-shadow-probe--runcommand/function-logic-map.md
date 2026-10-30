# Function Logic Map: `runCommand`

Source: `tools/a128-jev-shadow-probe/main.go`. Current AST and source digest: `ast.json`. Pre-edit snapshot: `../../pre-edit-ast/tools-a128-jev-shadow-probe--runcommand.json`.

## Inputs and invariants

CLI flags, environment and filesystem paths. Validates profile/output before model/broker clients; constructs read-only source and append-only ledger, restores pending then runs probe. Public-only data and isolated experimental output remain invariants.

## Branches and early returns

| Branch | Source condition |
|---|---|
| B1 | if at 92:2: `if err := flags.Parse(args); err != nil {` |
| B2 | if at 97:3: `if f.Name == "out" {` |
| B3 | if at 102:2: `if err != nil {` |
| B4 | if at 107:2: `if err != nil {` |
| B5 | if at 111:2: `if err != nil {` |
| B6 | if at 114:2: `if err := validateRunFlags(*top, *pool, *horizon, *interval); err != nil {` |
| B7 | if at 119:2: `if strings.TrimSpace(*sessionWindow) != "" {` |
| B8 | if at 121:3: `if err != nil {` |
| B9 | if at 128:2: `if err != nil {` |
| B10 | if at 132:2: `if err != nil {` |
| B11 | if at 135:2: `if credentials == nil {` |
| B12 | if at 144:2: `if err != nil {` |
| B13 | if at 155:2: `if err != nil {` |
| B14 | if at 159:2: `if err != nil {` |
| B15 | if at 165:2: `if *profile != "current" {` |
| B16 | if at 167:3: `if _, err := rand.Read(token[:]); err != nil {` |
| B17 | if at 179:2: `if err != nil {` |
| B18 | if at 191:2: `if manualSess != nil {` |

## Calls and live bindings

The complete current call enumeration is in `ast.json`. The current source is authoritative; observed call context includes the local ledger, profile helper, and existing probe/report operations appropriate to this function. CodeGraph source context located probe.cycle and assembleState; precise isolated checkout graph freshness was unavailable.

## State mutations and fallbacks

Validates profile/output before model/broker clients; constructs read-only source and append-only ledger, restores pending then runs probe. Failures remain explicit; no production order or runtime configuration changes are introduced.

## Safety conclusion

Shadow measurement only. Existing broker/source calls are read-only; validation uses fake transports. Full production high-risk mutation review is not applicable.
