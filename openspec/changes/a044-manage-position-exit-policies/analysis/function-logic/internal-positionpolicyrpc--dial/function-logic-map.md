# Function Logic Map: `Dial`

- Source: `internal/positionpolicyrpc/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| descriptor path | fixed `endpoint.json` under private control directory | engine descriptor locator | fail before network |
| descriptor fields | loopback address, PID>0, token length>=32 | strict descriptor JSON | reject invalid fields |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | empty descriptor path | none | error | client contract |
| B2-B3 | secure open or strict decode fails | close if opened | error | descriptor security tests |
| B4-B5 | fields or loopback validation fails | none | error | endpoint validation contract |
| B6-B7 | authenticated health fails or reports false | one bounded GET | wrapped unavailable | integration test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `openPrivateDescriptor` | securely bind read to validated inode | O_NOFOLLOW + identity validation | AST |
| `decodeDescriptor` | accept one <=4KiB known-field JSON value | unknown/trailing/oversize rejected | AST |
| `Client.call` | prove authenticated engine endpoint is live | 5s timeout, proxy disabled | AST |

## State mutations and fallbacks

- No journal/database constructor is reachable; descriptor possession only reaches fixed loopback RPC paths.

## Safety conclusion

- Safe edit boundary: validate filesystem authority before parsing/network and keep proxy disabled.
- High-risk impact: yes
